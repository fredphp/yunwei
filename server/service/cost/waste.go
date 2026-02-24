package cost

import (
        "context"
        "encoding/json"
        "fmt"
        "math"
        "sync"
        "time"

        "yunwei/model/cost"
)

// WasteDetectionService 资源浪费检测服务
type WasteDetectionService struct {
        mu sync.RWMutex
}

// NewWasteDetectionService 创建浪费检测服务
func NewWasteDetectionService() *WasteDetectionService {
        return &WasteDetectionService{}
}

// WasteSummary 浪费摘要
type WasteSummary struct {
        TotalWastedCost  float64              `json:"total_wasted_cost"`
        TotalResources   int                  `json:"total_resources"`
        WastedResources  int                  `json:"wasted_resources"`
        ByType           map[string]TypeWaste `json:"by_type"`
        ByProvider       map[string]float64   `json:"by_provider"`
        TopWasted        []WasteItem          `json:"top_wasted"`
        PotentialSavings float64              `json:"potential_savings"`
        SavingsPercent   float64              `json:"savings_percent"`
        Recommendations  []string             `json:"recommendations"`
}

// TypeWaste 类型浪费
type TypeWaste struct {
        Type         string  `json:"type"`
        WastedCost   float64 `json:"wasted_cost"`
        ResourceCount int    `json:"resource_count"`
        Percent      float64 `json:"percent"`
}

// WasteItem 浪费项
type WasteItem struct {
        ResourceID     string  `json:"resource_id"`
        ResourceName   string  `json:"resource_name"`
        ResourceType   string  `json:"resource_type"`
        WasteType      string  `json:"waste_type"`
        WastedCost     float64 `json:"wasted_cost"`
        MonthlyCost    float64 `json:"monthly_cost"`
        UsagePercent   float64 `json:"usage_percent"`
        Recommendation string  `json:"recommendation"`
        Priority       string  `json:"priority"`
}

// DetectWaste 检测资源浪费
func (s *WasteDetectionService) DetectWaste(ctx context.Context) (*WasteSummary, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        summary := &WasteSummary{
                ByType:          make(map[string]TypeWaste),
                ByProvider:      make(map[string]float64),
                TopWasted:       make([]WasteItem, 0),
                Recommendations: make([]string, 0),
        }

        // 检测各类浪费
        overProvisioned := s.detectOverProvisioned(ctx)
        idleResources := s.detectIdleResources(ctx)
        unattachedResources := s.detectUnattachedResources(ctx)
        oldSnapshots := s.detectOldSnapshots(ctx)
        unusedBandwidth := s.detectUnusedBandwidth(ctx)

        // 合并所有浪费项
        allWasted := make([]WasteItem, 0)
        allWasted = append(allWasted, overProvisioned...)
        allWasted = append(allWasted, idleResources...)
        allWasted = append(allWasted, unattachedResources...)
        allWasted = append(allWasted, oldSnapshots...)
        allWasted = append(allWasted, unusedBandwidth...)

        // 统计总浪费
        for _, item := range allWasted {
                summary.TotalWastedCost += item.WastedCost
                summary.WastedResources++

                // 按类型统计
                tw, ok := summary.ByType[item.ResourceType]
                if !ok {
                        tw = TypeWaste{Type: item.ResourceType}
                }
                tw.WastedCost += item.WastedCost
                tw.ResourceCount++
                summary.ByType[item.ResourceType] = tw
        }

        // 按云商统计 (模拟)
        summary.ByProvider = map[string]float64{
                "aliyun": summary.TotalWastedCost * 0.4,
                "tencent": summary.TotalWastedCost * 0.25,
                "aws":    summary.TotalWastedCost * 0.2,
                "azure":  summary.TotalWastedCost * 0.15,
        }

        // 排序并取Top 10
        sortedWasted := s.sortWasteItems(allWasted)
        if len(sortedWasted) > 10 {
                summary.TopWasted = sortedWasted[:10]
        } else {
                summary.TopWasted = sortedWasted
        }

        // 计算潜在节省
        summary.PotentialSavings = summary.TotalWastedCost * 0.8 // 假设能节省80%
        summary.TotalResources = 100 // 模拟总资源数
        summary.SavingsPercent = (summary.TotalWastedCost / 50000) * 100 // 假设总成本50000

        // 生成建议
        summary.Recommendations = s.generateWasteRecommendations(summary)

        return summary, nil
}

