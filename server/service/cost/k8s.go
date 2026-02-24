package cost

import (
        "context"
        "encoding/json"
        "fmt"
        "sync"
        "time"

        "yunwei/model/cost"
)

// K8sCostService K8s成本分析服务
type K8sCostService struct {
        mu sync.RWMutex
}

// NewK8sCostService 创建K8s成本服务
func NewK8sCostService() *K8sCostService {
        return &K8sCostService{}
}

// K8sCostSummary K8s成本摘要
type K8sCostSummary struct {
        ClusterID         string                   `json:"cluster_id"`
        ClusterName       string                   `json:"cluster_name"`
        TotalCost         float64                  `json:"total_cost"`
        ByNamespace       map[string]float64       `json:"by_namespace"`
        ByWorkloadType    map[string]float64       `json:"by_workload_type"`
        ByResourceType    map[string]float64       `json:"by_resource_type"`
        TopWorkloads      []WorkloadCost           `json:"top_workloads"`
        IdleResources     []IdleWorkload           `json:"idle_resources"`
        OptimizationOpportunities []K8sOptimization `json:"optimization_opportunities"`
        Efficiency        K8sEfficiency            `json:"efficiency"`
}

// WorkloadCost 工作负载成本
type WorkloadCost struct {
        Namespace      string  `json:"namespace"`
        Name           string  `json:"name"`
        Type           string  `json:"type"`
        CPUCost        float64 `json:"cpu_cost"`
        MemoryCost     float64 `json:"memory_cost"`
        StorageCost    float64 `json:"storage_cost"`
        NetworkCost    float64 `json:"network_cost"`
        TotalCost      float64 `json:"total_cost"`
        CPURequest     float64 `json:"cpu_request"`
        MemoryRequest  int64   `json:"memory_request"`
        CPUUsage       float64 `json:"cpu_usage"`
        MemoryUsage    int64   `json:"memory_usage"`
        CPUEfficiency  float64 `json:"cpu_efficiency"`
        MemoryEfficiency float64 `json:"memory_efficiency"`
}

// IdleWorkload 闲置工作负载
type IdleWorkload struct {
        Namespace     string  `json:"namespace"`
        Name          string  `json:"name"`
        Type          string  `json:"type"`
        CPURequest    float64 `json:"cpu_request"`
        MemoryRequest int64   `json:"memory_request"`
        CPUUsage      float64 `json:"cpu_usage"`
        MemoryUsage   int64   `json:"memory_usage"`
        MonthlyCost   float64 `json:"monthly_cost"`
        WastedCost    float64 `json:"wasted_cost"`
        IdleDays      int     `json:"idle_days"`
}

// K8sOptimization K8s优化建议
type K8sOptimization struct {
        Namespace      string  `json:"namespace"`
        Name           string  `json:"name"`
        Type           string  `json:"type"`
        CurrentCPU     float64 `json:"current_cpu"`
        RecommendedCPU float64 `json:"recommended_cpu"`
        CurrentMemory  int64   `json:"current_memory"`
        RecommendedMemory int64 `json:"recommended_memory"`
        MonthlySavings float64 `json:"monthly_savings"`
        Action         string  `json:"action"` // resize, scale_down, consolidate
}

// K8sEfficiency K8s效率
type K8sEfficiency struct {
        OverallEfficiency float64 `json:"overall_efficiency"`
        CPUEfficiency     float64 `json:"cpu_efficiency"`
        MemoryEfficiency  float64 `json:"memory_efficiency"`
        WastedCPU         float64 `json:"wasted_cpu"`       // 浪费的CPU核数
        WastedMemory      int64   `json:"wasted_memory"`    // 浪费的内存(字节)
        WastedCost        float64 `json:"wasted_cost"`
}

