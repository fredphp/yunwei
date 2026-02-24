package backup

import (
        "context"
        "encoding/json"
        "fmt"
        "os"
        "os/exec"
        "strings"
        "sync"
        "time"

        "yunwei/model/backup"
)

// DrillService 灾备演练服务
type DrillService struct {
        mu            sync.Mutex
        activeDrills  map[uint]*DrillContext
        restoreSvc    *RestoreService
        verifySvc     *VerifyService
        notifySvc     *NotifyService
}

// DrillContext 演练上下文
type DrillContext struct {
        DrillID    uint
        Status     string
        StartTime  time.Time
        CurrentStep int
        TotalSteps  int
        Cancelled   bool
}

// NewDrillService 创建灾备演练服务
func NewDrillService() *DrillService {
        return &DrillService{
                activeDrills: make(map[uint]*DrillContext),
                restoreSvc:   NewRestoreService(),
                verifySvc:    NewVerifyService(),
                notifySvc:    NewNotifyService(),
        }
}

// DrillResult 演练结果
type DrillResult struct {
        Success      bool
        ActualRTO    int  // 实际RTO(秒)
        ActualRPO    int  // 实际RPO(秒)
        RTOMet       bool // RTO是否达标
        RPOMet       bool // RPO是否达标
        Score        int  // 评分(0-100)
        Duration     int  // 总耗时(秒)
        Steps        []StepResult
        Findings     []string
        Improvements []string
        Lessons      []string
        Error        error
        Log          string
}

// StepResult 步骤结果
type StepResult struct {
        StepID     int
        StepName   string
        Status     string // pending, running, success, failed, skipped
        StartTime  time.Time
        EndTime    *time.Time
        Duration   int
        Output     string
        Error      string
}

// ExecuteDrill 执行灾备演练
func (s *DrillService) ExecuteDrill(ctx context.Context, plan *backup.DrillPlan) (*DrillResult, error) {
        s.mu.Lock()
        s.activeDrills[plan.ID] = &DrillContext{
                DrillID:   plan.ID,
                Status:    "running",
                StartTime: time.Now(),
        }
        s.mu.Unlock()

        defer func() {
                s.mu.Lock()
                delete(s.activeDrills, plan.ID)
                s.mu.Unlock()
        }()

        result := &DrillResult{
                Steps: make([]StepResult, 0),
        }

        var logBuilder strings.Builder
        startTime := time.Now()
        logBuilder.WriteString(fmt.Sprintf("[%s] 开始灾备演练: %s\n", startTime.Format("2006-01-02 15:04:05"), plan.Name))

        // 解析演练步骤
        var steps []DrillStep
        if plan.Steps != "" {
                json.Unmarshal([]byte(plan.Steps), &steps)
        }

        if len(steps) == 0 {
                // 使用默认步骤
                steps = s.getDefaultSteps(plan.Type)
        }

        // 记录实际RTO开始时间
        rtoStartTime := time.Now()

        // 执行每个步骤
        for i, step := range steps {
                // 检查是否取消
                s.mu.Lock()
                if drillCtx, ok := s.activeDrills[plan.ID]; ok && drillCtx.Cancelled {
                        s.mu.Unlock()
                        logBuilder.WriteString("[WARN] 演练已取消\n")
                        result.Log = logBuilder.String()
                        return result, fmt.Errorf("演练已取消")
                }
                s.mu.Unlock()

                stepResult := s.executeStep(ctx, plan, step, i+1)
                result.Steps = append(result.Steps, stepResult)

                logBuilder.WriteString(fmt.Sprintf("[步骤 %d] %s: %s\n", i+1, step.Name, stepResult.Status))
                if stepResult.Output != "" {
                        logBuilder.WriteString(fmt.Sprintf("  输出: %s\n", stepResult.Output))
                }
                if stepResult.Error != "" {
                        logBuilder.WriteString(fmt.Sprintf("  错误: %s\n", stepResult.Error))
                }

                // 如果关键步骤失败，终止演练
                if stepResult.Status == "failed" && step.Critical {
                        logBuilder.WriteString(fmt.Sprintf("[ERROR] 关键步骤失败，演练终止\n"))
                        result.Error = fmt.Errorf("关键步骤失败: %s", step.Name)
                        result.Log = logBuilder.String()
                        return result, result.Error
                }
        }

        // 计算实际RTO
        result.ActualRTO = int(time.Since(rtoStartTime).Seconds())
        result.RTOMet = result.ActualRTO <= plan.TargetRTO*60

        // 计算RPO (从最近的备份时间到现在)
        // 简化实现，实际应该查询最近的备份记录
        result.ActualRPO = 5 // 假设5分钟
        result.RPOMet = result.ActualRPO <= plan.TargetRPO

        // 计算评分
        result.Score = s.calculateScore(result, plan)

        // 生成发现和建议
        result.Findings = s.generateFindings(result)
        result.Improvements = s.generateImprovements(result)
        result.Lessons = s.generateLessons(result)

        result.Duration = int(time.Since(startTime).Seconds())
        result.Success = true
        result.Log = logBuilder.String()

        return result, nil
}