// detectOverProvisioned 检测过度配置
func (s *WasteDetectionService) detectOverProvisioned(ctx context.Context) []WasteItem {
        items := make([]WasteItem, 0)

        // 模拟检测过度配置的实例
        overProvisioned := []struct {
                id           string
                name         string
                resourceType string
                allocated    float64
                actual       float64
                monthlyCost  float64
        }{
                {"ecs-large-001", "大型应用服务器", "ecs", 16, 4, 2000},
                {"rds-large-001", "大型数据库实例", "rds", 32, 8, 5000},
                {"ecs-med-002", "中等API服务器", "ecs", 8, 2, 800},
                {"redis-large-001", "大型Redis实例", "kvstore", 32, 6, 1500},
                {"ecs-large-002", "Worker服务器", "ecs", 16, 3, 1800},
        }

        for _, r := range overProvisioned {
                usagePercent := (r.actual / r.allocated) * 100
                if usagePercent < 30 { // 使用率低于30%认为过度配置
                        wastedCost := r.monthlyCost * (1 - usagePercent/100)
                        items = append(items, WasteItem{
                                ResourceID:     r.id,
                                ResourceName:   r.name,
                                ResourceType:   r.resourceType,
                                WasteType:      "overprovisioned",
                                WastedCost:     wastedCost,
                                MonthlyCost:    r.monthlyCost,
                                UsagePercent:   usagePercent,
                                Recommendation: fmt.Sprintf("建议降级到更小规格，预计节省 %.0f 元/月", wastedCost),
                                Priority:       s.calculatePriority(wastedCost),
                        })
                }
        }

        return items
}

// detectIdleResources 检测闲置资源
func (s *WasteDetectionService) detectIdleResources(ctx context.Context) []WasteItem {
        items := make([]WasteItem, 0)

        // 模拟检测闲置资源
        idleResources := []struct {
                id           string
                name         string
                resourceType string
                idleDays     int
                monthlyCost  float64
        }{
                {"ecs-idle-001", "测试服务器A", "ecs", 30, 500},
                {"ecs-idle-002", "旧版API服务器", "ecs", 45, 800},
                {"rds-idle-001", "测试数据库", "rds", 20, 1200},
                {"elb-idle-001", "未使用负载均衡", "slb", 60, 200},
                {"eip-idle-001", "未绑定EIP", "eip", 15, 50},
        }

        for _, r := range idleResources {
                if r.idleDays >= 7 { // 闲置超过7天
                        items = append(items, WasteItem{
                                ResourceID:     r.id,
                                ResourceName:   r.name,
                                ResourceType:   r.resourceType,
                                WasteType:      "idle",
                                WastedCost:     r.monthlyCost,
                                MonthlyCost:    r.monthlyCost,
                                UsagePercent:   0,
                                Recommendation: fmt.Sprintf("已闲置 %d 天，建议释放或停用", r.idleDays),
                                Priority:       s.calculatePriority(r.monthlyCost),
                        })
                }
        }

        return items
}

// detectUnattachedResources 检测未挂载资源
func (s *WasteDetectionService) detectUnattachedResources(ctx context.Context) []WasteItem {
        items := make([]WasteItem, 0)

        // 模拟检测未挂载资源
        unattached := []struct {
                id           string
                name         string
                resourceType string
                size         int // GB
                monthlyCost  float64
        }{
                {"ebs-unattached-001", "旧数据盘A", "ebs", 500, 100},
                {"ebs-unattached-002", "备份数据盘", "ebs", 1000, 200},
                {"ebs-unattached-003", "旧日志盘", "ebs", 200, 40},
                {"oss-unused-001", "废弃存储桶", "oss", 2000, 300},
        }

        for _, r := range unattached {
                items = append(items, WasteItem{
                        ResourceID:     r.id,
                        ResourceName:   r.name,
                        ResourceType:   r.resourceType,
                        WasteType:      "unattached",
                        WastedCost:     r.monthlyCost,
                        MonthlyCost:    r.monthlyCost,
                        UsagePercent:   0,
                        Recommendation: fmt.Sprintf("未挂载存储(%d GB)，建议删除或挂载使用", r.size),
                        Priority:       s.calculatePriority(r.monthlyCost),
                })
        }

        return items
}

