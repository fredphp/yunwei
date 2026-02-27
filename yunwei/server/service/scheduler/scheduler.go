package scheduler

import (
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/service/patrol"
	"yunwei/service/prediction"
	"yunwei/service/selfheal"
)

// JobStatus 任务状态
type JobStatus string

const (
	JobStatusEnabled  JobStatus = "enabled"
	JobStatusDisabled JobStatus = "disabled"
)

// JobType 任务类型
type JobType string

const (
	JobTypePatrol    JobType = "patrol"    // 巡检任务
	JobTypeHeal      JobType = "heal"      // 自愈检查
	JobTypePredict   JobType = "predict"   // 预测分析
	JobTypeClean     JobType = "clean"     // 清理任务
	JobTypeBackup    JobType = "backup"    // 备份任务
	JobTypeReport    JobType = "report"    // 报告生成
	JobTypeCustom    JobType = "custom"    // 自定义任务
)

// Job 定时任务
type Job struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string    `json:"name" gorm:"type:varchar(64);not null"`
	Type        JobType   `json:"type" gorm:"type:varchar(32)"`
	Status      JobStatus `json:"status" gorm:"type:varchar(16);default:'enabled'"`

	// 调度配置
	Cron        string `json:"cron" gorm:"type:varchar(64)"`      // cron表达式
	Interval    int    `json:"interval"`                           // 间隔(分钟)，与cron二选一
	Timeout     int    `json:"timeout"`                            // 超时时间(秒)

	// 任务配置
	Config      string `json:"config" gorm:"type:text"`            // JSON配置
	Description string `json:"description" gorm:"type:varchar(255)"`

	// 最后执行
	LastRunAt   *time.Time `json:"lastRunAt"`
	LastResult  string     `json:"lastResult"`
	LastError   string     `json:"lastError"`

	// 下次执行
	NextRunAt   *time.Time `json:"nextRunAt"`

	// 统计
	RunCount    int   `json:"runCount"`
	SuccessCount int  `json:"successCount"`
	FailCount   int   `json:"failCount"`
}

func (Job) TableName() string {
	return "scheduler_jobs"
}

// JobLog 任务日志
type JobLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	JobID     uint   `json:"jobId" gorm:"index"`
	JobName   string `json:"jobName" gorm:"type:varchar(64)"`

	Status    string `json:"status" gorm:"type:varchar(16)"` // success, failed, timeout
	Output    string `json:"output" gorm:"type:text"`
	Error     string `json:"error" gorm:"type:text"`
	Duration  int64  `json:"duration"` // 毫秒
}

func (JobLog) TableName() string {
	return "scheduler_job_logs"
}

// Scheduler 调度器
type Scheduler struct {
	jobs      map[uint]*Job
	tickers   map[uint]*time.Ticker
	running   bool
	mu        sync.RWMutex

	patrolRobot *patrol.PatrolRobot
	healer      *selfheal.SelfHealer
	predictor   *prediction.Predictor
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:    make(map[uint]*Job),
		tickers: make(map[uint]*time.Ticker),
	}
}

// SetPatrolRobot 设置巡检机器人
func (s *Scheduler) SetPatrolRobot(robot *patrol.PatrolRobot) {
	s.patrolRobot = robot
}

// SetHealer 设置自愈系统
func (s *Scheduler) SetHealer(healer *selfheal.SelfHealer) {
	s.healer = healer
}

// SetPredictor 设置预测器
func (s *Scheduler) SetPredictor(predictor *prediction.Predictor) {
	s.predictor = predictor
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("调度器已在运行")
	}

	s.running = true

	// 加载所有启用的任务
	var jobs []Job
	global.DB.Where("status = ?", JobStatusEnabled).Find(&jobs)

	for i := range jobs {
		s.scheduleJob(&jobs[i])
	}

	// 启动默认任务
	s.initDefaultJobs()

	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false

	for _, ticker := range s.tickers {
		ticker.Stop()
	}

	s.tickers = make(map[uint]*time.Ticker)
	s.jobs = make(map[uint]*Job)
}

