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
        TaskStatusRetrying   TaskStatus = "retrying"
        TaskStatusTimeout    TaskStatus = "timeout"
        TaskStatusRolledback TaskStatus = "rolledback"
)

// TaskType 任务类型
type TaskType string

const (
        TaskTypeCommand    TaskType = "command"
        TaskTypeScript     TaskType = "script"
        TaskTypeBackup     TaskType = "backup"
        TaskTypeRestore    TaskType = "restore"
        TaskTypeDeploy     TaskType = "deploy"
        TaskTypeCheck      TaskType = "check"
        TaskTypeCleanup    TaskType = "cleanup"
        TaskTypeScheduled  TaskType = "scheduled"
        TaskTypeBatch      TaskType = "batch"
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

        // 目标
        ServerID    uint   `json:"serverId"`
        ServerName  string `json:"serverName" gorm:"type:varchar(64)"`
        QueueName   string `json:"queueName" gorm:"type:varchar(64)"`

        // 调度
        ScheduleType string     `json:"scheduleType" gorm:"type:varchar(16)"` // now, delayed, cron
        ScheduledAt  *time.Time `json:"scheduledAt"`
        ScheduleTime *time.Time `json:"scheduleTime"`

        // 执行信息
        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration    int64      `json:"duration"` // 毫秒

        // 结果
        Output string `json:"output" gorm:"type:text"`
        Error  string `json:"error" gorm:"type:text"`

        // 批次任务
        BatchID    uint  `json:"batchId"`
        BatchIndex int   `json:"batchIndex"`
        DependsOn  []uint `json:"dependsOn" gorm:"-"` // 依赖的任务ID

        // 回调
        CallbackURL  string `json:"callbackUrl" gorm:"type:varchar(255)"`
        CallbackData string `json:"callbackData" gorm:"type:text"`

        // 幂等性和去重
        IdempotentKey string `json:"idempotentKey" gorm:"type:varchar(64);index"`
        DedupWindow   int    `json:"dedupWindow"` // 去重窗口（秒）
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

// BatchTask 批次任务
type BatchTask struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`

        Name        string `json:"name" gorm:"type:varchar(128);not null"`
        Description string `json:"description" gorm:"type:varchar(255)"`

        TotalTasks     int    `json:"totalTasks"`
        CompletedTasks int    `json:"completedTasks"`
        FailedTasks    int    `json:"failedTasks"`
        Status         string `json:"status" gorm:"type:varchar(16)"`

        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
}

func (BatchTask) TableName() string {
        return "scheduler_batch_tasks"
}

// TaskTemplate 任务模板
type TaskTemplate struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`

        Name        string `json:"name" gorm:"type:varchar(128);not null"`
        Description string `json:"description" gorm:"type:varchar(255)"`

        Type     TaskType `json:"type" gorm:"type:varchar(32)"`
        Template string   `json:"template" gorm:"type:text"`

        // 默认值
        DefaultTimeout    int      `json:"defaultTimeout"`
        DefaultQueue      string   `json:"defaultQueue" gorm:"type:varchar(64)"`
        DefaultPriority   Priority `json:"defaultPriority"`
        DefaultMaxRetry   int      `json:"defaultMaxRetry"`
}

func (TaskTemplate) TableName() string {
        return "scheduler_task_templates"
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
}

func (TaskEvent) TableName() string {
        return "scheduler_task_events"
}

// WorkerStatus Worker 状态
type WorkerStatus struct {
        ID           uint       `json:"id" gorm:"primarykey"`
        CreatedAt    time.Time  `json:"createdAt"`
        UpdatedAt    time.Time  `json:"updatedAt"`

        WorkerID     string    `json:"workerId" gorm:"type:varchar(64);uniqueIndex"`
        QueueName    string    `json:"queueName" gorm:"type:varchar(64)"`
        Status       string    `json:"status" gorm:"type:varchar(16)"` // idle, busy, offline
        CurrentTask  uint      `json:"currentTask"`
        CompletedCount int      `json:"completedCount"`
        FailedCount    int      `json:"failedCount"`
        LastHeartbeat  *time.Time `json:"lastHeartbeat"`
}

func (WorkerStatus) TableName() string {
        return "scheduler_worker_status"
}

// TaskQueue 队列配置
type TaskQueue struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`

        Name       string `json:"name" gorm:"type:varchar(64);uniqueIndex"`
        MaxWorkers int    `json:"maxWorkers"`
        MaxPending int    `json:"maxPending"`
        Priority   int    `json:"priority"`
        Timeout    int    `json:"timeout"`
        MaxRetry   int    `json:"maxRetry"`
        Enabled    bool   `json:"enabled" gorm:"default:true"`
}

func (TaskQueue) TableName() string {
        return "scheduler_queues"
}