// detectOldSnapshots 检测过期快照
func (s *WasteDetectionService) detectOldSnapshots(ctx context.Context) []WasteItem {
        items := make([]WasteItem, 0)

        // 模拟检测过期快照
        oldSnapshots := []struct {
                id           string
                name         string
                size         int // GB
                age          int // days
                monthlyCost  float64
        }{
                {"snap-old-001", "旧快照A", 100, 90, 20},
                {"snap-old-002", "旧快照B", 200, 120, 40},
                {"snap-old-003", "废弃快照", 500, 180, 100},
                {"snap-old-004", "测试快照", 50, 60, 10},
        }

        for _, r := range oldSnapshots {
                if r.age > 30 { // 超过30天的快照
                        items = append(items, WasteItem{
                                ResourceID:     r.id,
                                ResourceName:   r.name,
                                ResourceType:   "snapshot",
                                WasteType:      "snapshot",
                                WastedCost:     r.monthlyCost,
                                MonthlyCost:    r.monthlyCost,
                                UsagePercent:   0,
                                Recommendation: fmt.Sprintf("快照已存在 %d 天(%d GB)，建议删除", r.age, r.size),
                                Priority:       "low",
                        })
                }
        }

        return items
}

// detectUnusedBandwidth 检测未使用带宽
func (s *WasteDetectionService) detectUnusedBandwidth(ctx context.Context) []WasteItem {
        items := make([]WasteItem, 0)

        // 模拟检测带宽浪费
        bandwidthWaste := []struct {
                id           string
                name         string
                purchased    int // Mbps
                used         int // Mbps
                monthlyCost  float64
        }{
                {"bandwidth-001", "主出口带宽", 100, 20, 3000},
                {"bandwidth-002", "备用带宽", 50, 5, 1500},
        }

        for _, r := range bandwidthWaste {
                usagePercent := float64(r.used) / float64(r.purchased) * 100
                if usagePercent < 30 {
                        wastedCost := r.monthlyCost * (1 - usagePercent/100)
                        items = append(items, WasteItem{
                                ResourceID:     r.id,
                                ResourceName:   r.name,
                                ResourceType:   "bandwidth",
                                WasteType:      "bandwidth",
                                WastedCost:     wastedCost,
                                MonthlyCost:    r.monthlyCost,
                                UsagePercent:   usagePercent,
                                Recommendation: fmt.Sprintf("带宽使用率仅 %.0f%%，建议降配", usagePercent),
                                Priority:       "high",
                        })
                }
        }

        return items
}

// DetectEC2Waste 检测EC2浪费
func (s *WasteDetectionService) DetectEC2Waste(ctx context.Context) (*EC2WasteReport, error) {
        report := &EC2WasteReport{
                OverProvisioned: make([]InstanceWaste, 0),
                UnderUtilized:   make([]InstanceWaste, 0),
                Stopped:         make([]InstanceWaste, 0),
        }

        // 检测过度配置
        report.OverProvisioned = []InstanceWaste{
                {InstanceID: "i-001", Name: "App-Server-1", InstanceType: "c5.4xlarge", CPUUtil: 15, MemUtil: 20, MonthlyCost: 2000, RecommendedType: "c5.xlarge", Savings: 1500},
                {InstanceID: "i-002", Name: "DB-Server-1", InstanceType: "r5.2xlarge", CPUUtil: 10, MemUtil: 30, MonthlyCost: 1500, RecommendedType: "r5.large", Savings: 1000},
        }

        // 检测低利用率
        report.UnderUtilized = []InstanceWaste{
                {InstanceID: "i-003", Name: "Worker-1", InstanceType: "m5.2xlarge", CPUUtil: 5, MemUtil: 10, MonthlyCost: 800, RecommendedType: "m5.large", Savings: 400},
        }

        // 检测已停止
        report.Stopped = []InstanceWaste{
                {InstanceID: "i-004", Name: "Old-Server", InstanceType: "m5.xlarge", CPUUtil: 0, MemUtil: 0, MonthlyCost: 200, Savings: 200},
        }

        // 计算总节省
        for _, i := range report.OverProvisioned {
                report.TotalSavings += i.Savings
        }
        for _, i := range report.UnderUtilized {
                report.TotalSavings += i.Savings
        }
        for _, i := range report.Stopped {
                report.TotalSavings += i.Savings
        }

        return report, nil
}

// EC2WasteReport EC2浪费报告
type EC2WasteReport struct {
        OverProvisioned []InstanceWaste `json:"over_provisioned"`
        UnderUtilized   []InstanceWaste `json:"under_utilized"`
        Stopped         []InstanceWaste `json:"stopped"`
        TotalSavings    float64         `json:"total_savings"`
}

// InstanceWaste 实例浪费
type InstanceWaste struct {
        InstanceID      string  `json:"instance_id"`
        Name            string  `json:"name"`
        InstanceType    string  `json:"instance_type"`
        CPUUtil         float64 `json:"cpu_util"`
        MemUtil         float64 `json:"mem_util"`
        MonthlyCost     float64 `json:"monthly_cost"`
        RecommendedType string  `json:"recommended_type"`
        Savings         float64 `json:"savings"`
}