// AnalyzeClusterCost 分析集群成本
func (s *K8sCostService) AnalyzeClusterCost(ctx context.Context, clusterID string) (*K8sCostSummary, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        summary := &K8sCostSummary{
                ClusterID:         clusterID,
                ClusterName:       "Production Cluster",
                ByNamespace:       make(map[string]float64),
                ByWorkloadType:    make(map[string]float64),
                ByResourceType:    make(map[string]float64),
                TopWorkloads:      make([]WorkloadCost, 0),
                IdleResources:     make([]IdleWorkload, 0),
                OptimizationOpportunities: make([]K8sOptimization, 0),
        }

        // 模拟数据
        summary.TotalCost = 12000

        // 按命名空间成本
        summary.ByNamespace = map[string]float64{
                "default":     4000,
                "monitoring":  2000,
                "logging":     1500,
                "istio-system": 1500,
                "kube-system": 1000,
                "other":       2000,
        }

        // 按工作负载类型
        summary.ByWorkloadType = map[string]float64{
                "deployment":  7000,
                "statefulset": 3000,
                "daemonset":   1500,
                "job":         500,
        }

        // 按资源类型
        summary.ByResourceType = map[string]float64{
                "cpu":     5000,
                "memory":  4000,
                "storage": 2000,
                "network": 1000,
        }

        // Top工作负载
        summary.TopWorkloads = []WorkloadCost{
                {
                        Namespace:        "default",
                        Name:             "api-server",
                        Type:             "deployment",
                        CPUCost:          1500,
                        MemoryCost:       1200,
                        TotalCost:        2700,
                        CPURequest:       4.0,
                        MemoryRequest:    8 * 1024 * 1024 * 1024,
                        CPUUsage:         2.5,
                        MemoryUsage:      5 * 1024 * 1024 * 1024,
                        CPUEfficiency:    62.5,
                        MemoryEfficiency: 62.5,
                },
                {
                        Namespace:        "monitoring",
                        Name:             "prometheus",
                        Type:             "statefulset",
                        CPUCost:          800,
                        MemoryCost:       1000,
                        TotalCost:        1800,
                        CPURequest:       2.0,
                        MemoryRequest:    16 * 1024 * 1024 * 1024,
                        CPUUsage:         1.2,
                        MemoryUsage:      10 * 1024 * 1024 * 1024,
                        CPUEfficiency:    60.0,
                        MemoryEfficiency: 62.5,
                },
        }

        // 闲置资源
        summary.IdleResources = []IdleWorkload{
                {
                        Namespace:     "test",
                        Name:          "test-api",
                        Type:          "deployment",
                        CPURequest:    2.0,
                        MemoryRequest: 4 * 1024 * 1024 * 1024,
                        CPUUsage:      0.1,
                        MemoryUsage:   200 * 1024 * 1024,
                        MonthlyCost:   500,
                        WastedCost:    450,
                        IdleDays:      15,
                },
        }

        // 优化建议
        summary.OptimizationOpportunities = []K8sOptimization{
                {
                        Namespace:       "default",
                        Name:            "api-server",
                        Type:            "deployment",
                        CurrentCPU:      4.0,
                        RecommendedCPU:  2.5,
                        CurrentMemory:   8 * 1024 * 1024 * 1024,
                        RecommendedMemory: 5 * 1024 * 1024 * 1024,
                        MonthlySavings:  800,
                        Action:          "resize",
                },
                {
                        Namespace:       "test",
                        Name:            "test-api",
                        Type:            "deployment",
                        CurrentCPU:      2.0,
                        RecommendedCPU:  0,
                        CurrentMemory:   4 * 1024 * 1024 * 1024,
                        RecommendedMemory: 0,
                        MonthlySavings:  500,
                        Action:          "scale_down",
                },
        }

        // 效率统计
        summary.Efficiency = K8sEfficiency{
                OverallEfficiency: 65.0,
                CPUEfficiency:     62.5,
                MemoryEfficiency:  68.0,
                WastedCPU:         15.0,  // 15核
                WastedMemory:      50 * 1024 * 1024 * 1024, // 50GB
                WastedCost:        3500,
        }

        return summary, nil
}

// GetNamespaceCostAnalysis 获取命名空间成本分析
func (s *K8sCostService) GetNamespaceCostAnalysis(ctx context.Context, clusterID, namespace string) (*NamespaceCostAnalysis, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        analysis := &NamespaceCostAnalysis{
                ClusterID:   clusterID,
                Namespace:   namespace,
                TotalCost:   4000,
                ByWorkload:  make(map[string]float64),
                ByLabel:     make(map[string]float64),
                Workloads:   make([]WorkloadCost, 0),
                Trend:       make([]TrendPoint, 0),
        }

        analysis.ByWorkload = map[string]float64{
                "api-server":    1500,
                "worker":        1200,
                "scheduler":     800,
                "other":         500,
        }

        analysis.ByLabel = map[string]float64{
                "app=api":      1500,
                "app=worker":   1200,
                "app=scheduler": 800,
        }

        analysis.Workloads = []WorkloadCost{
                {Namespace: namespace, Name: "api-server", Type: "deployment", TotalCost: 1500},
                {Namespace: namespace, Name: "worker", Type: "deployment", TotalCost: 1200},
                {Namespace: namespace, Name: "scheduler", Type: "deployment", TotalCost: 800},
        }

        return analysis, nil
}

