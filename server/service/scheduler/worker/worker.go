package worker

import (
        "context"
        "fmt"
        "sync"
        "sync/atomic"
        "time"

        schedulerModel "yunwei/model/scheduler"
        "yunwei/service/scheduler/queue"
)

// WorkerNodeStatus Worker 节点状态
type WorkerNodeStatus string

const (
        WorkerNodeStatusIdle     WorkerNodeStatus = "idle"
        WorkerNodeStatusBusy     WorkerNodeStatus = "busy"
        WorkerNodeStatusStopping WorkerNodeStatus = "stopping"
        WorkerNodeStatusStopped  WorkerNodeStatus = "stopped"
        WorkerNodeStatusError    WorkerNodeStatus = "error"
)

// WorkerNode Worker 节点
type WorkerNode struct {
        ID           string           `json:"id"`
        NodeName     string           `json:"nodeName"`
        Status       WorkerNodeStatus `json:"status"`

        // 执行统计
        TotalTasks   int64 `json:"totalTasks"`
        SuccessTasks int64 `json:"successTasks"`
        FailedTasks  int64 `json:"failedTasks"`

        // 当前任务
        CurrentTask *schedulerModel.Task `json:"currentTask"`
        TaskStarted time.Time            `json:"taskStarted"`

        // 配置
        MaxConcurrent int `json:"maxConcurrent"` // 最大并发数
        Timeout       int `json:"timeout"`       // 默认超时

        // 心跳
        LastHeartbeat time.Time `json:"lastHeartbeat"`

        // 控制
        ctx        context.Context
        cancel     context.CancelFunc
        taskChan   chan *schedulerModel.Task
        resultChan chan *WorkerTaskResult

        mu sync.RWMutex
}

// WorkerTaskResult 任务结果
type WorkerTaskResult struct {
        TaskID       uint                      `json:"taskId"`
        ExecutionID  string                    `json:"executionId"`
        ServerID     uint                      `json:"serverId"`
        Status       schedulerModel.TaskStatus `json:"status"`
        Output       string                    `json:"output"`
        ErrorMessage string                    `json:"errorMessage"`
        Duration     int64                     `json:"duration"`
        RetryCount   int                       `json:"retryCount"`
}

// NewWorkerNode 创建 Worker 节点
func NewWorkerNode(id string, maxConcurrent int) *WorkerNode {
        ctx, cancel := context.WithCancel(context.Background())

        return &WorkerNode{
                ID:            id,
                NodeName:      fmt.Sprintf("worker-%s", id),
                Status:        WorkerNodeStatusIdle,
                MaxConcurrent: maxConcurrent,
                Timeout:       300,
                ctx:           ctx,
                cancel:        cancel,
                taskChan:      make(chan *schedulerModel.Task, maxConcurrent*2),
                resultChan:    make(chan *WorkerTaskResult, maxConcurrent*2),
        }
}

// Start 启动 Worker
func (w *WorkerNode) Start() {
        w.mu.Lock()
        w.Status = WorkerNodeStatusIdle
        w.LastHeartbeat = time.Now()
        w.mu.Unlock()

        // 注册到队列
        queue.GetQueue().RegisterWorker(w.ID)

        // 启动任务处理协程
        for i := 0; i < w.MaxConcurrent; i++ {
                go w.runTaskProcessor()
        }

        // 启动心跳
        go w.heartbeatLoop()

        // 启动任务获取
        go w.taskFetcher()
}

// Stop 停止 Worker
func (w *WorkerNode) Stop() {
        w.mu.Lock()
        w.Status = WorkerNodeStatusStopping
        w.mu.Unlock()

        w.cancel()

        // 从队列注销
        queue.GetQueue().UnregisterWorker(w.ID)

        close(w.taskChan)
        close(w.resultChan)

        w.mu.Lock()
        w.Status = WorkerNodeStatusStopped
        w.mu.Unlock()
}

// runTaskProcessor 任务处理器
func (w *WorkerNode) runTaskProcessor() {
        for {
                select {
                case <-w.ctx.Done():
                        return
                case task, ok := <-w.taskChan:
                        if !ok {
                                return
                        }

                        w.processTask(task)
                }
        }
}

// processTask 处理任务
func (w *WorkerNode) processTask(task *schedulerModel.Task) {
        w.mu.Lock()
        w.Status = WorkerNodeStatusBusy
        w.CurrentTask = task
        w.TaskStarted = time.Now()
        w.mu.Unlock()

        // 执行任务
        result := w.executeTask(task)

        // 更新 Worker 统计
        atomic.AddInt64(&w.TotalTasks, 1)
        if result.Status == schedulerModel.TaskStatusSuccess {
                atomic.AddInt64(&w.SuccessTasks, 1)
        } else {
                atomic.AddInt64(&w.FailedTasks, 1)
        }

        // 发送结果
        w.sendResult(result)

        // 更新状态
        w.mu.Lock()
        w.Status = WorkerNodeStatusIdle
        w.CurrentTask = nil
        w.mu.Unlock()

        // 确认任务完成
        queue.GetQueue().Ack(task.ID)
}

