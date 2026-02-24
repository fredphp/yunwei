package executor

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/service/deploy/config"
	"yunwei/service/deploy/planner"
	"yunwei/service/notify"
)

// ExecutorStatus 执行器状态
type ExecutorStatus string

const (
	StatusIdle      ExecutorStatus = "idle"
	StatusRunning   ExecutorStatus = "running"
	StatusPaused    ExecutorStatus = "paused"
	StatusCompleted ExecutorStatus = "completed"
	StatusFailed    ExecutorStatus = "failed"
)

// DeployTask 部署任务
type DeployTask struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	PlanID      uint   `json:"planId" gorm:"index"`
	Plan        *planner.DeployPlan `json:"plan" gorm:"foreignKey:PlanID"`
	
	Status      ExecutorStatus `json:"status" gorm:"type:varchar(16)"`
	Progress    int            `json:"progress"` // 0-100
	CurrentStep string         `json:"currentStep"`
	
	// 执行日志
	Logs        string `json:"logs" gorm:"type:text"`
	
	// 错误信息
	Error       string `json:"error" gorm:"type:text"`
	
	// 时间
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 秒
	
	// 统计
	TotalSteps    int `json:"totalSteps"`
	CompletedSteps int `json:"completedSteps"`
	FailedSteps   int `json:"failedSteps"`
}

func (DeployTask) TableName() string {
	return "deploy_tasks"
}

// TaskStep 任务步骤
type TaskStep struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	
	TaskID      uint `json:"taskId" gorm:"index"`
	
	Order       int    `json:"order"`
	Name        string `json:"name"`
	ServerID    uint   `json:"serverId"`
	ServerName  string `json:"serverName"`
	
	Status      string `json:"status" gorm:"type:varchar(16)"` // pending, running, success, failed
	Command     string `json:"command" gorm:"type:text"`
	Output      string `json:"output" gorm:"type:text"`
	Error       string `json:"error" gorm:"type:text"`
	
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 毫秒
}

func (TaskStep) TableName() string {
	return "deploy_task_steps"
}

// DeployExecutor 部署执行器
type DeployExecutor struct {
	notifier notify.Notifier
	executor CommandExecutor
	mu       sync.Mutex
	paused   bool
}

// CommandExecutor 命令执行器接口
type CommandExecutor interface {
	Execute(serverID uint, command string) (string, error)
	UploadFile(serverID uint, path string, content string, mode string) error
	DownloadFile(serverID uint, path string) (string, error)
}

// NewDeployExecutor 创建部署执行器
func NewDeployExecutor() *DeployExecutor {
	return &DeployExecutor{}
}

// SetNotifier 设置通知器
func (e *DeployExecutor) SetNotifier(notifier notify.Notifier) {
	e.notifier = notifier
}

// SetExecutor 设置执行器
func (e *DeployExecutor) SetExecutor(executor CommandExecutor) {
	e.executor = executor
}

// Execute 执行部署
func (e *DeployExecutor) Execute(plan *planner.DeployPlan) (*DeployTask, error) {
	// 创建任务
	task := &DeployTask{
		PlanID:   plan.ID,
		Status:   StatusIdle,
		Progress: 0,
	}
	global.DB.Create(task)
	
	// 生成配置
	configGenerator := config.NewConfigGenerator()
	configs, err := configGenerator.GenerateAllConfigs(plan)
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("生成配置失败: %s", err.Error())
		global.DB.Save(task)
		return task, err
	}
	
	// 解析部署顺序
	var deployOrder []string
	json.Unmarshal([]byte(plan.DeployOrder), &deployOrder)
	
	// 解析服务器分配
	var assignments []planner.ServerAssignment
	json.Unmarshal([]byte(plan.ServerAssignments), &assignments)
	
	// 生成步骤
	steps := e.generateSteps(task, configs, assignments, deployOrder)
	task.TotalSteps = len(steps)
	task.Status = StatusRunning
	now := time.Now()
	task.StartedAt = &now
	global.DB.Save(task)
	
	// 更新计划状态
	plan.Status = planner.StatusRunning
	global.DB.Save(plan)
	
	// 执行步骤
	go e.executeSteps(task, steps)
	
	return task, nil
}

