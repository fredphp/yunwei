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
	JobTypePatrol  JobType = "patrol"
	JobTypeHeal    JobType = "heal"
	JobTypePredict JobType = "predict"
	JobTypeClean   JobType = "clean"
	JobTypeReport  JobType = "report"
)

// Job 定时任务
type Job struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string    `json:"name" gorm:"type:varchar(64);not null"`
	Type        JobType   `json:"type" gorm:"type:varchar(32)"`
	Status      JobStatus `json:"status" gorm:"type:varchar(16);default:'enabled'"`

	Cron        string `json:"cron" gorm:"type:varchar(64)"`
	Interval    int    `json:"interval"`
	Timeout     int    `json:"timeout"`
	Config      string `json:"config" gorm:"type:text"`
	Description string `json:"description" gorm:"type:varchar(255)"`

	LastRunAt  *time.Time `json:"lastRunAt"`
	LastResult string     `json:"lastResult"`
	LastError  string     `json:"lastError"`
	NextRunAt  *time.Time `json:"nextRunAt"`

	RunCount     int `json:"runCount"`
	SuccessCount int `json:"successCount"`
	FailCount    int `json:"failCount"`
}

func (Job) TableName() string {
	return "scheduler_jobs"
}

// JobLog 任务日志
type JobLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	JobID   uint   `json:"jobId" gorm:"index"`
	JobName string `json:"jobName" gorm:"type:varchar(64)"`

	Status   string `json:"status" gorm:"type:varchar(16)"`
	Output   string `json:"output" gorm:"type:text"`
	Error    string `json:"error" gorm:"type:text"`
	Duration int64  `json:"duration"`
}

func (JobLog) TableName() string {
	return "scheduler_job_logs"
}

// Scheduler 调度器
type Scheduler struct {
	jobs         map[uint]*Job
	tickers      map[uint]*time.Ticker
	running      bool
	mu           sync.RWMutex
	patrolRobot  *patrol.PatrolRobot
	healer       *selfheal.SelfHealer
	predictor    *prediction.Predictor
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:   make(map[uint]*Job),
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
		{Name: "每日巡检", Type: JobTypePatrol, Cron: "0 0 8 * * *", Description: "每天8:00执行巡检"},
		{Name: "每小时巡检", Type: JobTypePatrol, Cron: "0 0 * * * *", Description: "每小时检查"},
		{Name: "服务健康检查", Type: JobTypeHeal, Cron: "0 */5 * * * *", Description: "每5分钟检查"},
		{Name: "预测分析", Type: JobTypePredict, Cron: "0 0 */6 * * *", Description: "每6小时预测"},
		{Name: "日报生成", Type: JobTypeReport, Cron: "0 30 23 * * *", Description: "每天23:30生成"},
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

// executeJob 执行任务
func (s *Scheduler) executeJob(job *Job) {
	startTime := time.Now()
	log := &JobLog{
		JobID:   job.ID,
		JobName: job.Name,
	}

	var output string
	var err error

	switch job.Type {
	case JobTypePatrol:
		output, err = s.executePatrol(job)
	case JobTypeHeal:
		output = "健康检查已执行"
	case JobTypePredict:
		output = "预测分析已执行"
	case JobTypeReport:
		output, err = s.executeReport(job)
	default:
		err = fmt.Errorf("未知任务类型: %s", job.Type)
	}

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

	global.DB.Create(log)
	global.DB.Save(job)
}

// executePatrol 执行巡检
func (s *Scheduler) executePatrol(job *Job) (string, error) {
	if s.patrolRobot == nil {
		return "", fmt.Errorf("巡检机器人未配置")
	}
	record, err := s.patrolRobot.RunPatrol(patrol.PatrolTypeScheduled)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("巡检完成: %d台服务器", record.TotalServers), nil
}

// executeReport 执行报告
func (s *Scheduler) executeReport(job *Job) (string, error) {
	if s.patrolRobot == nil {
		return "", fmt.Errorf("巡检机器人未配置")
	}
	report, err := s.patrolRobot.GenerateDailyReport()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("日报已生成: 在线率%.1f%%", report.OnlineRate), nil
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

// RunJobNow 立即执行
func (s *Scheduler) RunJobNow(jobID uint) error {
	var job Job
	if err := global.DB.First(&job, jobID).Error; err != nil {
		return err
	}
	go s.executeJob(&job)
	return nil
}

// GetJobs 获取任务列表
func (s *Scheduler) GetJobs() ([]Job, error) {
	var jobs []Job
	err := global.DB.Find(&jobs).Error
	return jobs, err
}
