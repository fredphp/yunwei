package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/service/notify"
	"yunwei/service/scheduler/queue"
	"yunwei/service/scheduler/worker"
	"yunwei/service/scheduler/cron"
)

// JobCenter 任务中心
type JobCenter struct {
	queue      *queue.TaskQueue
	workerPool *worker.WorkerPool
	cron       *cron.CronScheduler
	notifier   notify.Notifier
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewJobCenter 创建任务中心
func NewJobCenter() *JobCenter {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建内存队列后端
	backend := queue.NewMemoryQueueBackend()
	taskQueue := queue.NewTaskQueue(backend)
	
	// 创建 Worker 池
	wp := worker.NewWorkerPool(taskQueue, nil)
	
	// 创建 Cron 调度器
	cronScheduler := cron.NewCronScheduler(taskQueue)
	
	return &JobCenter{
		queue:      taskQueue,
		workerPool: wp,
		cron:       cronScheduler,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetNotifier 设置通知器
func (jc *JobCenter) SetNotifier(notifier notify.Notifier) {
	jc.notifier = notifier
}

// Start 启动任务中心
func (jc *JobCenter) Start() error {
	// 初始化队列
	InitQueues()
	
	// 注册默认队列
	jc.queue.RegisterQueue(&queue.QueueConfig{
		Name:       "default",
		MaxWorkers: 10,
		MaxPending: 1000,
		Priority:   5,
		Timeout:    300,
		MaxRetry:   3,
	})
	jc.queue.RegisterQueue(&queue.QueueConfig{
		Name:       "critical",
		MaxWorkers: 5,
		MaxPending: 100,
		Priority:   20,
		Timeout:    600,
		MaxRetry:   5,
	})
	jc.queue.RegisterQueue(&queue.QueueConfig{
		Name:       "background",
		MaxWorkers: 20,
		MaxPending: 5000,
		Priority:   1,
		Timeout:    3600,
		MaxRetry:   2,
	})
	jc.queue.RegisterQueue(&queue.QueueConfig{
		Name:       "deploy",
		MaxWorkers: 3,
		MaxPending: 50,
		Priority:   10,
		Timeout:    1800,
		MaxRetry:   1,
	})
	jc.queue.RegisterQueue(&queue.QueueConfig{
		Name:       "batch",
		MaxWorkers: 15,
		MaxPending: 2000,
		Priority:   3,
		Timeout:    7200,
		MaxRetry:   3,
	})
	
	// 启动 Worker 池
	jc.workerPool.Start("default", 10)
	jc.workerPool.Start("critical", 5)
	jc.workerPool.Start("background", 20)
	jc.workerPool.Start("deploy", 3)
	jc.workerPool.Start("batch", 15)
	
	// 启动 Cron 调度器
	jc.cron.Start()
	
	return nil
}

// Stop 停止任务中心
func (jc *JobCenter) Stop() {
	jc.cancel()
	jc.workerPool.Stop()
	jc.cron.Stop()
}

// ==================== 任务提交 ====================

// SubmitTask 提交任务
func (jc *JobCenter) SubmitTask(task *Task) (uint, error) {
	if err := jc.queue.EnqueueTask(task); err != nil {
		return 0, err
	}
	return task.ID, nil
}

// SubmitTaskWithOptions 带选项提交任务
func (jc *JobCenter) SubmitTaskWithOptions(name string, taskType TaskType, command string, opts ...TaskOption) (uint, error) {
	task := &Task{
		Name:    name,
		Type:    taskType,
		Command: command,
		Status:  TaskStatusPending,
	}
	
	for _, opt := range opts {
		opt(task)
	}
	
	return jc.SubmitTask(task)
}

// TaskOption 任务选项
type TaskOption func(*Task)

// WithQueue 设置队列
func WithQueue(queueName string) TaskOption {
	return func(t *Task) {
		t.QueueName = queueName
	}
}

// WithPriority 设置优先级
func WithPriority(priority TaskPriority) TaskOption {
	return func(t *Task) {
		t.Priority = priority
	}
}

// WithTimeout 设置超时
func WithTimeout(timeout int) TaskOption {
	return func(t *Task) {
		t.Timeout = timeout
	}
}

// WithRetry 设置重试
func WithRetry(maxRetry, delay int, backoff string) TaskOption {
	return func(t *Task) {
		t.MaxRetry = maxRetry
		t.RetryDelay = delay
		t.RetryBackoff = backoff
	}
}

// WithTarget 设置目标服务器
func WithTarget(targetType string, targetIDs []uint) TaskOption {
	return func(t *Task) {
		t.TargetType = targetType
		idsJSON, _ := json.Marshal(targetIDs)
		t.TargetIDs = string(idsJSON)
	}
}

// WithRollback 设置回滚
func WithRollback(rollbackCommand string) TaskOption {
	return func(t *Task) {
		t.RollbackEnabled = true
		t.RollbackCommand = rollbackCommand
	}
}

// WithDependencies 设置依赖
func WithDependencies(depIDs []uint, dependsOn string) TaskOption {
	return func(t *Task) {
		depsJSON, _ := json.Marshal(depIDs)
		t.Dependencies = string(depsJSON)
		t.DependsOn = dependsOn
	}
}

// WithIdempotent 设置幂等
func WithIdempotent(key string, window int) TaskOption {
	return func(t *Task) {
		t.IdempotentKey = key
		t.DedupWindow = window
	}
}

// ==================== 批量任务 ====================

// SubmitBatch 提交批量任务
func (jc *JobCenter) SubmitBatch(batchName string, tasks []*Task, parallelism int, stopOnFail bool) (*TaskBatch, error) {
	// 创建批次
	batch := &TaskBatch{
		Name:        batchName,
		Status:      TaskStatusPending,
		TotalTasks:  len(tasks),
		Parallelism: parallelism,
		StopOnFail:  stopOnFail,
	}
	global.DB.Create(batch)
	
	// 设置任务批次ID
	for _, task := range tasks {
		task.BatchID = batch.ID
	}
	
	// 批量入队
	taskIDs, errors := jc.queue.BatchEnqueue(tasks)
	if len(errors) > 0 {
		batch.FailedTasks = len(errors)
	}
	batch.PendingTasks = len(taskIDs)
	batch.RunningTasks = 0
	batch.SuccessTasks = 0
	global.DB.Save(batch)
	
	// 更新批次状态
	now := time.Now()
	batch.Status = TaskStatusRunning
	batch.StartAt = &now
	global.DB.Save(batch)
	
	// 启动批次监控
	go jc.monitorBatch(batch.ID)
	
	return batch, nil
}

// monitorBatch 监控批次执行
func (jc *JobCenter) monitorBatch(batchID uint) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-jc.ctx.Done():
			return
		case <-ticker.C:
			// 获取批次任务统计
			var pending, running, success, failed int64
			global.DB.Model(&Task{}).Where("batch_id = ? AND status = ?", batchID, TaskStatusPending).Count(&pending)
			global.DB.Model(&Task{}).Where("batch_id = ? AND status = ?", batchID, TaskStatusRunning).Count(&running)
			global.DB.Model(&Task{}).Where("batch_id = ? AND status = ?", batchID, TaskStatusSuccess).Count(&success)
			global.DB.Model(&Task{}).Where("batch_id = ? AND status IN ?", batchID, []TaskStatus{TaskStatusFailed, TaskStatusCanceled}).Count(&failed)
			
			// 更新批次
			var batch TaskBatch
			global.DB.First(&batch, batchID)
			batch.PendingTasks = int(pending)
			batch.RunningTasks = int(running)
			batch.SuccessTasks = int(success)
			batch.FailedTasks = int(failed)
			
			// 检查是否完成
			if pending == 0 && running == 0 {
				now := time.Now()
				batch.EndAt = &now
				if batch.StartAt != nil {
					batch.Duration = now.Sub(*batch.StartAt).Milliseconds()
				}
				
				if failed > 0 {
					batch.Status = TaskStatusFailed
				} else {
					batch.Status = TaskStatusSuccess
				}
				
				// 发送通知
				if jc.notifier != nil && batch.NotifyOnComplete {
					jc.notifier.SendMessage("批量任务完成",
						fmt.Sprintf("批次 %s 执行完成: 成功 %d, 失败 %d",
							batch.Name, success, failed))
				}
			}
			
			global.DB.Save(&batch)
			
			// 如果已完成，停止监控
			if batch.Status != TaskStatusRunning {
				return
			}
		}
	}
}