// NamespaceCostAnalysis 命名空间成本分析
type NamespaceCostAnalysis struct {
        ClusterID   string              `json:"cluster_id"`
        Namespace   string              `json:"namespace"`
        TotalCost   float64             `json:"total_cost"`
        ByWorkload  map[string]float64  `json:"by_workload"`
        ByLabel     map[string]float64  `json:"by_label"`
        Workloads   []WorkloadCost      `json:"workloads"`
        Trend       []TrendPoint        `json:"trend"`
}

// GetPodCostAnalysis 获取Pod成本分析
func (s *K8sCostService) GetPodCostAnalysis(ctx context.Context, clusterID, namespace string) (*PodCostAnalysis, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        analysis := &PodCostAnalysis{
                ClusterID: clusterID,
                Namespace: namespace,
                Pods:      make([]PodCost, 0),
        }

        // 模拟Pod成本数据
        analysis.Pods = []PodCost{
                {
                        PodName:       "api-server-abc123",
                        WorkloadName:  "api-server",
                        NodeName:      "node-1",
                        CPURequest:    500,    // millicores
                        CPULimit:      1000,
                        MemoryRequest: 512 * 1024 * 1024,
                        MemoryLimit:   1 * 1024 * 1024 * 1024,
                        CPUUsage:      300,
                        MemoryUsage:   400 * 1024 * 1024,
                        HourlyCost:    0.15,
                        MonthlyCost:   108,
                },
                {
                        PodName:       "worker-xyz789",
                        WorkloadName:  "worker",
                        NodeName:      "node-2",
                        CPURequest:    1000,
                        CPULimit:      2000,
                        MemoryRequest: 1 * 1024 * 1024 * 1024,
                        MemoryLimit:   2 * 1024 * 1024 * 1024,
                        CPUUsage:      600,
                        MemoryUsage:   800 * 1024 * 1024,
                        HourlyCost:    0.25,
                        MonthlyCost:   180,
                },
        }

        return analysis, nil
}

// PodCostAnalysis Pod成本分析
type PodCostAnalysis struct {
        ClusterID string    `json:"cluster_id"`
        Namespace string    `json:"namespace"`
        Pods      []PodCost `json:"pods"`
}

// PodCost Pod成本
type PodCost struct {
        PodName       string  `json:"pod_name"`
        WorkloadName  string  `json:"workload_name"`
        NodeName      string  `json:"node_name"`
        CPURequest    int64   `json:"cpu_request"`     // millicores
        CPULimit      int64   `json:"cpu_limit"`
        MemoryRequest int64   `json:"memory_request"`
        MemoryLimit   int64   `json:"memory_limit"`
        CPUUsage      int64   `json:"cpu_usage"`
        MemoryUsage   int64   `json:"memory_usage"`
        HourlyCost    float64 `json:"hourly_cost"`
        MonthlyCost   float64 `json:"monthly_cost"`
}

// GetPVCCostAnalysis 获取PVC成本分析
func (s *K8sCostService) GetPVCCostAnalysis(ctx context.Context, clusterID string) (*PVCCostAnalysis, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        analysis := &PVCCostAnalysis{
                ClusterID: clusterID,
                PVCs:      make([]PVCCost, 0),
        }

        analysis.PVCs = []PVCCost{
                {
                        Name:          "data-pvc",
                        Namespace:     "default",
                        StorageClass:  "ssd",
                        Capacity:      100 * 1024 * 1024 * 1024, // 100GB
                        Used:          60 * 1024 * 1024 * 1024,
                        MonthlyCost:   20,
                        BoundPod:      "api-server-abc123",
                },
                {
                        Name:          "logs-pvc",
                        Namespace:     "logging",
                        StorageClass:  "standard",
                        Capacity:      200 * 1024 * 1024 * 1024,
                        Used:          50 * 1024 * 1024 * 1024,
                        MonthlyCost:   16,
                        BoundPod:      "fluentd-xyz",
                },
        }

        analysis.TotalCost = 36
        analysis.TotalCapacity = 300 * 1024 * 1024 * 1024
        analysis.TotalUsed = 110 * 1024 * 1024 * 1024

        return analysis, nil
}

