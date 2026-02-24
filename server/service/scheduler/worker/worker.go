package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"yunwei/service/scheduler"
	"yunwei/service/scheduler/queue"
)

// WorkerStatus Worker 状态
type WorkerStatus string

const (
	WorkerStatusIdle     WorkerStatus = "idle"
	WorkerStatusBusy     WorkerStatus = "busy"
	WorkerStatusStopping WorkerStatus = "stopping"
	WorkerStatusStopped  WorkerStatus = "stopped"
	WorkerStatusError    WorkerStatus = "error"
)

// Worker Worker 节点
type Worker struct {
	ID           string       `json:"id"`
	NodeName     string       `json:"nodeName"`
	Status       WorkerStatus `json:"status"`
	
	// 执行统计
	TotalTasks   int64 `json:"totalTasks"`
	SuccessTasks int64 `json:"successTasks"`
	FailedTasks  int64 `json:"failedTasks"`
	
	// 当前任务
	CurrentTask  *scheduler.TaskQueueItem `json:"currentTask"`
	TaskStarted  time.Time                `json:"taskStarted"`
	
	// 配置
	MaxConcurrent int `json:"maxConcurrent"` // 最大并发数
	Timeout       int `json:"timeout"`       // 默认超时
	
	// 心跳
	LastHeartbeat time.Time `json:"lastHeartbeat"`
	
	// 控制
	ctx        context.Context
	cancel     context.CancelFunc
	taskChan   chan *scheduler.TaskQueueItem
	resultChan chan *TaskResult
	
	mu         sync.RWMutex
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID       uint                   `json:"taskId"`
	ExecutionID  string                 `json:"executionId"`
	ServerID     uint                   `json:"serverId"`
	Status       scheduler.TaskStatus   `json:"status"`
	Output       string                 `json:"output"`
	ErrorMessage string                 `json:"errorMessage"`
	Duration     int64                  `json:"duration"`
	RetryCount   int                    `json:"retryCount"`
}

// NewWorker 创建 Worker
func NewWorker(id string, maxConcurrent int) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Worker{
		ID:            id,
		NodeName:      fmt.Sprintf("worker-%s", id),
		Status:        WorkerStatusIdle,
		MaxConcurrent: maxConcurrent,
		Timeout:       300,
		ctx:           ctx,
		cancel:        cancel,
		taskChan:      make(chan *scheduler.TaskQueueItem, maxConcurrent*2),
		resultChan:    make(chan *TaskResult, maxConcurrent*2),
	}
}

// Start 启动 Worker
func (w *Worker) Start() {
	w.mu.Lock()
	w.Status = WorkerStatusIdle
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
func (w *Worker) Stop() {
	w.mu.Lock()
	w.Status = WorkerStatusStopping
	w.mu.Unlock()
	
	w.cancel()
	
	// 从队列注销
	queue.GetQueue().UnregisterWorker(w.ID)
	
	close(w.taskChan)
	close(w.resultChan)
	
	w.mu.Lock()
	w.Status = WorkerStatusStopped
	w.mu.Unlock()
}

// runTaskProcessor 任务处理器
func (w *Worker) runTaskProcessor() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case item, ok := <-w.taskChan:
			if !ok {
				return
			}
			
			w.processTask(item)
		}
	}
}