// ==================== 定时任务 ====================

// CreateScheduledTask 创建定时任务
func (jc *JobCenter) CreateScheduledTask(name, cronExpr string, taskTemplate string) (*CronJob, error) {
	job := &CronJob{
		Name:         name,
		CronExpr:     cronExpr,
		TaskTemplate: taskTemplate,
		Enabled:      true,
	}
	
	if err := CreateCronJob(job); err != nil {
		return nil, err
	}
	
	// 添加到调度器
	if err := jc.cron.AddJob(job); err != nil {
		return nil, err
	}
	
	return job, nil
}

// UpdateScheduledTask 更新定时任务
func (jc *JobCenter) UpdateScheduledTask(job *CronJob) error {
	if err := UpdateCronJob(job); err != nil {
		return err
	}
	return jc.cron.UpdateJob(job)
}

// DeleteScheduledTask 删除定时任务
func (jc *JobCenter) DeleteScheduledTask(id uint) error {
	jc.cron.RemoveJob(id)
	return DeleteCronJob(id)
}

// TriggerScheduledTask 手动触发定时任务
func (jc *JobCenter) TriggerScheduledTask(id uint) error {
	return jc.cron.TriggerCronJob(id)
}

// ==================== 任务控制 ====================

// CancelTask 取消任务
func (jc *JobCenter) CancelTask(taskID uint) error {
	task, err := GetTask(taskID)
	if err != nil {
		return err
	}
	
	if task.Status == TaskStatusRunning {
		return fmt.Errorf("cannot cancel running task")
	}
	
	task.Status = TaskStatusCanceled
	global.DB.Save(task)
	
	RecordTaskEvent(taskID, "canceled", nil, "user", "任务已取消")
	
	return nil
}