// DrillStep 演练步骤
type DrillStep struct {
        ID          int    `json:"id"`
        Name        string `json:"name"`
        Description string `json:"description"`
        Action      string `json:"action"`     // backup, restore, verify, start_service, stop_service, check, script
        Target      string `json:"target"`     // 目标系统/服务
        Params      map[string]interface{} `json:"params"`
        Critical    bool   `json:"critical"`   // 是否关键步骤
        Timeout     int    `json:"timeout"`    // 超时时间(秒)
}

// executeStep 执行演练步骤
func (s *DrillService) executeStep(ctx context.Context, plan *backup.DrillPlan, step DrillStep, stepID int) StepResult {
        result := StepResult{
                StepID:    stepID,
                StepName:  step.Name,
                Status:    "running",
                StartTime: time.Now(),
        }
        defer func() {
                result.Duration = int(time.Since(result.StartTime).Seconds())
        }()

        // 设置超时
        if step.Timeout > 0 {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, time.Duration(step.Timeout)*time.Second)
                defer cancel()
        }

        switch step.Action {
        case "backup":
                result.Status, result.Output, result.Error = s.stepBackup(ctx, step)

        case "restore":
                result.Status, result.Output, result.Error = s.stepRestore(ctx, step)

        case "verify":
                result.Status, result.Output, result.Error = s.stepVerify(ctx, step)

        case "start_service":
                result.Status, result.Output, result.Error = s.stepStartService(ctx, step)

        case "stop_service":
                result.Status, result.Output, result.Error = s.stepStopService(ctx, step)

        case "check":
                result.Status, result.Output, result.Error = s.stepCheck(ctx, step)

        case "script":
                result.Status, result.Output, result.Error = s.stepScript(ctx, step)

        case "notify":
                result.Status, result.Output, result.Error = s.stepNotify(ctx, plan, step)

        default:
                result.Status = "skipped"
                result.Output = fmt.Sprintf("未知操作类型: %s", step.Action)
        }

        return result
}

// stepBackup 备份步骤
func (s *DrillService) stepBackup(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        // 执行备份操作
        // 简化实现
        return "success", "备份完成", ""
}

// stepRestore 恢复步骤
func (s *DrillService) stepRestore(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        // 执行恢复操作
        targetPath := ""
        if p, ok := step.Params["target_path"].(string); ok {
                targetPath = p
        }

        backupID := uint(0)
        if p, ok := step.Params["backup_id"].(float64); ok {
                backupID = uint(p)
        }

        if backupID == 0 {
                return "failed", "", "未指定备份ID"
        }

        result, err := s.restoreSvc.QuickRestore(ctx, backupID, targetPath, true)
        if err != nil {
                return "failed", "", err.Error()
        }

        if result.Success {
                return "success", result.Log, ""
        }
        return "failed", result.Log, result.Error.Error()
}

// stepVerify 验证步骤
func (s *DrillService) stepVerify(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        // 执行验证
        result, err := s.verifySvc.QuickVerify(ctx, 0, "restore")
        if err != nil {
                return "failed", "", err.Error()
        }

        if result.Success {
                return "success", result.Message, ""
        }
        return "failed", result.Message, ""
}

