package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	schedulerModel "yunwei/model/scheduler"
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
	CurrentTask  *schedulerModel.Task `json:"currentTask"`
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
	resultChan chan *TaskResult

	mu sync.RWMutex
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
		taskChan:      make(chan *schedulerModel.Task, maxConcurrent*2),
		resultChan:    make(chan *TaskResult, maxConcurrent*2),
	}
}

// Start 启动 Worker
func (w *Worker) Start() {
	w.mu.Lock()
	w.Status = WorkerStatusIdle
	w.LastHeartbeat = time.Now()
	w.mu.Unlock()

	// 启动任务处理协程
	for i := 0; i < w.MaxConcurrent; i++ {
		go w.runTaskProcessor()
	}

	// 启动心跳
	go w.heartbeatLoop()
}

// Stop 停止 Worker
func (w *Worker) Stop() {
	w.mu.Lock()
	w.Status = WorkerStatusStopping
	w.mu.Unlock()

	w.cancel()

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
		case task, ok := <-w.taskChan:
			if !ok {
				return
			}

			w.processTask(task)
		}
	}
}

// processTask 处理任务
func (w *Worker) processTask(task *schedulerModel.Task) {
	w.mu.Lock()
	w.Status = WorkerStatusBusy
	w.CurrentTask = task
	w.TaskStarted = time.Now()
	w.mu.Unlock()

	startTime := time.Now()

	// 执行任务
	result := w.executeTask(task)

	// 更新执行记录
	result.Duration = time.Since(startTime).Milliseconds()

	// 更新 Worker 统计
	atomic.AddInt64(&w.TotalTasks, 1)
	if result.Success {
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
}

// executeTask 执行任务
func (w *Worker) executeTask(task *schedulerModel.Task) *TaskResult {
	result := &TaskResult{}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 根据执行器类型执行
	switch task.Type {
	case schedulerModel.TaskTypeCommand, schedulerModel.TaskTypeScript:
		output, err := w.executeShell(task.Command)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Output = output
		}

	default:
		// 简化处理，其他类型都返回成功
		result.Success = true
		result.Output = fmt.Sprintf("Task %s executed", task.Name)
	}

	return result
}

// executeShell 执行 Shell 命令
func (w *Worker) executeShell(command string) (string, error) {
	// TODO: 实际执行 Shell 命令
	return fmt.Sprintf("Executed: %s", command), nil
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
