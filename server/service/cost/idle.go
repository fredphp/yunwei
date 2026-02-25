package cost

import (
        "context"
        "encoding/json"
        "fmt"
        "math"
        "strings"
        "sync"
        "time"

        "yunwei/model/cost"
)

// IdleDetectionService 闲置资源检测服务
type IdleDetectionService struct {
        mu sync.RWMutex
}

// NewIdleDetectionService 创建闲置检测服务
func NewIdleDetectionService() *IdleDetectionService {
        return &IdleDetectionService{}
}

// IdleSummary 闲置摘要
type IdleSummary struct {
        TotalIdleResources int                    `json:"total_idle_resources"`
        TotalWastedCost    float64                `json:"total_wasted_cost"`
        ByType             map[string]IdleTypeStat `json:"by_type"`
        ByProvider         map[string]int         `json:"by_provider"`
        TopIdle            []IdleResource         `json:"top_idle"`
        Recommendations    []string               `json:"recommendations"`
        PotentialSavings   float64                `json:"potential_savings"`
}

// IdleTypeStat 闲置类型统计
type IdleTypeStat struct {
        Type          string  `json:"type"`
        Count         int     `json:"count"`
        WastedCost    float64 `json:"wasted_cost"`
        AvgIdleDays   float64 `json:"avg_idle_days"`
}

// IdleResource 闲置资源
type IdleResource struct {
        ResourceID     string                 `json:"resource_id"`
        ResourceName   string                 `json:"resource_name"`
        ResourceType   string                 `json:"resource_type"`
        Provider       string                 `json:"provider"`
        Region         string                 `json:"region"`
        Status         string                 `json:"status"`
        IdleDays       int                    `json:"idle_days"`
        IdleReason     string                 `json:"idle_reason"`
        IdleScore      float64                `json:"idle_score"`
        MonthlyCost    float64                `json:"monthly_cost"`
        TotalWastedCost float64               `json:"total_wasted_cost"`
        CPUUsage       float64                `json:"cpu_usage"`
        MemoryUsage    float64                `json:"memory_usage"`
        NetworkUsage   float64                `json:"network_usage"`
        LastActiveTime string                 `json:"last_active_time"`
        Recommendation string                 `json:"recommendation"`
        Action         string                 `json:"action"`
        Priority       string                 `json:"priority"`
        Metrics        map[string]float64     `json:"metrics"`
}

// DetectIdleResources 检测闲置资源
func (s *IdleDetectionService) DetectIdleResources(ctx context.Context, threshold float64) (*IdleSummary, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        summary := &IdleSummary{
                ByType:          make(map[string]IdleTypeStat),
                ByProvider:      make(map[string]int),
                TopIdle:         make([]IdleResource, 0),
                Recommendations: make([]string, 0),
        }

        // 检测各类闲置资源
        computeIdle := s.detectIdleCompute(ctx, threshold)
        storageIdle := s.detectIdleStorage(ctx, threshold)
        databaseIdle := s.detectIdleDatabase(ctx, threshold)
        networkIdle := s.detectIdleNetwork(ctx, threshold)
        k8sIdle := s.detectIdleK8s(ctx, threshold)

        // 合并所有闲置资源
        allIdle := make([]IdleResource, 0)
        allIdle = append(allIdle, computeIdle...)
        allIdle = append(allIdle, storageIdle...)
        allIdle = append(allIdle, databaseIdle...)
        allIdle = append(allIdle, networkIdle...)
        allIdle = append(allIdle, k8sIdle...)

        // 统计
        for _, r := range allIdle {
                summary.TotalIdleResources++
                summary.TotalWastedCost += r.MonthlyCost

                // 按类型统计
                stat, ok := summary.ByType[r.ResourceType]
                if !ok {
                        stat = IdleTypeStat{Type: r.ResourceType}
                }
                stat.Count++
                stat.WastedCost += r.MonthlyCost
                stat.AvgIdleDays = (stat.AvgIdleDays*float64(stat.Count-1) + float64(r.IdleDays)) / float64(stat.Count)
                summary.ByType[r.ResourceType] = stat

                // 按云商统计
                summary.ByProvider[r.Provider]++
        }

        // 计算潜在节省
        summary.PotentialSavings = summary.TotalWastedCost * 0.9

        // 排序取Top
        s.sortedIdleResources(allIdle)
        if len(allIdle) > 10 {
                summary.TopIdle = allIdle[:10]
        } else {
                summary.TopIdle = allIdle
        }

        // 生成建议
        summary.Recommendations = s.generateIdleRecommendations(summary)

        return summary, nil
}

