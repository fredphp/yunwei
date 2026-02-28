package patrol

import (
        "encoding/json"
        "fmt"
        "strings"
        "time"

        "yunwei/global"
        patrolModel "yunwei/model/patrol"
        "yunwei/model/server"
        "yunwei/service/detector"
        "yunwei/service/notify"
        "yunwei/service/prediction"
)

// 类型别名，保持兼容性
type PatrolStatus = patrolModel.PatrolStatus
type PatrolType = patrolModel.PatrolType
type PatrolRecord = patrolModel.PatrolRecord
type DailyReport = patrolModel.DailyReport
type ServerUsage = patrolModel.ServerUsage
type TrendAnalysis = patrolModel.TrendAnalysis

// 常量别名
const (
        PatrolStatusRunning   = patrolModel.PatrolStatusRunning
        PatrolStatusCompleted = patrolModel.PatrolStatusCompleted
        PatrolStatusFailed    = patrolModel.PatrolStatusFailed
        PatrolTypeDaily       = patrolModel.PatrolTypeDaily
        PatrolTypeHourly      = patrolModel.PatrolTypeHourly
        PatrolTypeManual      = patrolModel.PatrolTypeManual
        PatrolTypeTrigger     = patrolModel.PatrolTypeTrigger
)

// ServerCheckResult 服务器检查结果
type ServerCheckResult struct {
        ServerID   uint                `json:"serverId"`
        ServerName string              `json:"serverName"`
        Status     string              `json:"status"` // healthy, warning, critical, offline
        Checks     []CheckItem         `json:"checks"`
        Metrics    *server.ServerMetric `json:"metrics"`
        Alerts     []detector.DetectionResult `json:"alerts"`
        Suggestions []string           `json:"suggestions"`
}

// CheckItem 检查项
type CheckItem struct {
        Name     string `json:"name"`
        Status   string `json:"status"` // pass, warning, fail
        Value    string `json:"value"`
        Message  string `json:"message"`
}

// PatrolRobot 巡检机器人
type PatrolRobot struct {
        detector   *detector.Detector
        predictor  *prediction.Predictor
        notifier   notify.Notifier
}

// NewPatrolRobot 创建巡检机器人
func NewPatrolRobot() *PatrolRobot {
        return &PatrolRobot{
                detector: detector.NewDetector(),
        }
}

// SetNotifier 设置通知器
func (r *PatrolRobot) SetNotifier(notifier notify.Notifier) {
        r.notifier = notifier
}

// RunDailyPatrol 执行每日巡检
func (r *PatrolRobot) RunDailyPatrol() (*PatrolRecord, error) {
        return r.RunPatrol(PatrolTypeDaily)
}

// RunHourlyPatrol 执行每小时巡检
func (r *PatrolRobot) RunHourlyPatrol() (*PatrolRecord, error) {
        return r.RunPatrol(PatrolTypeHourly)
}

// RunPatrol 执行巡检
func (r *PatrolRobot) RunPatrol(patrolType PatrolType) (*PatrolRecord, error) {
        record := &PatrolRecord{
                Type:    patrolType,
                Status:  PatrolStatusRunning,
        }
        
        now := time.Now()
        record.StartedAt = &now
        global.DB.Create(record)

        // 获取所有服务器
        var servers []server.Server
        global.DB.Find(&servers)
        record.TotalServers = len(servers)

        var healthyServers, warningServers, criticalServers, offlineServerList []ServerCheckResult
        var totalAlerts int

        // 检查每台服务器
        for _, srv := range servers {
                result := r.CheckServer(&srv)
                
                switch result.Status {
                case "healthy":
                        healthyServers = append(healthyServers, result)
                        record.OnlineServers++
                case "warning":
                        warningServers = append(warningServers, result)
                        record.OnlineServers++
                        record.WarningCount++
                case "critical":
                        criticalServers = append(criticalServers, result)
                        record.OnlineServers++
                        record.CriticalCount++
                case "offline":
                        offlineServerList = append(offlineServerList, result)
                        record.OfflineServers++
                }
                
                totalAlerts += len(result.Alerts)
        }

        record.AlertCount = totalAlerts

        // 序列化服务器列表
        healthyJSON, _ := json.Marshal(healthyServers)
        warningJSON, _ := json.Marshal(warningServers)
        criticalJSON, _ := json.Marshal(criticalServers)
        offlineJSON, _ := json.Marshal(offlineServerList)

        record.HealthyServers = string(healthyJSON)
        record.WarningServers = string(warningJSON)
        record.CriticalServers = string(criticalJSON)
        record.OfflineServerList = string(offlineJSON)

        // 生成摘要
        record.Summary = r.GenerateSummary(record)
        record.Suggestions = r.GenerateSuggestions(record)

        // 完成巡检
        completedAt := time.Now()
        record.CompletedAt = &completedAt
        record.Duration = completedAt.Sub(*record.StartedAt).Milliseconds()
        record.Status = PatrolStatusCompleted

        global.DB.Save(record)

        // 发送通知
        if r.notifier != nil {
                r.notifier.SendPatrolReport(record)
        }

        return record, nil
}