// processTask 处理任务
func (w *Worker) processTask(item *scheduler.TaskQueueItem) {
	w.mu.Lock()
	w.Status = WorkerStatusBusy
	w.CurrentTask = item
	w.TaskStarted = time.Now()
	w.mu.Unlock()
	
	startTime := time.Now()
	
	// 获取任务信息
	task, err := scheduler.GetTask(item.TaskID)
	if err != nil {
		w.sendResult(&TaskResult{
			TaskID:       item.TaskID,
			ExecutionID:  item.ExecutionID,
			ServerID:     item.ServerID,
			Status:       scheduler.TaskStatusFailed,
			ErrorMessage: fmt.Sprintf("获取任务失败: %s", err.Error()),
		})
		return
	}
	
	// 创建执行记录
	execution := &scheduler.TaskExecution{
		TaskID:          item.TaskID,
		ExecutionID:     item.ExecutionID,
		IdempotentKey:   task.IdempotentKey,
		Status:          scheduler.TaskStatusRunning,
		TargetServerID:  item.ServerID,
		BatchIndex:      item.BatchIndex,
		BatchTotal:      item.BatchTotal,
		Executor:        task.Executor,
		Action:          task.Action,
		Params:          task.Params,
		StartedAt:       &startTime,
		WorkerID:        w.ID,
		WorkerNode:      w.NodeName,
		IsRetry:         item.IsRetry,
		RetryCount:      item.RetryCount,
	}
	
	scheduler.CreateTaskExecution(execution)
	
	// 执行任务
	result := w.executeTask(task, item)
	
	// 更新执行记录
	endTime := time.Now()
	execution.Status = result.Status
	execution.Output = result.Output
	execution.ErrorMessage = result.ErrorMessage
	execution.Duration = result.Duration
	execution.CompletedAt = &endTime
	scheduler.UpdateTaskExecution(execution)
	
	// 更新 Worker 统计
	atomic.AddInt64(&w.TotalTasks, 1)
	if result.Status == scheduler.TaskStatusSuccess {
		atomic.AddInt64(&w.SuccessTasks, 1)
	} else {
		atomic.AddInt64(&w.FailedTasks, 1)
	}
	
	// 发送结果
	w.sendResult(result)
	
	// 更新状态
	w.mu.Lock()
	w.Status = WorkerStatusIdle
	w.CurrentTask = nil
	w.mu.Unlock()
	
	// 确认任务完成
	queue.GetQueue().Ack(item.ExecutionID)
}

// executeTask 执行任务
func (w *Worker) executeTask(task *scheduler.Task, item *scheduler.TaskQueueItem) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}
	
	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()
	
	// 根据执行器类型执行
	switch task.Executor {
	case "shell", "command":
		output, err := w.executeShell(task.Action)
		if err != nil {
			result.Status = scheduler.TaskStatusFailed
			result.ErrorMessage = err.Error()
		} else {
			result.Status = scheduler.TaskStatusSuccess
			result.Output = output
		}
		
	case "http", "api":
		output, err := w.executeHTTP(task.Action, task.Params)
		if err != nil {
			result.Status = scheduler.TaskStatusFailed
			result.ErrorMessage = err.Error()
		} else {
			result.Status = scheduler.TaskStatusSuccess
			result.Output = output
		}
		
	case "docker":
		output, err := w.executeDocker(task.Action, task.Params)
		if err != nil {
			result.Status = scheduler.TaskStatusFailed
			result.ErrorMessage = err.Error()
		} else {
			result.Status = scheduler.TaskStatusSuccess
			result.Output = output
		}
		
	default:
		result.Status = scheduler.TaskStatusFailed
		result.ErrorMessage = fmt.Sprintf("未知的执行器类型: %s", task.Executor)
	}
	
	result.RetryCount = item.RetryCount
	
	return result
}

// executeShell 执行 Shell 命令
func (w *Worker) executeShell(command string) (string, error) {
	// TODO: 实际执行 Shell 命令
	// 这里需要通过 SSH 或本地执行
	return fmt.Sprintf("Executed: %s", command), nil
}

// executeHTTP 执行 HTTP 请求
func (w *Worker) executeHTTP(url string, params string) (string, error) {
	// TODO: 实际执行 HTTP 请求
	return fmt.Sprintf("HTTP Request to: %s", url), nil
}

// executeDocker 执行 Docker 命令
func (w *Worker) executeDocker(action string, params string) (string, error) {
	// TODO: 实际执行 Docker 命令
	return fmt.Sprintf("Docker: %s", action), nil
}

// sendResult 发送结果
func (w *Worker) sendResult(result *TaskResult) {
	select {
	case w.resultChan <- result:
	default:
		// 结果通道已满，记录日志
	}
}

