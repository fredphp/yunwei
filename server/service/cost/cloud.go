package cost

import (
        "context"
        "encoding/json"
        "fmt"
        "sync"
        "time"

        "yunwei/model/cost"
)

// CloudCostService 云成本分析服务
type CloudCostService struct {
        mu sync.RWMutex
}

// NewCloudCostService 创建云成本服务
func NewCloudCostService() *CloudCostService {
        return &CloudCostService{}
}

// TrendPoint 趋势点
type TrendPoint struct {
        Date  string  `json:"date"`
        Cost  float64 `json:"cost"`
        Label string  `json:"label"`
}

// CloudCostSummary 云成本摘要
type CloudCostSummary struct {
        Provider       string             `json:"provider"`
        TotalCost      float64            `json:"total_cost"`
        PreviousCost   float64            `json:"previous_cost"`
        ChangePercent  float64            `json:"change_percent"`
        ByRegion       map[string]float64 `json:"by_region"`
        ByService      map[string]float64 `json:"by_service"`
        ByInstanceType map[string]float64 `json:"by_instance_type"`
        TopResources   []CloudResource    `json:"top_resources"`
        Trend          []TrendPoint       `json:"trend"`
        Savings        CloudSavings       `json:"savings"`
}

// CloudResource 云资源
type CloudResource struct {
        ResourceID   string  `json:"resource_id"`
        ResourceName string  `json:"resource_name"`
        ResourceType string  `json:"resource_type"`
        Region       string  `json:"region"`
        MonthlyCost  float64 `json:"monthly_cost"`
        Status       string  `json:"status"`
}

// CloudSavings 云节省
type CloudSavings struct {
        ReservedSavings  float64 `json:"reserved_savings"`
        SpotSavings      float64 `json:"spot_savings"`
        RightSizeSavings float64 `json:"right_size_savings"`
        TotalSavings     float64 `json:"total_savings"`
}

// AnalyzeAliyunCost 分析阿里云成本
func (s *CloudCostService) AnalyzeAliyunCost(ctx context.Context, accountID string, startTime, endTime time.Time) (*CloudCostSummary, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        summary := &CloudCostSummary{
                Provider:       "aliyun",
                ByRegion:       make(map[string]float64),
                ByService:      make(map[string]float64),
                ByInstanceType: make(map[string]float64),
                TopResources:   make([]CloudResource, 0),
                Trend:          make([]TrendPoint, 0),
        }

        // 模拟数据
        summary.TotalCost = 25000
        summary.PreviousCost = 22000
        summary.ChangePercent = (summary.TotalCost - summary.PreviousCost) / summary.PreviousCost * 100

        // 按区域
        summary.ByRegion = map[string]float64{
                "cn-hangzhou": 10000,
                "cn-shanghai": 8000,
                "cn-beijing":  5000,
                "cn-shenzhen": 2000,
        }

        // 按服务
        summary.ByService = map[string]float64{
                "ecs": 12000,
                "rds": 6000,
                "oss": 3000,
                "slb": 2000,
                "other": 2000,
        }

        // 按实例类型
        summary.ByInstanceType = map[string]float64{
                "ecs.g6.large":    5000,
                "ecs.g6.xlarge":   4000,
                "ecs.c6.large":    3000,
                "rds.mysql.s2":    6000,
                "other":           7000,
        }

        // Top资源
        summary.TopResources = []CloudResource{
                {ResourceID: "ecs-001", ResourceName: "主应用服务器", ResourceType: "ecs", Region: "cn-hangzhou", MonthlyCost: 3000, Status: "running"},
                {ResourceID: "rds-001", ResourceName: "主数据库", ResourceType: "rds", Region: "cn-hangzhou", MonthlyCost: 2500, Status: "running"},
                {ResourceID: "ecs-002", ResourceName: "API服务器", ResourceType: "ecs", Region: "cn-shanghai", MonthlyCost: 2000, Status: "running"},
        }

        // 趋势
        for i := 0; i < 7; i++ {
                d := startTime.AddDate(0, 0, i)
                summary.Trend = append(summary.Trend, TrendPoint{
                        Date:  d.Format("2006-01-02"),
                        Cost:  800 + float64(i)*50,
                        Label: d.Format("01-02"),
                })
        }

        // 节省建议
        summary.Savings = CloudSavings{
                ReservedSavings:  3000,
                SpotSavings:      1000,
                RightSizeSavings: 2000,
                TotalSavings:     6000,
        }

        return summary, nil
}