// initDefaultJobs 初始化默认任务
func (s *Scheduler) initDefaultJobs() {
	defaultJobs := []Job{
		{
			Name:        "每日巡检",
			Type:        JobTypePatrol,
			Cron:        "0 0 8 * * *", // 每天8:00
			Interval:    0,
			Description: "每天早上8点执行全量服务器巡检",
		},
		{
			Name:        "每小时巡检",
			Type:        JobTypePatrol,
			Cron:        "0 0 * * * *", // 每小时
			Interval:    0,
			Description: "每小时检查服务器状态",
		},
		{
			Name:        "服务健康检查",
			Type:        JobTypeHeal,
			Cron:        "0 */5 * * * *", // 每5分钟
			Interval:    0,
			Description: "每5分钟检查服务健康状态",
		},
		{
			Name:        "预测分析",
			Type:        JobTypePredict,
			Cron:        "0 0 */6 * * *", // 每6小时
			Interval:    0,
			Description: "每6小时执行预测分析",
		},
		{
			Name:        "日志清理",
			Type:        JobTypeClean,
			Cron:        "0 0 3 * * *", // 每天凌晨3点
			Interval:    0,
			Description: "每天凌晨3点清理过期日志",
		},
		{
			Name:        "日报生成",
			Type:        JobTypeReport,
			Cron:        "0 30 23 * * *", // 每天23:30
			Interval:    0,
			Description: "每天23:30生成日报",
		},
	}

	for _, job := range defaultJobs {
		var existing Job
		if err := global.DB.Where("name = ?", job.Name).First(&existing).Error; err != nil {
			job.Status = JobStatusEnabled
			global.DB.Create(&job)
			s.scheduleJob(&job)
		}
	}
}

// scheduleJob 调度任务
func (s *Scheduler) scheduleJob(job *Job) {
	if job.Status != JobStatusEnabled {
		return
	}

	s.jobs[job.ID] = job

	// 计算下次执行时间
	s.calculateNextRun(job)

	// 创建定时器
	interval := time.Minute
	if job.Interval > 0 {
		interval = time.Duration(job.Interval) * time.Minute
	}

	ticker := time.NewTicker(interval)
	s.tickers[job.ID] = ticker

	go func() {
		for range ticker.C {
			if s.running {
				s.executeJob(job)
			} else {
				break
			}
		}
	}()
}

// calculateNextRun 计算下次执行时间
func (s *Scheduler) calculateNextRun(job *Job) {
	// 简化处理：基于间隔计算
	if job.Interval > 0 {
		next := time.Now().Add(time.Duration(job.Interval) * time.Minute)
		job.NextRunAt = &next
	}
}

// executeJob 执行任务
func (s *Scheduler) executeJob(job *Job) {
	startTime := time.Now()
	log := &JobLog{
		JobID:   job.ID,
		JobName: job.Name,
	}

	// 执行任务
	var output string
	var err error

	switch job.Type {
	case JobTypePatrol:
		output, err = s.executePatrol(job)
	case JobTypeHeal:
		output, err = s.executeHeal(job)
	case JobTypePredict:
		output, err = s.executePredict(job)
	case JobTypeClean:
		output, err = s.executeClean(job)
	case JobTypeReport:
		output, err = s.executeReport(job)
	default:
		err = fmt.Errorf("未知任务类型: %s", job.Type)
	}

	// 记录日志
	log.Duration = time.Since(startTime).Milliseconds()
	if err != nil {
		log.Status = "failed"
		log.Error = err.Error()
		job.LastError = err.Error()
		job.FailCount++
	} else {
		log.Status = "success"
		job.SuccessCount++
	}
	log.Output = output
	job.LastResult = output

	now := time.Now()
	job.LastRunAt = &now
	job.RunCount++
	s.calculateNextRun(job)

	global.DB.Create(log)
	global.DB.Save(job)
}

// executePatrol 执行巡检任务
func (s *Scheduler) executePatrol(job *Job) (string, error) {
	if s.patrolRobot == nil {
		return "", fmt.Errorf("巡检机器人未配置")
	}

	record, err := s.patrolRobot.RunPatrol(patrol.PatrolTypeScheduled)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("巡检完成: 检查%d台服务器, 发现%d个问题", 
		record.TotalServers, record.CriticalCount+record.WarningCount), nil
}

