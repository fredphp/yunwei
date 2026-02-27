package workflow

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/model/server"
	"yunwei/service/ai/decision"
	"yunwei/service/ai/llm"
	"yunwei/service/detector"
	"yunwei/service/executor"
	"yunwei/model/notify"
	"yunwei/service/security"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
)

// WorkflowType 工作流类型
type WorkflowType string

const (
	WorkflowTypeAutoFix   WorkflowType = "auto_fix"
	WorkflowTypeAlert     WorkflowType = "alert"
	WorkflowTypeManual    WorkflowType = "manual"
)

// WorkflowRecord 工作流记录
type WorkflowRecord struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"createdAt"`
	Type         WorkflowType   `json:"type" gorm:"type:varchar(32)"`
	Status       WorkflowStatus `json:"status" gorm:"type:varchar(16)"`
	TriggerSource string         `json:"triggerSource" gorm:"type:varchar(64)"`
	ServerID     uint           `json:"serverId" gorm:"index"`
	Server       *server.Server `json:"server" gorm:"foreignKey:ServerID"`
	Steps        string         `json:"steps" gorm:"type:text"`
	Result       string         `json:"result" gorm:"type:text"`
	AIAnalysis   string         `json:"aiAnalysis" gorm:"type:text"`
	Commands     string         `json:"commands" gorm:"type:text"`
	Output       string         `json:"output" gorm:"type:text"`
	StartedAt    *time.Time     `json:"startedAt"`
	CompletedAt  *time.Time     `json:"completedAt"`
	Duration     int64          `json:"duration"`
	AutoMode     bool           `json:"autoMode"`
	NeedApprove  bool           `json:"needApprove"`
	ApprovedBy   uint           `json:"approvedBy"`
}

func (WorkflowRecord) TableName() string {
	return "workflow_records"
}

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	detector   *detector.Detector
	executor   *executor.Executor
	decision   *decision.Engine
	security   *security.SecurityChecker
	notifier   notify.Notifier
	llmClient  *llm.GLM5Client
	runningWorkflows sync.Map
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		detector: detector.NewDetector(),
		executor: executor.NewExecutor(),
		security: security.NewSecurityChecker(),
	}
}

// SetLLMClient 设置LLM客户端
func (e *WorkflowEngine) SetLLMClient(client *llm.GLM5Client) {
	e.llmClient = client
	e.decision = decision.NewEngine(client)
}

// SetNotifier 设置通知器
func (e *WorkflowEngine) SetNotifier(notifier notify.Notifier) {
	e.notifier = notifier
}

// TriggerFromAlert 从告警触发
func (e *WorkflowEngine) TriggerFromAlert(alert *detector.Alert) (*WorkflowRecord, error) {
	workflow := &WorkflowRecord{
		Type:          WorkflowTypeAutoFix,
		Status:        WorkflowStatusPending,
		TriggerSource: "alert",
		ServerID:      alert.ServerID,
		AutoMode:      true,
	}
	
	if err := global.DB.Create(workflow).Error; err != nil {
		return nil, err
	}
	
	go e.Execute(workflow)
	return workflow, nil
}

// Execute 执行工作流
func (e *WorkflowEngine) Execute(workflow *WorkflowRecord) error {
	if _, exists := e.runningWorkflows.Load(workflow.ID); exists {
		return fmt.Errorf("工作流已在运行")
	}

	e.runningWorkflows.Store(workflow.ID, true)
	defer e.runningWorkflows.Delete(workflow.ID)

	now := time.Now()
	workflow.Status = WorkflowStatusRunning
	workflow.StartedAt = &now
	global.DB.Save(workflow)

	// 获取服务器
	var srv server.Server
	if err := global.DB.First(&srv, workflow.ServerID).Error; err != nil {
		return e.failWorkflow(workflow, "服务器不存在")
	}

	// 执行自动化流程
	// 1. 数据采集
	metric, err := e.stepDetect(&srv)
	if err != nil {
		return e.failWorkflow(workflow, err.Error())
	}

	// 2. AI分析
	aiDecision, err := e.stepAnalyze(&srv, metric)
	if err != nil {
		return e.failWorkflow(workflow, err.Error())
	}
	workflow.AIAnalysis = aiDecision.Analysis

	// 3. 安全检查
	commands := e.parseCommands(aiDecision.Commands)
	result := e.security.ValidateCommands(commands)
	if !result.Allowed {
		return e.failWorkflow(workflow, "安全检查未通过: "+result.Message)
	}
	workflow.NeedApprove = result.RequiresApproval

	// 4. 执行命令
	output, err := e.stepExecute(&srv, commands)
	if err != nil {
		return e.failWorkflow(workflow, err.Error())
	}
	workflow.Output = output
	workflow.Commands = aiDecision.Commands

	// 5. 记录日志
	log := server.ServerLog{
		ServerID: srv.ID,
		Type:     "workflow",
		Content:  "自动修复工作流",
		Output:   output,
	}
	global.DB.Create(&log)

	// 6. 发送通知
	if e.notifier != nil {
		e.notifier.SendMessage("自动修复完成", 
			fmt.Sprintf("服务器 %s 自动修复完成", srv.Name))
	}

	return e.completeWorkflow(workflow)
}

func (e *WorkflowEngine) stepDetect(srv *server.Server) (*server.ServerMetric, error) {
	var metric server.ServerMetric
	if err := global.DB.Where("server_id = ?", srv.ID).Order("created_at DESC").First(&metric).Error; err != nil {
		return nil, fmt.Errorf("无法获取指标")
	}
	return &metric, nil
}

func (e *WorkflowEngine) stepAnalyze(srv *server.Server, metric *server.ServerMetric) (*decision.AIDecision, error) {
	if e.llmClient == nil {
		return nil, fmt.Errorf("AI客户端未配置")
	}
	return e.decision.QuickAnalyze(srv, metric)
}

func (e *WorkflowEngine) stepExecute(srv *server.Server, commands []string) (string, error) {
	// TODO: 通过SSH或Agent执行
	var outputs []string
	for _, cmd := range commands {
		outputs = append(outputs, fmt.Sprintf("[执行] %s", cmd))
	}
	return strings.Join(outputs, "\n"), nil
}

func (e *WorkflowEngine) parseCommands(commandsJSON string) []string {
	var commands []string
	json.Unmarshal([]byte(commandsJSON), &commands)
	return commands
}

func (e *WorkflowEngine) failWorkflow(workflow *WorkflowRecord, reason string) error {
	now := time.Now()
	workflow.Status = WorkflowStatusFailed
	workflow.CompletedAt = &now
	workflow.Result = reason
	global.DB.Save(workflow)
	
	if e.notifier != nil {
		e.notifier.SendMessage("工作流失败", reason)
	}
	return fmt.Errorf(reason)
}

func (e *WorkflowEngine) completeWorkflow(workflow *WorkflowRecord) error {
	now := time.Now()
	workflow.Status = WorkflowStatusCompleted
	workflow.CompletedAt = &now
	workflow.Duration = now.Sub(*workflow.StartedAt).Milliseconds()
	workflow.Result = "执行成功"
	return global.DB.Save(workflow).Error
}

// GetHistory 获取历史
func (e *WorkflowEngine) GetHistory(serverID uint, limit int) ([]WorkflowRecord, error) {
	var records []WorkflowRecord
	query := global.DB.Model(&WorkflowRecord{}).Order("created_at DESC")
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	return records, query.Find(&records).Error
}
