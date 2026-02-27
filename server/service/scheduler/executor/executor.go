package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"yunwei/global"
	schedulerModel "yunwei/model/scheduler"
	"yunwei/service/ssh"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusSuccess    TaskStatus = "success"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusRetrying   TaskStatus = "retrying"
	TaskStatusCanceled   TaskStatus = "canceled"
	TaskStatusTimeout    TaskStatus = "timeout"
	TaskStatusRolledback TaskStatus = "rolledback"
)

// TaskResult 任务结果
type TaskResult struct {
	TaskID       uint       `json:"taskId"`
	ExecutionID  string     `json:"executionId"`
	ServerID     uint       `json:"serverId"`
	Status       TaskStatus `json:"status"`
	Output       string     `json:"output"`
	ErrorMessage string     `json:"errorMessage"`
	Duration     int64      `json:"duration"`
	RetryCount   int        `json:"retryCount"`
}

// TaskQueueItem 任务队列项
type TaskQueueItem struct {
	TaskID      uint   `json:"taskId"`
	ExecutionID string `json:"executionId"`
	ServerID    uint   `json:"serverId"`
	Priority    int    `json:"priority"`
}

// TaskExecutor 任务执行器
type TaskExecutor struct {
	sshPool    *ssh.SSHPool
	executions map[string]*ExecutionContext
	mu         sync.RWMutex
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	ExecutionID string
	TaskID      uint
	ServerID    uint
	CancelFunc  context.CancelFunc
	StartTime   time.Time
	Status      TaskStatus
	Output      strings.Builder
	Error       error
}

// NewTaskExecutor 创建任务执行器
func NewTaskExecutor() *TaskExecutor {
	return &TaskExecutor{
		sshPool:    ssh.NewSSHPool(),
		executions: make(map[string]*ExecutionContext),
	}
}

// Execute 执行任务
func (e *TaskExecutor) Execute(item *TaskQueueItem) (*TaskResult, error) {
	// 获取任务信息
	var task schedulerModel.Task
	if err := global.DB.First(&task, item.TaskID).Error; err != nil {
		return nil, err
	}

	// 创建执行上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Timeout)*time.Second)

	execCtx := &ExecutionContext{
		ExecutionID: item.ExecutionID,
		TaskID:      item.TaskID,
		ServerID:    item.ServerID,
		CancelFunc:  cancel,
		StartTime:   time.Now(),
		Status:      TaskStatusRunning,
	}

	e.mu.Lock()
	e.executions[item.ExecutionID] = execCtx
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.executions, item.ExecutionID)
		e.mu.Unlock()
		cancel()
	}()

	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
	}

	// 解析参数
	var params map[string]interface{}
	if task.Params != "" {
		json.Unmarshal([]byte(task.Params), &params)
	}

	// 根据执行器类型执行
	switch task.Type {
	case "command", "shell":
		result = e.executeShell(ctx, &task, item, params)
	case "ssh":
		result = e.executeSSH(ctx, &task, item, params)
	case "http", "api":
		result = e.executeHTTP(ctx, &task, item, params)
	case "docker":
		result = e.executeDocker(ctx, &task, item, params)
	case "kubernetes", "k8s":
		result = e.executeKubernetes(ctx, &task, item, params)
	case "script":
		result = e.executeScript(ctx, &task, item, params)
	default:
		result.Status = TaskStatusFailed
		result.ErrorMessage = fmt.Sprintf("未知的执行器类型: %s", task.Type)
	}

	// 更新执行记录
	e.updateExecution(item.ExecutionID, result)

	return result, nil
}

// executeShell 执行 Shell 命令
func (e *TaskExecutor) executeShell(ctx context.Context, task *schedulerModel.Task, item *TaskQueueItem, params map[string]interface{}) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 替换参数
	command := e.replaceParams(task.Command, params)

	// 本地执行 - 模拟
	output := fmt.Sprintf("Executed: %s", command)

	result.Status = TaskStatusSuccess
	result.Output = output

	return result
}

// executeSSH 通过 SSH 执行
func (e *TaskExecutor) executeSSH(ctx context.Context, task *schedulerModel.Task, item *TaskQueueItem, params map[string]interface{}) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 替换参数
	command := e.replaceParams(task.Command, params)

	// 获取服务器信息
	if item.ServerID == 0 {
		result.Status = TaskStatusFailed
		result.ErrorMessage = "未指定目标服务器"
		return result
	}

	// 通过 SSH 执行 - 模拟
	output := fmt.Sprintf("SSH Executed on server %d: %s", item.ServerID, command)

	result.Status = TaskStatusSuccess
	result.Output = output

	return result
}

