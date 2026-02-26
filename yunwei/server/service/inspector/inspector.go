package inspector

import (
	"encoding/json"
	"fmt"
	"time"

	"yunwei/global"
	"yunwei/model/server"
	"yunwei/service/ai/llm"
	"yunwei/service/notifier"
)

// InspectionType 巡检类型
type InspectionType string

const (
	InspectionDaily   InspectionType = "daily"
	InspectionWeekly  InspectionType = "weekly"
	InspectionMonthly InspectionType = "monthly"
	InspectionManual  InspectionType = "manual"
)

// InspectionStatus 巡检状态
type InspectionStatus string

const (
	InspectionStatusRunning   InspectionStatus = "running"
	InspectionStatusCompleted InspectionStatus = "completed"
	InspectionStatusFailed    InspectionStatus = "failed"
)

// InspectionReport 巡检报告
type InspectionReport struct {
	ID           uint            `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time       `json:"createdAt"`
	Type         InspectionType  `json:"type" gorm:"type:varchar(16)"`
	Status       InspectionStatus `json:"status" gorm:"type:varchar(16)"`

	// 统计信息
	TotalServers    int `json:"totalServers"`
	OnlineServers   int `json:"onlineServers"`
	OfflineServers  int `json:"offlineServers"`
	WarningServers  int `json:"warningServers"`
	CriticalServers int `json:"criticalServers"`

	// 检查项统计
	CPUWarnings    int `json:"cpuWarnings"`
	MemoryWarnings int `json:"memoryWarnings"`
	DiskWarnings   int `json:"diskWarnings"`
	ServiceIssues  int `json:"serviceIssues"`

	// 详细信息
	ServerDetails  string `json:"serverDetails" gorm:"type:text"` // JSON
	Issues         string `json:"issues" gorm:"type:text"`        // JSON
	Recommendations string `json:"recommendations" gorm:"type:text"`

	// AI 分析
	AIAnalysis     string `json:"aiAnalysis" gorm:"type:text"`

	// 通知状态
	Notified       bool   `json:"notified"`
	NotifyChannels string `json:"notifyChannels" gorm:"type:varchar(255)"`
}

func (InspectionReport) TableName() string {
	return "inspection_reports"
}

// ServerInspectionResult 服务器巡检结果
type ServerInspectionResult struct {
	ServerID      uint      `json:"serverId"`
	ServerName    string    `json:"serverName"`
	Host          string    `json:"host"`
	Status        string    `json:"status"` // online, offline, warning, critical

	// 指标
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
	DiskUsage     float64 `json:"diskUsage"`
	Load1         float64 `json:"load1"`

	// 服务状态
	NginxStatus   string `json:"nginxStatus"`
	DockerStatus  string `json:"dockerStatus"`
	RedisStatus   string `json:"redisStatus"`
	MySQLStatus   string `json:"mysqlStatus"`

	// 问题列表
	Issues        []string `json:"issues"`

	// 最后检查时间
	LastCheck     time.Time `json:"lastCheck"`
}

// Inspector 巡检机器人
type Inspector struct {
	llmClient  *llm.GLM5Client
	notifier   *notifier.NotifierService
}

// NewInspector 创建巡检机器人
func NewInspector(llmClient *llm.GLM5Client) *Inspector {
	return &Inspector{
		llmClient: llmClient,
		notifier:  notifier.NewNotifierService(),
	}
}

// RunDailyInspection 执行每日巡检
func (i *Inspector) RunDailyInspection() (*InspectionReport, error) {
	report := &InspectionReport{
		Type:      InspectionDaily,
		Status:    InspectionStatusRunning,
		CreatedAt: time.Now(),
	}

	// 保存报告
	global.DB.Create(report)

	// 获取所有服务器
	var servers []server.Server
	global.DB.Find(&servers)

	report.TotalServers = len(servers)

	var results []ServerInspectionResult
	var allIssues []string

	for _, srv := range servers {
		result := i.inspectServer(&srv)
		results = append(results, result)

		// 统计
		switch result.Status {
		case "online":
			report.OnlineServers++
		case "offline":
			report.OfflineServers++
		case "warning":
			report.WarningServers++
		case "critical":
			report.CriticalServers++
		}

		// 统计告警
		if result.CPUUsage > 80 {
			report.CPUWarnings++
		}
		if result.MemoryUsage > 80 {
			report.MemoryWarnings++
		}
		if result.DiskUsage > 80 {
			report.DiskWarnings++
		}

		// 收集问题
		for _, issue := range result.Issues {
			allIssues = append(allIssues, fmt.Sprintf("[%s] %s", srv.Name, issue))
		}
	}

	// 保存详情
	detailsJSON, _ := json.Marshal(results)
	report.ServerDetails = string(detailsJSON)

	issuesJSON, _ := json.Marshal(allIssues)
	report.Issues = string(issuesJSON)

	// AI 分析
	if i.llmClient != nil {
		report.AIAnalysis = i.aiAnalyze(report, results)
	}

	// 生成建议
	report.Recommendations = i.generateRecommendations(report)

	// 更新状态
	report.Status = InspectionStatusCompleted
	global.DB.Save(report)

	// 发送通知
	go i.sendNotification(report)

	return report, nil
}

// inspectServer 检查单个服务器
func (i *Inspector) inspectServer(srv *server.Server) ServerInspectionResult {
	result := ServerInspectionResult{
		ServerID:   srv.ID,
		ServerName: srv.Name,
		Host:       srv.Host,
		Issues:     []string{},
		LastCheck:  time.Now(),
	}

	// 检查在线状态
	if !srv.AgentOnline {
		result.Status = "offline"
		result.Issues = append(result.Issues, "服务器离线")
		return result
	}

	result.Status = "online"

	// 获取最新指标
	var metric server.ServerMetric
	if err := global.DB.Where("server_id = ?", srv.ID).Order("created_at DESC").First(&metric).Error; err == nil {
		result.CPUUsage = metric.CPUUsage
		result.MemoryUsage = metric.MemoryUsage
		result.DiskUsage = metric.DiskUsage
		result.Load1 = metric.Load1

		// CPU 检查
		if metric.CPUUsage > 90 {
			result.Status = "critical"
			result.Issues = append(result.Issues, fmt.Sprintf("CPU使用率过高: %.1f%%", metric.CPUUsage))
		} else if metric.CPUUsage > 80 {
			if result.Status == "online" {
				result.Status = "warning"
			}
			result.Issues = append(result.Issues, fmt.Sprintf("CPU使用率警告: %.1f%%", metric.CPUUsage))
		}

		// 内存检查
		if metric.MemoryUsage > 90 {
			result.Status = "critical"
			result.Issues = append(result.Issues, fmt.Sprintf("内存使用率过高: %.1f%%", metric.MemoryUsage))
		} else if metric.MemoryUsage > 80 {
			if result.Status == "online" {
				result.Status = "warning"
			}
			result.Issues = append(result.Issues, fmt.Sprintf("内存使用率警告: %.1f%%", metric.MemoryUsage))
		}

		// 磁盘检查
		if metric.DiskUsage > 90 {
			result.Status = "critical"
			result.Issues = append(result.Issues, fmt.Sprintf("磁盘空间不足: %.1f%%", metric.DiskUsage))
		} else if metric.DiskUsage > 80 {
			if result.Status == "online" {
				result.Status = "warning"
			}
			result.Issues = append(result.Issues, fmt.Sprintf("磁盘空间警告: %.1f%%", metric.DiskUsage))
		}

		// 负载检查
		if metric.Load1 > float64(srv.CPUCores)*2 {
			if result.Status == "online" {
				result.Status = "warning"
			}
			result.Issues = append(result.Issues, fmt.Sprintf("系统负载过高: %.2f", metric.Load1))
		}
	}

	// 检查服务状态
	result.NginxStatus = i.checkService(srv, "nginx")
	result.DockerStatus = i.checkService(srv, "docker")
	result.RedisStatus = i.checkService(srv, "redis")
	result.MySQLStatus = i.checkService(srv, "mysql")

	// 服务异常
	if result.NginxStatus == "stopped" {
		result.Issues = append(result.Issues, "Nginx服务已停止")
		report.ServiceIssues++
	}
	if result.DockerStatus == "stopped" {
		result.Issues = append(result.Issues, "Docker服务已停止")
		report.ServiceIssues++
	}
	if result.RedisStatus == "stopped" {
		result.Issues = append(result.Issues, "Redis服务已停止")
		report.ServiceIssues++
	}
	if result.MySQLStatus == "stopped" {
		result.Issues = append(result.Issues, "MySQL服务已停止")
		report.ServiceIssues++
	}

	return result
}

// checkService 检查服务状态
func (i *Inspector) checkService(srv *server.Server, serviceName string) string {
	// TODO: 通过SSH或Agent检查服务状态
	// 这里返回模拟状态
	return "running"
}

// aiAnalyze AI分析巡检结果
func (i *Inspector) aiAnalyze(report *InspectionReport, results []ServerInspectionResult) string {
	if i.llmClient == nil {
		return ""
	}

	// 构建分析提示
	var summary string
	summary += fmt.Sprintf("总服务器: %d, 在线: %d, 离线: %d, 告警: %d, 严重: %d\n",
		report.TotalServers, report.OnlineServers, report.OfflineServers,
		report.WarningServers, report.CriticalServers)
	summary += fmt.Sprintf("CPU告警: %d, 内存告警: %d, 磁盘告警: %d, 服务异常: %d\n",
		report.CPUWarnings, report.MemoryWarnings, report.DiskWarnings, report.ServiceIssues)

	// 严重问题列表
	var criticalIssues []string
	for _, r := range results {
		if r.Status == "critical" {
			criticalIssues = append(criticalIssues, fmt.Sprintf("%s: %v", r.ServerName, r.Issues))
		}
	}

	prompt := fmt.Sprintf(`作为运维专家，分析今日服务器巡检报告：

%s

严重问题:
%s

请提供:
1. 问题优先级排序
2. 紧急处理建议
3. 长期优化建议

简要回复(300字以内)。`, summary, criticalIssues)

	response, err := i.llmClient.QuickChat(prompt)
	if err != nil {
		return ""
	}

	return response
}

// generateRecommendations 生成建议
func (i *Inspector) generateRecommendations(report *InspectionReport) string {
	var recommendations []string

	if report.CriticalServers > 0 {
		recommendations = append(recommendations, fmt.Sprintf("⚠️ 有%d台服务器处于严重状态，需立即处理", report.CriticalServers))
	}

	if report.OfflineServers > 0 {
		recommendations = append(recommendations, fmt.Sprintf("🔌 有%d台服务器离线，请检查网络连接", report.OfflineServers))
	}

	if report.CPUWarnings > 3 {
		recommendations = append(recommendations, "📈 多台服务器CPU使用率过高，建议评估扩容需求")
	}

	if report.MemoryWarnings > 3 {
		recommendations = append(recommendations, "💾 多台服务器内存不足，建议增加内存或优化应用")
	}

	if report.DiskWarnings > 0 {
		recommendations = append(recommendations, "💿 有磁盘空间不足警告，建议执行清理任务")
	}

	if report.ServiceIssues > 0 {
		recommendations = append(recommendations, "🔧 有服务异常，建议检查服务状态和日志")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ 所有服务器运行正常，无异常")
	}

	// JSON格式返回
	jsonStr, _ := json.Marshal(recommendations)
	return string(jsonStr)
}

// sendNotification 发送通知
func (i *Inspector) sendNotification(report *InspectionReport) {
	if i.notifier == nil {
		return
	}

	// 构建通知内容
	title := fmt.Sprintf("📊 服务器每日巡检报告 - %s", report.CreatedAt.Format("2006-01-02"))
	content := i.formatReport(report)

	// 发送到各渠道
	var channels []string

	// Telegram
	if err := i.notifier.SendTelegram(title, content); err == nil {
		channels = append(channels, "telegram")
	}

	// 企业微信
	if err := i.notifier.SendWechat(title, content); err == nil {
		channels = append(channels, "wechat")
	}

	// 更新通知状态
	report.Notified = true
	channelsJSON, _ := json.Marshal(channels)
	report.NotifyChannels = string(channelsJSON)
	global.DB.Save(report)
}

// formatReport 格式化报告
func (i *Inspector) formatReport(report *InspectionReport) string {
	var content string

	content += fmt.Sprintf("📅 巡检时间: %s\n\n", report.CreatedAt.Format("2006-01-02 15:04:05"))

	content += "📊 服务器概览:\n"
	content += fmt.Sprintf("  • 总数: %d\n", report.TotalServers)
	content += fmt.Sprintf("  • 在线: %d\n", report.OnlineServers)
	content += fmt.Sprintf("  • 离线: %d\n", report.OfflineServers)
	content += fmt.Sprintf("  • 告警: %d\n", report.WarningServers)
	content += fmt.Sprintf("  • 严重: %d\n\n", report.CriticalServers)

	content += "⚠️ 告警统计:\n"
	content += fmt.Sprintf("  • CPU告警: %d\n", report.CPUWarnings)
	content += fmt.Sprintf("  • 内存告警: %d\n", report.MemoryWarnings)
	content += fmt.Sprintf("  • 磁盘告警: %d\n", report.DiskWarnings)
	content += fmt.Sprintf("  • 服务异常: %d\n\n", report.ServiceIssues)

	// 建议
	var recommendations []string
	json.Unmarshal([]byte(report.Recommendations), &recommendations)
	if len(recommendations) > 0 {
		content += "💡 处理建议:\n"
		for _, r := range recommendations {
			content += fmt.Sprintf("  • %s\n", r)
		}
		content += "\n"
	}

	// AI 分析
	if report.AIAnalysis != "" {
		content += "🤖 AI分析:\n"
		content += report.AIAnalysis
	}

	return content
}

// GetReports 获取巡检报告列表
func (i *Inspector) GetReports(limit int) ([]InspectionReport, error) {
	var reports []InspectionReport
	err := global.DB.Order("created_at DESC").Limit(limit).Find(&reports).Error
	return reports, err
}

// GetReport 获取单个报告
func (i *Inspector) GetReport(id uint) (*InspectionReport, error) {
	var report InspectionReport
	err := global.DB.First(&report, id).Error
	return &report, err
}

// ScheduleInspection 定时巡检
func (i *Inspector) ScheduleInspection() {
	// 每天凌晨 6:00 执行
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 6, 0, 0, 0, now.Location())
			time.Sleep(next.Sub(now))

			global.Logger.Info("开始执行每日巡检...")
			_, err := i.RunDailyInspection()
			if err != nil {
				global.Logger.Error(fmt.Sprintf("巡检失败: %v", err))
			}
		}
	}()
}

// RunWeeklyInspection 执行每周巡检
func (i *Inspector) RunWeeklyInspection() (*InspectionReport, error) {
	report, err := i.RunDailyInspection()
	if err != nil {
		return nil, err
	}
	report.Type = InspectionWeekly
	global.DB.Save(report)
	return report, nil
}

// RunManualInspection 手动巡检
func (i *Inspector) RunManualInspection() (*InspectionReport, error) {
	report, err := i.RunDailyInspection()
	if err != nil {
		return nil, err
	}
	report.Type = InspectionManual
	global.DB.Save(report)
	return report, nil
}

// QuickHealthCheck 快速健康检查
func (i *Inspector) QuickHealthCheck() map[string]interface{} {
	var servers []server.Server
	global.DB.Find(&servers)

	online := 0
	offline := 0
	warning := 0
	critical := 0

	for _, srv := range servers {
		if !srv.AgentOnline {
			offline++
			continue
		}

		online++

		if srv.CPUUsage > 90 || srv.MemoryUsage > 90 || srv.DiskUsage > 90 {
			critical++
		} else if srv.CPUUsage > 80 || srv.MemoryUsage > 80 || srv.DiskUsage > 80 {
			warning++
		}
	}

	return map[string]interface{}{
		"timestamp": time.Now(),
		"total":     len(servers),
		"online":    online,
		"offline":   offline,
		"warning":   warning,
		"critical":  critical,
		"healthScore": float64(online*100-offline*50-warning*10-critical*30) / float64(len(servers)),
	}
}
