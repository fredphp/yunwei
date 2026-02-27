package cron

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"yunwei/global"
	schedulerModel "yunwei/model/scheduler"
	"yunwei/service/scheduler/queue"
)

// CronScheduler Cron 调度器
type CronScheduler struct {
	queue    *queue.TaskQueue
	jobs     map[uint]*CronJobEntry
	mu       sync.RWMutex
	stopChan chan struct{}
	running  bool
}

// CronJobEntry Cron 任务条目
type CronJobEntry struct {
	Job      *schedulerModel.CronJob
	NextRun  time.Time
	PrevRun  time.Time
	Schedule Schedule
}

// Schedule 调度接口
type Schedule interface {
	Next(time.Time) time.Time
}

// NewCronScheduler 创建 Cron 调度器
func NewCronScheduler(taskQueue *queue.TaskQueue) *CronScheduler {
	return &CronScheduler{
		queue:    taskQueue,
		jobs:     make(map[uint]*CronJobEntry),
		stopChan: make(chan struct{}),
	}
}

// Start 启动调度器
func (cs *CronScheduler) Start() {
	cs.mu.Lock()
	if cs.running {
		cs.mu.Unlock()
		return
	}
	cs.running = true
	cs.mu.Unlock()

	// 加载所有启用的 Cron 任务
	cs.loadJobs()

	// 启动调度循环
	go cs.run()
}

// Stop 停止调度器
func (cs *CronScheduler) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.running {
		return
	}

	cs.running = false
	close(cs.stopChan)
}

// loadJobs 加载任务
func (cs *CronScheduler) loadJobs() {
	var jobs []schedulerModel.CronJob
	global.DB.Where("enabled = ?", true).Find(&jobs)

	for i := range jobs {
		cs.AddJob(&jobs[i])
	}
}

// run 调度循环
func (cs *CronScheduler) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopChan:
			return
		case <-ticker.C:
			cs.checkAndRun()
		}
	}
}

// checkAndRun 检查并执行任务
func (cs *CronScheduler) checkAndRun() {
	now := time.Now()

	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for jobID, entry := range cs.jobs {
		if entry.NextRun.Before(now) || entry.NextRun.Equal(now) {
			// 执行任务
			go cs.executeJob(jobID, entry)
		}
	}
}

// executeJob 执行任务
func (cs *CronScheduler) executeJob(jobID uint, entry *CronJobEntry) {
	job := entry.Job

	// 检查并发策略
	switch job.ConcurrentPolicy {
	case "forbid":
		// 禁止并发，检查是否有运行中的任务
		var runningCount int64
		global.DB.Model(&schedulerModel.Task{}).Where(
			"status = ? AND json_extract(params, '$.cron_job_id') = ?",
			schedulerModel.TaskStatusRunning, jobID,
		).Count(&runningCount)
		if runningCount > 0 {
			return
		}

	case "replace":
		// 取消之前的任务
		var runningTasks []schedulerModel.Task
		global.DB.Where(
			"status = ? AND json_extract(params, '$.cron_job_id') = ?",
			schedulerModel.TaskStatusRunning, jobID,
		).Find(&runningTasks)
		for _, t := range runningTasks {
			t.Status = schedulerModel.TaskStatusCanceled
			global.DB.Save(&t)
		}
	}

	// 创建执行记录
	execution := &schedulerModel.CronExecution{
		CronJobID:   jobID,
		ScheduledAt: entry.NextRun,
		Status:      schedulerModel.TaskStatusPending,
	}
	global.DB.Create(execution)

	// 解析任务模板
	var taskDef map[string]interface{}
	if err := json.Unmarshal([]byte(job.TaskTemplate), &taskDef); err != nil {
		execution.Status = schedulerModel.TaskStatusFailed
		execution.Error = "Invalid task template"
		global.DB.Save(execution)
		return
	}

	// 创建任务
	task := &schedulerModel.Task{
		Name:         fmt.Sprintf("%s-%d", job.Name, time.Now().Unix()),
		Type:         schedulerModel.TaskTypeScheduled,
		Priority:     schedulerModel.PriorityNormal,
		Status:       schedulerModel.TaskStatusPending,
		ScheduleType: "cron",
		Params:       fmt.Sprintf(`{"cron_job_id": %d, "execution_id": %d}`, jobID, execution.ID),
	}

	// 从模板填充任务属性
	if name, ok := taskDef["name"].(string); ok {
		task.Name = name
	}
	if taskType, ok := taskDef["type"].(string); ok {
		task.Type = schedulerModel.TaskType(taskType)
	}
	if command, ok := taskDef["command"].(string); ok {
		task.Command = command
	}
	if queueName, ok := taskDef["queueName"].(string); ok {
		task.QueueName = queueName
	}
	if timeout, ok := taskDef["timeout"].(float64); ok {
		task.Timeout = int(timeout)
	}

	// 入队执行
	if err := cs.queue.EnqueueTask(task); err != nil {
		execution.Status = schedulerModel.TaskStatusFailed
		execution.Error = err.Error()
		global.DB.Save(execution)
		return
	}

	execution.TaskID = task.ID
	execution.Status = schedulerModel.TaskStatusRunning
	now := time.Now()
	execution.StartedAt = &now
	global.DB.Save(execution)

	// 更新 Cron 任务状态
	entry.PrevRun = entry.NextRun
	entry.NextRun = entry.Schedule.Next(now)
	job.LastRunAt = &now
	job.NextRunAt = &entry.NextRun
	job.RunCount++
	global.DB.Save(job)

	// 记录事件
	event := &schedulerModel.TaskEvent{
		TaskID:  task.ID,
		Type:    "cron_triggered",
		Source:  "cron",
		Message: "Cron 任务触发",
		Data:    fmt.Sprintf(`{"cron_job_id": %d, "scheduled": "%s"}`, jobID, entry.PrevRun.Format(time.RFC3339)),
	}
	global.DB.Create(event)
}