// detectIdleCompute 检测闲置计算资源
func (s *IdleDetectionService) detectIdleCompute(ctx context.Context, threshold float64) []IdleResource {
        resources := make([]IdleResource, 0)

        // 模拟检测闲置ECS/EC2
        idleInstances := []struct {
                id         string
                name       string
                provider   string
                region     string
                cpuUsage   float64
                memUsage   float64
                netUsage   float64
                idleDays   int
                monthlyCost float64
                status     string
        }{
                {"ecs-001", "测试服务器A", "aliyun", "cn-hangzhou", 0, 0, 0, 30, 500, "stopped"},
                {"ecs-002", "旧版API服务器", "aliyun", "cn-shanghai", 2, 5, 0.5, 45, 800, "running"},
                {"ec2-001", "Dev环境服务器", "aws", "us-east-1", 3, 8, 1, 60, 600, "running"},
                {"ecs-003", "备份服务器", "aliyun", "cn-beijing", 0, 0, 0, 90, 400, "stopped"},
                {"vm-001", "临时测试VM", "azure", "eastus", 1, 2, 0, 15, 300, "running"},
        }

        for _, i := range idleInstances {
                avgUsage := (i.cpuUsage + i.memUsage + i.netUsage) / 3
                if avgUsage < threshold {
                        idleScore := s.calculateIdleScore(i.cpuUsage, i.memUsage, i.netUsage, i.idleDays)
                        action := "terminate"
                        if i.status == "running" && avgUsage > 0 {
                                action = "stop_or_downsize"
                        }

                        resources = append(resources, IdleResource{
                                ResourceID:     i.id,
                                ResourceName:   i.name,
                                ResourceType:   "ecs",
                                Provider:       i.provider,
                                Region:         i.region,
                                Status:         i.status,
                                IdleDays:       i.idleDays,
                                IdleReason:     s.determineIdleReason(i.cpuUsage, i.memUsage, i.netUsage, i.idleDays),
                                IdleScore:      idleScore,
                                MonthlyCost:    i.monthlyCost,
                                TotalWastedCost: i.monthlyCost * float64(i.idleDays) / 30,
                                CPUUsage:       i.cpuUsage,
                                MemoryUsage:    i.memUsage,
                                NetworkUsage:   i.netUsage,
                                LastActiveTime: time.Now().AddDate(0, 0, -i.idleDays).Format("2006-01-02"),
                                Recommendation: s.generateComputeRecommendation(i.idleDays, i.monthlyCost),
                                Action:         action,
                                Priority:       s.calculateIdlePriority(i.idleDays, i.monthlyCost),
                                Metrics: map[string]float64{
                                        "cpu_usage":    i.cpuUsage,
                                        "memory_usage": i.memUsage,
                                        "network_usage": i.netUsage,
                                },
                        })
                }
        }

        return resources
}

// detectIdleStorage 检测闲置存储资源
func (s *IdleDetectionService) detectIdleStorage(ctx context.Context, threshold float64) []IdleResource {
        resources := make([]IdleResource, 0)

        // 模拟检测闲置存储
        idleStorage := []struct {
                id          string
                name        string
                provider    string
                region      string
                type_       string
                size        int // GB
                usagePercent float64
                idleDays    int
                monthlyCost float64
        }{
                {"ebs-001", "旧数据盘A", "aws", "us-east-1", "ebs", 500, 0, 30, 100},
                {"oss-001", "废弃存储桶", "aliyun", "cn-hangzhou", "oss", 1000, 0, 60, 150},
                {"disk-001", "未挂载磁盘", "aliyun", "cn-shanghai", "disk", 200, 0, 45, 50},
                {"cos-001", "旧备份存储", "tencent", "ap-guangzhou", "cos", 500, 5, 90, 80},
        }

        for _, s_ := range idleStorage {
                if s_.usagePercent < threshold {
                        resources = append(resources, IdleResource{
                                ResourceID:     s_.id,
                                ResourceName:   s_.name,
                                ResourceType:   s_.type_,
                                Provider:       s_.provider,
                                Region:         s_.region,
                                Status:         "unattached",
                                IdleDays:       s_.idleDays,
                                IdleReason:     fmt.Sprintf("存储使用率仅 %.1f%%", s_.usagePercent),
                                IdleScore:      100 - s_.usagePercent,
                                MonthlyCost:    s_.monthlyCost,
                                TotalWastedCost: s_.monthlyCost * float64(s_.idleDays) / 30,
                                LastActiveTime: time.Now().AddDate(0, 0, -s_.idleDays).Format("2006-01-02"),
                                Recommendation: fmt.Sprintf("建议删除或归档，可节省 %.0f 元/月", s_.monthlyCost),
                                Action:         "delete_or_archive",
                                Priority:       "medium",
                                Metrics: map[string]float64{
                                        "size_gb":       float64(s_.size),
                                        "usage_percent": s_.usagePercent,
                                },
                        })
                }
        }

        return resources
}