// heartbeatLoop 心跳循环
func (w *Worker) heartbeatLoop() {
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
func (w *Worker) taskFetcher() {
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
			item, err := q.Dequeue()
			if err != nil || item == nil {
				continue
			}
			
			// 发送到任务通道
			select {
			case w.taskChan <- item:
			default:
				// 任务通道已满，放回队列
				q.Nack(item.ExecutionID)
			}
		}
	}
}

// GetResultChan 获取结果通道
func (w *Worker) GetResultChan() <-chan *TaskResult {
	return w.resultChan
}

// GetStats 获取统计信息
func (w *Worker) GetStats() map[string]interface{} {
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

// WorkerPool Worker 池
type WorkerPool struct {
	workers    map[string]*Worker
	mu         sync.RWMutex
	resultChan chan *TaskResult
	ctx        context.Context
	cancel     context.CancelFunc
	
	// 配置
	maxWorkers    int
	workerTimeout time.Duration
}

// NewWorkerPool 创建 Worker 池
func NewWorkerPool(maxWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workers:       make(map[string]*Worker),
		resultChan:    make(chan *TaskResult, 1000),
		ctx:           ctx,
		cancel:        cancel,
		maxWorkers:    maxWorkers,
		workerTimeout: 30 * time.Second,
	}
}

// AddWorker 添加 Worker
func (p *WorkerPool) AddWorker(id string, maxConcurrent int) (*Worker, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.workers) >= p.maxWorkers {
		return nil, fmt.Errorf("worker pool is full")
	}
	
	if _, exists := p.workers[id]; exists {
		return nil, fmt.Errorf("worker already exists: %s", id)
	}
	
	worker := NewWorker(id, maxConcurrent)
	p.workers[id] = worker
	
	return worker, nil
}

// RemoveWorker 移除 Worker
func (p *WorkerPool) RemoveWorker(id string) error {
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
func (p *WorkerPool) StartAll() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	for _, worker := range p.workers {
		worker.Start()
		go p.collectResults(worker)
	}
}

// StopAll 停止所有 Worker
func (p *WorkerPool) StopAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, worker := range p.workers {
		worker.Stop()
	}
	
	p.cancel()
	close(p.resultChan)
}

// collectResults 收集结果
func (p *WorkerPool) collectResults(worker *Worker) {
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
func (p *WorkerPool) GetResultChan() <-chan *TaskResult {
	return p.resultChan
}

// GetWorker 获取 Worker
func (p *WorkerPool) GetWorker(id string) (*Worker, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	worker, exists := p.workers[id]
	if !exists {
		return nil, fmt.Errorf("worker not found: %s", id)
	}
	
	return worker, nil
}

// GetWorkers 获取所有 Worker
func (p *WorkerPool) GetWorkers() []*Worker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	workers := make([]*Worker, 0, len(p.workers))
	for _, worker := range p.workers {
		workers = append(workers, worker)
	}
	
	return workers
}

// GetStats 获取统计信息
func (p *WorkerPool) GetStats() map[string]interface{} {
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
		if worker.Status == WorkerStatusIdle {
			idleWorkers++
		} else if worker.Status == WorkerStatusBusy {
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
func (p *WorkerPool) GetIdleWorker() *Worker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	for _, worker := range p.workers {
		worker.mu.RLock()
		if worker.Status == WorkerStatusIdle {
			worker.mu.RUnlock()
			return worker
		}
		worker.mu.RUnlock()
	}
	
	return nil
}

// ScaleUp 扩容 Worker
func (p *WorkerPool) ScaleUp(count int) error {
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
func (p *WorkerPool) ScaleDown(count int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	var toRemove []string
	for id, worker := range p.workers {
		worker.mu.RLock()
		if worker.Status == WorkerStatusIdle {
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

// GlobalWorkerPool 全局 Worker 池
var GlobalWorkerPool *WorkerPool

// InitWorkerPool 初始化 Worker 池
func InitWorkerPool(maxWorkers int) {
	GlobalWorkerPool = NewWorkerPool(maxWorkers)
}

// GetWorkerPool 获取 Worker 池
func GetWorkerPool() *WorkerPool {
	if GlobalWorkerPool == nil {
		InitWorkerPool(100)
	}
	return GlobalWorkerPool
}
