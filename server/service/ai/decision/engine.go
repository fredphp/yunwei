package decision

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yunwei/model/server"
	"yunwei/service/ai/llm"
	"yunwei/service/detector"
	"yunwei/service/optimizer"
)

// DecisionType 决策类型
type DecisionType string

const (
	DecisionTypeAuto     DecisionType = "auto"     // 自动执行
	DecisionTypeManual   DecisionType = "manual"   // 人工确认
	DecisionTypeIgnore   DecisionType = "ignore"   // 忽略
	DecisionTypeSchedule DecisionType = "schedule" // 定时执行
)

// DecisionStatus 决策状态
type DecisionStatus string

const (
	DecisionStatusPending   DecisionStatus = "pending"
	DecisionStatusApproved  DecisionStatus = "approved"
	DecisionStatusRejected  DecisionStatus = "rejected"
	DecisionStatusExecuted  DecisionStatus = "executed"
	DecisionStatusFailed    DecisionStatus = "failed"
)

// AIDecision AI 决策记录
type AIDecision struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	
	ServerID  uint           `json:"serverId" gorm:"index"`
	Server    *server.Server `json:"server" gorm:"foreignKey:ServerID"`
	AlertID   uint           `json:"alertId" gorm:"index"`
	
	// 决策信息
	Type          DecisionType   `json:"type" gorm:"type:varchar(16)"`
	Status        DecisionStatus `json:"status" gorm:"type:varchar(16);default:'pending'"`
	Confidence    float64        `json:"confidence"` // 置信度 0-1
	
	// AI 分析
	Summary       string         `json:"summary" gorm:"type:text"`
	Analysis      string         `json:"analysis" gorm:"type:text"`
	Suggestions   string         `json:"suggestions" gorm:"type:text"`
	Commands      string         `json:"commands" gorm:"type:text"` // 推荐命令(JSON数组)
	
	// 执行信息
	SelectedCommand string        `json:"selectedCommand" gorm:"type:text"`
	ExecutionResult string        `json:"executionResult" gorm:"type:text"`
	ExecutionError  string        `json:"executionError" gorm:"type:text"`
	ExecutedAt      *time.Time    `json:"executedAt"`
	
	// 审核信息
	ApprovedBy    uint           `json:"approvedBy"`
	ApprovedAt    *time.Time    `json:"approvedAt"`
	RejectedBy    uint           `json:"rejectedBy"`
	RejectedAt    *time.Time    `json:"rejectedAt"`
	RejectReason  string         `json:"rejectReason" gorm:"type:text"`
}

func (AIDecision) TableName() string {
	return "ai_decisions"
}

// Engine 决策引擎
type Engine struct {
	llmClient   *llm.GLM5Client
	detector    *detector.Detector
	optimizer   *optimizer.Optimizer
	actionGen   *optimizer.AIActionGenerator
}

// NewEngine 创建决策引擎
func NewEngine(llmClient *llm.GLM5Client) *Engine {
	return &Engine{
		llmClient: llmClient,
		detector:  detector.NewDetector(),
		optimizer: optimizer.NewOptimizer(),
		actionGen: optimizer.NewAIActionGenerator(),
	}
}

// ServerStatusSummary 服务器状态摘要
type ServerStatusSummary struct {
	ServerID    uint      `json:"serverId"`
	ServerName  string    `json:"serverName"`
	Timestamp   time.Time `json:"timestamp"`
	
	// 基本信息
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	CPUCores    int    `json:"cpuCores"`
	MemoryTotal uint64 `json:"memoryTotal"`
	DiskTotal   uint64 `json:"diskTotal"`
	
	// 实时指标
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	DiskUsage   float64 `json:"diskUsage"`
	Load1       float64 `json:"load1"`
	Load5       float64 `json:"load5"`
	Load15      float64 `json:"load15"`
	
	// 进程和容器
	ProcessCount    int `json:"processCount"`
	ContainerCount  int `json:"containerCount"`
	RunningContainers int `json:"runningContainers"`
	
	// 网络状态
	NetInSpeed  uint64 `json:"netInSpeed"`
	NetOutSpeed uint64 `json:"netOutSpeed"`
	
	// 告警
	ActiveAlerts int `json:"activeAlerts"`
	AlertTypes   []string `json:"alertTypes"`
}

