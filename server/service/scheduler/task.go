package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"yunwei/global"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"    // 等待执行
	TaskStatusQueued    TaskStatus = "queued"     // 已入队
	TaskStatusRunning   TaskStatus = "running"    // 执行中
	TaskStatusSuccess   TaskStatus = "success"    // 执行成功
	TaskStatusFailed    TaskStatus = "failed"     // 执行失败
	TaskStatusRetrying  TaskStatus = "retrying"   // 重试中
	TaskStatusCanceled  TaskStatus = "canceled"   // 已取消
	TaskStatusTimeout   TaskStatus = "timeout"    // 执行超时
	TaskStatusRolledback TaskStatus = "rolledback" // 已回滚
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow      TaskPriority = 1
	PriorityNormal   TaskPriority = 5
	PriorityHigh     TaskPriority = 10
	PriorityCritical TaskPriority = 20
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeCommand    TaskType = "command"    // 命令执行
	TaskTypeScript     TaskType = "script"     // 脚本执行
	TaskTypeDeploy     TaskType = "deploy"     // 部署任务
	TaskTypeBackup     TaskType = "backup"     // 备份任务
	TaskTypeCleanup    TaskType = "cleanup"    // 清理任务
	TaskTypeMonitor    TaskType = "monitor"    // 监控任务
	TaskTypeReport     TaskType = "report"     // 报告生成
	TaskTypeSync       TaskType = "sync"       // 同步任务
	TaskTypeBatch      TaskType = "batch"      // 批量任务
	TaskTypeWorkflow   TaskType = "workflow"   // 工作流任务
	TaskTypeScheduled  TaskType = "scheduled"  // 定时任务
)

// Task 任务定义
type Task struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// 基本信息
	Name        string       `json:"name" gorm:"type:varchar(128);index"`
	Type        TaskType     `json:"type" gorm:"type:varchar(32);index"`
	Priority    TaskPriority `json:"priority"`
	Status      TaskStatus   `json:"status" gorm:"type:varchar(16);index"`
	
	// 执行目标
	TargetType  string `json:"targetType" gorm:"type:varchar(32)"`  // server, group, all
	TargetIDs   string `json:"targetIds" gorm:"type:text"`           // JSON 数组
	
	// 执行内容
	Executor    string `json:"executor" gorm:"type:varchar(32)"`     // 执行器类型
	Command     string `json:"command" gorm:"type:text"`             // 命令/脚本
	Params      string `json:"params" gorm:"type:text"`              // JSON 参数
	
	// 调度配置
	ScheduleType  string `json:"scheduleType" gorm:"type:varchar(16)"` // immediate, delayed, cron
	ScheduleTime  *time.Time `json:"scheduleTime"`                    // 延迟执行时间
	CronExpr      string `json:"cronExpr" gorm:"type:varchar(64)"`    // Cron 表达式
	CronTimezone  string `json:"cronTimezone" gorm:"type:varchar(32)"` // 时区
	
	// 超时与重试
	Timeout      int `json:"timeout"`         // 超时时间(秒)
	MaxRetry     int `json:"maxRetry"`        // 最大重试次数
	RetryCount   int `json:"retryCount"`      // 当前重试次数
	RetryDelay   int `json:"retryDelay"`      // 重试延迟(秒)
	RetryBackoff string `json:"retryBackoff" gorm:"type:varchar(16)"` // linear, exponential
	
	// 幂等控制
	IdempotentKey string `json:"idempotentKey" gorm:"type:varchar(64);uniqueIndex"` // 幂等键
	DedupWindow   int    `json:"dedupWindow"` // 去重窗口(秒)
	
	// 回滚配置
	RollbackEnabled bool   `json:"rollbackEnabled"`
	RollbackCommand string `json:"rollbackCommand" gorm:"type:text"`
	
	// 依赖控制
	Dependencies string `json:"dependencies" gorm:"type:text"` // JSON 依赖任务ID列表
	DependsOn    string `json:"dependsOn" gorm:"type:text"`    // 依赖条件: all_success, any_success
	
	// 队列信息
	QueueName   string `json:"queueName" gorm:"type:varchar(32)"`
	WorkerID    string `json:"workerId" gorm:"type:varchar(32)"`
	QueueAt     *time.Time `json:"queueAt"`
	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	
	// 执行结果
	Result      string `json:"result" gorm:"type:text"`
	Error       string `json:"error" gorm:"type:text"`
	ExitCode    int    `json:"exitCode"`
	Duration    int64  `json:"duration"` // 毫秒
	
	// 输出
	Stdout      string `json:"stdout" gorm:"type:text"`
	Stderr      string `json:"stderr" gorm:"type:text"`
	
	// 标签和元数据
	Tags        string `json:"tags" gorm:"type:text"` // JSON 数组
	Metadata    string `json:"metadata" gorm:"type:text"` // JSON 对象
	
	// 创建者
	CreatedBy   uint `json:"createdBy"`
	
	// 关联任务
	ParentID    uint `json:"parentId" gorm:"index"`
	BatchID     uint `json:"batchId" gorm:"index"` // 批次ID
}

func (Task) TableName() string {
	return "scheduler_tasks"
}

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	
	TaskID      uint      `json:"taskId" gorm:"index"`
	Task        *Task     `json:"task" gorm:"foreignKey:TaskID"`
	
	// 执行信息
	Attempt     int       `json:"attempt"`      // 第几次尝试
	Status      TaskStatus `json:"status" gorm:"type:varchar(16)"`
	WorkerID    string    `json:"workerId" gorm:"type:varchar(32)"`
	ServerID    uint      `json:"serverId"`     // 执行服务器
	
	// 时间
	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	Duration    int64      `json:"duration"` // 毫秒
	
	// 结果
	ExitCode    int    `json:"exitCode"`
	Stdout      string `json:"stdout" gorm:"type:text"`
	Stderr      string `json:"stderr" gorm:"type:text"`
	Error       string `json:"error" gorm:"type:text"`
	
	// 回滚
	RollbackAt  *time.Time `json:"rollbackAt"`
	RollbackResult string `json:"rollbackResult" gorm:"type:text"`
}

