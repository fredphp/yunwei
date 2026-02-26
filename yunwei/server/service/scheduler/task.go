package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"yunwei/global"
	schedulerModel "yunwei/model/scheduler"
)

// 类型别名 - 引用 model/scheduler 中的类型
type Task = schedulerModel.Task
type TaskStatus = schedulerModel.TaskStatus
type TaskType = schedulerModel.TaskType
type TaskPriority = schedulerModel.Priority
type TaskExecution = schedulerModel.TaskExecution
type TaskBatch = schedulerModel.TaskBatch
type TaskTemplate = schedulerModel.TaskTemplate
type TaskQueue = schedulerModel.TaskQueue
type CronJob = schedulerModel.CronJob
type CronExecution = schedulerModel.CronExecution
type TaskEvent = schedulerModel.TaskEvent

// 常量别名
const (
	TaskStatusPending    = schedulerModel.TaskStatusPending
	TaskStatusQueued     = schedulerModel.TaskStatusQueued
	TaskStatusRunning    = schedulerModel.TaskStatusRunning
	TaskStatusSuccess    = schedulerModel.TaskStatusSuccess
	TaskStatusFailed     = schedulerModel.TaskStatusFailed
	TaskStatusRetrying   = schedulerModel.TaskStatusRetrying
	TaskStatusCanceled   = schedulerModel.TaskStatusCanceled
	TaskStatusTimeout    = schedulerModel.TaskStatusTimeout
	TaskStatusRolledback = schedulerModel.TaskStatusRolledback
)

const (
	PriorityLow      = schedulerModel.PriorityLow
	PriorityNormal   = schedulerModel.PriorityNormal
	PriorityHigh     = schedulerModel.PriorityHigh
	PriorityCritical = schedulerModel.PriorityCritical
)

const (
	TaskTypeCommand   = schedulerModel.TaskTypeCommand
	TaskTypeScript    = schedulerModel.TaskTypeScript
	TaskTypeDeploy    = schedulerModel.TaskTypeDeploy
	TaskTypeBackup    = schedulerModel.TaskTypeBackup
	TaskTypeCleanup   = schedulerModel.TaskTypeCleanup
	TaskTypeMonitor   = schedulerModel.TaskTypeCheck
	TaskTypeReport    = schedulerModel.TaskTypeScheduled
	TaskTypeSync      = schedulerModel.TaskTypeRestore
	TaskTypeBatch     = schedulerModel.TaskTypeBatch
	TaskTypeWorkflow  = schedulerModel.TaskTypeScheduled
	TaskTypeScheduled = schedulerModel.TaskTypeScheduled
)

// DefaultQueues 默认队列配置
func DefaultQueues() []TaskQueue {
	return []TaskQueue{
		{
			Name:        "default",
			Description: "默认任务队列",
			MaxWorkers:  10,
			MaxPending:  1000,
			Priority:    5,
			Timeout:     300,
			MaxRetry:    3,
			Enabled:     true,
		},
		{
			Name:        "critical",
			Description: "关键任务队列",
			MaxWorkers:  5,
			MaxPending:  100,
			Priority:    20,
			Timeout:     600,
			MaxRetry:    5,
			Enabled:     true,
		},
		{
			Name:        "background",
			Description: "后台任务队列",
			MaxWorkers:  20,
			MaxPending:  5000,
			Priority:    1,
			Timeout:     3600,
			MaxRetry:    2,
			Enabled:     true,
		},
		{
			Name:        "deploy",
			Description: "部署任务队列",
			MaxWorkers:  3,
			MaxPending:  50,
			Priority:    10,
			Timeout:     1800,
			MaxRetry:    1,
			Enabled:     true,
		},
		{
			Name:        "batch",
			Description: "批量任务队列",
			MaxWorkers:  15,
			MaxPending:  2000,
			Priority:    3,
			Timeout:     7200,
			MaxRetry:    3,
			Enabled:     true,
		},
	}
}

// InitQueues 初始化队列
func InitQueues() error {
	for _, queue := range DefaultQueues() {
		var existing TaskQueue
		result := global.DB.Where("name = ?", queue.Name).First(&existing)
		if result.Error != nil {
			global.DB.Create(&queue)
		}
	}
	return nil
}

