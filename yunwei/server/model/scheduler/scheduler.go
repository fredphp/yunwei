package scheduler

import "time"

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusSuccess    TaskStatus = "success"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCanceled   TaskStatus = "canceled"
	TaskStatusCancelled  TaskStatus = "canceled" // alias
	TaskStatusRetrying   TaskStatus = "retrying"
	TaskStatusTimeout    TaskStatus = "timeout"
	TaskStatusRolledback TaskStatus = "rolledback"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeCommand   TaskType = "command"
	TaskTypeScript    TaskType = "script"
	TaskTypeBackup    TaskType = "backup"
	TaskTypeRestore   TaskType = "restore"
	TaskTypeDeploy    TaskType = "deploy"
	TaskTypeCheck     TaskType = "check"
	TaskTypeCleanup   TaskType = "cleanup"
	TaskTypeScheduled TaskType = "scheduled"
	TaskTypeBatch     TaskType = "batch"
)

// Priority 任务优先级
type Priority int

const (
	PriorityLow      Priority = 1
	PriorityNormal   Priority = 5
	PriorityHigh     Priority = 10
	PriorityUrgent   Priority = 20
	PriorityCritical Priority = 50
)

// Task 任务定义
type Task struct {
	ID           uint       `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	Name         string     `json:"name" gorm:"type:varchar(128);not null"`
	Type         TaskType   `json:"type" gorm:"type:varchar(32)"`
	Priority     Priority   `json:"priority"`
	Status       TaskStatus `json:"status" gorm:"type:varchar(16);index"`

	// 执行配置
	Command      string `json:"command" gorm:"type:text"`
	Script       string `json:"script" gorm:"type:text"`
	Params       string `json:"params" gorm:"type:text"`
	Timeout      int    `json:"timeout"` // 秒
	MaxRetry     int    `json:"maxRetry"`
	RetryCount   int    `json:"retryCount"`
	RetryDelay   int    `json:"retryDelay"`     // 重试延迟（秒）
	RetryBackoff string `json:"retryBackoff"`   // linear, exponential

	// 执行器
	Executor     string `json:"executor" gorm:"type:varchar(32)"`
	Action       string `json:"action" gorm:"type:text"`

	// 目标
	TargetType   string `json:"targetType" gorm:"type:varchar(32)"`
	TargetIDs    string `json:"targetIds" gorm:"type:text"`
	ServerID     uint   `json:"serverId"`
	ServerName   string `json:"serverName" gorm:"type:varchar(64)"`
	QueueName    string `json:"queueName" gorm:"type:varchar(64)"`

	// 调度
	ScheduleType string     `json:"scheduleType" gorm:"type:varchar(16)"` // now, delayed, cron
	ScheduledAt  *time.Time `json:"scheduledAt"`
	ScheduleTime *time.Time `json:"scheduleTime"`

	// 执行信息
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 毫秒

	// 结果
	Output       string `json:"output" gorm:"type:text"`
	Result       string `json:"result" gorm:"type:text"`
	Error        string `json:"error" gorm:"type:text"`
	Stdout       string `json:"stdout" gorm:"type:text"`
	Stderr       string `json:"stderr" gorm:"type:text"`
	ExitCode     int    `json:"exitCode"`
	ErrorMessage string `json:"errorMessage" gorm:"type:text"`

	// 回滚
	RollbackEnabled bool   `json:"rollbackEnabled"`
	RollbackCommand string `json:"rollbackCommand" gorm:"type:text"`

	// 依赖
	Dependencies string `json:"dependencies" gorm:"type:text"`
	DependsOn    string `json:"dependsOn" gorm:"type:text"`

	// 批次任务
	BatchID    uint   `json:"batchId"`
	BatchIndex int    `json:"batchIndex"`
	ParentID   uint   `json:"parentId" gorm:"index"`

	// 队列
	QueueAt  *time.Time `json:"queueAt"`
	StartAt  *time.Time `json:"startAt"`
	EndAt    *time.Time `json:"endAt"`
	WorkerID string     `json:"workerId" gorm:"type:varchar(32)"`

	// 回调
	CallbackURL  string `json:"callbackUrl" gorm:"type:varchar(255)"`
	CallbackData string `json:"callbackData" gorm:"type:text"`

	// 幂等性和去重
	IdempotentKey string `json:"idempotentKey" gorm:"type:varchar(64);index"`
	DedupWindow   int    `json:"dedupWindow"` // 去重窗口（秒）

	// 标签和元数据
	Tags     string `json:"tags" gorm:"type:text"`
	Metadata string `json:"metadata" gorm:"type:text"`

	// 创建者
	CreatedBy uint `json:"createdBy"`
}

func (Task) TableName() string {
	return "scheduler_tasks"
}

// CronJob 定时任务
type CronJob struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string `json:"name" gorm:"type:varchar(128);not null"`
	Description string `json:"description" gorm:"type:varchar(255)"`

	// Cron 配置
	CronExpr    string `json:"cronExpr" gorm:"type:varchar(64);not null"`
	Timezone    string `json:"timezone" gorm:"type:varchar(32)"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	// 任务模板
	TaskTemplate string `json:"taskTemplate" gorm:"type:text"`

	// 并发策略
	ConcurrentPolicy string `json:"concurrentPolicy" gorm:"type:varchar(16)"` // allow, forbid, replace

	// 统计
	RunCount     int        `json:"runCount"`
	SuccessCount int        `json:"successCount"`
	FailCount    int        `json:"failCount"`
	LastRunAt    *time.Time `json:"lastRunAt"`
	NextRunAt    *time.Time `json:"nextRunAt"`
	LastError    string     `json:"lastError" gorm:"type:text"`

	// 通知
	NotifyOnSuccess bool `json:"notifyOnSuccess"`
	NotifyOnFail    bool `json:"notifyOnFail"`

	// 创建者
	CreatedBy uint `json:"createdBy"`
}