// PVCCostAnalysis PVC成本分析
type PVCCostAnalysis struct {
        ClusterID     string    `json:"cluster_id"`
        PVCs          []PVCCost `json:"pvcs"`
        TotalCost     float64   `json:"total_cost"`
        TotalCapacity int64     `json:"total_capacity"`
        TotalUsed     int64     `json:"total_used"`
}

// PVCCost PVC成本
type PVCCost struct {
        Name         string  `json:"name"`
        Namespace    string  `json:"namespace"`
        StorageClass string  `json:"storage_class"`
        Capacity     int64   `json:"capacity"`
        Used         int64   `json:"used"`
        MonthlyCost  float64 `json:"monthly_cost"`
        BoundPod     string  `json:"bound_pod"`
}

// CalculateResourceEfficiency 计算资源效率
func (s *K8sCostService) CalculateResourceEfficiency(ctx context.Context, clusterID string) (*ResourceEfficiency, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        efficiency := &ResourceEfficiency{
                ClusterID: clusterID,
        }

        // CPU效率
        efficiency.CPURequested = 50.0    // 请求50核
        efficiency.CPUUsed = 30.0         // 实际使用30核
        efficiency.CPULimit = 100.0       // 限制100核
        efficiency.CPUEfficiency = efficiency.CPUUsed / efficiency.CPURequested * 100

        // 内存效率
        efficiency.MemoryRequested = 100 * 1024 * 1024 * 1024 // 请求100GB
        efficiency.MemoryUsed = 70 * 1024 * 1024 * 1024       // 实际使用70GB
        efficiency.MemoryLimit = 200 * 1024 * 1024 * 1024     // 限制200GB
        efficiency.MemoryEfficiency = float64(efficiency.MemoryUsed) / float64(efficiency.MemoryRequested) * 100

        // 总体效率
        efficiency.OverallEfficiency = (efficiency.CPUEfficiency + efficiency.MemoryEfficiency) / 2

        // 浪费计算
        efficiency.WastedCPU = efficiency.CPURequested - efficiency.CPUUsed
        efficiency.WastedMemory = efficiency.MemoryRequested - efficiency.MemoryUsed
        efficiency.WastedCost = 3500 // 模拟浪费成本

        return efficiency, nil
}

// ResourceEfficiency 资源效率
type ResourceEfficiency struct {
        ClusterID          string  `json:"cluster_id"`
        CPURequested       float64 `json:"cpu_requested"`       // 核
        CPUUsed            float64 `json:"cpu_used"`            // 核
        CPULimit           float64 `json:"cpu_limit"`           // 核
        CPUEfficiency      float64 `json:"cpu_efficiency"`      // 百分比
        MemoryRequested    int64   `json:"memory_requested"`    // 字节
        MemoryUsed         int64   `json:"memory_used"`         // 字节
        MemoryLimit        int64   `json:"memory_limit"`        // 字节
        MemoryEfficiency   float64 `json:"memory_efficiency"`   // 百分比
        OverallEfficiency  float64 `json:"overall_efficiency"`  // 百分比
        WastedCPU          float64 `json:"wasted_cpu"`          // 浪费的CPU核数
        WastedMemory       int64   `json:"wasted_memory"`       // 浪费的内存
        WastedCost         float64 `json:"wasted_cost"`         // 浪费的成本
}