// executeHTTP 执行 HTTP 请求
func (e *TaskExecutor) executeHTTP(ctx context.Context, task *schedulerModel.Task, item *TaskQueueItem, params map[string]interface{}) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 模拟 HTTP 请求执行
	output := fmt.Sprintf("HTTP Request: %s", task.Command)

	result.Status = TaskStatusSuccess
	result.Output = output

	return result
}

// executeDocker 执行 Docker 命令
func (e *TaskExecutor) executeDocker(ctx context.Context, task *schedulerModel.Task, item *TaskQueueItem, params map[string]interface{}) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 替换参数
	command := e.replaceParams(task.Command, params)

	// 模拟 Docker 命令执行
	output := fmt.Sprintf("Docker Executed: %s", command)

	result.Status = TaskStatusSuccess
	result.Output = output

	return result
}

// executeKubernetes 执行 Kubernetes 命令
func (e *TaskExecutor) executeKubernetes(ctx context.Context, task *schedulerModel.Task, item *TaskQueueItem, params map[string]interface{}) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 模拟 Kubernetes 命令执行
	output := fmt.Sprintf("K8s Executed: %s", task.Command)

	result.Status = TaskStatusSuccess
	result.Output = output

	return result
}

// executeScript 执行脚本
func (e *TaskExecutor) executeScript(ctx context.Context, task *schedulerModel.Task, item *TaskQueueItem, params map[string]interface{}) *TaskResult {
	result := &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		ServerID:    item.ServerID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime).Milliseconds()
	}()

	// 模拟脚本执行
	output := fmt.Sprintf("Script Executed: %s", task.Command)

	result.Status = TaskStatusSuccess
	result.Output = output

	return result
}

// replaceParams 替换参数
func (e *TaskExecutor) replaceParams(template string, params map[string]interface{}) string {
	result := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// updateExecution 更新执行记录
func (e *TaskExecutor) updateExecution(executionID string, result *TaskResult) {
	now := time.Now()
	global.DB.Model(&schedulerModel.Task{}).
		Where("id = ?", result.TaskID).
		Updates(map[string]interface{}{
			"status":  string(result.Status),
			"output":  result.Output,
			"error":   result.ErrorMessage,
			"duration": result.Duration,
		})

	// 记录事件
	event := &schedulerModel.TaskEvent{
		TaskID:  result.TaskID,
		Type:    string(result.Status),
		Source:  "executor",
		Message: result.Output,
	}
	if result.ErrorMessage != "" {
		event.Message = result.ErrorMessage
	}
	global.DB.Create(event)
}

// Cancel 取消执行
func (e *TaskExecutor) Cancel(executionID string) error {
	e.mu.RLock()
	execCtx, exists := e.executions[executionID]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("执行记录不存在: %s", executionID)
	}

	execCtx.CancelFunc()
	execCtx.Status = TaskStatusCanceled

	return nil
}

// GetStatus 获取执行状态
func (e *TaskExecutor) GetStatus(executionID string) (TaskStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	execCtx, exists := e.executions[executionID]
	if !exists {
		return "", fmt.Errorf("执行记录不存在: %s", executionID)
	}

	return execCtx.Status, nil
}

// BatchExecutor 批量执行器
type BatchExecutor struct {
	executor   *TaskExecutor
	batchSize  int
	interval   time.Duration
}

// NewBatchExecutor 创建批量执行器
func NewBatchExecutor(batchSize int, interval time.Duration) *BatchExecutor {
	return &BatchExecutor{
		executor:  NewTaskExecutor(),
		batchSize: batchSize,
		interval:  interval,
	}
}

// ExecuteBatch 批量执行
func (e *BatchExecutor) ExecuteBatch(items []*TaskQueueItem) []*TaskResult {
	var results []*TaskResult

	for i := 0; i < len(items); i += e.batchSize {
		end := i + e.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		// 并发执行当前批次
		var wg sync.WaitGroup
		batchResults := make([]*TaskResult, len(batch))

		for j, item := range batch {
			wg.Add(1)
			go func(idx int, taskItem *TaskQueueItem) {
				defer wg.Done()
				result, _ := e.executor.Execute(taskItem)
				batchResults[idx] = result
			}(j, item)
		}

		wg.Wait()
		results = append(results, batchResults...)

		// 批次间隔
		if i+e.batchSize < len(items) && e.interval > 0 {
			time.Sleep(e.interval)
		}
	}

	return results
}