func (CronJob) TableName() string {
	return "scheduler_cron_jobs"
}

// CronExecution Cron 执行记录
type CronExecution struct {
	ID           uint       `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time  `json:"createdAt"`

	CronJobID    uint       `json:"cronJobId" gorm:"index"`
	TaskID       uint       `json:"taskId"`
	ScheduledAt  time.Time  `json:"scheduledAt"`
	StartedAt    *time.Time `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	Status       TaskStatus `json:"status" gorm:"type:varchar(16)"`
	Error        string     `json:"error" gorm:"type:text"`
}

func (CronExecution) TableName() string {
	return "scheduler_cron_executions"
}

// TaskEvent 任务事件
type TaskEvent struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	TaskID    uint   `json:"taskId" gorm:"index"`
	Type      string `json:"type" gorm:"type:varchar(32)"`
	Source    string `json:"source" gorm:"type:varchar(32)"`
	Message   string `json:"message" gorm:"type:varchar(255)"`
	Data      string `json:"data" gorm:"type:text"`

	EventType string `json:"eventType" gorm:"type:varchar(32)"`
	EventData string `json:"eventData" gorm:"type:text"`
	Operator  string `json:"operator" gorm:"type:varchar(32)"`
	Remark    string `json:"remark" gorm:"type:text"`
}

func (TaskEvent) TableName() string {
	return "scheduler_task_events"
}

// TaskQueue 队列配置
type TaskQueue struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string `json:"name" gorm:"type:varchar(64);uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`
	MaxWorkers  int    `json:"maxWorkers"`
	MaxPending  int    `json:"maxPending"`
	Priority    int    `json:"priority"`
	Timeout     int    `json:"timeout"`
	MaxRetry    int    `json:"maxRetry"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`

	PendingCount int `json:"pendingCount"`
	RunningCount int `json:"runningCount"`
}

func (TaskQueue) TableName() string {
	return "scheduler_queues"
}

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time  `json:"createdAt"`

	TaskID      uint       `json:"taskId" gorm:"index"`
	Attempt     int        `json:"attempt"`
	Status      TaskStatus `json:"status" gorm:"type:varchar(16)"`
	WorkerID    string     `json:"workerId" gorm:"type:varchar(32)"`
	ServerID    uint       `json:"serverId"`

	ExecutionID string     `json:"executionId" gorm:"type:varchar(64);index"`

	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"`

	ExitCode    int    `json:"exitCode"`
	Stdout      string `json:"stdout" gorm:"type:text"`
	Stderr      string `json:"stderr" gorm:"type:text"`
	Output      string `json:"output" gorm:"type:text"`
	Error       string `json:"error" gorm:"type:text"`
	ErrorMessage string `json:"errorMessage" gorm:"type:text"`

	RollbackAt     *time.Time `json:"rollbackAt"`
	RollbackResult string     `json:"rollbackResult" gorm:"type:text"`
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

	TotalTasks   int `json:"totalTasks"`
	PendingTasks int `json:"pendingTasks"`
	RunningTasks int `json:"runningTasks"`
	SuccessTasks int `json:"successTasks"`
	FailedTasks  int `json:"failedTasks"`

	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	Duration    int64      `json:"duration"`

	Parallelism     int  `json:"parallelism"`
	StopOnFail      bool `json:"stopOnFail"`
	NotifyOnComplete bool `json:"notifyOnComplete"`

	CreatedBy uint `json:"createdBy"`
}

func (TaskBatch) TableName() string {
	return "scheduler_task_batches"
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string `json:"name" gorm:"type:varchar(128)"`
	Category    string `json:"category" gorm:"type:varchar(32)"`
	Description string `json:"description" gorm:"type:text"`

	TaskDef string `json:"taskDef" gorm:"type:text"`
	Params  string `json:"params" gorm:"type:text"`

	UseCount int  `json:"useCount"`
	Enabled  bool `json:"enabled"`
}

func (TaskTemplate) TableName() string {
	return "scheduler_task_templates"
}

// TaskQueueItem 队列项（用于执行器）
type TaskQueueItem struct {
	TaskID     uint   `json:"taskId"`
	ExecutionID string `json:"executionId"`
	ServerID   uint   `json:"serverId"`
	QueueName  string `json:"queueName"`
	Priority   int    `json:"priority"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID       uint       `json:"taskId"`
	ExecutionID  string     `json:"executionId"`
	ServerID     uint       `json:"serverId"`
	Status       TaskStatus `json:"status"`
	Output       string     `json:"output"`
	ErrorMessage string     `json:"errorMessage"`
	Duration     int64      `json:"duration"`
	RetryCount   int        `json:"retryCount"`
	ExitCode     int        `json:"exitCode"`
	Stdout       string     `json:"stdout"`
	Stderr       string     `json:"stderr"`
}

// TaskLog 任务日志
type TaskLog struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`

	ExecutionID string `json:"executionId" gorm:"type:varchar(64);index"`
	TaskID      uint   `json:"taskId" gorm:"index"`

	Level   string `json:"level" gorm:"type:varchar(16)"`
	Message string `json:"message" gorm:"type:text"`
	Data    string `json:"data" gorm:"type:text"`
}

func (TaskLog) TableName() string {
	return "scheduler_task_logs"
}