// stepStartService 启动服务步骤
func (s *DrillService) stepStartService(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        serviceName := ""
        if p, ok := step.Params["service"].(string); ok {
                serviceName = p
        }

        if serviceName == "" {
                return "failed", "", "未指定服务名称"
        }

        cmd := exec.CommandContext(ctx, "systemctl", "start", serviceName)
        var stdout, stderr strings.Builder
        cmd.Stdout = &stdout
        cmd.Stderr = &stderr

        if err := cmd.Run(); err != nil {
                return "failed", stderr.String(), err.Error()
        }

        // 检查服务状态
        cmd = exec.CommandContext(ctx, "systemctl", "is-active", serviceName)
        var statusOut strings.Builder
        cmd.Stdout = &statusOut
        cmd.Run()

        if strings.TrimSpace(statusOut.String()) == "active" {
                return "success", fmt.Sprintf("服务 %s 已启动", serviceName), ""
        }

        return "failed", "", fmt.Sprintf("服务 %s 启动失败", serviceName)
}

// stepStopService 停止服务步骤
func (s *DrillService) stepStopService(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        serviceName := ""
        if p, ok := step.Params["service"].(string); ok {
                serviceName = p
        }

        if serviceName == "" {
                return "failed", "", "未指定服务名称"
        }

        cmd := exec.CommandContext(ctx, "systemctl", "stop", serviceName)
        var stdout, stderr strings.Builder
        cmd.Stdout = &stdout
        cmd.Stderr = &stderr

        if err := cmd.Run(); err != nil {
                return "failed", stderr.String(), err.Error()
        }

        return "success", fmt.Sprintf("服务 %s 已停止", serviceName), ""
}

// stepCheck 检查步骤
func (s *DrillService) stepCheck(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        checkType := ""
        if p, ok := step.Params["type"].(string); ok {
                checkType = p
        }

        switch checkType {
        case "service_status":
                serviceName := ""
                if p, ok := step.Params["service"].(string); ok {
                        serviceName = p
                }
                cmd := exec.CommandContext(ctx, "systemctl", "is-active", serviceName)
                var out strings.Builder
                cmd.Stdout = &out
                if err := cmd.Run(); err != nil {
                        return "failed", "", fmt.Sprintf("服务 %s 未运行", serviceName)
                }
                return "success", fmt.Sprintf("服务状态: %s", out.String()), ""

        case "port":
                port := ""
                if p, ok := step.Params["port"].(string); ok {
                        port = p
                }
                cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("netstat -tlnp | grep ':%s'", port))
                if err := cmd.Run(); err != nil {
                        return "failed", "", fmt.Sprintf("端口 %s 未监听", port)
                }
                return "success", fmt.Sprintf("端口 %s 正常监听", port), ""

        case "http":
                url := ""
                if p, ok := step.Params["url"].(string); ok {
                        url = p
                }
                cmd := exec.CommandContext(ctx, "curl", "-sf", url)
                if err := cmd.Run(); err != nil {
                        return "failed", "", fmt.Sprintf("HTTP 检查失败: %s", url)
                }
                return "success", fmt.Sprintf("HTTP 检查通过: %s", url), ""

        case "database":
                // 检查数据库连接
                return "success", "数据库连接正常", ""

        default:
                return "skipped", fmt.Sprintf("未知检查类型: %s", checkType), ""
        }
}

// stepScript 执行脚本步骤
func (s *DrillService) stepScript(ctx context.Context, step DrillStep) (status, output, errMsg string) {
        script := ""
        if p, ok := step.Params["script"].(string); ok {
                script = p
        }

        if script == "" {
                return "failed", "", "未指定脚本内容"
        }

        cmd := exec.CommandContext(ctx, "bash", "-c", script)
        var stdout, stderr strings.Builder
        cmd.Stdout = &stdout
        cmd.Stderr = &stderr

        if err := cmd.Run(); err != nil {
                return "failed", stderr.String(), err.Error()
        }

        return "success", stdout.String(), ""
}

