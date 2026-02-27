package scheduler

import (
	"net/http"

	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/service/scheduler"

	"github.com/gin-gonic/gin"
)

// 全局任务中心实例
var jobCenter *scheduler.JobCenter

// InitJobCenter 初始化任务中心
func InitJobCenter() {
	jobCenter = scheduler.NewJobCenter()
	jobCenter.Start()
}

// GetJobCenter 获取任务中心实例
func GetJobCenter() *scheduler.JobCenter {
	return jobCenter
}

// ==================== 任务管理 ====================

// SubmitTask 提交任务
func SubmitTask(c *gin.Context) {
	var task scheduler.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	taskID, err := jobCenter.SubmitTask(&task)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(gin.H{
		"taskId": taskID,
		"message": "任务已提交",
	}, c)
}

// SubmitTaskWithOptions 带选项提交任务
func SubmitTaskWithOptions(c *gin.Context) {
	var req struct {
		Name       string                 `json:"name" binding:"required"`
		Type       scheduler.TaskType     `json:"type" binding:"required"`
		Command    string                 `json:"command" binding:"required"`
		QueueName  string                 `json:"queueName"`
		Priority   scheduler.TaskPriority `json:"priority"`
		Timeout    int                    `json:"timeout"`
		MaxRetry   int                    `json:"maxRetry"`
		RetryDelay int                    `json:"retryDelay"`
		TargetType string                 `json:"targetType"`
		TargetIDs  []uint                 `json:"targetIds"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	var opts []scheduler.TaskOption
	if req.QueueName != "" {
		opts = append(opts, scheduler.WithQueue(req.QueueName))
	}
	if req.Priority != 0 {
		opts = append(opts, scheduler.WithPriority(req.Priority))
	}
	if req.Timeout > 0 {
		opts = append(opts, scheduler.WithTimeout(req.Timeout))
	}
	if req.MaxRetry > 0 {
		opts = append(opts, scheduler.WithRetry(req.MaxRetry, req.RetryDelay, "exponential"))
	}
	if req.TargetType != "" && len(req.TargetIDs) > 0 {
		opts = append(opts, scheduler.WithTarget(req.TargetType, req.TargetIDs))
	}
	
	taskID, err := jobCenter.SubmitTaskWithOptions(req.Name, req.Type, req.Command, opts...)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(gin.H{
		"taskId": taskID,
	}, c)
}

// GetTask 获取任务详情
func GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := jobCenter.GetTaskStatus(parseInt(id))
	if err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	response.OkWithData(task, c)
}

// ListTasks 列出任务
func ListTasks(c *gin.Context) {
	filter := &scheduler.TaskFilter{
		Status:    c.Query("status"),
		Type:      c.Query("type"),
		QueueName: c.Query("queueName"),
		Limit:     50,
		Offset:    0,
	}
	
	if c.Query("batchId") != "" {
		filter.BatchID = parseInt(c.Query("batchId"))
	}
	
	tasks, total, err := jobCenter.ListTasks(filter)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(gin.H{
		"list":  tasks,
		"total": total,
	}, c)
}

// CancelTask 取消任务
func CancelTask(c *gin.Context) {
	id := c.Param("id")
	if err := jobCenter.CancelTask(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// RetryTask 重试任务
func RetryTask(c *gin.Context) {
	id := c.Param("id")
	if err := jobCenter.RetryTask(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// RollbackTask 回滚任务
func RollbackTask(c *gin.Context) {
	id := c.Param("id")
	if err := jobCenter.RollbackTask(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetTaskExecutions 获取执行历史
func GetTaskExecutions(c *gin.Context) {
	id := c.Param("id")
	executions, err := jobCenter.GetTaskExecutions(parseInt(id), 20)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(executions, c)
}

// ==================== 批量任务 ====================

// SubmitBatch 提交批量任务
func SubmitBatch(c *gin.Context) {
	var req struct {
		Name        string          `json:"name" binding:"required"`
		Tasks       []scheduler.Task `json:"tasks" binding:"required"`
		Parallelism int             `json:"parallelism"`
		StopOnFail  bool            `json:"stopOnFail"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	batch, err := jobCenter.SubmitBatch(req.Name, req.Tasks, req.Parallelism, req.StopOnFail)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(batch, c)
}

// GetBatch 获取批次详情
func GetBatch(c *gin.Context) {
	id := c.Param("id")
	var batch scheduler.TaskBatch
	if err := global.DB.First(&batch, id).Error; err != nil {
		response.FailWithMessage("批次不存在", c)
		return
	}
	response.OkWithData(batch, c)
}

// ListBatches 列出批次
func ListBatches(c *gin.Context) {
	var batches []scheduler.TaskBatch
	global.DB.Order("created_at DESC").Limit(50).Find(&batches)
	response.OkWithData(batches, c)
}

// GetBatchTasks 获取批次任务
func GetBatchTasks(c *gin.Context) {
	id := c.Param("id")
	tasks, err := scheduler.GetTasksByBatch(parseInt(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tasks, c)
}

// ==================== 定时任务 ====================

// CreateCronJob 创建定时任务
func CreateCronJob(c *gin.Context) {
	var job scheduler.CronJob
	if err := c.ShouldBindJSON(&job); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	created, err := jobCenter.CreateScheduledTask(job.Name, job.CronExpr, job.TaskTemplate)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(created, c)
}

// GetCronJob 获取定时任务
func GetCronJob(c *gin.Context) {
	id := c.Param("id")
	job, err := scheduler.GetCronJob(parseInt(id))
	if err != nil {
		response.FailWithMessage("定时任务不存在", c)
		return
	}
	response.OkWithData(job, c)
}

// ListCronJobs 列出定时任务
func ListCronJobs(c *gin.Context) {
	jobs, err := scheduler.GetCronJobs()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(jobs, c)
}

// UpdateCronJob 更新定时任务
func UpdateCronJob(c *gin.Context) {
	var job scheduler.CronJob
	if err := c.ShouldBindJSON(&job); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	if err := jobCenter.UpdateScheduledTask(&job); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(job, c)
}

// DeleteCronJob 删除定时任务
func DeleteCronJob(c *gin.Context) {
	id := c.Param("id")
	if err := jobCenter.DeleteScheduledTask(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// TriggerCronJob 手动触发定时任务
func TriggerCronJob(c *gin.Context) {
	id := c.Param("id")
	if err := jobCenter.TriggerScheduledTask(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetCronExecutions 获取执行历史
func GetCronExecutions(c *gin.Context) {
	id := c.Param("id")
	executions, err := scheduler.GetCronExecutions(parseInt(id), 20)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(executions, c)
}

// ==================== 队列管理 ====================

// GetQueues 获取队列列表
func GetQueues(c *gin.Context) {
	var queues []scheduler.TaskQueue
	global.DB.Find(&queues)
	response.OkWithData(queues, c)
}

// GetQueueStats 获取队列统计
func GetQueueStats(c *gin.Context) {
	queueName := c.Query("queue")
	
	if queueName != "" {
		stats, err := jobCenter.GetQueueStats(queueName)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.OkWithData(stats, c)
	} else {
		stats, err := jobCenter.GetAllQueueStats()
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.OkWithData(stats, c)
	}
}

// ==================== Worker 管理 ====================

// GetWorkers 获取 Worker 列表
func GetWorkers(c *gin.Context) {
	queueName := c.Query("queue")
	workers := jobCenter.GetAllWorkerStats()
	
	if queueName != "" {
		if stats, ok := workers[queueName]; ok {
			response.OkWithData(stats, c)
			return
		}
	}
	
	response.OkWithData(workers, c)
}

// ScaleWorkers 调整 Worker 数量
func ScaleWorkers(c *gin.Context) {
	var req struct {
		QueueName string `json:"queueName" binding:"required"`
		Workers   int    `json:"workers" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	if err := jobCenter.ScaleWorkers(req.QueueName, req.Workers); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.Ok(c)
}

// ==================== 模板管理 ====================

// CreateTemplate 创建模板
func CreateTemplate(c *gin.Context) {
	var template scheduler.TaskTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	if err := jobCenter.CreateTemplate(&template); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(template, c)
}

// ListTemplates 列出模板
func ListTemplates(c *gin.Context) {
	category := c.Query("category")
	templates, err := jobCenter.ListTemplates(category)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(templates, c)
}

// SubmitFromTemplate 从模板创建任务
func SubmitFromTemplate(c *gin.Context) {
	var req struct {
		TemplateID uint                   `json:"templateId" binding:"required"`
		Params     map[string]interface{} `json:"params"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	taskID, err := jobCenter.SubmitFromTemplate(req.TemplateID, req.Params)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	
	response.OkWithData(gin.H{
		"taskId": taskID,
	}, c)
}

// ==================== 仪表盘统计 ====================

// GetDashboard 获取仪表盘统计
func GetDashboard(c *gin.Context) {
	// 任务统计
	var pending, running, success, failed int64
	global.DB.Model(&scheduler.Task{}).Where("status = ?", scheduler.TaskStatusPending).Count(&pending)
	global.DB.Model(&scheduler.Task{}).Where("status = ?", scheduler.TaskStatusRunning).Count(&running)
	global.DB.Model(&scheduler.Task{}).Where("status = ?", scheduler.TaskStatusSuccess).Count(&success)
	global.DB.Model(&scheduler.Task{}).Where("status = ?", scheduler.TaskStatusFailed).Count(&failed)
	
	// 今日统计
	today := "2006-01-02"
	var todayTasks, todaySuccess, todayFailed int64
	global.DB.Model(&scheduler.Task{}).Where("DATE(created_at) = ?", today).Count(&todayTasks)
	global.DB.Model(&scheduler.Task{}).Where("status = ? AND DATE(created_at) = ?", scheduler.TaskStatusSuccess, today).Count(&todaySuccess)
	global.DB.Model(&scheduler.Task{}).Where("status = ? AND DATE(created_at) = ?", scheduler.TaskStatusFailed, today).Count(&todayFailed)
	
	// 队列统计
	queueStats, _ := jobCenter.GetAllQueueStats()
	
	// Worker 统计
	workerStats := jobCenter.GetAllWorkerStats()
	
	response.OkWithData(gin.H{
		"taskStats": gin.H{
			"pending": pending,
			"running": running,
			"success": success,
			"failed":  failed,
		},
		"todayStats": gin.H{
			"total":   todayTasks,
			"success": todaySuccess,
			"failed":  todayFailed,
		},
		"queueStats":  queueStats,
		"workerStats": workerStats,
	}, c)
}

func parseInt(s string) uint {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		}
	}
	return result
}