// RetryTask 重试任务
func (jc *JobCenter) RetryTask(taskID uint) error {
	task, err := GetTask(taskID)
	if err != nil {
		return err
	}
	
	if task.Status != TaskStatusFailed {
		return fmt.Errorf("only failed tasks can be retried")
	}
	
	task.Status = TaskStatusPending
	task.RetryCount = 0
	global.DB.Save(task)
	
	return jc.queue.EnqueueTask(task)
}

// RollbackTask 回滚任务
func (jc *JobCenter) RollbackTask(taskID uint) error {
	task, err := GetTask(taskID)
	if err != nil {
		return err
	}
	
	if !task.RollbackEnabled || task.RollbackCommand == "" {
		return fmt.Errorf("rollback not enabled for this task")
	}
	
	// 创建回滚任务
	rollbackTask := &Task{
		Name:        fmt.Sprintf("rollback-%d", taskID),
		Type:        task.Type,
		Command:     task.RollbackCommand,
		TargetType:  task.TargetType,
		TargetIDs:   task.TargetIDs,
		Priority:    PriorityHigh,
		QueueName:   task.QueueName,
		ParentID:    taskID,
	}
	
	if err := jc.SubmitTask(rollbackTask); err != nil {
		return err
	}
	
	// 更新原任务状态
	task.Status = TaskStatusRolledback
	global.DB.Save(task)
	
	RecordTaskEvent(taskID, "rolled_back", map[string]interface{}{
		"rollback_task_id": rollbackTask.ID,
	}, "user", "任务已回滚")
	
	return nil
}