// RightSizeWorkload 工作负载右 sizing
func (s *K8sCostService) RightSizeWorkload(ctx context.Context, clusterID, namespace, workloadName string) (*RightSizeRecommendation, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        recommendation := &RightSizeRecommendation{
                ClusterID:  clusterID,
                Namespace:  namespace,
                Workload:   workloadName,
                Current:    ResourceSpec{CPU: 4.0, Memory: 8 * 1024 * 1024 * 1024},
                Recommended: ResourceSpec{CPU: 2.5, Memory: 5 * 1024 * 1024 * 1024},
                MonthlySavings: 800,
                Reason:     "基于过去30天的使用数据，当前资源配置过高",
                Confidence: 0.85,
        }

        // 历史数据
        recommendation.UsageHistory = []UsageData{
                {Date: "2024-01-01", CPU: 2.1, Memory: 4.5 * 1024 * 1024 * 1024},
                {Date: "2024-01-02", CPU: 2.3, Memory: 4.8 * 1024 * 1024 * 1024},
                {Date: "2024-01-03", CPU: 2.0, Memory: 4.2 * 1024 * 1024 * 1024},
        }

        return recommendation, nil
}

// RightSizeRecommendation 右sizing推荐
type RightSizeRecommendation struct {
        ClusterID      string       `json:"cluster_id"`
        Namespace      string       `json:"namespace"`
        Workload       string       `json:"workload"`
        Current        ResourceSpec `json:"current"`
        Recommended    ResourceSpec `json:"recommended"`
        MonthlySavings float64      `json:"monthly_savings"`
        Reason         string       `json:"reason"`
        Confidence     float64      `json:"confidence"`
        UsageHistory   []UsageData  `json:"usage_history"`
}

// ResourceSpec 资源规格
type ResourceSpec struct {
        CPU    float64 `json:"cpu"`    // 核
        Memory int64   `json:"memory"` // 字节
}

// UsageData 使用数据
type UsageData struct {
        Date   string  `json:"date"`
        CPU    float64 `json:"cpu"`
        Memory int64   `json:"memory"`
}

// SetNamespaceBudget 设置命名空间预算
func (s *K8sCostService) SetNamespaceBudget(ctx context.Context, clusterID, namespace string, budget float64) error {
        s.mu.Lock()
        defer s.mu.Unlock()

        // 实际应保存到数据库
        return nil
}

// GetNamespaceBudget 获取命名空间预算
func (s *K8sCostService) GetNamespaceBudget(ctx context.Context, clusterID, namespace string) (*NamespaceBudget, error) {
        budget := &NamespaceBudget{
                ClusterID:  clusterID,
                Namespace:  namespace,
                Budget:     5000,
                Used:       3500,
                Remaining:  1500,
                UsedPercent: 70,
                Forecast:   4200,
                AlertLevel: "warning",
        }

        return budget, nil
}

// NamespaceBudget 命名空间预算
type NamespaceBudget struct {
        ClusterID   string  `json:"cluster_id"`
        Namespace   string  `json:"namespace"`
        Budget      float64 `json:"budget"`
        Used        float64 `json:"used"`
        Remaining   float64 `json:"remaining"`
        UsedPercent float64 `json:"used_percent"`
        Forecast    float64 `json:"forecast"`
        AlertLevel  string  `json:"alert_level"`
}

// CreateK8sCostRecord 创建K8s成本记录
func (s *K8sCostService) CreateK8sCostRecord(ctx context.Context, record *cost.K8sCostRecord) error {
        return nil
}

// GetK8sCostTrend 获取K8s成本趋势
func (s *K8sCostService) GetK8sCostTrend(ctx context.Context, clusterID string, days int) (*K8sCostTrend, error) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        trend := &K8sCostTrend{
                ClusterID: clusterID,
                Points:    make([]K8sTrendPoint, 0),
        }

        now := time.Now()
        for i := 0; i < days; i++ {
                date := now.AddDate(0, 0, -i)
                trend.Points = append(trend.Points, K8sTrendPoint{
                        Date:      date.Format("2006-01-02"),
                        TotalCost: 400 + float64(i)*2,
                        CPUCost:   200 + float64(i),
                        MemoryCost: 150 + float64(i)*0.5,
                        StorageCost: 50,
                })
        }

        return trend, nil
}

// K8sCostTrend K8s成本趋势
type K8sCostTrend struct {
        ClusterID string          `json:"cluster_id"`
        Points    []K8sTrendPoint `json:"points"`
}

// K8sTrendPoint K8s趋势点
type K8sTrendPoint struct {
        Date        string  `json:"date"`
        TotalCost   float64 `json:"total_cost"`
        CPUCost     float64 `json:"cpu_cost"`
        MemoryCost  float64 `json:"memory_cost"`
        StorageCost float64 `json:"storage_cost"`
}