// CheckServer 检查单台服务器
func (r *PatrolRobot) CheckServer(srv *server.Server) ServerCheckResult {
        result := ServerCheckResult{
                ServerID:   srv.ID,
                ServerName: srv.Name,
                Checks:     []CheckItem{},
                Status:     "healthy",
        }

        // 检查服务器是否在线
        if !srv.AgentOnline {
                result.Status = "offline"
                result.Checks = append(result.Checks, CheckItem{
                        Name:    "连接状态",
                        Status:  "fail",
                        Value:   "离线",
                        Message: "服务器Agent未连接",
                })
                return result
        }

        // 获取最新指标
        var metric server.ServerMetric
        if err := global.DB.Where("server_id = ?", srv.ID).Order("created_at DESC").First(&metric).Error; err != nil {
                result.Status = "warning"
                result.Checks = append(result.Checks, CheckItem{
                        Name:    "指标采集",
                        Status:  "fail",
                        Value:   "无数据",
                        Message: "无法获取服务器指标",
                })
                return result
        }
        result.Metrics = &metric

        // CPU 检查
        cpuStatus := "pass"
        if metric.CPUUsage > 90 {
                cpuStatus = "fail"
                result.Status = "critical"
        } else if metric.CPUUsage > 80 {
                cpuStatus = "warning"
                if result.Status == "healthy" {
                        result.Status = "warning"
                }
        }
        result.Checks = append(result.Checks, CheckItem{
                Name:    "CPU使用率",
                Status:  cpuStatus,
                Value:   fmt.Sprintf("%.1f%%", metric.CPUUsage),
                Message: r.getCPUMessage(metric.CPUUsage),
        })

        // 内存检查
        memStatus := "pass"
        if metric.MemoryUsage > 90 {
                memStatus = "fail"
                if result.Status != "critical" {
                        result.Status = "critical"
                }
        } else if metric.MemoryUsage > 80 {
                memStatus = "warning"
                if result.Status == "healthy" {
                        result.Status = "warning"
                }
        }
        result.Checks = append(result.Checks, CheckItem{
                Name:    "内存使用率",
                Status:  memStatus,
                Value:   fmt.Sprintf("%.1f%%", metric.MemoryUsage),
                Message: r.getMemoryMessage(metric.MemoryUsage),
        })

        // 磁盘检查
        diskStatus := "pass"
        if metric.DiskUsage > 90 {
                diskStatus = "fail"
                if result.Status != "critical" {
                        result.Status = "critical"
                }
        } else if metric.DiskUsage > 80 {
                diskStatus = "warning"
                if result.Status == "healthy" {
                        result.Status = "warning"
                }
        }
        result.Checks = append(result.Checks, CheckItem{
                Name:    "磁盘使用率",
                Status:  diskStatus,
                Value:   fmt.Sprintf("%.1f%%", metric.DiskUsage),
                Message: r.getDiskMessage(metric.DiskUsage),
        })

        // 负载检查
        loadStatus := "pass"
        if metric.Load1 > float64(srv.CPUCores) {
                loadStatus = "warning"
                if result.Status == "healthy" {
                        result.Status = "warning"
                }
        }
        result.Checks = append(result.Checks, CheckItem{
                Name:    "系统负载",
                Status:  loadStatus,
                Value:   fmt.Sprintf("%.2f", metric.Load1),
                Message: r.getLoadMessage(metric.Load1, srv.CPUCores),
        })

        // 运行检测规则
        processes := []detector.ProcessInfo{} // TODO: 从Agent获取
        containers := []server.DockerContainer{}
        ports := []server.PortInfo{}

        detectionResults := r.detector.Detect(srv, &metric, processes, containers, ports)
        result.Alerts = detectionResults

        // 生成建议
        result.Suggestions = r.generateServerSuggestions(result)

        return result
}

