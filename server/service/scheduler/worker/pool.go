package worker

import (
        "context"
        "fmt"
        "sync"
        "sync/atomic"
        "time"

        "yunwei/global"
        "yunwei/service/scheduler"
        "yunwei/service/scheduler/queue"
)

// PoolWorkerStatus Worker 状态
type PoolWorkerStatus string

const (
        PoolWorkerStatusIdle     PoolWorkerStatus = "idle"
        PoolWorkerStatusBusy     PoolWorkerStatus = "busy"
        PoolWorkerStatusStopped  PoolWorkerStatus = "stopped"
        PoolWorkerStatusError    PoolWorkerStatus = "error"
)

// PoolWorker Worker 实例
type PoolWorker struct {
        ID           string            `json:"id"`
        QueueName    string            `json:"queueName"`
        Status       PoolWorkerStatus  `json:"status"`
        CurrentTask  uint              `json:"currentTask"`
        TasksHandled int64             `json:"tasksHandled"`
        LastError    string            `json:"lastError"`
        StartedAt    time.Time         `json:"startedAt"`
        LastActiveAt time.Time         `json:"lastActiveAt"`

        // 内部状态
        ctx      context.Context
        cancel   context.CancelFunc
        taskChan chan *scheduler.Task
        executor TaskExecutor
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
        Execute(task *scheduler.Task) (*TaskResult, error)
        Cancel(taskID uint) error
        IsRunning(taskID uint) bool
}

// TaskResult 任务执行结果
type TaskResult struct {
        Success  bool   `json:"success"`
        Output   string `json:"output"`
        Error    string `json:"error"`
        ExitCode int    `json:"exitCode"`
        Duration int64  `json:"duration"` // 毫秒
}

// WorkerPool Worker 池
type WorkerPool struct {
        queue      *queue.TaskQueue
        workers    map[string]*PoolWorker
        executor   TaskExecutor
        mu         sync.RWMutex
        ctx        context.Context
        cancel     context.CancelFunc

        // 统计
        totalTasksHandled int64
        totalErrors       int64
}

// NewWorkerPool 创建 Worker 池
func NewWorkerPool(taskQueue *queue.TaskQueue, executor TaskExecutor) *WorkerPool {
        ctx, cancel := context.WithCancel(context.Background())
        return &WorkerPool{
                queue:    taskQueue,
                workers:  make(map[string]*PoolWorker),
                executor: executor,
                ctx:      ctx,
                cancel:   cancel,
        }
}

// Start 启动 Worker 池
func (wp *WorkerPool) Start(queueName string, numWorkers int) error {
        wp.mu.Lock()
        defer wp.mu.Unlock()

        for i := 0; i < numWorkers; i++ {
                workerID := fmt.Sprintf("%s-worker-%d-%d", queueName, time.Now().Unix(), i)
                worker := &PoolWorker{
                        ID:        workerID,
                        QueueName: queueName,
                        Status:    PoolWorkerStatusIdle,
                        StartedAt: time.Now(),
                        taskChan:  make(chan *scheduler.Task, 1),
                        executor:  wp.executor,
                }

                workerCtx, workerCancel := context.WithCancel(wp.ctx)
                worker.ctx = workerCtx
                worker.cancel = workerCancel

                wp.workers[workerID] = worker

                // 启动 Worker 协程
                go wp.runWorker(worker)
        }

        return nil
}

// runWorker 运行 Worker
func (wp *WorkerPool) runWorker(worker *PoolWorker) {
        for {
                select {
                case <-worker.ctx.Done():
                        worker.Status = PoolWorkerStatusStopped
                        return

                default:
                        // 从队列获取任务
                        task, err := wp.queue.DequeueTask(worker.QueueName, 5*time.Second)
                        if err != nil {
                                continue
                        }

                        // 执行任务
                        wp.executeTask(worker, task)
                }
        }
}