// executeTask 执行任务
func (w *WorkerNode) executeTask(task *schedulerModel.Task) *WorkerTaskResult {
        result := &WorkerTaskResult{
                TaskID: task.ID,
        }

        startTime := time.Now()
        defer func() {
                result.Duration = time.Since(startTime).Milliseconds()
        }()

        // 根据执行器类型执行
        switch task.Type {
        case schedulerModel.TaskTypeCommand, schedulerModel.TaskTypeScript:
                output, err := w.executeShell(task.Command)
                if err != nil {
                        result.Status = schedulerModel.TaskStatusFailed
                        result.ErrorMessage = err.Error()
                } else {
                        result.Status = schedulerModel.TaskStatusSuccess
                        result.Output = output
                }

        default:
                output, err := w.executeShell(task.Command)
                if err != nil {
                        result.Status = schedulerModel.TaskStatusFailed
                        result.ErrorMessage = err.Error()
                } else {
                        result.Status = schedulerModel.TaskStatusSuccess
                        result.Output = output
                }
        }

        result.RetryCount = task.RetryCount

        return result
}

// executeShell 执行 Shell 命令
func (w *WorkerNode) executeShell(command string) (string, error) {
        // TODO: 实际执行 Shell 命令
        // 这里需要通过 SSH 或本地执行
        return fmt.Sprintf("Executed: %s", command), nil
}

// sendResult 发送结果
func (w *WorkerNode) sendResult(result *WorkerTaskResult) {
        select {
        case w.resultChan <- result:
        default:
                // 结果通道已满，记录日志
        }
}

// heartbeatLoop 心跳循环
func (w *WorkerNode) heartbeatLoop() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for {
                select {
                case <-w.ctx.Done():
                        return
                case <-ticker.C:
                        w.mu.Lock()
                        w.LastHeartbeat = time.Now()
                        w.mu.Unlock()

                        // 更新队列中的心跳
                        queue.GetQueue().Heartbeat(w.ID)
                }
        }
}

// taskFetcher 任务获取器
func (w *WorkerNode) taskFetcher() {
        ticker := time.NewTicker(100 * time.Millisecond)
        defer ticker.Stop()

        q := queue.GetQueue()

        for {
                select {
                case <-w.ctx.Done():
                        return
                case <-ticker.C:
                        // 检查是否有空闲槽位
                        if len(w.taskChan) >= w.MaxConcurrent {
                                continue
                        }

                        // 从队列获取任务
                        task, err := q.Dequeue()
                        if err != nil || task == nil {
                                continue
                        }

                        // 发送到任务通道
                        select {
                        case w.taskChan <- task:
                        default:
                                // 任务通道已满，放回队列
                                q.Nack(task.ID)
                        }
                }
        }
}

// GetResultChan 获取结果通道
func (w *WorkerNode) GetResultChan() <-chan *WorkerTaskResult {
        return w.resultChan
}

// GetStats 获取统计信息
func (w *WorkerNode) GetStats() map[string]interface{} {
        w.mu.RLock()
        defer w.mu.RUnlock()

        return map[string]interface{}{
                "id":           w.ID,
                "nodeName":     w.NodeName,
                "status":       w.Status,
                "totalTasks":   w.TotalTasks,
                "successTasks": w.SuccessTasks,
                "failedTasks":  w.FailedTasks,
                "currentTask":  w.CurrentTask,
                "lastHeartbeat": w.LastHeartbeat,
        }
}

// NodeWorkerPool Worker 池
type NodeWorkerPool struct {
        workers    map[string]*WorkerNode
        mu         sync.RWMutex
        resultChan chan *WorkerTaskResult
        ctx        context.Context
        cancel     context.CancelFunc

        // 配置
        maxWorkers    int
        workerTimeout time.Duration
}

// NewNodeWorkerPool 创建 Worker 池
func NewNodeWorkerPool(maxWorkers int) *NodeWorkerPool {
        ctx, cancel := context.WithCancel(context.Background())

        return &NodeWorkerPool{
                workers:       make(map[string]*WorkerNode),
                resultChan:    make(chan *WorkerTaskResult, 1000),
                ctx:           ctx,
                cancel:        cancel,
                maxWorkers:    maxWorkers,
                workerTimeout: 30 * time.Second,
        }
}

// AddWorker 添加 Worker
func (p *NodeWorkerPool) AddWorker(id string, maxConcurrent int) (*WorkerNode, error) {
        p.mu.Lock()
        defer p.mu.Unlock()

        if len(p.workers) >= p.maxWorkers {
                return nil, fmt.Errorf("worker pool is full")
        }

        if _, exists := p.workers[id]; exists {
                return nil, fmt.Errorf("worker already exists: %s", id)
        }

        worker := NewWorkerNode(id, maxConcurrent)
        p.workers[id] = worker

        return worker, nil
}

// RemoveWorker 移除 Worker
func (p *NodeWorkerPool) RemoveWorker(id string) error {
        p.mu.Lock()
        defer p.mu.Unlock()

        worker, exists := p.workers[id]
        if !exists {
                return fmt.Errorf("worker not found: %s", id)
        }

        worker.Stop()
        delete(p.workers, id)

        return nil
}