// detectIdleDatabase 检测闲置数据库
func (s *IdleDetectionService) detectIdleDatabase(ctx context.Context, threshold float64) []IdleResource {
        resources := make([]IdleResource, 0)

        // 模拟检测闲置数据库
        idleDBs := []struct {
                id          string
                name        string
                provider    string
                region      string
                type_       string
                connCount   int
                cpuUsage    float64
                idleDays    int
                monthlyCost float64
                status      string
        }{
                {"rds-001", "测试数据库", "aliyun", "cn-hangzhou", "rds", 0, 0, 20, 1200, "running"},
                {"rds-002", "旧业务库", "aliyun", "cn-shanghai", "rds", 2, 1, 60, 800, "running"},
                {"mysql-001", "Dev数据库", "aws", "us-east-1", "rds", 0, 0, 30, 600, "stopped"},
                {"mongo-001", "测试MongoDB", "aliyun", "cn-beijing", "mongodb", 1, 2, 45, 500, "running"},
        }

        for _, db := range idleDBs {
                if db.cpuUsage < threshold {
                        resources = append(resources, IdleResource{
                                ResourceID:     db.id,
                                ResourceName:   db.name,
                                ResourceType:   db.type_,
                                Provider:       db.provider,
                                Region:         db.region,
                                Status:         db.status,
                                IdleDays:       db.idleDays,
                                IdleReason:     fmt.Sprintf("连接数 %d, CPU使用率 %.1f%%", db.connCount, db.cpuUsage),
                                IdleScore:      100 - db.cpuUsage,
                                MonthlyCost:    db.monthlyCost,
                                TotalWastedCost: db.monthlyCost * float64(db.idleDays) / 30,
                                CPUUsage:       db.cpuUsage,
                                LastActiveTime: time.Now().AddDate(0, 0, -db.idleDays).Format("2006-01-02"),
                                Recommendation: s.generateDBRecommendation(db.idleDays, db.monthlyCost, db.status),
                                Action:         "stop_or_downsize",
                                Priority:       s.calculateIdlePriority(db.idleDays, db.monthlyCost),
                                Metrics: map[string]float64{
                                        "cpu_usage":   db.cpuUsage,
                                        "connections": float64(db.connCount),
                                },
                        })
                }
        }

        return resources
}

// detectIdleNetwork 检测闲置网络资源
func (s *IdleDetectionService) detectIdleNetwork(ctx context.Context, threshold float64) []IdleResource {
        resources := make([]IdleResource, 0)

        // 模拟检测闲置网络资源
        idleNetwork := []struct {
                id          string
                name        string
                provider    string
                region      string
                type_       string
                usagePercent float64
                idleDays    int
                monthlyCost float64
        }{
                {"eip-001", "未绑定EIP", "aliyun", "cn-hangzhou", "eip", 0, 15, 50},
                {"slb-001", "未使用负载均衡", "aliyun", "cn-shanghai", "slb", 0, 30, 200},
                {"nat-001", "闲置NAT网关", "aliyun", "cn-beijing", "nat", 5, 20, 300},
                {"elb-001", "废弃ELB", "aws", "us-east-1", "elb", 0, 60, 150},
        }

        for _, n := range idleNetwork {
                if n.usagePercent < threshold {
                        resources = append(resources, IdleResource{
                                ResourceID:     n.id,
                                ResourceName:   n.name,
                                ResourceType:   n.type_,
                                Provider:       n.provider,
                                Region:         n.region,
                                Status:         "unused",
                                IdleDays:       n.idleDays,
                                IdleReason:     fmt.Sprintf("使用率 %.1f%%", n.usagePercent),
                                IdleScore:      100 - n.usagePercent,
                                MonthlyCost:    n.monthlyCost,
                                TotalWastedCost: n.monthlyCost * float64(n.idleDays) / 30,
                                LastActiveTime: time.Now().AddDate(0, 0, -n.idleDays).Format("2006-01-02"),
                                Recommendation: fmt.Sprintf("建议释放，可节省 %.0f 元/月", n.monthlyCost),
                                Action:         "release",
                                Priority:       "high",
                        })
                }
        }

        return resources
}