// GenerateSummary 生成服务器状态摘要
func (e *Engine) GenerateSummary(srv *server.Server, metric *server.ServerMetric, containers []server.DockerContainer) ServerStatusSummary {
	summary := ServerStatusSummary{
		ServerID:        srv.ID,
		ServerName:      srv.Name,
		Timestamp:       time.Now(),
		OS:              srv.OS,
		Arch:            srv.Arch,
		CPUCores:        srv.CPUCores,
		MemoryTotal:     srv.MemoryTotal,
		DiskTotal:       srv.DiskTotal,
		CPUUsage:        metric.CPUUsage,
		MemoryUsage:     metric.MemoryUsage,
		DiskUsage:       metric.DiskUsage,
		Load1:           metric.Load1,
		Load5:           metric.Load5,
		Load15:          metric.Load15,
		ProcessCount:    metric.ProcessCount,
		ContainerCount:  len(containers),
	}

	runningCount := 0
	for _, c := range containers {
		if c.State == "running" {
			runningCount++
		}
	}
	summary.RunningContainers = runningCount

	return summary
}

// Analyze 分析服务器状态并生成决策
func (e *Engine) Analyze(summary ServerStatusSummary, alerts []detector.DetectionResult) (*AIDecision, error) {
	// 构建 prompt
	prompt := e.buildAnalysisPrompt(summary, alerts)

	// 调用 GLM5
	response, err := e.llmClient.QuickChat(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析响应
	decision := &AIDecision{
		ServerID: summary.ServerID,
		Type:     DecisionTypeManual, // 默认需要人工确认
		Status:   DecisionStatusPending,
	}

	// 解析AI响应
	e.parseAIResponse(response, decision)

	// 根据告警级别决定执行类型
	if len(alerts) > 0 {
		highestLevel := e.getHighestAlertLevel(alerts)
		if highestLevel == detector.AlertLevelEmergency || highestLevel == detector.AlertLevelCritical {
			// 高危告警，检查是否可以自动执行
			if !e.hasDangerousCommands(decision.Commands) {
				decision.Type = DecisionTypeAuto
			}
		}
	}

	return decision, nil
}

// buildAnalysisPrompt 构建分析提示词
func (e *Engine) buildAnalysisPrompt(summary ServerStatusSummary, alerts []detector.DetectionResult) string {
	var sb strings.Builder

	sb.WriteString("你是一个专业的Linux运维专家。请分析以下服务器状态并给出优化建议。\n\n")

	// 服务器基本信息
	sb.WriteString("## 服务器基本信息\n")
	sb.WriteString(fmt.Sprintf("- 名称: %s\n", summary.ServerName))
	sb.WriteString(fmt.Sprintf("- 系统: %s %s\n", summary.OS, summary.Arch))
	sb.WriteString(fmt.Sprintf("- CPU核心: %d\n", summary.CPUCores))
	sb.WriteString(fmt.Sprintf("- 内存总量: %d MB\n", summary.MemoryTotal))
	sb.WriteString(fmt.Sprintf("- 磁盘总量: %d GB\n", summary.DiskTotal))

	// 实时指标
	sb.WriteString("\n## 当前状态指标\n")
	sb.WriteString(fmt.Sprintf("- CPU使用率: %.2f%%\n", summary.CPUUsage))
	sb.WriteString(fmt.Sprintf("- 内存使用率: %.2f%%\n", summary.MemoryUsage))
	sb.WriteString(fmt.Sprintf("- 磁盘使用率: %.2f%%\n", summary.DiskUsage))
	sb.WriteString(fmt.Sprintf("- 系统负载: %.2f, %.2f, %.2f (1/5/15分钟)\n", summary.Load1, summary.Load5, summary.Load15))
	sb.WriteString(fmt.Sprintf("- 进程数: %d\n", summary.ProcessCount))
	sb.WriteString(fmt.Sprintf("- Docker容器: %d (运行中: %d)\n", summary.ContainerCount, summary.RunningContainers))

	// 告警信息
	if len(alerts) > 0 {
		sb.WriteString("\n## 当前告警\n")
		for _, alert := range alerts {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", alert.Level, alert.Title, alert.Message))
		}
	}

	// 请求格式
	sb.WriteString("\n## 请按以下格式回复\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "summary": "一句话总结当前状态",
  "analysis": "详细分析问题原因",
  "suggestions": "优化建议列表",
  "commands": ["可执行的Shell命令1", "可执行的Shell命令2"],
  "risk_level": "low/medium/high",
  "auto_execute": true/false
}` + "\n")
	sb.WriteString("```\n")

	return sb.String()
}

// parseAIResponse 解析AI响应
func (e *Engine) parseAIResponse(response string, decision *AIDecision) {
	// 提取JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		// 无法解析JSON，直接使用响应
		decision.Summary = "AI分析结果"
		decision.Analysis = response
		decision.Suggestions = "请手动检查服务器状态"
		return
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var result struct {
		Summary     string   `json:"summary"`
		Analysis    string   `json:"analysis"`
		Suggestions string   `json:"suggestions"`
		Commands    []string `json:"commands"`
		RiskLevel   string   `json:"risk_level"`
		AutoExecute bool     `json:"auto_execute"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		decision.Summary = "AI分析结果"
		decision.Analysis = response
		return
	}

	decision.Summary = result.Summary
	decision.Analysis = result.Analysis
	decision.Suggestions = result.Suggestions

	// 序列化命令
	if len(result.Commands) > 0 {
		commandsJSON, _ := json.Marshal(result.Commands)
		decision.Commands = string(commandsJSON)
	}

	// 根据风险等级设置置信度
	switch result.RiskLevel {
	case "low":
		decision.Confidence = 0.9
	case "medium":
		decision.Confidence = 0.6
	case "high":
		decision.Confidence = 0.3
	default:
		decision.Confidence = 0.5
	}

	// 如果AI建议自动执行且风险低
	if result.AutoExecute && result.RiskLevel == "low" {
		decision.Type = DecisionTypeAuto
	}
}

// getHighestAlertLevel 获取最高告警级别
func (e *Engine) getHighestAlertLevel(alerts []detector.DetectionResult) detector.AlertLevel {
	levelPriority := map[detector.AlertLevel]int{
		detector.AlertLevelInfo:      1,
		detector.AlertLevelWarning:   2,
		detector.AlertLevelCritical:  3,
		detector.AlertLevelEmergency: 4,
	}

	highest := detector.AlertLevelInfo
	for _, alert := range alerts {
		if levelPriority[alert.Level] > levelPriority[highest] {
			highest = alert.Level
		}
	}
	return highest
}

// hasDangerousCommands 检查是否有危险命令
func (e *Engine) hasDangerousCommands(commandsJSON string) bool {
	var commands []string
	if err := json.Unmarshal([]byte(commandsJSON), &commands); err != nil {
		return true // 解析失败，认为是危险的
	}

	dangerousKeywords := []string{
		"rm -rf /",
		"mkfs",
		"dd if=",
		":(){ :|:& };:",
		"shutdown",
		"reboot",
		"init 0",
		"halt",
		"poweroff",
		"> /dev/sda",
		"chmod -R 777 /",
		"chown -R",
		"DROP DATABASE",
		"DROP TABLE",
		"TRUNCATE",
	}

	for _, cmd := range commands {
		for _, keyword := range dangerousKeywords {
			if strings.Contains(cmd, keyword) {
				return true
			}
		}
	}

	return false
}

// QuickAnalyze 快速分析（简化版）
func (e *Engine) QuickAnalyze(srv *server.Server, metric *server.ServerMetric) (*AIDecision, error) {
	summary := e.GenerateSummary(srv, metric, nil)

	prompt := fmt.Sprintf(`当前服务器状态如下：
CPU: %.2f%%
内存: %.2f%%
负载: %.2f
Docker容器数: %d

请给出优化建议，并输出可执行的Shell命令。要求：
1. 简洁明了，直接给出问题
2. 提供具体可执行的命令
3. 说明命令的作用`,
		metric.CPUUsage, metric.MemoryUsage, metric.Load1, 0)

	response, err := e.llmClient.QuickChat(prompt)
	if err != nil {
		return nil, err
	}

	decision := &AIDecision{
		ServerID:  srv.ID,
		Type:      DecisionTypeManual,
		Status:    DecisionStatusPending,
		Summary:   "服务器状态分析",
		Analysis:  response,
	}

	// 提取命令
	commands := e.actionGen.ParseAIResponse(response)
	if len(commands) > 0 {
		commandsJSON, _ := json.Marshal(commands)
		decision.Commands = string(commandsJSON)
	}

	return decision, nil
}

// ExecuteDecision 执行决策
func (e *Engine) ExecuteDecision(decision *AIDecision, executor CommandExecutor) error {
	var commands []string
	if err := json.Unmarshal([]byte(decision.Commands), &commands); err != nil {
		return fmt.Errorf("解析命令失败: %w", err)
	}

	if len(commands) == 0 {
		return fmt.Errorf("没有可执行的命令")
	}

	var results []string
	for _, cmd := range commands {
		output, err := executor.Execute(cmd)
		if err != nil {
			results = append(results, fmt.Sprintf("命令: %s\n错误: %s", cmd, err.Error()))
		} else {
			results = append(results, fmt.Sprintf("命令: %s\n输出: %s", cmd, output))
		}
	}

	now := time.Now()
	decision.ExecutedAt = &now
	decision.ExecutionResult = strings.Join(results, "\n---\n")
	decision.Status = DecisionStatusExecuted

	return nil
}

// CommandExecutor 命令执行器接口
type CommandExecutor interface {
	Execute(command string) (string, error)
}

// GetDecisionHistory 获取决策历史
func GetDecisionHistory(serverID uint, limit int) []AIDecision {
	// 这里应该从数据库查询，返回模拟数据
	return []AIDecision{}
}

// AnalyzeTrend 分析趋势
func (e *Engine) AnalyzeTrend(serverID uint, metrics []server.ServerMetric) string {
	if len(metrics) < 2 {
		return "数据不足，无法分析趋势"
	}

	var cpuTrend, memTrend, diskTrend string
	first := metrics[0]
	last := metrics[len(metrics)-1]

	// CPU 趋势
	if last.CPUUsage > first.CPUUsage+10 {
		cpuTrend = "CPU使用率上升趋势，需关注"
	} else if last.CPUUsage < first.CPUUsage-10 {
		cpuTrend = "CPU使用率下降趋势，状态良好"
	} else {
		cpuTrend = "CPU使用率稳定"
	}

	// 内存趋势
	if last.MemoryUsage > first.MemoryUsage+10 {
		memTrend = "内存使用率上升趋势，可能存在内存泄漏"
	} else if last.MemoryUsage < first.MemoryUsage-10 {
		memTrend = "内存使用率下降趋势"
	} else {
		memTrend = "内存使用率稳定"
	}

	// 磁盘趋势
	if last.DiskUsage > first.DiskUsage+5 {
		diskTrend = "磁盘使用率上升，建议清理"
	} else {
		diskTrend = "磁盘使用率稳定"
	}

	return fmt.Sprintf("趋势分析:\n- %s\n- %s\n- %s", cpuTrend, memTrend, diskTrend)
}
