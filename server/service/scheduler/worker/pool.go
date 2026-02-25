package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"yunwei/global"
	schedulerModel "yunwei/model/scheduler"
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
	ID           string           `json:"id"`
	QueueName    string           `json:"queueName"`
	Status       PoolWorkerStatus `json:"status"`
	CurrentTask  uint             `json:"currentTask"`
	TasksHandled int64            `json:"tasksHandled"`
	LastError    string           `json:"lastError"`
	StartedAt    time.Time        `json:"startedAt"`
	LastActiveAt time.Time        `json:"lastActiveAt"`

	// 内部状态
	ctx      context.Context
	cancel   context.CancelFunc
	taskChan chan *schedulerModel.Task
	executor TaskExecutor
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Execute(task *schedulerModel.Task) (*TaskResult, error)
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
			taskChan:  make(chan *schedulerModel.Task, 1),
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
func (wp *WorkerPool) executeTask(worker *PoolWorker, task *schedulerModel.Task) {
	// 更新 Worker 状态
	worker.Status = PoolWorkerStatusBusy
	worker.CurrentTask = task.ID
	worker.LastActiveAt = time.Now()

	// 更新任务状态
	now := time.Now()
	task.Status = schedulerModel.TaskStatusRunning
	task.ServerName = worker.ID // 使用 ServerName 字段存储 WorkerID
	global.DB.Save(task)

	// 创建执行记录
	execution := &TaskExecution{
		TaskID:   task.ID,
		Attempt:  task.RetryCount + 1,
		Status:   schedulerModel.TaskStatusRunning,
		WorkerID: worker.ID,
		StartAt:  &now,
	}
	global.DB.Create(execution)

	// 记录事件
	recordTaskEvent(task.ID, "started", map[string]interface{}{
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
		execution.Status = schedulerModel.TaskStatusFailed
		execution.Error = result.Error
		execution.Stderr = result.Error

		// 处理失败
		wp.handleTaskFailure(worker, task, result, execErr)
	} else {
		execution.Status = schedulerModel.TaskStatusSuccess
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
func (wp *WorkerPool) handleTaskSuccess(worker *PoolWorker, task *schedulerModel.Task, result *TaskResult) {
	now := time.Now()
	task.Status = schedulerModel.TaskStatusSuccess
	task.CompletedAt = &now
	task.Duration = now.Sub(*task.StartedAt).Milliseconds()
	task.Output = result.Output
	global.DB.Save(task)

	// 确认任务
	wp.queue.AckTask(task.ID)

	// 记录事件
	recordTaskEvent(task.ID, "completed", map[string]interface{}{
		"duration":  task.Duration,
		"exit_code": result.ExitCode,
	}, worker.ID, "任务执行成功")
}

// handleTaskFailure 处理任务失败
func (wp *WorkerPool) handleTaskFailure(worker *PoolWorker, task *schedulerModel.Task, result *TaskResult, execErr error) {
	errMsg := result.Error
	if execErr != nil {
		errMsg = execErr.Error()
	}

	// 记录错误
	atomic.AddInt64(&wp.totalErrors, 1)
	worker.LastError = errMsg

	// 更新任务状态
	task.Status = schedulerModel.TaskStatusFailed
	task.Error = errMsg
	task.RetryCount++
	global.DB.Save(task)

	// 拒绝任务
	wp.queue.NackTask(task.ID, errMsg)

	// 记录事件
	recordTaskEvent(task.ID, "failed", map[string]interface{}{
		"error":   errMsg,
		"attempt": task.RetryCount,
	}, worker.ID, "任务执行失败")
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
				taskChan:  make(chan *schedulerModel.Task, 1),
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
	QueueName         string `json:"queueName"`
	TotalWorkers      int    `json:"totalWorkers"`
	IdleWorkers       int    `json:"idleWorkers"`
	BusyWorkers       int    `json:"busyWorkers"`
	StoppedWorkers    int    `json:"stoppedWorkers"`
	TotalTasksHandled int64  `json:"totalTasksHandled"`
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

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID        uint                   `json:"id" gorm:"primarykey"`
	CreatedAt time.Time              `json:"createdAt"`
	TaskID    uint                   `json:"taskId" gorm:"index"`
	Attempt   int                    `json:"attempt"`
	Status    schedulerModel.TaskStatus `json:"status" gorm:"type:varchar(16)"`
	WorkerID  string                 `json:"workerId" gorm:"type:varchar(32)"`
	StartAt   *time.Time             `json:"startAt"`
	EndAt     *time.Time             `json:"endAt"`
	Duration  int64                  `json:"duration"`
	Error     string                 `json:"error" gorm:"type:text"`
	Stdout    string                 `json:"stdout" gorm:"type:text"`
	Stderr    string                 `json:"stderr" gorm:"type:text"`
}

func (TaskExecution) TableName() string {
	return "scheduler_task_executions"
}

// recordTaskEvent 记录任务事件
func recordTaskEvent(taskID uint, eventType string, eventData interface{}, source, message string) {
	dataJSON, _ := json.Marshal(eventData)
	event := schedulerModel.TaskEvent{
		TaskID:  taskID,
		Type:    eventType,
		Source:  source,
		Message: message,
		Data:    string(dataJSON),
	}
	global.DB.Create(&event)
}