// AddJob 添加任务
func (cs *CronScheduler) AddJob(job *schedulerModel.CronJob) error {
	// 解析 Cron 表达式
	schedule, err := ParseCron(job.CronExpr, job.Timezone)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	entry := &CronJobEntry{
		Job:      job,
		Schedule: schedule,
		NextRun:  schedule.Next(time.Now()),
	}

	cs.mu.Lock()
	cs.jobs[job.ID] = entry
	cs.mu.Unlock()

	// 更新下次运行时间
	job.NextRunAt = &entry.NextRun
	global.DB.Save(job)

	return nil
}

// RemoveJob 移除任务
func (cs *CronScheduler) RemoveJob(jobID uint) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.jobs, jobID)
}

// UpdateJob 更新任务
func (cs *CronScheduler) UpdateJob(job *schedulerModel.CronJob) error {
	cs.RemoveJob(job.ID)
	if job.Enabled {
		return cs.AddJob(job)
	}
	return nil
}

// GetNextRuns 获取下次执行时间列表
func (cs *CronScheduler) GetNextRuns(jobID uint, count int) []time.Time {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, exists := cs.jobs[jobID]
	if !exists {
		return nil
	}

	var runs []time.Time
	next := entry.NextRun
	for i := 0; i < count; i++ {
		next = entry.Schedule.Next(next)
		runs = append(runs, next)
	}
	return runs
}

// ParseCron 解析 Cron 表达式
func ParseCron(expr, timezone string) (Schedule, error) {
	// 支持标准 5 字段和扩展 6 字段表达式
	fields := strings.Fields(expr)

	var loc *time.Location
	if timezone != "" {
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			loc = time.Local
		}
	} else {
		loc = time.Local
	}

	if len(fields) == 5 {
		// 标准格式: 分 时 日 月 周
		return &cronSchedule{
			minute:  parseField(fields[0], 0, 59),
			hour:    parseField(fields[1], 0, 23),
			day:     parseField(fields[2], 1, 31),
			month:   parseField(fields[3], 1, 12),
			weekday: parseField(fields[4], 0, 6),
			loc:     loc,
		}, nil
	}

	if len(fields) == 6 {
		// 扩展格式: 秒 分 时 日 月 周
		return &cronScheduleWithSeconds{
			second:  parseField(fields[0], 0, 59),
			minute:  parseField(fields[1], 0, 59),
			hour:    parseField(fields[2], 0, 23),
			day:     parseField(fields[3], 1, 31),
			month:   parseField(fields[4], 1, 12),
			weekday: parseField(fields[5], 0, 6),
			loc:     loc,
		}, nil
	}

	return nil, fmt.Errorf("invalid cron expression: %s", expr)
}

// cronSchedule 标准 Cron 调度
type cronSchedule struct {
	minute  map[int]bool
	hour    map[int]bool
	day     map[int]bool
	month   map[int]bool
	weekday map[int]bool
	loc     *time.Location
}

// Next 计算下次执行时间
func (s *cronSchedule) Next(t time.Time) time.Time {
	t = t.In(s.loc)

	// 从下一分钟开始
	t = t.Add(time.Minute - time.Duration(t.Second())*time.Second)

	for {
		// 检查月份
		if !s.month[int(t.Month())] {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, s.loc)
			continue
		}

		// 检查日期
		if !s.day[t.Day()] && !s.weekday[int(t.Weekday())] {
			t = t.AddDate(0, 0, 1)
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, s.loc)
			continue
		}

		// 检查小时
		if !s.hour[t.Hour()] {
			t = t.Add(time.Hour - time.Duration(t.Minute())*time.Minute)
			continue
		}

		// 检查分钟
		if !s.minute[t.Minute()] {
			t = t.Add(time.Minute)
			continue
		}

		return t
	}
}