// DetectRDSWaste 检测RDS浪费
func (s *WasteDetectionService) DetectRDSWaste(ctx context.Context) (*RDSWasteReport, error) {
        report := &RDSWasteReport{
                OverProvisioned: make([]DBInstanceWaste, 0),
                Idle:            make([]DBInstanceWaste, 0),
                UnusedReplicas:  make([]DBInstanceWaste, 0),
        }

        report.OverProvisioned = []DBInstanceWaste{
                {InstanceID: "db-001", Name: "主数据库", InstanceClass: "db.r5.4xlarge", CPUUtil: 10, StorageUsed: 100, StorageAllocated: 500, MonthlyCost: 5000, Savings: 3000},
        }

        report.Idle = []DBInstanceWaste{
                {InstanceID: "db-002", Name: "测试数据库", InstanceClass: "db.t3.medium", CPUUtil: 0, StorageUsed: 10, StorageAllocated: 100, MonthlyCost: 200, Savings: 200},
        }

        report.UnusedReplicas = []DBInstanceWaste{
                {InstanceID: "db-003", Name: "只读副本1", InstanceClass: "db.r5.large", CPUUtil: 2, MonthlyCost: 500, Savings: 500},
        }

        return report, nil
}

// RDSWasteReport RDS浪费报告
type RDSWasteReport struct {
        OverProvisioned []DBInstanceWaste `json:"over_provisioned"`
        Idle            []DBInstanceWaste `json:"idle"`
        UnusedReplicas  []DBInstanceWaste `json:"unused_replicas"`
        TotalSavings    float64           `json:"total_savings"`
}

// DBInstanceWaste 数据库实例浪费
type DBInstanceWaste struct {
        InstanceID       string  `json:"instance_id"`
        Name             string  `json:"name"`
        InstanceClass    string  `json:"instance_class"`
        CPUUtil          float64 `json:"cpu_util"`
        StorageUsed      int     `json:"storage_used"`
        StorageAllocated int     `json:"storage_allocated"`
        MonthlyCost      float64 `json:"monthly_cost"`
        Savings          float64 `json:"savings"`
}

// DetectStorageWaste 检测存储浪费
func (s *WasteDetectionService) DetectStorageWaste(ctx context.Context) (*StorageWasteReport, error) {
        report := &StorageWasteReport{
                UnattachedVolumes: make([]VolumeWaste, 0),
                LowUtilVolumes:    make([]VolumeWaste, 0),
                OldSnapshots:      make([]SnapshotWaste, 0),
        }

        report.UnattachedVolumes = []VolumeWaste{
                {VolumeID: "vol-001", Name: "旧数据盘", Size: 500, IOPS: 3000, MonthlyCost: 100, DaysUnattached: 30},
                {VolumeID: "vol-002", Name: "备份盘", Size: 1000, IOPS: 5000, MonthlyCost: 200, DaysUnattached: 60},
        }

        report.LowUtilVolumes = []VolumeWaste{
                {VolumeID: "vol-003", Name: "日志盘", Size: 200, IOPS: 1000, MonthlyCost: 40, IOPSUtil: 5, ThroughputUtil: 3},
        }

        report.OldSnapshots = []SnapshotWaste{
                {SnapshotID: "snap-001", Name: "旧快照", Size: 100, Age: 90, MonthlyCost: 20},
        }

        return report, nil
}

// StorageWasteReport 存储浪费报告
type StorageWasteReport struct {
        UnattachedVolumes []VolumeWaste   `json:"unattached_volumes"`
        LowUtilVolumes    []VolumeWaste   `json:"low_util_volumes"`
        OldSnapshots      []SnapshotWaste `json:"old_snapshots"`
        TotalSavings      float64         `json:"total_savings"`
}

// VolumeWaste 卷浪费
type VolumeWaste struct {
        VolumeID        string  `json:"volume_id"`
        Name            string  `json:"name"`
        Size            int     `json:"size"` // GB
        IOPS            int     `json:"iops"`
        MonthlyCost     float64 `json:"monthly_cost"`
        DaysUnattached  int     `json:"days_unattached"`
        IOPSUtil        float64 `json:"iops_util"`
        ThroughputUtil  float64 `json:"throughput_util"`
}

// SnapshotWaste 快照浪费
type SnapshotWaste struct {
        SnapshotID  string  `json:"snapshot_id"`
        Name        string  `json:"name"`
        Size        int     `json:"size"` // GB
        Age         int     `json:"age"`  // days
        MonthlyCost float64 `json:"monthly_cost"`
}

// calculatePriority 计算优先级
func (s *WasteDetectionService) calculatePriority(cost float64) string {
        if cost >= 1000 {
                return "high"
        } else if cost >= 500 {
                return "medium"
        }
        return "low"
}