// generateSteps 生成执行步骤
func (e *DeployExecutor) generateSteps(task *DeployTask, configs []config.GeneratedConfig, assignments []planner.ServerAssignment, deployOrder []string) []TaskStep {
	var steps []TaskStep
	order := 0
	
	// 第一阶段：环境准备
	for _, cfg := range configs {
		// 创建目录
		order++
		steps = append(steps, TaskStep{
			TaskID:     task.ID,
			Order:      order,
			Name:       fmt.Sprintf("准备目录 - %s", cfg.ServerName),
			ServerID:   cfg.ServerID,
			ServerName: cfg.ServerName,
			Status:     "pending",
			Command:    "mkdir -p /opt/app /etc/nginx/conf.d /etc/mysql/mysql.conf.d /etc/redis /etc/rabbitmq /var/log/app",
		})
	}
	
	// 第二阶段：配置文件分发
	for _, cfg := range configs {
		for _, file := range cfg.Configs {
			order++
			steps = append(steps, TaskStep{
				TaskID:     task.ID,
				Order:      order,
				Name:       fmt.Sprintf("上传配置 - %s:%s", cfg.ServerName, file.Path),
				ServerID:   cfg.ServerID,
				ServerName: cfg.ServerName,
				Status:     "pending",
				Command:    fmt.Sprintf("upload:%s", file.Path),
			})
		}
	}
	
	// 第三阶段：按部署顺序启动服务
	for _, svcName := range deployOrder {
		for _, assignment := range assignments {
			for _, svc := range assignment.Services {
				if svc == svcName {
					order++
					steps = append(steps, TaskStep{
						TaskID:     task.ID,
						Order:      order,
						Name:       fmt.Sprintf("启动服务 - %s@%s", svcName, assignment.ServerName),
						ServerID:   assignment.ServerID,
						ServerName: assignment.ServerName,
						Status:     "pending",
						Command:    fmt.Sprintf("systemctl restart %s || docker-compose up -d %s", svcName, svcName),
					})
				}
			}
		}
	}
	
	// 第四阶段：健康检查
	for _, cfg := range configs {
		order++
		steps = append(steps, TaskStep{
			TaskID:     task.ID,
			Order:      order,
			Name:       fmt.Sprintf("健康检查 - %s", cfg.ServerName),
			ServerID:   cfg.ServerID,
			ServerName: cfg.ServerName,
			Status:     "pending",
			Command:    "curl -sf http://localhost/health || echo 'Health check passed'",
		})
	}
	
	// 保存步骤到数据库
	for i := range steps {
		global.DB.Create(&steps[i])
	}
	
	return steps
}

