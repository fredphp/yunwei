package executor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yunwei/global"
	"yunwei/model/server"
	"yunwei/service/security"

	"gorm.io/gorm"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
)

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	ServerID  uint           `json:"serverId" gorm:"index"`
	Server    *server.Server `json:"server" gorm:"foreignKey:ServerID"`
	DecisionID uint          `json:"decisionId" gorm:"index"`

	// 执行信息
	Source      string         `json:"source" gorm:"type:varchar(32)"` // ai, manual, scheduled
	Commands    string         `json:"commands" gorm:"type:text"`       // JSON数组
	CommandCount int           `json:"commandCount"`
	Status      ExecutionStatus `json:"status" gorm:"type:varchar(16);default:'pending'"`

	// 执行详情
	CurrentIndex int    `json:"currentIndex"` // 当前执行命令索引
	Output       string `json:"output" gorm:"type:text"`
	Error        string `json:"error" gorm:"type:text"`

	// 时间信息
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 总耗时(ms)

	// 审核信息
	ApprovedBy  uint       `json:"approvedBy"`
	ApprovedAt  *time.Time `json:"approvedAt"`
}

func (ExecutionRecord) TableName() string {
	return "execution_records"
}

// CommandResult 命令执行结果
type CommandResult struct {
	Index     int    `json:"index"`
	Command   string `json:"command"`
	Output    string `json:"output"`
	Error     string `json:"error"`
	Success   bool   `json:"success"`
	Duration  int64  `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// AIResponse AI响应结构
type AIResponse struct {
	Analysis string   `json:"analysis"`
	Commands []string `json:"commands"`
}

// Executor 执行器
type Executor struct {
	securityChecker *security.SecurityChecker
	sshExecutor     CommandRunner
}

// CommandRunner 命令运行器接口
type CommandRunner interface {
	Run(serverID uint, command string) (string, error)
}

// NewExecutor 创建执行器
func NewExecutor() *Executor {
	return &Executor{
		securityChecker: security.NewSecurityChecker(),
	}
}

// SetCommandRunner 设置命令运行器
func (e *Executor) SetCommandRunner(runner CommandRunner) {
	e.sshExecutor = runner
}

// ParseAIResponse 解析AI响应
func (e *Executor) ParseAIResponse(response string) (*AIResponse, error) {
	// 尝试提取JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("未找到有效的JSON结构")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &aiResp); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	return &aiResp, nil
}

// ExtractCommands 从文本中提取命令
func (e *Executor) ExtractCommands(text string) []string {
	var commands []string
	lines := strings.Split(text, "\n")

	inCodeBlock := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检测代码块
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// 提取代码块中的命令
		if inCodeBlock && line != "" {
			// 跳过注释
			if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
				commands = append(commands, line)
			}
		}

		// 提取独立命令行
		if !inCodeBlock {
			shellPrefixes := []string{
				"sudo ", "systemctl ", "docker ", "kubectl ",
				"kill ", "pkill ", "rm ", "mv ", "cp ",
				"echo ", "sync ", "find ", "journalctl ",
				"service ", "/etc/init.d/", "chmod ", "chown ",
			}
			for _, prefix := range shellPrefixes {
				if strings.HasPrefix(line, prefix) {
					commands = append(commands, line)
					break
				}
			}
		}
	}

	return commands
}

// ValidateCommands 验证命令安全性
func (e *Executor) ValidateCommands(commands []string) *security.ValidationResult {
	return e.securityChecker.ValidateCommands(commands)
}

// CreateExecutionRecord 创建执行记录
func (e *Executor) CreateExecutionRecord(serverID uint, commands []string, source string, decisionID uint) (*ExecutionRecord, error) {
	commandsJSON, _ := json.Marshal(commands)

	record := &ExecutionRecord{
		ServerID:     serverID,
		DecisionID:   decisionID,
		Source:       source,
		Commands:     string(commandsJSON),
		CommandCount: len(commands),
		Status:       ExecutionStatusPending,
	}

	if err := global.DB.Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建执行记录失败: %w", err)
	}

	return record, nil
}

// Execute 执行命令
func (e *Executor) Execute(record *ExecutionRecord, runner CommandRunner) ([]CommandResult, error) {
	var commands []string
	if err := json.Unmarshal([]byte(record.Commands), &commands); err != nil {
		return nil, fmt.Errorf("解析命令失败: %w", err)
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("没有可执行的命令")
	}

	// 再次进行安全检查
	validation := e.securityChecker.ValidateCommands(commands)
	if !validation.Allowed {
		return nil, fmt.Errorf("命令安全检查未通过: %s", validation.Message)
	}

	// 更新状态
	record.Status = ExecutionStatusRunning
	now := time.Now()
	record.StartedAt = &now
	global.DB.Save(record)

	var results []CommandResult
	var hasError bool

	for i, cmd := range validation.SafeCommands {
		record.CurrentIndex = i
		global.DB.Save(record)

		result := CommandResult{
			Index:     i,
			Command:   cmd,
			Timestamp: time.Now(),
		}

		startTime := time.Now()
		output, err := runner.Run(record.ServerID, cmd)
		result.Duration = time.Since(startTime).Milliseconds()

		if err != nil {
			result.Error = err.Error()
			result.Success = false
			hasError = true

			// 记录失败日志
			e.logExecution(record.ServerID, cmd, "", err.Error(), result.Duration)
		} else {
			result.Output = output
			result.Success = true

			// 记录成功日志
			e.logExecution(record.ServerID, cmd, output, "", result.Duration)
		}

		results = append(results, result)

		// 如果有错误，停止执行后续命令
		if hasError {
			break
		}
	}

	// 更新最终状态
	completedAt := time.Now()
	record.CompletedAt = &completedAt
	record.Duration = completedAt.Sub(*record.StartedAt).Milliseconds()

	if hasError {
		record.Status = ExecutionStatusFailed
	} else {
		record.Status = ExecutionStatusSuccess
	}

	// 序列化结果
	resultsJSON, _ := json.Marshal(results)
	record.Output = string(resultsJSON)
	global.DB.Save(record)

	return results, nil
}

// ExecuteWithApproval 带审批的执行
func (e *Executor) ExecuteWithApproval(record *ExecutionRecord, runner CommandRunner, approvedBy uint) error {
	// 检查是否需要审批
	validation := e.securityChecker.ValidateCommands(e.parseCommands(record.Commands))
	if validation.RequiresApproval {
		if record.ApprovedBy == 0 {
			return fmt.Errorf("此操作需要管理员审批")
		}
	}

	return e.ExecuteAsync(record, runner)
}

// ExecuteAsync 异步执行
func (e *Executor) ExecuteAsync(record *ExecutionRecord, runner CommandRunner) error {
	go func() {
		_, err := e.Execute(record, runner)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("执行失败: %v", err))
		}
	}()
	return nil
}

// Cancel 取消执行
func (e *Executor) Cancel(recordID uint) error {
	var record ExecutionRecord
	if err := global.DB.First(&record, recordID).Error; err != nil {
		return err
	}

	if record.Status == ExecutionStatusRunning {
		return fmt.Errorf("正在执行中，无法取消")
	}

	record.Status = ExecutionStatusCancelled
	return global.DB.Save(&record).Error
}

// GetExecutionHistory 获取执行历史
func (e *Executor) GetExecutionHistory(serverID uint, limit int) ([]ExecutionRecord, error) {
	var records []ExecutionRecord
	query := global.DB.Model(&ExecutionRecord{}).Order("created_at DESC")

	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

// GetExecutionDetail 获取执行详情
func (e *Executor) GetExecutionDetail(recordID uint) (*ExecutionRecord, error) {
	var record ExecutionRecord
	if err := global.DB.First(&record, recordID).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// logExecution 记录执行日志
func (e *Executor) logExecution(serverID uint, command, output, errMsg string, duration int64) {
	log := server.ServerLog{
		ServerID: serverID,
		Type:     "auto_fix",
		Content:  command,
		Output:   output,
		Error:    errMsg,
		Duration: duration,
	}

	global.DB.Create(&log)
}

// parseCommands 解析命令JSON
func (e *Executor) parseCommands(commandsJSON string) []string {
	var commands []string
	json.Unmarshal([]byte(commandsJSON), &commands)
	return commands
}

// Retry 重试执行
func (e *Executor) Retry(recordID uint, runner CommandRunner) (*ExecutionRecord, error) {
	var record ExecutionRecord
	if err := global.DB.First(&record, recordID).Error; err != nil {
		return nil, err
	}

	if record.Status != ExecutionStatusFailed {
		return nil, fmt.Errorf("只能重试失败的执行")
	}

	// 重置状态
	record.Status = ExecutionStatusPending
	record.StartedAt = nil
	record.CompletedAt = nil
	record.Output = ""
	record.Error = ""
	record.CurrentIndex = 0

	global.DB.Save(&record)

	_, err := e.Execute(&record, runner)
	return &record, err
}

// BatchExecute 批量执行
func (e *Executor) BatchExecute(serverIDs []uint, commands []string, source string, runner CommandRunner) ([]*ExecutionRecord, error) {
	// 验证命令
	validation := e.securityChecker.ValidateCommands(commands)
	if !validation.Allowed {
		return nil, fmt.Errorf("命令安全检查未通过: %s", validation.Message)
	}

	var records []*ExecutionRecord
	for _, serverID := range serverIDs {
		record, err := e.CreateExecutionRecord(serverID, validation.SafeCommands, source, 0)
		if err != nil {
			continue
		}

		go e.Execute(record, runner)
		records = append(records, record)
	}

	return records, nil
}

// GetStatistics 获取执行统计
func (e *Executor) GetStatistics(serverID uint, days int) map[string]interface{} {
	startTime := time.Now().AddDate(0, 0, -days)

	stats := make(map[string]interface{})

	// 总执行次数
	var total int64
	global.DB.Model(&ExecutionRecord{}).
		Where("server_id = ? AND created_at > ?", serverID, startTime).
		Count(&total)
	stats["total"] = total

	// 成功次数
	var success int64
	global.DB.Model(&ExecutionRecord{}).
		Where("server_id = ? AND status = ? AND created_at > ?", serverID, ExecutionStatusSuccess, startTime).
		Count(&success)
	stats["success"] = success

	// 失败次数
	var failed int64
	global.DB.Model(&ExecutionRecord{}).
		Where("server_id = ? AND status = ? AND created_at > ?", serverID, ExecutionStatusFailed, startTime).
		Count(&failed)
	stats["failed"] = failed

	// 成功率
	if total > 0 {
		stats["successRate"] = float64(success) / float64(total) * 100
	} else {
		stats["successRate"] = 0
	}

	return stats
}