// cronScheduleWithSeconds 带秒的 Cron 调度
type cronScheduleWithSeconds struct {
	second  map[int]bool
	minute  map[int]bool
	hour    map[int]bool
	day     map[int]bool
	month   map[int]bool
	weekday map[int]bool
	loc     *time.Location
}

// Next 计算下次执行时间
func (s *cronScheduleWithSeconds) Next(t time.Time) time.Time {
	t = t.In(s.loc)

	// 从下一秒开始
	t = t.Add(time.Second)

	for {
		// 检查月份
		if !s.month[int(t.Month())] {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, s.loc)
			continue
		}

		// 检查日期
		if !s.day[t.Day()] && !s.weekday[int(t.Weekday())] {
			t = t.AddDate(0, 0, 1)
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, s.loc)
			continue
		}

		// 检查小时
		if !s.hour[t.Hour()] {
			t = t.Add(time.Hour - time.Duration(t.Minute())*time.Minute - time.Duration(t.Second())*time.Second)
			continue
		}

		// 检查分钟
		if !s.minute[t.Minute()] {
			t = t.Add(time.Minute - time.Duration(t.Second())*time.Second)
			continue
		}

		// 检查秒
		if !s.second[t.Second()] {
			t = t.Add(time.Second)
			continue
		}

		return t
	}
}

// parseField 解析字段
func parseField(field string, min, max int) map[int]bool {
	result := make(map[int]bool)

	// 处理 *
	if field == "*" {
		for i := min; i <= max; i++ {
			result[i] = true
		}
		return result
	}

	// 处理逗号分隔
	for _, part := range strings.Split(field, ",") {
		// 处理步长
		if strings.Contains(part, "/") {
			parts := strings.Split(part, "/")
			rangePart := parts[0]
			step := parseInt(parts[1])

			var start, end int
			if rangePart == "*" {
				start, end = min, max
			} else if strings.Contains(rangePart, "-") {
				rangeParts := strings.Split(rangePart, "-")
				start, end = parseInt(rangeParts[0]), parseInt(rangeParts[1])
			} else {
				start, end = parseInt(rangePart), max
			}

			for i := start; i <= end; i += step {
				result[i] = true
			}
			continue
		}

		// 处理范围
		if strings.Contains(part, "-") {
			parts := strings.Split(part, "-")
			start, end := parseInt(parts[0]), parseInt(parts[1])
			for i := start; i <= end; i++ {
				result[i] = true
			}
			continue
		}

		// 单个值
		result[parseInt(part)] = true
	}

	return result
}

// parseInt 解析整数
func parseInt(s string) int {
	var result int
	for _, c := range strings.TrimSpace(s) {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// CreateCronJob 创建 Cron 任务
func CreateCronJob(job *schedulerModel.CronJob) error {
	return global.DB.Create(job).Error
}

// GetCronJob 获取 Cron 任务
func GetCronJob(id uint) (*schedulerModel.CronJob, error) {
	var job schedulerModel.CronJob
	err := global.DB.First(&job, id).Error
	return &job, err
}

// GetCronJobs 获取 Cron 任务列表
func GetCronJobs() ([]schedulerModel.CronJob, error) {
	var jobs []schedulerModel.CronJob
	err := global.DB.Order("created_at DESC").Find(&jobs).Error
	return jobs, err
}

// UpdateCronJob 更新 Cron 任务
func UpdateCronJob(job *schedulerModel.CronJob) error {
	return global.DB.Save(job).Error
}

// DeleteCronJob 删除 Cron 任务
func DeleteCronJob(id uint) error {
	return global.DB.Delete(&schedulerModel.CronJob{}, id).Error
}

// EnableCronJob 启用 Cron 任务
func EnableCronJob(id uint) error {
	return global.DB.Model(&schedulerModel.CronJob{}).Where("id = ?", id).Update("enabled", true).Error
}

// DisableCronJob 禁用 Cron 任务
func DisableCronJob(id uint) error {
	return global.DB.Model(&schedulerModel.CronJob{}).Where("id = ?", id).Update("enabled", false).Error
}

// TriggerCronJob 手动触发 Cron 任务
func (cs *CronScheduler) TriggerCronJob(id uint) error {
	cs.mu.RLock()
	entry, exists := cs.jobs[id]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cron job not found: %d", id)
	}

	go cs.executeJob(id, entry)
	return nil
}

// GetCronExecutions 获取执行历史
func GetCronExecutions(cronJobID uint, limit int) ([]schedulerModel.CronExecution, error) {
	var executions []schedulerModel.CronExecution
	query := global.DB.Where("cron_job_id = ?", cronJobID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&executions).Error
	return executions, err
}