// AnalyzeAWSCost 分析AWS成本
func (s *CloudCostService) AnalyzeAWSCost(ctx context.Context, accountID string, startTime, endTime time.Time) (*CloudCostSummary, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        summary := &CloudCostSummary{
                Provider:       "aws",
                ByRegion:       make(map[string]float64),
                ByService:      make(map[string]float64),
                ByInstanceType: make(map[string]float64),
                TopResources:   make([]CloudResource, 0),
                Trend:          make([]TrendPoint, 0),
        }

        summary.TotalCost = 15000
        summary.PreviousCost = 14000
        summary.ChangePercent = (summary.TotalCost - summary.PreviousCost) / summary.PreviousCost * 100

        summary.ByRegion = map[string]float64{
                "us-east-1":      6000,
                "us-west-2":      4000,
                "eu-west-1":      3000,
                "ap-northeast-1": 2000,
        }

        summary.ByService = map[string]float64{
                "ec2":    7000,
                "rds":    4000,
                "s3":     2000,
                "cloudwatch": 1000,
                "other":  1000,
        }

        summary.Savings = CloudSavings{
                ReservedSavings:  2500,
                SpotSavings:      1500,
                RightSizeSavings: 1000,
                TotalSavings:     5000,
        }

        return summary, nil
}

// AnalyzeTencentCost 分析腾讯云成本
func (s *CloudCostService) AnalyzeTencentCost(ctx context.Context, accountID string, startTime, endTime time.Time) (*CloudCostSummary, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        summary := &CloudCostSummary{
                Provider:       "tencent",
                ByRegion:       make(map[string]float64),
                ByService:      make(map[string]float64),
                ByInstanceType: make(map[string]float64),
                TopResources:   make([]CloudResource, 0),
        }

        summary.TotalCost = 8000
        summary.PreviousCost = 7500
        summary.ChangePercent = (summary.TotalCost - summary.PreviousCost) / summary.PreviousCost * 100

        summary.ByRegion = map[string]float64{
                "ap-guangzhou": 4000,
                "ap-shanghai":  2500,
                "ap-beijing":   1500,
        }

        summary.ByService = map[string]float64{
                "cvm":   4000,
                "cdb":   2500,
                "cos":   1000,
                "other": 500,
        }

        return summary, nil
}

// GetMultiCloudCost 获取多云成本对比
func (s *CloudCostService) GetMultiCloudCost(ctx context.Context, startTime, endTime time.Time) (*MultiCloudSummary, error) {
        summary := &MultiCloudSummary{
                ByProvider: make(map[string]ProviderCost),
                Trend:      make([]MultiCloudTrend, 0),
        }

        // 获取各云商成本
        aliyun, _ := s.AnalyzeAliyunCost(ctx, "", startTime, endTime)
        aws, _ := s.AnalyzeAWSCost(ctx, "", startTime, endTime)
        tencent, _ := s.AnalyzeTencentCost(ctx, "", startTime, endTime)

        summary.ByProvider["aliyun"] = ProviderCost{
                Cost:          aliyun.TotalCost,
                ChangePercent: aliyun.ChangePercent,
                ResourceCount: 50,
        }
        summary.ByProvider["aws"] = ProviderCost{
                Cost:          aws.TotalCost,
                ChangePercent: aws.ChangePercent,
                ResourceCount: 30,
        }
        summary.ByProvider["tencent"] = ProviderCost{
                Cost:          tencent.TotalCost,
                ChangePercent: tencent.ChangePercent,
                ResourceCount: 20,
        }

        summary.TotalCost = aliyun.TotalCost + aws.TotalCost + tencent.TotalCost

        // 生成趋势
        for i := 0; i < 7; i++ {
                d := startTime.AddDate(0, 0, i)
                summary.Trend = append(summary.Trend, MultiCloudTrend{
                        Date:    d.Format("2006-01-02"),
                        Aliyun:  800 + float64(i)*30,
                        AWS:     500 + float64(i)*20,
                        Tencent: 300 + float64(i)*10,
                })
        }

        return summary, nil
}

// MultiCloudSummary 多云摘要
type MultiCloudSummary struct {
        TotalCost  float64                    `json:"total_cost"`
        ByProvider map[string]ProviderCost    `json:"by_provider"`
        Trend      []MultiCloudTrend          `json:"trend"`
}

// ProviderCost 云商成本
type ProviderCost struct {
        Cost          float64 `json:"cost"`
        ChangePercent float64 `json:"change_percent"`
        ResourceCount int     `json:"resource_count"`
}

// MultiCloudTrend 多云趋势
type MultiCloudTrend struct {
        Date    string  `json:"date"`
        Aliyun  float64 `json:"aliyun"`
        AWS     float64 `json:"aws"`
        Tencent float64 `json:"tencent"`
}