// executeTask 执行任务
func (wp *WorkerPool) executeTask(worker *PoolWorker, task *scheduler.Task) {
        // 更新 Worker 状态
        worker.Status = PoolWorkerStatusBusy
        worker.CurrentTask = task.ID
        worker.LastActiveAt = time.Now()
        
        // 更新任务状态
        now := time.Now()
        task.Status = scheduler.TaskStatusRunning
        task.WorkerID = worker.ID
        task.StartAt = &now
        global.DB.Save(task)
        
        // 创建执行记录
        execution := &scheduler.TaskExecution{
                TaskID:   task.ID,
                Attempt:  task.RetryCount + 1,
                Status:   scheduler.TaskStatusRunning,
                WorkerID: worker.ID,
                StartAt:  &now,
        }
        scheduler.CreateTaskExecution(execution)
        
        // 记录事件
        scheduler.RecordTaskEvent(task.ID, "started", map[string]interface{}{
                "worker_id": worker.ID,
                "attempt":   execution.Attempt,
        }, worker.ID, "任务开始执行")
        
        // 设置超时
        ctx, cancel := context.WithTimeout(worker.ctx, time.Duration(task.Timeout)*time.Second)
        defer cancel()
        
        // 执行任务
        var result *TaskResult
        var execErr error
        
        done := make(chan struct{})
        go func() {
                result, execErr = wp.executor.Execute(task)
                close(done)
        }()
        
        select {
        case <-done:
                // 执行完成
        case <-ctx.Done():
                // 超时
                execErr = fmt.Errorf("task timeout after %d seconds", task.Timeout)
                result = &TaskResult{
                        Success: false,
                        Error:   execErr.Error(),
                }
        }
        
        // 更新执行记录
        endNow := time.Now()
        execution.EndAt = &endNow
        execution.Duration = endNow.Sub(*execution.StartAt).Milliseconds()
        
        if execErr != nil || !result.Success {
                execution.Status = scheduler.TaskStatusFailed
                execution.Error = result.Error
                execution.Stderr = result.Error
                
                // 处理失败
                wp.handleTaskFailure(worker, task, result, execErr)
        } else {
                execution.Status = scheduler.TaskStatusSuccess
                execution.Stdout = result.Output
                execution.ExitCode = result.ExitCode
                
                // 更新任务成功
                wp.handleTaskSuccess(worker, task, result)
        }
        
        global.DB.Save(execution)
        
        // 更新 Worker 状态
        worker.Status = PoolWorkerStatusIdle
        worker.CurrentTask = 0
        worker.TasksHandled++
        worker.LastActiveAt = time.Now()
        atomic.AddInt64(&wp.totalTasksHandled, 1)
}

// handleTaskSuccess 处理任务成功
func (wp *WorkerPool) handleTaskSuccess(worker *PoolWorker, task *scheduler.Task, result *TaskResult) {
        now := time.Now()
        task.Status = scheduler.TaskStatusSuccess
        task.EndAt = &now
        task.Duration = now.Sub(*task.StartAt).Milliseconds()
        task.Result = result.Output
        task.Stdout = result.Output
        task.ExitCode = result.ExitCode
        global.DB.Save(task)
        
        // 确认任务
        wp.queue.AckTask(task.ID)
        
        // 记录事件
        scheduler.RecordTaskEvent(task.ID, "completed", map[string]interface{}{
                "duration": task.Duration,
                "exit_code": result.ExitCode,
        }, worker.ID, "任务执行成功")
        
        // 检查依赖此任务的其他任务
        wp.checkDependentTasks(task)
}

// handleTaskFailure 处理任务失败
func (wp *WorkerPool) handleTaskFailure(worker *PoolWorker, task *scheduler.Task, result *TaskResult, execErr error) {
        errMsg := result.Error
        if execErr != nil {
                errMsg = execErr.Error()
        }
        
        // 记录错误
        atomic.AddInt64(&wp.totalErrors, 1)
        worker.LastError = errMsg
        
        // 拒绝任务（可能触发重试）
        wp.queue.NackTask(task.ID, errMsg)
        
        // 记录事件
        scheduler.RecordTaskEvent(task.ID, "failed", map[string]interface{}{
                "error":   errMsg,
                "attempt": task.RetryCount,
        }, worker.ID, "任务执行失败")
}

// checkDependentTasks 检查依赖任务
func (wp *WorkerPool) checkDependentTasks(completedTask *scheduler.Task) {
        // 查找依赖此任务的其他任务
        var dependentTasks []scheduler.Task
        global.DB.Where("status = ? AND dependencies LIKE ?", scheduler.TaskStatusPending, fmt.Sprintf("%%\"%d\"%%", completedTask.ID)).Find(&dependentTasks)
        
        for _, depTask := range dependentTasks {
                // 检查所有依赖是否满足
                var depIDs []uint
                if depTask.Dependencies != "" {
                        // 解析依赖
                        // 简化处理，实际需要解析 JSON
                }
                
                // 检查依赖条件
                allSuccess := true
                for _, depID := range depIDs {
                        var dep scheduler.Task
                        if global.DB.First(&dep, depID).Error == nil {
                                if dep.Status != scheduler.TaskStatusSuccess {
                                        allSuccess = false
                                        break
                                }
                        }
                }
                
                if allSuccess {
                        // 所有依赖满足，可以执行
                        wp.queue.EnqueueTask(&depTask)
                }
        }
}

// Stop 停止 Worker 池
func (wp *WorkerPool) Stop() {
        wp.cancel()
        
        wp.mu.Lock()
        defer wp.mu.Unlock()
        
        for _, worker := range wp.workers {
                if worker.cancel != nil {
                        worker.cancel()
                }
                worker.Status = PoolWorkerStatusStopped
        }
}