// CreateTask 创建任务
func CreateTask(task *Task) error {
	// 生成幂等键
	if task.IdempotentKey == "" {
		task.IdempotentKey = generateIdempotentKey(task)
	}

	// 设置默认值
	if task.Status == "" {
		task.Status = TaskStatusPending
	}
	if task.Priority == 0 {
		task.Priority = PriorityNormal
	}
	if task.QueueName == "" {
		task.QueueName = "default"
	}
	if task.Timeout == 0 {
		task.Timeout = 300
	}

	return global.DB.Create(task).Error
}

// generateIdempotentKey 生成幂等键
func generateIdempotentKey(task *Task) string {
	data := fmt.Sprintf("%s-%s-%s-%d", task.Type, task.Name, task.Command, time.Now().UnixNano())
	return fmt.Sprintf("%x", data)[:32]
}

// GetTask 获取任务
func GetTask(id uint) (*Task, error) {
	var task Task
	err := global.DB.First(&task, id).Error
	return &task, err
}

// GetTaskByIdempotentKey 通过幂等键获取任务
func GetTaskByIdempotentKey(key string) (*Task, error) {
	var task Task
	err := global.DB.Where("idempotent_key = ?", key).First(&task).Error
	return &task, err
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(id uint, status TaskStatus, result string) error {
	return global.DB.Model(&Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status,
		"result": result,
	}).Error
}

// GetPendingTasks 获取等待执行的任务
func GetPendingTasks(queueName string, limit int) ([]Task, error) {
	var tasks []Task
	query := global.DB.Where("status = ?", TaskStatusQueued)
	if queueName != "" {
		query = query.Where("queue_name = ?", queueName)
	}
	err := query.Order("priority DESC, created_at ASC").Limit(limit).Find(&tasks).Error
	return tasks, err
}

// GetRunningTasks 获取执行中的任务
func GetRunningTasks() ([]Task, error) {
	var tasks []Task
	err := global.DB.Where("status = ?", TaskStatusRunning).Find(&tasks).Error
	return tasks, err
}

// GetTasksByBatch 获取批次任务
func GetTasksByBatch(batchID uint) ([]Task, error) {
	var tasks []Task
	err := global.DB.Where("batch_id = ?", batchID).Find(&tasks).Error
	return tasks, err
}

// CreateTaskExecution 创建执行记录
func CreateTaskExecution(execution *TaskExecution) error {
	return global.DB.Create(execution).Error
}

// GetTaskExecutions 获取执行历史
func GetTaskExecutions(taskID uint, limit int) ([]TaskExecution, error) {
	var executions []TaskExecution
	query := global.DB.Where("task_id = ?", taskID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&executions).Error
	return executions, err
}

// RecordTaskEvent 记录任务事件
func RecordTaskEvent(taskID uint, eventType string, eventData interface{}, operator, remark string) error {
	dataJSON, _ := json.Marshal(eventData)
	event := &TaskEvent{
		TaskID:    taskID,
		Type:      eventType,
		EventData: string(dataJSON),
		Operator:  operator,
		Remark:    remark,
	}
	return global.DB.Create(event).Error
}

// CreateCronJob 创建定时任务
func CreateCronJob(job *CronJob) error {
	return global.DB.Create(job).Error
}

// UpdateCronJob 更新定时任务
func UpdateCronJob(job *CronJob) error {
	return global.DB.Save(job).Error
}

// DeleteCronJob 删除定时任务
func DeleteCronJob(id uint) error {
	return global.DB.Delete(&CronJob{}, id).Error
}

// GetCronJob 获取定时任务
func GetCronJob(id uint) (*CronJob, error) {
	var job CronJob
	err := global.DB.First(&job, id).Error
	return &job, err
}

// ListCronJobs 列出定时任务
func ListCronJobs() ([]CronJob, error) {
	var jobs []CronJob
	err := global.DB.Find(&jobs).Error
	return jobs, err
}

// GetTaskExecutionByExecutionID 通过 ExecutionID 获取执行记录
func GetTaskExecutionByExecutionID(executionID string) (*TaskExecution, error) {
	var execution TaskExecution
	err := global.DB.Where("execution_id = ?", executionID).First(&execution).Error
	return &execution, err
}

// UpdateTaskExecution 更新执行记录
func UpdateTaskExecution(execution *TaskExecution) error {
	return global.DB.Save(execution).Error
}

// AddTaskLog 添加任务日志
func AddTaskLog(executionID string, taskID uint, level, message string, data map[string]interface{}) {
	dataJSON, _ := json.Marshal(data)
	log := &schedulerModel.TaskLog{
		ExecutionID: executionID,
		TaskID:      taskID,
		Level:       level,
		Message:     message,
		Data:        string(dataJSON),
	}
	global.DB.Create(log)
}