// GenerateDailyReport 生成日报
func (r *PatrolRobot) GenerateDailyReport() (*DailyReport, error) {
        report := &DailyReport{
                Date:        time.Now().Format("2006-01-02"),
                GeneratedAt: time.Now(),
        }

        // 获取所有服务器
        var servers []server.Server
        global.DB.Find(&servers)
        report.TotalServers = len(servers)

        // 在线率
        onlineCount := 0
        for _, srv := range servers {
                if srv.AgentOnline {
                        onlineCount++
                }
        }
        if len(servers) > 0 {
                report.OnlineRate = float64(onlineCount) / float64(len(servers)) * 100
        }

        // 获取最近24小时指标
        yesterday := time.Now().Add(-24 * time.Hour)
        var metrics []server.ServerMetric
        global.DB.Where("created_at > ?", yesterday).Find(&metrics)

        // 计算平均值
        if len(metrics) > 0 {
                var cpuSum, memSum, diskSum float64
                for _, m := range metrics {
                        cpuSum += m.CPUUsage
                        memSum += m.MemoryUsage
                        diskSum += m.DiskUsage
                }
                report.AvgCPUUsage = cpuSum / float64(len(metrics))
                report.AvgMemoryUsage = memSum / float64(len(metrics))
                report.AvgDiskUsage = diskSum / float64(len(metrics))
        }

        // 获取告警统计
        var alerts []detector.Alert
        global.DB.Where("created_at > ?", yesterday).Find(&alerts)
        report.TotalAlerts = len(alerts)

        for _, alert := range alerts {
                switch alert.Level {
                case detector.AlertLevelCritical:
                        report.CriticalAlerts++
                case detector.AlertLevelWarning:
                        report.WarningAlerts++
                }
                if alert.Status == "resolved" {
                        report.ResolvedAlerts++
                }
        }

        // 获取资源使用率最高的服务器
        report.TopCPUServers = r.getTopUsageServers(servers, "cpu")
        report.TopMemoryServers = r.getTopUsageServers(servers, "memory")
        report.TopDiskServers = r.getTopUsageServers(servers, "disk")

        // 趋势分析
        report.Trends = r.analyzeTrends(yesterday)

        // 生成建议
        report.Recommendations = r.generateDailyRecommendations(report)

        return report, nil
}

// GenerateSummary 生成摘要
func (r *PatrolRobot) GenerateSummary(record *PatrolRecord) string {
        var sb strings.Builder

        sb.WriteString(fmt.Sprintf("## 巡检报告 - %s\n\n", record.CreatedAt.Format("2006-01-02 15:04")))
        sb.WriteString(fmt.Sprintf("**巡检类型**: %s\n", record.Type))
        sb.WriteString(fmt.Sprintf("**执行时间**: %s\n\n", record.StartedAt.Format("15:04:05")))

        sb.WriteString("### 服务器概览\n")
        sb.WriteString(fmt.Sprintf("- 总服务器数: %d\n", record.TotalServers))
        sb.WriteString(fmt.Sprintf("- 🟢 在线: %d\n", record.OnlineServers))
        sb.WriteString(fmt.Sprintf("- 🔴 离线: %d\n", record.OfflineServers))
        sb.WriteString(fmt.Sprintf("- ⚠️ 警告: %d\n", record.WarningCount))
        sb.WriteString(fmt.Sprintf("- 🔥 严重: %d\n\n", record.CriticalCount))

        sb.WriteString("### 告警统计\n")
        sb.WriteString(fmt.Sprintf("- 总告警数: %d\n", record.AlertCount))

        if record.Duration > 0 {
                sb.WriteString(fmt.Sprintf("\n**巡检耗时**: %dms\n", record.Duration))
        }

        return sb.String()
}

// GenerateSuggestions 生成建议
func (r *PatrolRobot) GenerateSuggestions(record *PatrolRecord) string {
        var suggestions []string

        if record.OfflineServers > 0 {
                suggestions = append(suggestions, "🔴 有服务器离线，请检查网络连接和Agent状态")
        }
        if record.CriticalCount > 0 {
                suggestions = append(suggestions, "🔥 发现严重问题，建议立即处理")
        }
        if record.WarningCount > 0 {
                suggestions = append(suggestions, "⚠️ 存在警告级别的异常，建议尽快关注")
        }

        // 解析严重服务器列表
        var criticalServers []ServerCheckResult
        json.Unmarshal([]byte(record.CriticalServers), &criticalServers)
        for _, srv := range criticalServers {
                for _, check := range srv.Checks {
                        if check.Status == "fail" {
                                suggestions = append(suggestions, fmt.Sprintf("- [%s] %s: %s", srv.ServerName, check.Name, check.Message))
                        }
                }
        }

        if len(suggestions) == 0 {
                suggestions = append(suggestions, "✅ 所有服务器运行正常")
        }

        return strings.Join(suggestions, "\n")
}