// ==================== 任务查询 ====================

// GetTaskStatus 获取任务状态
func (jc *JobCenter) GetTaskStatus(taskID uint) (*Task, error) {
	return GetTask(taskID)
}

// GetTaskExecutions 获取执行历史
func (jc *JobCenter) GetTaskExecutions(taskID uint, limit int) ([]TaskExecution, error) {
	return GetTaskExecutions(taskID, limit)
}

// ListTasks 列出任务
func (jc *JobCenter) ListTasks(filter *TaskFilter) ([]Task, int64, error) {
	query := global.DB.Model(&Task{})
	
	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Type != "" {
			query = query.Where("type = ?", filter.Type)
		}
		if filter.QueueName != "" {
			query = query.Where("queue_name = ?", filter.QueueName)
		}
		if filter.BatchID > 0 {
			query = query.Where("batch_id = ?", filter.BatchID)
		}
	}
	
	var total int64
	query.Count(&total)
	
	var tasks []Task
	err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&tasks).Error
	
	return tasks, total, err
}

// TaskFilter 任务过滤器
type TaskFilter struct {
	Status    string `json:"status"`
	Type      string `json:"type"`
	QueueName string `json:"queueName"`
	BatchID   uint   `json:"batchId"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// ==================== 统计 ====================

// GetQueueStats 获取队列统计
func (jc *JobCenter) GetQueueStats(queueName string) (*queue.QueueStats, error) {
	return jc.queue.GetQueueStats(queueName)
}

// GetAllQueueStats 获取所有队列统计
func (jc *JobCenter) GetAllQueueStats() (map[string]*queue.QueueStats, error) {
	return jc.queue.GetAllQueueStats()
}

// GetWorkerStats 获取 Worker 统计
func (jc *JobCenter) GetWorkerStats(queueName string) *worker.WorkerPoolStats {
	return jc.workerPool.GetWorkerStats(queueName)
}

// GetAllWorkerStats 获取所有 Worker 统计
func (jc *JobCenter) GetAllWorkerStats() map[string]*worker.WorkerPoolStats {
	return jc.workerPool.GetAllWorkerStats()
}

// ScaleWorkers 调整 Worker 数量
func (jc *JobCenter) ScaleWorkers(queueName string, targetWorkers int) error {
	return jc.workerPool.Scale(queueName, targetWorkers)
}

// ==================== 模板 ====================

// CreateTemplate 创建任务模板
func (jc *JobCenter) CreateTemplate(template *TaskTemplate) error {
	return global.DB.Create(template).Error
}

// GetTemplate 获取任务模板
func (jc *JobCenter) GetTemplate(id uint) (*TaskTemplate, error) {
	var template TaskTemplate
	err := global.DB.First(&template, id).Error
	return &template, err
}

// ListTemplates 列出任务模板
func (jc *JobCenter) ListTemplates(category string) ([]TaskTemplate, error) {
	var templates []TaskTemplate
	query := global.DB.Model(&TaskTemplate{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&templates).Error
	return templates, err
}

// SubmitFromTemplate 从模板创建任务
func (jc *JobCenter) SubmitFromTemplate(templateID uint, params map[string]interface{}) (uint, error) {
	template, err := jc.GetTemplate(templateID)
	if err != nil {
		return 0, err
	}
	
	// 解析任务定义
	var taskDef map[string]interface{}
	if err := json.Unmarshal([]byte(template.TaskDef), &taskDef); err != nil {
		return 0, err
	}
	
	// 合并参数
	for k, v := range params {
		taskDef[k] = v
	}
	
	// 创建任务
	taskDefJSON, _ := json.Marshal(taskDef)
	var task Task
	if err := json.Unmarshal(taskDefJSON, &task); err != nil {
		return 0, err
	}
	
	// 更新使用次数
	global.DB.Model(&TaskTemplate{}).Where("id = ?", templateID).UpdateColumn("use_count", template.UseCount+1)
	
	return jc.SubmitTask(&task)
}