// GetReservedInstanceRecommendation 获取预留实例推荐
func (s *CloudCostService) GetReservedInstanceRecommendation(ctx context.Context, provider string) ([]ReservedInstanceRecommendation, error) {
        recommendations := make([]ReservedInstanceRecommendation, 0)

        // 模拟推荐
        recommendations = []ReservedInstanceRecommendation{
                {
                        Provider:       "aliyun",
                        InstanceType:   "ecs.g6.xlarge",
                        Region:         "cn-hangzhou",
                        RunningCount:   5,
                        OnDemandCost:   1200,
                        ReservedCost:   720,
                        MonthlySavings: 480,
                        YearlySavings:  5760,
                        Utilization:    95,
                },
                {
                        Provider:       "aws",
                        InstanceType:   "m5.xlarge",
                        Region:         "us-east-1",
                        RunningCount:   3,
                        OnDemandCost:   140,
                        ReservedCost:   85,
                        MonthlySavings: 165,
                        YearlySavings:  1980,
                        Utilization:    90,
                },
        }

        return recommendations, nil
}

// ReservedInstanceRecommendation 预留实例推荐
type ReservedInstanceRecommendation struct {
        Provider       string  `json:"provider"`
        InstanceType   string  `json:"instance_type"`
        Region         string  `json:"region"`
        RunningCount   int     `json:"running_count"`
        OnDemandCost   float64 `json:"on_demand_cost"`   // 月按需成本
        ReservedCost   float64 `json:"reserved_cost"`    // 月预留成本
        MonthlySavings float64 `json:"monthly_savings"`
        YearlySavings  float64 `json:"yearly_savings"`
        Utilization    float64 `json:"utilization"` // 利用率
}

// GetSpotInstanceRecommendation 获取Spot实例推荐
func (s *CloudCostService) GetSpotInstanceRecommendation(ctx context.Context, provider string) ([]SpotInstanceRecommendation, error) {
        recommendations := []SpotInstanceRecommendation{
                {
                        Provider:        "aws",
                        InstanceType:    "m5.xlarge",
                        Region:          "us-east-1",
                        OnDemandPrice:   0.192,
                        SpotPrice:       0.058,
                        Discount:        70,
                        InterruptionRate: 5,
                        Recommended:      true,
                },
                {
                        Provider:        "aliyun",
                        InstanceType:    "ecs.g6.large",
                        Region:          "cn-hangzhou",
                        OnDemandPrice:   0.35,
                        SpotPrice:       0.10,
                        Discount:        71,
                        InterruptionRate: 8,
                        Recommended:      true,
                },
        }

        return recommendations, nil
}

// SpotInstanceRecommendation Spot实例推荐
type SpotInstanceRecommendation struct {
        Provider         string  `json:"provider"`
        InstanceType     string  `json:"instance_type"`
        Region           string  `json:"region"`
        OnDemandPrice    float64 `json:"on_demand_price"`    // 小时按需价格
        SpotPrice        float64 `json:"spot_price"`         // 小时Spot价格
        Discount         float64 `json:"discount"`           // 折扣百分比
        InterruptionRate float64 `json:"interruption_rate"`  // 中断率
        Recommended      bool    `json:"recommended"`
}

// SyncCloudBill 同步云账单
func (s *CloudCostService) SyncCloudBill(ctx context.Context, accountID string, provider string) (*SyncResult, error) {
        s.mu.Lock()
        defer s.mu.Unlock()

        result := &SyncResult{
                AccountID: accountID,
                Provider:  provider,
                StartTime: time.Now(),
        }

        // 模拟同步
        result.RecordsSynced = 150
        result.NewRecords = 25
        result.Status = "completed"
        result.EndTime = time.Now()

        return result, nil
}

// SyncResult 同步结果
type SyncResult struct {
        AccountID     string    `json:"account_id"`
        Provider      string    `json:"provider"`
        StartTime     time.Time `json:"start_time"`
        EndTime       time.Time `json:"end_time"`
        RecordsSynced int       `json:"records_synced"`
        NewRecords    int       `json:"new_records"`
        Status        string    `json:"status"`
        Error         string    `json:"error,omitempty"`
}

// CreateCostRecord 创建成本记录
func (s *CloudCostService) CreateCostRecord(ctx context.Context, record *cost.CostRecord) error {
        // 实际应保存到数据库
        return nil
}

// GetCostRecords 获取成本记录
func (s *CloudCostService) GetCostRecords(ctx context.Context, provider string, startTime, endTime time.Time) ([]cost.CostRecord, error) {
        records := make([]cost.CostRecord, 0)

        // 模拟数据
        records = []cost.CostRecord{
                {
                        Provider:     provider,
                        ResourceType: "ecs",
                        ResourceID:   "ecs-001",
                        ResourceName: "主应用服务器",
                        NetCost:      1000,
                        RecordDate:   startTime,
                },
                {
                        Provider:     provider,
                        ResourceType: "rds",
                        ResourceID:   "rds-001",
                        ResourceName: "主数据库",
                        NetCost:      800,
                        RecordDate:   startTime,
                },
        }

        return records, nil
}