// detectIdleK8s 检测闲置K8s资源
func (s *IdleDetectionService) detectIdleK8s(ctx context.Context, threshold float64) []IdleResource {
        resources := make([]IdleResource, 0)

        // 模拟检测闲置K8s资源
        idleK8s := []struct {
                clusterID   string
                namespace   string
                name        string
                type_       string
                cpuRequest  float64
                cpuUsage    float64
                memRequest  int64
                memUsage    int64
                idleDays    int
                monthlyCost float64
        }{
                {"k8s-001", "test", "test-deployment", "deployment", 4, 0.2, 8*1024*1024*1024, 1*1024*1024*1024, 20, 500},
                {"k8s-001", "dev", "dev-api", "deployment", 2, 0.1, 4*1024*1024*1024, 512*1024*1024, 30, 300},
                {"k8s-002", "staging", "old-service", "deployment", 8, 0.5, 16*1024*1024*1024, 2*1024*1024*1024, 45, 800},
                {"k8s-001", "default", "unused-pod", "pod", 1, 0, 2*1024*1024*1024, 0, 7, 100},
        }

        for _, k := range idleK8s {
                cpuUsagePercent := (k.cpuUsage / k.cpuRequest) * 100
                memUsagePercent := float64(k.memUsage) / float64(k.memRequest) * 100
                avgUsage := (cpuUsagePercent + memUsagePercent) / 2

                if avgUsage < threshold {
                        resources = append(resources, IdleResource{
                                ResourceID:     fmt.Sprintf("%s/%s/%s", k.clusterID, k.namespace, k.name),
                                ResourceName:   fmt.Sprintf("%s (%s)", k.name, k.namespace),
                                ResourceType:   "k8s_" + k.type_,
                                Provider:       "kubernetes",
                                Region:         k.clusterID,
                                Status:         "running",
                                IdleDays:       k.idleDays,
                                IdleReason:     fmt.Sprintf("CPU使用 %.1f%%, 内存使用 %.1f%%", cpuUsagePercent, memUsagePercent),
                                IdleScore:      100 - avgUsage,
                                MonthlyCost:    k.monthlyCost,
                                TotalWastedCost: k.monthlyCost * float64(k.idleDays) / 30,
                                CPUUsage:       cpuUsagePercent,
                                MemoryUsage:    memUsagePercent,
                                LastActiveTime: time.Now().AddDate(0, 0, -k.idleDays).Format("2006-01-02"),
                                Recommendation: "建议缩减副本数或调整资源请求",
                                Action:         "scale_down",
                                Priority:       "medium",
                                Metrics: map[string]float64{
                                        "cpu_request": k.cpuRequest,
                                        "cpu_usage":   k.cpuUsage,
                                        "mem_request_gb": float64(k.memRequest) / 1024 / 1024 / 1024,
                                        "mem_usage_gb":   float64(k.memUsage) / 1024 / 1024 / 1024,
                                },
                        })
                }
        }

        return resources
}

// AnalyzeResourceUsage 分析资源使用情况
func (s *IdleDetectionService) AnalyzeResourceUsage(ctx context.Context, resourceID string, duration time.Duration) (*UsageAnalysis, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        analysis := &UsageAnalysis{
                ResourceID: resourceID,
                Duration:   duration.String(),
                DataPoints: make([]UsagePoint, 0),
        }

        // 模拟生成使用数据点
        now := time.Now()
        interval := duration / 24
        for i := 0; i < 24; i++ {
                t := now.Add(-duration + time.Duration(i)*interval)
                analysis.DataPoints = append(analysis.DataPoints, UsagePoint{
                        Timestamp: t.Unix(),
                        CPU:       5 + float64(i%10),
                        Memory:    10 + float64(i%15),
                        Network:   2 + float64(i%5),
                        DiskIO:    1 + float64(i%3),
                })
        }

        // 计算统计数据
        var cpuSum, memSum, netSum float64
        for _, p := range analysis.DataPoints {
                cpuSum += p.CPU
                memSum += p.Memory
                netSum += p.Network
        }
        n := float64(len(analysis.DataPoints))
        analysis.AvgCPU = cpuSum / n
        analysis.AvgMemory = memSum / n
        analysis.AvgNetwork = netSum / n
        analysis.IsIdle = analysis.AvgCPU < 10 && analysis.AvgMemory < 20

        return analysis, nil
}