// RetryHandler 重试处理器
type RetryHandler struct {
	maxRetries     int
	retryInterval  time.Duration
	retryOnTimeout bool
	retryOnFail    bool
}

// NewRetryHandler 创建重试处理器
func NewRetryHandler(maxRetries int, retryInterval time.Duration) *RetryHandler {
	return &RetryHandler{
		maxRetries:     maxRetries,
		retryInterval:  retryInterval,
		retryOnTimeout: true,
		retryOnFail:    true,
	}
}

// ExecuteWithRetry 带重试的执行
func (h *RetryHandler) ExecuteWithRetry(executor *TaskExecutor, item *TaskQueueItem) *TaskResult {
	var result *TaskResult

	for retry := 0; retry <= h.maxRetries; retry++ {
		if retry > 0 {
			time.Sleep(h.retryInterval)
		}

		result, _ = executor.Execute(item)

		// 检查是否需要重试
		if result.Status == TaskStatusSuccess {
			return result
		}

		if result.Status == TaskStatusTimeout && !h.retryOnTimeout {
			return result
		}

		if result.Status == TaskStatusFailed && !h.retryOnFail {
			return result
		}

		result.RetryCount = retry
	}

	return result
}

// IdempotentHandler 幂等处理器
type IdempotentHandler struct {
	cache map[string]*IdempotentRecord
	mu    sync.RWMutex
}

// IdempotentRecord 幂等记录
type IdempotentRecord struct {
	Key        string
	Status     TaskStatus
	Result     *TaskResult
	ExpireAt   time.Time
	ExecutedAt time.Time
}

// NewIdempotentHandler 创建幂等处理器
func NewIdempotentHandler() *IdempotentHandler {
	return &IdempotentHandler{
		cache: make(map[string]*IdempotentRecord),
	}
}

// Check 检查幂等
func (h *IdempotentHandler) Check(key string) (*IdempotentRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	record, exists := h.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(record.ExpireAt) {
		return nil, false
	}

	return record, true
}

// Set 设置幂等记录
func (h *IdempotentHandler) Set(key string, result *TaskResult, ttl time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.cache[key] = &IdempotentRecord{
		Key:        key,
		Status:     result.Status,
		Result:     result,
		ExpireAt:   time.Now().Add(ttl),
		ExecutedAt: time.Now(),
	}
}

// ExecuteWithIdempotent 带幂等的执行
func (h *IdempotentHandler) ExecuteWithIdempotent(executor *TaskExecutor, item *TaskQueueItem, key string, ttl time.Duration) *TaskResult {
	// 检查幂等
	if record, exists := h.Check(key); exists {
		// 如果任务还在执行中，等待完成
		if record.Status == TaskStatusRunning {
			// 等待执行完成
			for i := 0; i < 60; i++ {
				time.Sleep(time.Second)
				if r, ok := h.Check(key); ok && r.Status != TaskStatusRunning {
					return r.Result
				}
			}
		}

		// 返回已执行的结果
		return record.Result
	}

	// 标记为执行中
	h.Set(key, &TaskResult{
		TaskID:      item.TaskID,
		ExecutionID: item.ExecutionID,
		Status:      TaskStatusRunning,
	}, ttl)

	// 执行任务
	result, _ := executor.Execute(item)

	// 更新结果
	h.Set(key, result, ttl)

	return result
}

// Cleanup 清理过期记录
func (h *IdempotentHandler) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for key, record := range h.cache {
		if now.After(record.ExpireAt) {
			delete(h.cache, key)
		}
	}
}

// GlobalExecutor 全局执行器
var GlobalExecutor *TaskExecutor

// InitExecutor 初始化执行器
func InitExecutor() {
	GlobalExecutor = NewTaskExecutor()
}

// GetExecutor 获取执行器
func GetExecutor() *TaskExecutor {
	if GlobalExecutor == nil {
		InitExecutor()
	}
	return GlobalExecutor
}