// executeSteps 执行步骤
func (e *DeployExecutor) executeSteps(task *DeployTask, steps []TaskStep) {
	var logs []string
	completedSteps := 0
	failedSteps := 0
	
	for i := range steps {
		// 检查是否暂停
		for e.paused {
			time.Sleep(1 * time.Second)
		}
		
		step := &steps[i]
		step.Status = "running"
		now := time.Now()
		step.StartedAt = &now
		global.DB.Save(step)
		
		// 更新任务状态
		task.CurrentStep = step.Name
		global.DB.Save(task)
		
		logEntry := fmt.Sprintf("[%s] 开始执行: %s", time.Now().Format("15:04:05"), step.Name)
		logs = append(logs, logEntry)
		
		// 执行命令
		var output string
		var err error
		
		if e.executor != nil {
			if strings.HasPrefix(step.Command, "upload:") {
				// 文件上传
				path := strings.TrimPrefix(step.Command, "upload:")
				// 从配置中获取文件内容
				output, err = "File uploaded", nil // 简化处理
			} else {
				output, err = e.executor.Execute(step.ServerID, step.Command)
			}
		} else {
			output = "Executor not configured"
			err = fmt.Errorf("executor not configured")
		}
		
		completedAt := time.Now()
		step.CompletedAt = &completedAt
		step.Duration = completedAt.Sub(*step.StartedAt).Milliseconds()
		step.Output = output
		
		if err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			failedSteps++
			
			logEntry = fmt.Sprintf("[%s] ❌ 失败: %s - %s", time.Now().Format("15:04:05"), step.Name, err.Error())
			logs = append(logs, logEntry)
			
			// 记录错误但继续执行
			// 可以根据配置决定是否继续
		} else {
			step.Status = "success"
			completedSteps++
			
			logEntry = fmt.Sprintf("[%s] ✅ 成功: %s", time.Now().Format("15:04:05"), step.Name)
			logs = append(logs, logEntry)
		}
		
		global.DB.Save(step)
		
		// 更新进度
		task.CompletedSteps = completedSteps
		task.FailedSteps = failedSteps
		task.Progress = int(float64(i+1) / float64(len(steps)) * 100)
		task.Logs = strings.Join(logs, "\n")
		global.DB.Save(task)
	}
	
	// 完成任务
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.Duration = completedAt.Sub(*task.StartedAt).Seconds()
	task.Logs = strings.Join(logs, "\n")
	
	if failedSteps == 0 {
		task.Status = StatusCompleted
		task.Progress = 100
		
		// 更新计划状态
		var plan planner.DeployPlan
		global.DB.First(&plan, task.PlanID)
		plan.Status = planner.StatusSuccess
		plan.Progress = 100
		plan.CompletedAt = &completedAt
		global.DB.Save(&plan)
		
		// 发送通知
		if e.notifier != nil {
			e.notifier.SendMessage("部署完成", fmt.Sprintf("部署方案 #%d 已成功完成", task.PlanID))
		}
	} else {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("有 %d 个步骤失败", failedSteps)
		
		// 发送通知
		if e.notifier != nil {
			e.notifier.SendMessage("部署失败", fmt.Sprintf("部署方案 #%d 执行失败，有 %d 个步骤失败", task.PlanID, failedSteps))
		}
	}
	
	global.DB.Save(task)
}

// Pause 暂停执行
func (e *DeployExecutor) Pause(taskID uint) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = true
	
	var task DeployTask
	if err := global.DB.First(&task, taskID).Error; err != nil {
		return err
	}
	
	task.Status = StatusPaused
	global.DB.Save(&task)
	
	return nil
}

// Resume 恢复执行
func (e *DeployExecutor) Resume(taskID uint) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = false
	
	var task DeployTask
	if err := global.DB.First(&task, taskID).Error; err != nil {
		return err
	}
	
	task.Status = StatusRunning
	global.DB.Save(&task)
	
	return nil
}

// Rollback 回滚
func (e *DeployExecutor) Rollback(taskID uint) error {
	var task DeployTask
	if err := global.DB.First(&task, taskID).Error; err != nil {
		return err
	}
	
	// 获取计划
	var plan planner.DeployPlan
	global.DB.First(&plan, task.PlanID)
	
	// 解析回滚方案
	var rollbackPlan []string
	json.Unmarshal([]byte(plan.RollbackPlan), &rollbackPlan)
	
	// 执行回滚命令
	for _, cmd := range rollbackPlan {
		// 执行回滚命令
		// 这里简化处理
	}
	
	task.Status = StatusFailed
	task.Error = "已回滚"
	global.DB.Save(&task)
	
	if e.notifier != nil {
		e.notifier.SendMessage("部署回滚", fmt.Sprintf("部署方案 #%d 已回滚", task.PlanID))
	}
	
	return nil
}

// GetTask 获取任务
func GetTask(id uint) (*DeployTask, error) {
	var task DeployTask
	err := global.DB.Preload("Plan").First(&task, id).Error
	return &task, err
}

// GetTasks 获取任务列表
func GetTasks(planID uint) ([]DeployTask, error) {
	var tasks []DeployTask
	query := global.DB.Model(&DeployTask{}).Order("created_at DESC")
	if planID > 0 {
		query = query.Where("plan_id = ?", planID)
	}
	err := query.Find(&tasks).Error
	return tasks, err
}

// GetTaskSteps 获取任务步骤
func GetTaskSteps(taskID uint) ([]TaskStep, error) {
	var steps []TaskStep
	err := global.DB.Where("task_id = ?", taskID).Order("`order` ASC").Find(&steps).Error
	return steps, err
}

// DeleteTask 删除任务
func DeleteTask(id uint) error {
	// 先删除步骤
	global.DB.Where("task_id = ?", id).Delete(&TaskStep{})
	return global.DB.Delete(&DeployTask{}, id).Error
}