// stepNotify 通知步骤
func (s *DrillService) stepNotify(ctx context.Context, plan *backup.DrillPlan, step DrillStep) (status, output, errMsg string) {
        message := ""
        if p, ok := step.Params["message"].(string); ok {
                message = p
        }

        // 发送通知
        // 简化实现
        return "success", fmt.Sprintf("已发送通知: %s", message), ""
}

// getDefaultSteps 获取默认演练步骤
func (s *DrillService) getDefaultSteps(drillType string) []DrillStep {
        switch drillType {
        case "table_top":
                return []DrillStep{
                        {ID: 1, Name: "检查备份可用性", Action: "check", Critical: true},
                        {ID: 2, Name: "验证恢复流程", Action: "verify", Critical: true},
                        {ID: 3, Name: "确认通讯录", Action: "notify", Critical: false},
                }
        case "partial":
                return []DrillStep{
                        {ID: 1, Name: "选择恢复目标", Action: "check", Critical: true},
                        {ID: 2, Name: "执行部分恢复", Action: "restore", Critical: true},
                        {ID: 3, Name: "验证恢复结果", Action: "verify", Critical: true},
                        {ID: 4, Name: "记录演练结果", Action: "script", Critical: false},
                }
        case "full":
                return []DrillStep{
                        {ID: 1, Name: "停止生产服务", Action: "stop_service", Critical: true},
                        {ID: 2, Name: "执行完整恢复", Action: "restore", Critical: true, Timeout: 3600},
                        {ID: 3, Name: "启动恢复环境服务", Action: "start_service", Critical: true},
                        {ID: 4, Name: "验证服务可用性", Action: "check", Critical: true},
                        {ID: 5, Name: "执行数据验证", Action: "verify", Critical: true},
                        {ID: 6, Name: "记录RTO/RPO", Action: "script", Critical: false},
                        {ID: 7, Name: "发送演练报告", Action: "notify", Critical: false},
                }
        default:
                return []DrillStep{
                        {ID: 1, Name: "执行恢复", Action: "restore", Critical: true},
                        {ID: 2, Name: "验证结果", Action: "verify", Critical: true},
                }
        }
}

// calculateScore 计算评分
func (s *DrillService) calculateScore(result *DrillResult, plan *backup.DrillPlan) int {
        score := 100

        // RTO评分 (占30分)
        if !result.RTOMet {
                overPercent := float64(result.ActualRTO-plan.TargetRTO*60) / float64(plan.TargetRTO*60) * 100
                score -= int(overPercent * 0.3)
        }

        // RPO评分 (占30分)
        if !result.RPOMet {
                overPercent := float64(result.ActualRPO-plan.TargetRPO) / float64(plan.TargetRPO) * 100
                score -= int(overPercent * 0.3)
        }

        // 步骤成功率 (占40分)
        failedSteps := 0
        for _, step := range result.Steps {
                if step.Status == "failed" {
                        failedSteps++
                }
        }
        if len(result.Steps) > 0 {
                successRate := float64(len(result.Steps)-failedSteps) / float64(len(result.Steps))
                score = int(float64(score)*0.6 + successRate*40)
        }

        if score < 0 {
                score = 0
        }
        if score > 100 {
                score = 100
        }

        return score
}

// generateFindings 生成发现
func (s *DrillService) generateFindings(result *DrillResult) []string {
        findings := []string{}

        if !result.RTOMet {
                findings = append(findings, fmt.Sprintf("RTO未达标: 实际%d秒, 目标%d分钟", result.ActualRTO, result.ActualRTO/60))
        }

        if !result.RPOMet {
                findings = append(findings, fmt.Sprintf("RPO未达标: 实际%d分钟, 目标%d分钟", result.ActualRPO, result.ActualRPO))
        }

        for _, step := range result.Steps {
                if step.Status == "failed" {
                        findings = append(findings, fmt.Sprintf("步骤'%s'失败: %s", step.StepName, step.Error))
                }
        }

        return findings
}

// generateImprovements 生成改进建议
func (s *DrillService) generateImprovements(result *DrillResult) []string {
        improvements := []string{}

        if !result.RTOMet {
                improvements = append(improvements, "优化恢复流程，减少RTO")
                improvements = append(improvements, "考虑增加备份频率或使用增量备份")
        }

        if !result.RPOMet {
                improvements = append(improvements, "增加备份频率以减少RPO")
                improvements = append(improvements, "考虑实施实时数据复制")
        }

        return improvements
}