// Helper functions
func (r *PatrolRobot) getCPUMessage(usage float64) string {
        if usage > 90 {
                return "CPU使用率过高，可能影响服务性能"
        } else if usage > 80 {
                return "CPU使用率较高，建议关注"
        }
        return "正常"
}

func (r *PatrolRobot) getMemoryMessage(usage float64) string {
        if usage > 90 {
                return "内存严重不足，可能导致OOM"
        } else if usage > 80 {
                return "内存使用率较高"
        }
        return "正常"
}

func (r *PatrolRobot) getDiskMessage(usage float64) string {
        if usage > 90 {
                return "磁盘空间严重不足，请立即清理"
        } else if usage > 80 {
                return "磁盘空间紧张，建议清理"
        }
        return "正常"
}

func (r *PatrolRobot) getLoadMessage(load float64, cores int) string {
        if load > float64(cores) {
                return "系统负载较高，超过CPU核心数"
        }
        return "正常"
}

func (r *PatrolRobot) generateServerSuggestions(result ServerCheckResult) []string {
        var suggestions []string

        for _, check := range result.Checks {
                if check.Status == "fail" {
                        switch check.Name {
                        case "CPU使用率":
                                suggestions = append(suggestions, "检查CPU密集型进程，考虑清理缓存或扩容")
                        case "内存使用率":
                                suggestions = append(suggestions, "释放内存缓存，检查内存泄漏")
                        case "磁盘使用率":
                                suggestions = append(suggestions, "清理Docker镜像、日志文件，或扩容磁盘")
                        }
                }
        }

        return suggestions
}

func (r *PatrolRobot) getTopUsageServers(servers []server.Server, metricType string) []ServerUsage {
        var usages []ServerUsage
        
        for _, srv := range servers {
                usage := ServerUsage{
                        ServerID:   srv.ID,
                        ServerName: srv.Name,
                }
                
                switch metricType {
                case "cpu":
                        usage.Usage = srv.CPUUsage
                case "memory":
                        usage.Usage = srv.MemoryUsage
                case "disk":
                        usage.Usage = srv.DiskUsage
                }
                
                usages = append(usages, usage)
        }

        // 简单排序（冒泡）
        for i := 0; i < len(usages); i++ {
                for j := i + 1; j < len(usages); j++ {
                        if usages[j].Usage > usages[i].Usage {
                                usages[i], usages[j] = usages[j], usages[i]
                        }
                }
        }

        // 返回前5
        if len(usages) > 5 {
                usages = usages[:5]
        }

        return usages
}

func (r *PatrolRobot) analyzeTrends(since time.Time) TrendAnalysis {
        // 简化的趋势分析
        return TrendAnalysis{
                CPUTrend:    "stable",
                MemoryTrend: "stable",
                DiskTrend:   "stable",
                AlertTrend:  "stable",
        }
}

func (r *PatrolRobot) generateDailyRecommendations(report *DailyReport) []string {
        var recommendations []string

        if report.AvgCPUUsage > 70 {
                recommendations = append(recommendations, "平均CPU使用率较高，建议评估扩容需求")
        }
        if report.AvgMemoryUsage > 75 {
                recommendations = append(recommendations, "平均内存使用率较高，建议优化内存配置")
        }
        if report.AvgDiskUsage > 70 {
                recommendations = append(recommendations, "平均磁盘使用率较高，建议制定清理计划")
        }
        if report.OnlineRate < 100 {
                recommendations = append(recommendations, "有服务器离线，请检查网络和Agent状态")
        }
        if report.CriticalAlerts > 0 {
                recommendations = append(recommendations, fmt.Sprintf("今日有%d个严重告警，建议优先处理", report.CriticalAlerts))
        }

        if len(recommendations) == 0 {
                recommendations = append(recommendations, "系统运行稳定，继续保持监控")
        }

        return recommendations
}

// GetPatrolHistory 获取巡检历史
func (r *PatrolRobot) GetPatrolHistory(limit int) ([]PatrolRecord, error) {
        var records []PatrolRecord
        err := global.DB.Order("created_at DESC").Limit(limit).Find(&records).Error
        return records, err
}

// GetPatrolRecord 获取巡检记录详情
func (r *PatrolRobot) GetPatrolRecord(id uint) (*PatrolRecord, error) {
        var record PatrolRecord
        err := global.DB.First(&record, id).Error
        return &record, err
}