// UsageAnalysis 使用分析
type UsageAnalysis struct {
        ResourceID string       `json:"resource_id"`
        Duration   string       `json:"duration"`
        DataPoints []UsagePoint `json:"data_points"`
        AvgCPU     float64      `json:"avg_cpu"`
        AvgMemory  float64      `json:"avg_memory"`
        AvgNetwork float64      `json:"avg_network"`
        IsIdle     bool         `json:"is_idle"`
}

// UsagePoint 使用点
type UsagePoint struct {
        Timestamp int64   `json:"timestamp"`
        CPU       float64 `json:"cpu"`
        Memory    float64 `json:"memory"`
        Network   float64 `json:"network"`
        DiskIO    float64 `json:"disk_io"`
}

// GetIdleTimeline 获取闲置时间线
func (s *IdleDetectionService) GetIdleTimeline(ctx context.Context, resourceID string) (*IdleTimeline, error) {
        timeline := &IdleTimeline{
                ResourceID: resourceID,
                Events:     make([]IdleEvent, 0),
        }

        // 模拟时间线事件
        now := time.Now()
        timeline.Events = []IdleEvent{
                {Time: now.Add(-30*24*time.Hour).Unix(), Event: "resource_created", Details: "资源创建"},
                {Time: now.Add(-25*24*time.Hour).Unix(), Event: "usage_high", Details: "使用率 80%"},
                {Time: now.Add(-20*24*time.Hour).Unix(), Event: "usage_drop", Details: "使用率下降至 30%"},
                {Time: now.Add(-15*24*time.Hour).Unix(), Event: "idle_detected", Details: "检测到闲置"},
                {Time: now.Add(-10*24*time.Hour).Unix(), Event: "alert_sent", Details: "发送闲置告警"},
                {Time: now.Add(-5*24*time.Hour).Unix(), Event: "still_idle", Details: "持续闲置"},
        }

        timeline.TotalIdleDays = 15
        timeline.EstimatedWaste = 750.0

        return timeline, nil
}

// IdleTimeline 闲置时间线
type IdleTimeline struct {
        ResourceID    string       `json:"resource_id"`
        Events        []IdleEvent  `json:"events"`
        TotalIdleDays int          `json:"total_idle_days"`
        EstimatedWaste float64     `json:"estimated_waste"`
}

// IdleEvent 闲置事件
type IdleEvent struct {
        Time    int64  `json:"time"`
        Event   string `json:"event"`
        Details string `json:"details"`
}

// Helper methods

func (s *IdleDetectionService) calculateIdleScore(cpu, mem, net float64, idleDays int) float64 {
        // 使用率越低，闲置天数越长，分数越高
        usageScore := 100 - (cpu+mem+net)/3
        timeScore := math.Min(float64(idleDays)/30*100, 100)
        return (usageScore + timeScore) / 2
}

func (s *IdleDetectionService) determineIdleReason(cpu, mem, net float64, idleDays int) string {
        reasons := make([]string, 0)
        if cpu < 5 {
                reasons = append(reasons, fmt.Sprintf("CPU使用率 %.1f%%", cpu))
        }
        if mem < 10 {
                reasons = append(reasons, fmt.Sprintf("内存使用率 %.1f%%", mem))
        }
        if net < 1 {
                reasons = append(reasons, "网络流量几乎为零")
        }
        if idleDays > 7 {
                reasons = append(reasons, fmt.Sprintf("已闲置 %d 天", idleDays))
        }
        if len(reasons) == 0 {
                return "资源使用率低"
        }
        return strings.Join(reasons, ", ")
}