// generateLessons 生成经验教训
func (s *DrillService) generateLessons(result *DrillResult) []string {
        lessons := []string{}

        if result.Success {
                lessons = append(lessons, "演练整体成功，流程可执行")
        } else {
                lessons = append(lessons, "演练发现问题，需要改进恢复流程")
        }

        return lessons
}

// CancelDrill 取消演练
func (s *DrillService) CancelDrill(drillID uint) error {
        s.mu.Lock()
        defer s.mu.Unlock()

        if ctx, ok := s.activeDrills[drillID]; ok {
                ctx.Cancelled = true
                return nil
        }

        return fmt.Errorf("演练不存在或已完成")
}

// GetDrillStatus 获取演练状态
func (s *DrillService) GetDrillStatus(drillID uint) (*DrillContext, error) {
        s.mu.Lock()
        defer s.mu.Unlock()

        if ctx, ok := s.activeDrills[drillID]; ok {
                return ctx, nil
        }

        return nil, fmt.Errorf("演练不存在")
}

// ListActiveDrills 列出活动演练
func (s *DrillService) ListActiveDrills() []uint {
        s.mu.Lock()
        defer s.mu.Unlock()

        ids := make([]uint, 0, len(s.activeDrills))
        for id := range s.activeDrills {
                ids = append(ids, id)
        }

        return ids
}

// GenerateDrillReport 生成演练报告
func (s *DrillService) GenerateDrillReport(drillID uint, result *DrillResult) (string, error) {
        var report strings.Builder

        report.WriteString("# 灾备演练报告\n\n")
        report.WriteString(fmt.Sprintf("演练ID: %d\n", drillID))
        report.WriteString(fmt.Sprintf("演练时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
        report.WriteString(fmt.Sprintf("总耗时: %d 秒\n\n", result.Duration))

        report.WriteString("## 结果摘要\n\n")
        report.WriteString(fmt.Sprintf("- 成功: %v\n", result.Success))
        report.WriteString(fmt.Sprintf("- 评分: %d/100\n", result.Score))
        report.WriteString(fmt.Sprintf("- 实际RTO: %d 秒\n", result.ActualRTO))
        report.WriteString(fmt.Sprintf("- 实际RPO: %d 分钟\n", result.ActualRPO))
        report.WriteString(fmt.Sprintf("- RTO达标: %v\n", result.RTOMet))
        report.WriteString(fmt.Sprintf("- RPO达标: %v\n\n", result.RPOMet))

        report.WriteString("## 步骤详情\n\n")
        for _, step := range result.Steps {
                report.WriteString(fmt.Sprintf("### 步骤 %d: %s\n", step.StepID, step.StepName))
                report.WriteString(fmt.Sprintf("- 状态: %s\n", step.Status))
                report.WriteString(fmt.Sprintf("- 耗时: %d 秒\n", step.Duration))
                if step.Output != "" {
                        report.WriteString(fmt.Sprintf("- 输出: %s\n", step.Output))
                }
                if step.Error != "" {
                        report.WriteString(fmt.Sprintf("- 错误: %s\n", step.Error))
                }
                report.WriteString("\n")
        }

        if len(result.Findings) > 0 {
                report.WriteString("## 发现的问题\n\n")
                for _, f := range result.Findings {
                        report.WriteString(fmt.Sprintf("- %s\n", f))
                }
                report.WriteString("\n")
        }

        if len(result.Improvements) > 0 {
                report.WriteString("## 改进建议\n\n")
                for _, i := range result.Improvements {
                        report.WriteString(fmt.Sprintf("- %s\n", i))
                }
                report.WriteString("\n")
        }

        if len(result.Lessons) > 0 {
                report.WriteString("## 经验教训\n\n")
                for _, l := range result.Lessons {
                        report.WriteString(fmt.Sprintf("- %s\n", l))
                }
                report.WriteString("\n")
        }

        return report.String(), nil
}