func (TaskExecution) TableName() string {
	return "scheduler_task_executions"
}

// TaskBatch 任务批次
type TaskBatch struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	Name        string     `json:"name" gorm:"type:varchar(128)"`
	Description string     `json:"description" gorm:"type:text"`
	Status      TaskStatus `json:"status" gorm:"type:varchar(16)"`
	
	// 统计
	TotalTasks   int `json:"totalTasks"`
	PendingTasks int `json:"pendingTasks"`
	RunningTasks int `json:"runningTasks"`
	SuccessTasks int `json:"successTasks"`
	FailedTasks  int `json:"failedTasks"`
	
	// 时间
	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	Duration    int64      `json:"duration"`
	
	// 配置
	Parallelism int  `json:"parallelism"` // 并行数
	StopOnFail  bool `json:"stopOnFail"`  // 失败时停止
	NotifyOnComplete bool `json:"notifyOnComplete"`
	
	// 创建者
	CreatedBy   uint `json:"createdBy"`
}

func (TaskBatch) TableName() string {
	return "scheduler_task_batches"
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	Name        string   `json:"name" gorm:"type:varchar(128)"`
	Category    string   `json:"category" gorm:"type:varchar(32)"`
	Description string   `json:"description" gorm:"type:text"`
	
	// 模板内容
	TaskDef     string `json:"taskDef" gorm:"type:text"` // JSON 任务定义
	
	// 参数模板
	Params      string `json:"params" gorm:"type:text"` // JSON 参数定义
	
	// 统计
	UseCount    int `json:"useCount"`
	
	Enabled     bool `json:"enabled"`
}

func (TaskTemplate) TableName() string {
	return "scheduler_task_templates"
}

// TaskQueue 任务队列定义
type TaskQueue struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	Name        string `json:"name" gorm:"type:varchar(32);uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`
	
	// 队列配置
	MaxWorkers   int `json:"maxWorkers"`   // 最大 Worker 数
	MaxPending   int `json:"maxPending"`   // 最大等待任务数
	Priority     int `json:"priority"`     // 队列优先级
	
	// 处理配置
	Timeout      int  `json:"timeout"`      // 默认超时
	MaxRetry     int  `json:"maxRetry"`     // 默认最大重试
	
	// 统计
	PendingCount int `json:"pendingCount"`
	RunningCount int `json:"runningCount"`
	
	Enabled      bool `json:"enabled"`
}

func (TaskQueue) TableName() string {
	return "scheduler_queues"
}

// CronJob 定时任务
type CronJob struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	Name        string `json:"name" gorm:"type:varchar(128)"`
	Description string `json:"description" gorm:"type:text"`
	
	// Cron 表达式
	CronExpr     string `json:"cronExpr" gorm:"type:varchar(64)"`
	Timezone     string `json:"timezone" gorm:"type:varchar(32)"`
	
	// 任务定义
	TaskTemplate string `json:"taskTemplate" gorm:"type:text"` // JSON 任务模板
	
	// 执行配置
	MisfirePolicy string `json:"misfirePolicy" gorm:"type:varchar(16)"` // fire_now, ignore, fire_once
	ConcurrentPolicy string `json:"concurrentPolicy" gorm:"type:varchar(16)"` // allow, forbid, replace
	
	// 状态
	Enabled      bool `json:"enabled"`
	LastRunAt    *time.Time `json:"lastRunAt"`
	NextRunAt    *time.Time `json:"nextRunAt"`
	LastStatus   TaskStatus `json:"lastStatus" gorm:"type:varchar(16)"`
	
	// 统计
	RunCount    int `json:"runCount"`
	SuccessCount int `json:"successCount"`
	FailCount   int `json:"failCount"`
	
	// 创建者
	CreatedBy   uint `json:"createdBy"`
}

func (CronJob) TableName() string {
	return "scheduler_cron_jobs"
}

// CronExecution Cron 执行记录
type CronExecution struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	
	CronJobID   uint      `json:"cronJobId" gorm:"index"`
	TaskID      uint      `json:"taskId" gorm:"index"` // 生成的任务ID
	
	ScheduledAt time.Time `json:"scheduledAt"` // 计划执行时间
	StartedAt   *time.Time `json:"startedAt"`
	EndedAt     *time.Time `json:"endedAt"`
	
	Status      TaskStatus `json:"status" gorm:"type:varchar(16)"`
	Error       string     `json:"error" gorm:"type:text"`
}

func (CronExecution) TableName() string {
	return "scheduler_cron_executions"
}

// TaskEvent 任务事件（用于事件溯源）
type TaskEvent struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	
	TaskID      uint      `json:"taskId" gorm:"index"`
	EventType   string    `json:"eventType" gorm:"type:varchar(32)"`
	EventData   string    `json:"eventData" gorm:"type:text"` // JSON
	
	Operator    string    `json:"operator" gorm:"type:varchar(32)"`
	Remark      string    `json:"remark" gorm:"type:text"`
}

func (TaskEvent) TableName() string {
	return "scheduler_task_events"
}

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
		EventType: eventType,
		EventData: string(dataJSON),
		Operator:  operator,
		Remark:    remark,
	}
	return global.DB.Create(event).Error
}