// Scale 调整 Worker 数量
func (wp *WorkerPool) Scale(queueName string, targetWorkers int) error {
        wp.mu.Lock()
        defer wp.mu.Unlock()

        // 统计当前 Worker 数量
        currentWorkers := 0
        for _, w := range wp.workers {
                if w.QueueName == queueName && w.Status != PoolWorkerStatusStopped {
                        currentWorkers++
                }
        }

        if targetWorkers > currentWorkers {
                // 增加 Worker
                for i := 0; i < targetWorkers-currentWorkers; i++ {
                        workerID := fmt.Sprintf("%s-worker-%d-%d", queueName, time.Now().Unix(), i)
                        worker := &PoolWorker{
                                ID:        workerID,
                                QueueName: queueName,
                                Status:    PoolWorkerStatusIdle,
                                StartedAt: time.Now(),
                                taskChan:  make(chan *scheduler.Task, 1),
                                executor:  wp.executor,
                        }
                        
                        workerCtx, workerCancel := context.WithCancel(wp.ctx)
                        worker.ctx = workerCtx
                        worker.cancel = workerCancel
                        
                        wp.workers[workerID] = worker
                        go wp.runWorker(worker)
                }
        } else if targetWorkers < currentWorkers {
                // 减少 Worker
                toRemove := currentWorkers - targetWorkers
                removed := 0
                for id, w := range wp.workers {
                        if w.QueueName == queueName && w.Status == PoolWorkerStatusIdle {
                                w.cancel()
                                w.Status = PoolWorkerStatusStopped
                                delete(wp.workers, id)
                                removed++
                                if removed >= toRemove {
                                        break
                                }
                        }
                }
        }

        return nil
}

// GetWorkerStats 获取 Worker 统计
func (wp *WorkerPool) GetWorkerStats(queueName string) *WorkerPoolStats {
        wp.mu.RLock()
        defer wp.mu.RUnlock()

        stats := &WorkerPoolStats{
                QueueName: queueName,
        }

        for _, w := range wp.workers {
                if w.QueueName == queueName {
                        stats.TotalWorkers++
                        switch w.Status {
                        case PoolWorkerStatusIdle:
                                stats.IdleWorkers++
                        case PoolWorkerStatusBusy:
                                stats.BusyWorkers++
                        case PoolWorkerStatusStopped:
                                stats.StoppedWorkers++
                        }
                        stats.TotalTasksHandled += w.TasksHandled
                }
        }

        return stats
}

// WorkerPoolStats Worker 池统计
type WorkerPoolStats struct {
        QueueName        string `json:"queueName"`
        TotalWorkers     int    `json:"totalWorkers"`
        IdleWorkers      int    `json:"idleWorkers"`
        BusyWorkers      int    `json:"busyWorkers"`
        StoppedWorkers   int    `json:"stoppedWorkers"`
        TotalTasksHandled int64 `json:"totalTasksHandled"`
}

// GetAllWorkerStats 获取所有 Worker 统计
func (wp *WorkerPool) GetAllWorkerStats() map[string]*WorkerPoolStats {
        wp.mu.RLock()
        defer wp.mu.RUnlock()

        stats := make(map[string]*WorkerPoolStats)

        for _, w := range wp.workers {
                if _, exists := stats[w.QueueName]; !exists {
                        stats[w.QueueName] = &WorkerPoolStats{
                                QueueName: w.QueueName,
                        }
                }

                stats[w.QueueName].TotalWorkers++
                switch w.Status {
                case PoolWorkerStatusIdle:
                        stats[w.QueueName].IdleWorkers++
                case PoolWorkerStatusBusy:
                        stats[w.QueueName].BusyWorkers++
                case PoolWorkerStatusStopped:
                        stats[w.QueueName].StoppedWorkers++
                }
                stats[w.QueueName].TotalTasksHandled += w.TasksHandled
        }

        return stats
}

// GetWorker 获取 Worker
func (wp *WorkerPool) GetWorker(workerID string) (*PoolWorker, error) {
        wp.mu.RLock()
        defer wp.mu.RUnlock()

        worker, exists := wp.workers[workerID]
        if !exists {
                return nil, fmt.Errorf("worker not found: %s", workerID)
        }
        return worker, nil
}

// ListWorkers 列出所有 Worker
func (wp *WorkerPool) ListWorkers(queueName string) []*PoolWorker {
        wp.mu.RLock()
        defer wp.mu.RUnlock()

        var workers []*PoolWorker
        for _, w := range wp.workers {
                if queueName == "" || w.QueueName == queueName {
                        workers = append(workers, w)
                }
        }
        return workers
}

// GetTotalStats 获取总统计
func (wp *WorkerPool) GetTotalStats() (int64, int64) {
        return atomic.LoadInt64(&wp.totalTasksHandled), atomic.LoadInt64(&wp.totalErrors)
}