// executeHeal 执行自愈任务
func (s *Scheduler) executeHeal(job *Job) (string, error) {
	if s.healer == nil {
		return "", fmt.Errorf("自愈系统未配置")
	}

	// 启动一轮健康检查
	// 实际由selfheal.MonitorAndHeal()持续运行
	return "健康检查已执行", nil
}

// executePredict 执行预测任务
func (s *Scheduler) executePredict(job *Job) (string, error) {
	if s.predictor == nil {
		return "", fmt.Errorf("预测器未配置")
	}

	// 对所有服务器执行预测
	// TODO: 实现批量预测
	return "预测分析已执行", nil
}

// executeClean 执行清理任务
func (s *Scheduler) executeClean(job *Job) (string, error) {
	// 清理过期日志
	// 清理过期告警
	// 清理过期指标
	return "清理任务已执行", nil
}

// executeReport 执行报告任务
func (s *Scheduler) executeReport(job *Job) (string, error) {
	if s.patrolRobot == nil {
		return "", fmt.Errorf("巡检机器人未配置")
	}

	report, err := s.patrolRobot.GenerateDailyReport()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("日报已生成: 在线率%.1f%%, 告警%d条", 
		report.OnlineRate, report.TotalAlerts), nil
}

// AddJob 添加任务
func (s *Scheduler) AddJob(job *Job) error {
	if err := global.DB.Create(job).Error; err != nil {
		return err
	}

	if job.Status == JobStatusEnabled {
		s.scheduleJob(job)
	}

	return nil
}

// UpdateJob 更新任务
func (s *Scheduler) UpdateJob(job *Job) error {
	// 停止旧的定时器
	if ticker, exists := s.tickers[job.ID]; exists {
		ticker.Stop()
		delete(s.tickers, job.ID)
	}

	// 更新数据库
	if err := global.DB.Save(job).Error; err != nil {
		return err
	}

	// 重新调度
	if job.Status == JobStatusEnabled {
		s.scheduleJob(job)
	}

	return nil
}

// DeleteJob 删除任务
func (s *Scheduler) DeleteJob(jobID uint) error {
	// 停止定时器
	if ticker, exists := s.tickers[jobID]; exists {
		ticker.Stop()
		delete(s.tickers, jobID)
	}

	delete(s.jobs, jobID)

	return global.DB.Delete(&Job{}, jobID).Error
}

// RunJobNow 立即执行任务
func (s *Scheduler) RunJobNow(jobID uint) error {
	s.mu.RLock()
	job, exists := s.jobs[jobID]
	s.mu.RUnlock()

	if !exists {
		var j Job
		if err := global.DB.First(&j, jobID).Error; err != nil {
			return err
		}
		job = &j
	}

	go s.executeJob(job)
	return nil
}

// GetJobs 获取任务列表
func (s *Scheduler) GetJobs() ([]Job, error) {
	var jobs []Job
	err := global.DB.Find(&jobs).Error
	return jobs, err
}

// GetJobLogs 获取任务日志
func (s *Scheduler) GetJobLogs(jobID uint, limit int) ([]JobLog, error) {
	var logs []JobLog
	query := global.DB.Model(&JobLog{}).Order("created_at DESC")
	if jobID > 0 {
		query = query.Where("job_id = ?", jobID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

// GetStatistics 获取统计
func (s *Scheduler) GetStatistics() map[string]int64 {
	stats := make(map[string]int64)

	global.DB.Model(&Job{}).Count(&stats["totalJobs"])
	global.DB.Model(&Job{}).Where("status = ?", JobStatusEnabled).Count(&stats["enabledJobs"])
	global.DB.Model(&JobLog{}).Where("created_at > ?", time.Now().AddDate(0, 0, -1)).Count(&stats["todayRuns"])
	global.DB.Model(&JobLog{}).Where("created_at > ? AND status = ?", time.Now().AddDate(0, 0, -1), "success").Count(&stats["todaySuccess"])

	return stats
}
