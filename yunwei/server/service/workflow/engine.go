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
        "yunwei/service/notify"
        "yunwei/service/security"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
        WorkflowStatusPending   WorkflowStatus = "pending"
        WorkflowStatusRunning   WorkflowStatus = "running"
        WorkflowStatusCompleted WorkflowStatus = "completed"
        WorkflowStatusFailed    WorkflowStatus = "failed"
        WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// WorkflowType 工作流类型
type WorkflowType string

const (
        WorkflowTypeAutoFix   WorkflowType = "auto_fix"   // 自动修复
        WorkflowTypeAlert     WorkflowType = "alert"      // 告警处理
        WorkflowTypeScheduled WorkflowType = "scheduled"  // 定时任务
        WorkflowTypeManual    WorkflowType = "manual"     // 手动触发
)

// WorkflowRecord 工作流记录
type WorkflowRecord struct {
        ID        uint           `json:"id" gorm:"primarykey"`
        CreatedAt time.Time      `json:"createdAt"`
        UpdatedAt time.Time      `json:"updatedAt"`

        Type      WorkflowType   `json:"type" gorm:"type:varchar(32)"`
        Status    WorkflowStatus `json:"status" gorm:"type:varchar(16)"`

        // 触发信息
        TriggerSource string `json:"triggerSource" gorm:"type:varchar(64)"` // alert, scheduled, manual, ai
        TriggerID     uint   `json:"triggerId"`                             // 触发源ID
        TriggerDetail string `json:"triggerDetail" gorm:"type:text"`        // 触发详情

        // 服务器信息
        ServerID   uint           `json:"serverId" gorm:"index"`
        Server     *server.Server `json:"server" gorm:"foreignKey:ServerID"`

        // 步骤
        CurrentStep int    `json:"currentStep"`
        TotalSteps  int    `json:"totalSteps"`
        Steps       string `json:"steps" gorm:"type:text"` // JSON数组

        // 结果
        Result      string `json:"result" gorm:"type:text"`
        AIAnalysis  string `json:"aiAnalysis" gorm:"type:text"`
        Commands    string `json:"commands" gorm:"type:text"`
        Output      string `json:"output" gorm:"type:text"`

        // 时间
        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration    int64      `json:"duration"` // 毫秒

        // 自动化标记
        AutoMode    bool `json:"autoMode"`    // 是否自动模式
        NeedApprove bool `json:"needApprove"` // 是否需要审批
        ApprovedBy  uint `json:"approvedBy"`
        ApprovedAt  *time.Time `json:"approvedAt"`
}