// StartAll 启动所有 Worker
func (p *NodeWorkerPool) StartAll() {
        p.mu.RLock()
        defer p.mu.RUnlock()

        for _, worker := range p.workers {
                worker.Start()
                go p.collectResults(worker)
        }
}

// StopAll 停止所有 Worker
func (p *NodeWorkerPool) StopAll() {
        p.mu.Lock()
        defer p.mu.Unlock()

        for _, worker := range p.workers {
                worker.Stop()
        }

        p.cancel()
        close(p.resultChan)
}

// collectResults 收集结果
func (p *NodeWorkerPool) collectResults(worker *WorkerNode) {
        for {
                select {
                case <-p.ctx.Done():
                        return
                case result, ok := <-worker.GetResultChan():
                        if !ok {
                                return
                        }

                        // 发送到池的结果通道
                        select {
                        case p.resultChan <- result:
                        default:
                                // 结果通道已满
                        }
                }
        }
}

// GetResultChan 获取结果通道
func (p *NodeWorkerPool) GetResultChan() <-chan *WorkerTaskResult {
        return p.resultChan
}

// GetWorker 获取 Worker
func (p *NodeWorkerPool) GetWorker(id string) (*WorkerNode, error) {
        p.mu.RLock()
        defer p.mu.RUnlock()

        worker, exists := p.workers[id]
        if !exists {
                return nil, fmt.Errorf("worker not found: %s", id)
        }

        return worker, nil
}

// GetWorkers 获取所有 Worker
func (p *NodeWorkerPool) GetWorkers() []*WorkerNode {
        p.mu.RLock()
        defer p.mu.RUnlock()

        workers := make([]*WorkerNode, 0, len(p.workers))
        for _, worker := range p.workers {
                workers = append(workers, worker)
        }

        return workers
}

// GetStats 获取统计信息
func (p *NodeWorkerPool) GetStats() map[string]interface{} {
        p.mu.RLock()
        defer p.mu.RUnlock()

        var totalTasks, successTasks, failedTasks int64
        var activeWorkers, idleWorkers int

        workerStats := make([]map[string]interface{}, 0, len(p.workers))

        for _, worker := range p.workers {
                stats := worker.GetStats()
                workerStats = append(workerStats, stats)

                totalTasks += worker.TotalTasks
                successTasks += worker.SuccessTasks
                failedTasks += worker.FailedTasks

                worker.mu.RLock()
                if worker.Status == WorkerNodeStatusIdle {
                        idleWorkers++
                } else if worker.Status == WorkerNodeStatusBusy {
                        activeWorkers++
                }
                worker.mu.RUnlock()
        }

        return map[string]interface{}{
                "totalWorkers":  len(p.workers),
                "activeWorkers": activeWorkers,
                "idleWorkers":   idleWorkers,
                "totalTasks":    totalTasks,
                "successTasks":  successTasks,
                "failedTasks":   failedTasks,
                "workers":       workerStats,
        }
}

// GetIdleWorker 获取空闲 Worker
func (p *NodeWorkerPool) GetIdleWorker() *WorkerNode {
        p.mu.RLock()
        defer p.mu.RUnlock()

        for _, worker := range p.workers {
                worker.mu.RLock()
                if worker.Status == WorkerNodeStatusIdle {
                        worker.mu.RUnlock()
                        return worker
                }
                worker.mu.RUnlock()
        }

        return nil
}

// ScaleUp 扩容 Worker
func (p *NodeWorkerPool) ScaleUp(count int) error {
        for i := 0; i < count; i++ {
                id := fmt.Sprintf("worker-%d", time.Now().UnixNano())
                _, err := p.AddWorker(id, 10)
                if err != nil {
                        return err
                }

                worker, _ := p.GetWorker(id)
                worker.Start()
                go p.collectResults(worker)
        }

        return nil
}

// ScaleDown 缩容 Worker
func (p *NodeWorkerPool) ScaleDown(count int) error {
        p.mu.Lock()
        defer p.mu.Unlock()

        var toRemove []string
        for id, worker := range p.workers {
                worker.mu.RLock()
                if worker.Status == WorkerNodeStatusIdle {
                        toRemove = append(toRemove, id)
                }
                worker.mu.RUnlock()

                if len(toRemove) >= count {
                        break
                }
        }

        if len(toRemove) < count {
                return fmt.Errorf("not enough idle workers to remove")
        }

        for _, id := range toRemove {
                p.workers[id].Stop()
                delete(p.workers, id)
        }

        return nil
}

// GlobalNodeWorkerPool 全局 Worker 池
var GlobalNodeWorkerPool *NodeWorkerPool

// InitNodeWorkerPool 初始化 Worker 池
func InitNodeWorkerPool(maxWorkers int) {
        GlobalNodeWorkerPool = NewNodeWorkerPool(maxWorkers)
}

// GetNodeWorkerPool 获取 Worker 池
func GetNodeWorkerPool() *NodeWorkerPool {
        if GlobalNodeWorkerPool == nil {
                InitNodeWorkerPool(100)
        }
        return GlobalNodeWorkerPool
}