// sortWasteItems 排序浪费项
func (s *WasteDetectionService) sortWasteItems(items []WasteItem) []WasteItem {
        // 按浪费成本降序排序
        for i := 0; i < len(items); i++ {
                for j := i + 1; j < len(items); j++ {
                        if items[i].WastedCost < items[j].WastedCost {
                                items[i], items[j] = items[j], items[i]
                        }
                }
        }
        return items
}

// generateWasteRecommendations 生成浪费建议
func (s *WasteDetectionService) generateWasteRecommendations(summary *WasteSummary) []string {
        recommendations := make([]string, 0)

        if summary.TotalWastedCost > 10000 {
                recommendations = append(recommendations, fmt.Sprintf("总浪费成本 %.0f 元，建议优先处理高优先级资源", summary.TotalWastedCost))
        }

        for t, w := range summary.ByType {
                if w.WastedCost > 1000 {
                        recommendations = append(recommendations, fmt.Sprintf("%s 类型浪费严重(%.0f 元)，建议重点优化", t, w.WastedCost))
                }
        }

        recommendations = append(recommendations, "建议实施自动化资源生命周期管理")
        recommendations = append(recommendations, "考虑使用预留实例或Spot实例降低成本")

        return recommendations
}

// CreateWasteRecord 创建浪费记录
func (s *WasteDetectionService) CreateWasteRecord(ctx context.Context, item WasteItem) *cost.WasteDetection {
        record := &cost.WasteDetection{
                ResourceID:    item.ResourceID,
                ResourceName:  item.ResourceName,
                ResourceType:  item.ResourceType,
                WasteType:     item.WasteType,
                MetricValue:   item.MonthlyCost,
                MonthlyWasteCost: item.WastedCost,
                Recommendation: item.Recommendation,
                RiskLevel:     item.Priority,
        }

        return record
}

// CalculateSavings 计算节省
func (s *WasteDetectionService) CalculateSavings(ctx context.Context, wasteItems []WasteItem) *SavingsCalculation {
        calc := &SavingsCalculation{
                ByAction: make(map[string]ActionSavings),
        }

        totalSavings := 0.0
        for _, item := range wasteItems {
                savings := item.WastedCost * 0.8 // 假设优化后能节省80%
                totalSavings += savings

                action := "optimize"
                if item.WasteType == "idle" || item.WasteType == "unattached" {
                        action = "terminate"
                } else if item.WasteType == "overprovisioned" {
                        action = "resize"
                }

                as, ok := calc.ByAction[action]
                if !ok {
                        as = ActionSavings{Action: action}
                }
                as.Savings += savings
                as.ResourceCount++
                calc.ByAction[action] = as
        }

        calc.TotalSavings = totalSavings
        calc.MonthlySavings = totalSavings
        calc.YearlySavings = totalSavings * 12

        return calc
}

// SavingsCalculation 节省计算
type SavingsCalculation struct {
        TotalSavings    float64                  `json:"total_savings"`
        MonthlySavings  float64                  `json:"monthly_savings"`
        YearlySavings   float64                  `json:"yearly_savings"`
        ByAction        map[string]ActionSavings `json:"by_action"`
}

// ActionSavings 行动节省
type ActionSavings struct {
        Action        string  `json:"action"`
        Savings       float64 `json:"savings"`
        ResourceCount int     `json:"resource_count"`
}

// OptimizeResource 优化资源
func (s *WasteDetectionService) OptimizeResource(ctx context.Context, resourceID string, action string) (*OptimizationResult, error) {
        result := &OptimizationResult{
                ResourceID: resourceID,
                Action:     action,
                StartTime:  time.Now(),
        }

        // 模拟优化操作
        switch action {
        case "resize":
                result.Status = "success"
                result.Savings = 500.0
                result.Message = "实例已成功降配"
        case "terminate":
                result.Status = "success"
                result.Savings = 800.0
                result.Message = "资源已成功释放"
        case "optimize":
                result.Status = "success"
                result.Savings = 300.0
                result.Message = "资源已优化配置"
        default:
                result.Status = "failed"
                result.Message = "未知操作类型"
        }

        result.EndTime = time.Now()
        return result, nil
}

// OptimizationResult 优化结果
type OptimizationResult struct {
        ResourceID string    `json:"resource_id"`
        Action     string    `json:"action"`
        Status     string    `json:"status"`
        Savings    float64   `json:"savings"`
        Message    string    `json:"message"`
        StartTime  time.Time `json:"start_time"`
        EndTime    time.Time `json:"end_time"`
}