func (WorkflowRecord) TableName() string {
        return "workflow_records"
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
        StepID     int       `json:"stepId"`
        Name       string    `json:"name"`
        Type       string    `json:"type"`       // detect, analyze, decide, execute, notify
        Status     string    `json:"status"`     // pending, running, success, failed, skipped
        StartedAt  *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration   int64     `json:"duration"`
        Input      string    `json:"input"`      // JSON
        Output     string    `json:"output"`     // JSON
        Error      string    `json:"error"`
}

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
        detector   *detector.Detector
        executor   *executor.Executor
        decision   *decision.Engine
        security   *security.SecurityChecker
        notifier   notify.Notifier
        llmClient  *llm.GLM5Client

        runningWorkflows sync.Map // 正在运行的工作流
        maxConcurrent    int      // 最大并发数
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine() *WorkflowEngine {
        return &WorkflowEngine{
                detector:      detector.NewDetector(),
                executor:      executor.NewExecutor(),
                security:      security.NewSecurityChecker(),
                maxConcurrent: 10,
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

// TriggerFromAlert 从告警触发工作流
func (e *WorkflowEngine) TriggerFromAlert(alert *detector.Alert) (*WorkflowRecord, error) {
        // 创建工作流
        workflow := &WorkflowRecord{
                Type:          WorkflowTypeAutoFix,
                Status:        WorkflowStatusPending,
                TriggerSource: "alert",
                TriggerID:     alert.ID,
                TriggerDetail: fmt.Sprintf("%s: %s", alert.Title, alert.Message),
                ServerID:      alert.ServerID,
                AutoMode:      true,
        }

        alertDetail, _ := json.Marshal(alert)
        workflow.TriggerDetail = string(alertDetail)

        if err := global.DB.Create(workflow).Error; err != nil {
                return nil, err
        }

        // 异步执行
        go e.Execute(workflow)

        return workflow, nil
}

// TriggerFromMetric 从指标异常触发工作流
func (e *WorkflowEngine) TriggerFromMetric(srv *server.Server, metric *server.ServerMetric, results []detector.DetectionResult) (*WorkflowRecord, error) {
        workflow := &WorkflowRecord{
                Type:          WorkflowTypeAutoFix,
                Status:        WorkflowStatusPending,
                TriggerSource: "metric_anomaly",
                ServerID:      srv.ID,
                AutoMode:      true,
        }

        detail, _ := json.Marshal(map[string]interface{}{
                "serverId": srv.ID,
                "serverName": srv.Name,
                "metric": metric,
                "detections": results,
        })
        workflow.TriggerDetail = string(detail)

        if err := global.DB.Create(workflow).Error; err != nil {
                return nil, err
        }

        go e.Execute(workflow)

        return workflow, nil
}

// TriggerManual 手动触发工作流
func (e *WorkflowEngine) TriggerManual(serverID uint, reason string, userID uint) (*WorkflowRecord, error) {
        workflow := &WorkflowRecord{
                Type:          WorkflowTypeManual,
                Status:        WorkflowStatusPending,
                TriggerSource: "manual",
                ServerID:      serverID,
                AutoMode:      false,
                NeedApprove:   false,
        }

        detail, _ := json.Marshal(map[string]interface{}{
                "reason":  reason,
                "userId":  userID,
                "time":    time.Now(),
        })
        workflow.TriggerDetail = string(detail)

        if err := global.DB.Create(workflow).Error; err != nil {
                return nil, err
        }

        go e.Execute(workflow)

        return workflow, nil
}

// Execute 执行工作流
func (e *WorkflowEngine) Execute(workflow *WorkflowRecord) error {
        // 检查是否已经在运行
        if _, exists := e.runningWorkflows.Load(workflow.ID); exists {
                return fmt.Errorf("工作流已在运行中")
        }

        e.runningWorkflows.Store(workflow.ID, true)
        defer e.runningWorkflows.Delete(workflow.ID)

        // 更新状态
        now := time.Now()
        workflow.Status = WorkflowStatusRunning
        workflow.StartedAt = &now
        global.DB.Save(workflow)

        // 获取服务器信息
        var srv server.Server
        if err := global.DB.First(&srv, workflow.ServerID).Error; err != nil {
                return e.failWorkflow(workflow, "服务器不存在")
        }
        workflow.Server = &srv

        // 初始化步骤
        steps := []WorkflowStep{
                {StepID: 1, Name: "数据采集", Type: "detect", Status: "pending"},
                {StepID: 2, Name: "AI分析", Type: "analyze", Status: "pending"},
                {StepID: 3, Name: "决策生成", Type: "decide", Status: "pending"},
                {StepID: 4, Name: "安全检查", Type: "security", Status: "pending"},
                {StepID: 5, Name: "执行操作", Type: "execute", Status: "pending"},
                {StepID: 6, Name: "结果验证", Type: "verify", Status: "pending"},
                {StepID: 7, Name: "记录通知", Type: "notify", Status: "pending"},
        }
        workflow.TotalSteps = len(steps)

        // 执行步骤
        var lastOutput string
        var aiDecision *decision.AIDecision
        var commands []string

        for i := range steps {
                step := &steps[i]
                workflow.CurrentStep = step.StepID

                stepStart := time.Now()
                step.StartedAt = &stepStart
                step.Status = "running"
                e.saveSteps(workflow, steps)

                var err error
                switch step.Type {
                case "detect":
                        lastOutput, err = e.stepDetect(workflow, step, &srv)
                case "analyze":
                        aiDecision, err = e.stepAnalyze(workflow, step, &srv, lastOutput)
                case "decide":
                        commands, err = e.stepDecide(workflow, step, aiDecision)
                case "security":
                        err = e.stepSecurity(workflow, step, commands)
                case "execute":
                        lastOutput, err = e.stepExecute(workflow, step, &srv, commands)
                case "verify":
                        err = e.stepVerify(workflow, step, &srv, lastOutput)
                case "notify":
                        err = e.stepNotify(workflow, step, &srv, lastOutput)
                }

                stepComplete := time.Now()
                step.CompletedAt = &stepComplete
                step.Duration = stepComplete.Sub(*step.StartedAt).Milliseconds()

                if err != nil {
                        step.Status = "failed"
                        step.Error = err.Error()
                        e.saveSteps(workflow, steps)
                        return e.failWorkflow(workflow, fmt.Sprintf("步骤[%s]失败: %s", step.Name, err.Error()))
                }

                step.Status = "success"
                e.saveSteps(workflow, steps)
        }

        // 完成
        return e.completeWorkflow(workflow)
}

// stepDetect 步骤1: 数据采集
func (e *WorkflowEngine) stepDetect(workflow *WorkflowRecord, step *WorkflowStep, srv *server.Server) (string, error) {
        // 获取最新指标
        var metric server.ServerMetric
        if err := global.DB.Where("server_id = ?", srv.ID).Order("created_at DESC").First(&metric).Error; err != nil {
                return "", fmt.Errorf("无法获取指标数据")
        }

        // 执行检测
        processes := []detector.ProcessInfo{} // TODO: 从Agent获取
        containers := []server.DockerContainer{}
        ports := []server.PortInfo{}

        results := e.detector.Detect(srv, &metric, processes, containers, ports)

        // 序列化输出
        output, _ := json.Marshal(map[string]interface{}{
                "metric":    metric,
                "detections": results,
        })
        step.Output = string(output)

        return string(output), nil
}

// stepAnalyze 步骤2: AI分析
func (e *WorkflowEngine) stepAnalyze(workflow *WorkflowRecord, step *WorkflowStep, srv *server.Server, input string) (*decision.AIDecision, error) {
        if e.llmClient == nil {
                return nil, fmt.Errorf("AI客户端未配置")
        }

        // 解析输入
        var inputData struct {
                Metric     server.ServerMetric        `json:"metric"`
                Detections []detector.DetectionResult `json:"detections"`
        }
        json.Unmarshal([]byte(input), &inputData)

        // 调用AI分析
        aiDecision, err := e.decision.QuickAnalyze(srv, &inputData.Metric)
        if err != nil {
                return nil, fmt.Errorf("AI分析失败: %w", err)
        }

        // 保存分析结果
        workflow.AIAnalysis = aiDecision.Analysis
        aiOutput, _ := json.Marshal(aiDecision)
        step.Output = string(aiOutput)

        return aiDecision, nil
}

// stepDecide 步骤3: 决策生成
func (e *WorkflowEngine) stepDecide(workflow *WorkflowRecord, step *WorkflowStep, aiDecision *decision.AIDecision) ([]string, error) {
        if aiDecision == nil {
                return nil, fmt.Errorf("无AI决策")
        }

        // 解析命令
        var commands []string
        if aiDecision.Commands != "" {
                if err := json.Unmarshal([]byte(aiDecision.Commands), &commands); err != nil {
                        return nil, fmt.Errorf("解析命令失败: %w", err)
                }
        }

        // 检查是否需要审批
        if aiDecision.Type == decision.DecisionTypeManual || e.hasDangerousCommands(commands) {
                workflow.NeedApprove = true
        }

        workflow.Commands = aiDecision.Commands
        step.Output = aiDecision.Commands

        return commands, nil
}

// stepSecurity 步骤4: 安全检查
func (e *WorkflowEngine) stepSecurity(workflow *WorkflowRecord, step *WorkflowStep, commands []string) error {
        if len(commands) == 0 {
                return fmt.Errorf("无命令需要执行")
        }

        // 执行安全检查
        result := e.security.ValidateCommands(commands)
        if !result.Allowed {
                return fmt.Errorf("安全检查未通过: %s", result.Message)
        }

        workflow.NeedApprove = result.RequiresApproval
        output, _ := json.Marshal(result)
        step.Output = string(output)

        return nil
}

// stepExecute 步骤5: 执行操作
func (e *WorkflowEngine) stepExecute(workflow *WorkflowRecord, step *WorkflowStep, srv *server.Server, commands []string) (string, error) {
        // 检查是否需要审批
        if workflow.NeedApprove && workflow.ApprovedBy == 0 {
                // 等待审批
                return "", fmt.Errorf("需要人工审批")
        }

        // TODO: 实际执行命令（通过SSH或Agent）
        // 这里模拟执行
        var outputs []string
        for _, cmd := range commands {
                outputs = append(outputs, fmt.Sprintf("[执行] %s", cmd))
        }

        output := strings.Join(outputs, "\n")
        step.Output = output
        workflow.Output = output

        return output, nil
}

// stepVerify 步骤6: 结果验证
func (e *WorkflowEngine) stepVerify(workflow *WorkflowRecord, step *WorkflowStep, srv *server.Server, output string) error {
        // 获取最新指标验证修复效果
        var metric server.ServerMetric
        if err := global.DB.Where("server_id = ?", srv.ID).Order("created_at DESC").First(&metric).Error; err != nil {
                step.Output = "无法验证"
                return nil
        }

        // 检查是否改善
        improved := true
        message := "验证通过，问题已解决"
        
        // 简单判断（实际应该根据具体问题类型判断）
        if metric.CPUUsage > 90 || metric.MemoryUsage > 90 || metric.DiskUsage > 90 {
                improved = false
                message = "问题可能未完全解决，建议持续监控"
        }

        step.Output = message
        workflow.Result = message

        if !improved {
                // 可以触发重试或升级处理
        }

        return nil
}

// stepNotify 步骤7: 记录通知
func (e *WorkflowEngine) stepNotify(workflow *WorkflowRecord, step *WorkflowStep, srv *server.Server, output string) error {
        // 记录日志
        log := server.ServerLog{
                ServerID: srv.ID,
                Type:     "workflow",
                Content:  fmt.Sprintf("工作流执行完成: %s", workflow.Type),
                Output:   output,
        }
        global.DB.Create(&log)

        // 发送通知
        if e.notifier != nil {
                e.notifier.SendMessage(
                        "自动修复完成",
                        fmt.Sprintf("服务器 %s 自动修复工作流执行完成\n%s", srv.Name, workflow.Result),
                )
        }

        step.Output = "通知已发送"
        return nil
}

// Helper methods
func (e *WorkflowEngine) saveSteps(workflow *WorkflowRecord, steps []WorkflowStep) {
        stepsJSON, _ := json.Marshal(steps)
        workflow.Steps = string(stepsJSON)
        global.DB.Save(workflow)
}

func (e *WorkflowEngine) failWorkflow(workflow *WorkflowRecord, reason string) error {
        now := time.Now()
        workflow.Status = WorkflowStatusFailed
        workflow.CompletedAt = &now
        workflow.Result = reason
        workflow.Duration = now.Sub(*workflow.StartedAt).Milliseconds()
        global.DB.Save(workflow)

        // 发送失败通知
        if e.notifier != nil {
                e.notifier.SendMessage("工作流执行失败", reason)
        }

        return fmt.Errorf(reason)
}

func (e *WorkflowEngine) completeWorkflow(workflow *WorkflowRecord) error {
        now := time.Now()
        workflow.Status = WorkflowStatusCompleted
        workflow.CompletedAt = &now
        workflow.Duration = now.Sub(*workflow.StartedAt).Milliseconds()
        return global.DB.Save(workflow).Error
}

func (e *WorkflowEngine) hasDangerousCommands(commands []string) bool {
        result := e.security.ValidateCommands(commands)
        return result.SecurityLevel == security.SecurityLevelDangerous ||
                result.SecurityLevel == security.SecurityLevelForbidden
}

// ApproveWorkflow 审批工作流
func (e *WorkflowEngine) ApproveWorkflow(workflowID, userID uint) error {
        var workflow WorkflowRecord
        if err := global.DB.First(&workflow, workflowID).Error; err != nil {
                return err
        }

        if !workflow.NeedApprove {
                return fmt.Errorf("该工作流不需要审批")
        }

        now := time.Now()
        workflow.ApprovedBy = userID
        workflow.ApprovedAt = &now
        global.DB.Save(&workflow)

        // 继续执行
        go e.Execute(&workflow)

        return nil
}

// CancelWorkflow 取消工作流
func (e *WorkflowEngine) CancelWorkflow(workflowID uint) error {
        var workflow WorkflowRecord
        if err := global.DB.First(&workflow, workflowID).Error; err != nil {
                return err
        }

        if workflow.Status == WorkflowStatusCompleted {
                return fmt.Errorf("工作流已完成，无法取消")
        }

        workflow.Status = WorkflowStatusCancelled
        now := time.Now()
        workflow.CompletedAt = &now
        return global.DB.Save(&workflow).Error
}

// GetWorkflowHistory 获取工作流历史
func (e *WorkflowEngine) GetWorkflowHistory(serverID uint, limit int) ([]WorkflowRecord, error) {
        var records []WorkflowRecord
        query := global.DB.Model(&WorkflowRecord{}).Order("created_at DESC")
        if serverID > 0 {
                query = query.Where("server_id = ?", serverID)
        }
        if limit > 0 {
                query = query.Limit(limit)
        }
        err := query.Find(&records).Error
        return records, err
}

// GetStatistics 获取统计
func (e *WorkflowEngine) GetStatistics(days int) map[string]int64 {
        stats := make(map[string]int64)
        since := time.Now().AddDate(0, 0, -days)

        global.DB.Model(&WorkflowRecord{}).Where("created_at > ?", since).Count(&stats["total"])
        global.DB.Model(&WorkflowRecord{}).Where("created_at > ? AND status = ?", since, WorkflowStatusCompleted).Count(&stats["completed"])
        global.DB.Model(&WorkflowRecord{}).Where("created_at > ? AND status = ?", since, WorkflowStatusFailed).Count(&stats["failed"])
        global.DB.Model(&WorkflowRecord{}).Where("created_at > ? AND auto_mode = ?", since, true).Count(&stats["autoExecuted"])

        return stats
}