func (s *IdleDetectionService) generateComputeRecommendation(idleDays int, monthlyCost float64) string {
        if idleDays > 30 {
                return fmt.Sprintf("建议立即释放，每月可节省 %.0f 元", monthlyCost)
        } else if idleDays > 7 {
                return fmt.Sprintf("建议停止实例，需要时再启动，每月可节省 %.0f 元", monthlyCost)
        }
        return "建议监控使用情况，考虑降配"
}

func (s *IdleDetectionService) generateDBRecommendation(idleDays int, monthlyCost float64, status string) string {
        if status == "stopped" {
                return fmt.Sprintf("数据库已停止，建议释放，每月可节省 %.0f 元", monthlyCost)
        }
        if idleDays > 30 {
                return fmt.Sprintf("建议停止或释放，每月可节省 %.0f 元", monthlyCost)
        }
        return "建议降配或迁移到更小的实例"
}

func (s *IdleDetectionService) calculateIdlePriority(idleDays int, monthlyCost float64) string {
        if idleDays > 30 || monthlyCost > 1000 {
                return "high"
        } else if idleDays > 14 || monthlyCost > 500 {
                return "medium"
        }
        return "low"
}

func (s *IdleDetectionService) sortedIdleResources(resources []IdleResource) {
        for i := 0; i < len(resources); i++ {
                for j := i + 1; j < len(resources); j++ {
                        if resources[i].MonthlyCost < resources[j].MonthlyCost {
                                resources[i], resources[j] = resources[j], resources[i]
                        }
                }
        }
}

func (s *IdleDetectionService) generateIdleRecommendations(summary *IdleSummary) []string {
        recommendations := make([]string, 0)

        if summary.TotalIdleResources > 10 {
                recommendations = append(recommendations, fmt.Sprintf("发现 %d 个闲置资源，建议优先处理高成本资源", summary.TotalIdleResources))
        }

        if summary.TotalWastedCost > 5000 {
                recommendations = append(recommendations, fmt.Sprintf("每月浪费 %.0f 元，建议实施自动化闲置资源检测和清理", summary.TotalWastedCost))
        }

        for t, stat := range summary.ByType {
                if stat.WastedCost > 1000 {
                        recommendations = append(recommendations, fmt.Sprintf("%s 类型闲置浪费严重(%.0f 元/月)", t, stat.WastedCost))
                }
        }

        recommendations = append(recommendations, "建议配置自动停止/释放策略")
        recommendations = append(recommendations, "建议为测试环境设置自动关机时间")

        return recommendations
}

// CreateIdleRecord 创建闲置记录
func (s *IdleDetectionService) CreateIdleRecord(ctx context.Context, r IdleResource) *cost.IdleResource {
        record := &cost.IdleResource{
                ResourceID:       r.ResourceID,
                ResourceName:     r.ResourceName,
                ResourceType:     r.ResourceType,
                Provider:         r.Provider,
                IdleStatus:       r.IdleReason,
                IdleDays:         r.IdleDays,
                IdleScore:        r.IdleScore,
                MonthlyCost:      r.MonthlyCost,
                AccumulatedCost:  r.TotalWastedCost,
                CPUUtilization:   r.CPUUsage,
                MemoryUtilization: r.MemoryUsage,
                NetworkThroughput: r.NetworkUsage,
                Recommendation:   r.Recommendation,
                ScheduledAction:  r.Action,
                Status:           "active",
                FirstDetectedAt:  time.Now(),
        }

        if r.LastActiveTime != "" {
                t, _ := time.Parse("2006-01-02", r.LastActiveTime)
                record.LastActiveAt = &t
        }

        return record
}

// SetIdlePolicy 设置闲置策略
func (s *IdleDetectionService) SetIdlePolicy(ctx context.Context, policy IdlePolicy) error {
        // 实际应保存到数据库
        return nil
}

// IdlePolicy 闲置策略
type IdlePolicy struct {
        ID                  uint              `json:"id"`
        Name                string            `json:"name"`
        ResourceType        string            `json:"resource_type"`
        IdleThreshold       float64           `json:"idle_threshold"`
        IdleDaysThreshold   int               `json:"idle_days_threshold"`
        AutoAction          string            `json:"auto_action"` // none, stop, terminate, notify
        ExcludeTags         map[string]string `json:"exclude_tags"`
        NotifyBeforeAction  bool              `json:"notify_before_action"`
        NotifyDaysBefore    int               `json:"notify_days_before"`
        Enabled             bool              `json:"enabled"`
}
