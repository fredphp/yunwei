package cost

import (
	"time"
)

// ==================== 成本统计 ====================

// CostRecord 成本记录
type CostRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RecordDate   time.Time `gorm:"type:date;not null;index" json:"record_date"`
	RecordHour   int       `gorm:"not null" json:"record_hour"` // 0-23

	// 资源信息
	Provider     string `gorm:"size:50;not null;index" json:"provider"`     // aws, azure, gcp, aliyun, tencent, huawei
	ResourceType string `gorm:"size:50;not null;index" json:"resource_type"` // ec2, rds, s3, ebs, lambda, eks, ecs
	ResourceID   string `gorm:"size:100;not null;index" json:"resource_id"`
	ResourceName string `gorm:"size:200" json:"resource_name"`
	Region       string `gorm:"size:50" json:"region"`
	Zone         string `gorm:"size:50" json:"zone"`

	// 计费信息
	BillingMode  string  `gorm:"size:20" json:"billing_mode"`   // on_demand, reserved, spot, savings_plan
	InstanceType string  `gorm:"size:50" json:"instance_type"`
	UsageAmount  float64 `json:"usage_amount"`                  // 使用量
	UsageUnit    string  `gorm:"size:20" json:"usage_unit"`     // hours, gb, requests, etc.

	// 成本信息
	ListPrice    float64 `json:"list_price"`     // 刊例价
	DiscountRate float64 `json:"discount_rate"`  // 折扣率
	BlendedCost  float64 `json:"blended_cost"`   // 混合成本
	NetCost      float64 `json:"net_cost"`       // 净成本
	Currency     string  `gorm:"size:10;default:'USD'" json:"currency"`

	// 标签
	Tags string `gorm:"type:text" json:"tags"` // JSON格式标签

	// 归属信息
	AccountID   string `gorm:"size:100" json:"account_id"`
	AccountName string `gorm:"size:200" json:"account_name"`
	ProjectID   string `gorm:"size:100;index" json:"project_id"`
	ProjectName string `gorm:"size:200" json:"project_name"`
	Department  string `gorm:"size:100;index" json:"department"`
	Owner       string `gorm:"size:100;index" json:"owner"`
	Environment string `gorm:"size:50;index" json:"environment"` // prod, staging, dev, test

	// 成本中心
	CostCenter  string `gorm:"size:100;index" json:"cost_center"`
	BusinessUnit string `gorm:"size:100;index" json:"business_unit"`

	CreatedAt time.Time `json:"created_at"`
}

// CostSummary 成本汇总
type CostSummary struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SummaryType string    `gorm:"size:20;not null;index" json:"summary_type"` // daily, weekly, monthly
	PeriodStart time.Time `gorm:"type:date;not null;index" json:"period_start"`
	PeriodEnd   time.Time `gorm:"type:date;not null" json:"period_end"`

	// 维度
	Provider     string `gorm:"size:50;index" json:"provider"`
	ResourceType string `gorm:"size:50;index" json:"resource_type"`
	ProjectID    string `gorm:"size:100;index" json:"project_id"`
	Department   string `gorm:"size:100;index" json:"department"`
	Environment  string `gorm:"size:50;index" json:"environment"`
	CostCenter   string `gorm:"size:100;index" json:"cost_center"`

	// 汇总数据
	TotalCost      float64 `json:"total_cost"`
	TotalUsage     float64 `json:"total_usage"`
	ResourceCount  int     `json:"resource_count"`
	AvgDailyCost   float64 `json:"avg_daily_cost"`
	MaxDailyCost   float64 `json:"max_daily_cost"`
	MinDailyCost   float64 `json:"min_daily_cost"`

	// 环比
	PrevPeriodCost float64 `json:"prev_period_cost"`
	CostChange     float64 `json:"cost_change"`      // 变化金额
	CostChangeRate float64 `json:"cost_change_rate"` // 变化率(%)

	// 预算
	BudgetAmount   float64 `json:"budget_amount"`
	BudgetUsage    float64 `json:"budget_usage"`    // 预算使用率(%)
	BudgetStatus   string  `gorm:"size:20" json:"budget_status"` // normal, warning, exceeded

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==================== 成本预测 ====================

// CostForecast 成本预测
type CostForecast struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ForecastDate time.Time `gorm:"type:date;not null;index" json:"forecast_date"`

	// 预测范围
	DimensionType  string `gorm:"size:50;not null" json:"dimension_type"`  // total, provider, project, department
	DimensionValue string `gorm:"size:200" json:"dimension_value"`

	// 预测周期
	PeriodType string    `gorm:"size:20;not null" json:"period_type"` // daily, weekly, monthly, quarterly
	StartDate  time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:date;not null" json:"end_date"`

	// 预测值
	PredictedCost float64 `json:"predicted_cost"`
	LowerBound    float64 `json:"lower_bound"`   // 置信区间下限
	UpperBound    float64 `json:"upper_bound"`   // 置信区间上限
	Confidence    float64 `json:"confidence"`    // 置信度(0-1)

	// 预测模型
	ModelType    string `gorm:"size:50" json:"model_type"`    // linear, arima, prophet, ml
	ModelVersion string `gorm:"size:50" json:"model_version"`
	Features     string `gorm:"type:text" json:"features"`    // JSON特征列表

	// 准确度追踪
	ActualCost    *float64 `json:"actual_cost"`
	ErrorRate     *float64 `json:"error_rate"`
	AbsoluteError *float64 `json:"absolute_error"`

	// 趋势
	TrendDirection string  `gorm:"size:20" json:"trend_direction"` // up, down, stable
	TrendStrength  float64 `json:"trend_strength"`

	// 季节性
	SeasonalityPattern string  `gorm:"size:50" json:"seasonality_pattern"` // weekly, monthly, yearly
	SeasonalityFactor  float64 `json:"seasonality_factor"`

	CreatedAt time.Time `json:"created_at"`
}

// PredictionModel 预测模型
type PredictionModel struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"size:100;not null" json:"name"`
	Type         string `gorm:"size:50;not null" json:"type"` // linear, arima, prophet, ml
	Version      string `gorm:"size:50" json:"version"`
	Description  string `gorm:"type:text" json:"description"`

	// 模型配置
	Config string `gorm:"type:text" json:"config"` // JSON配置

	// 训练配置
	TrainingPeriod int `gorm:"default:90" json:"training_period"` // 训练数据周期(天)
	ValidationSplit float64 `gorm:"default:0.2" json:"validation_split"`

	// 特征
	Features string `gorm:"type:text" json:"features"` // JSON特征列表

	// 性能指标
	MAE    float64 `json:"mae"`    // 平均绝对误差
	RMSE   float64 `json:"rmse"`   // 均方根误差
	MAPE   float64 `json:"mape"`   // 平均绝对百分比误差
	R2     float64 `json:"r2"`     // R²分数
	AIC    float64 `json:"aic"`    // AIC
	BIC    float64 `json:"bic"`    // BIC

	// 状态
	Status      string     `gorm:"size:20;default:'active'" json:"status"` // active, inactive, training
	LastTrained *time.Time `json:"last_trained"`
	NextTrain   *time.Time `json:"next_train"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==================== 资源浪费检测 ====================

// WasteDetection 浪费检测
type WasteDetection struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DetectedAt time.Time `gorm:"not null;index" json:"detected_at"`

	// 资源信息
	Provider     string `gorm:"size:50;not null;index" json:"provider"`
	ResourceType string `gorm:"size:50;not null;index" json:"resource_type"`
	ResourceID   string `gorm:"size:100;not null;index" json:"resource_id"`
	ResourceName string `gorm:"size:200" json:"resource_name"`
	Region       string `gorm:"size:50" json:"region"`

	// 浪费类型
	WasteType    string `gorm:"size:50;not null" json:"waste_type"` // idle, oversized, orphaned, unattached, expired
	WasteCategory string `gorm:"size:50" json:"waste_category"` // compute, storage, network, database

	// 检测详情
	DetectionRule  string `gorm:"size:100" json:"detection_rule"`
	MetricName     string `gorm:"size:100" json:"metric_name"`
	MetricValue    float64 `json:"metric_value"`
	Threshold      float64 `json:"threshold"`
	DaysDetected   int     `json:"days_detected"` // 持续天数

	// 成本影响
	DailyWasteCost   float64 `json:"daily_waste_cost"`
	MonthlyWasteCost float64 `json:"monthly_waste_cost"`
	AnnualWasteCost  float64 `json:"annual_waste_cost"`
	Currency         string  `gorm:"size:10" json:"currency"`

	// 建议操作
	Recommendation   string `gorm:"type:text" json:"recommendation"`
	ActionType       string `gorm:"size:50" json:"action_type"` // terminate, resize, snapshot, release
	PotentialSavings float64 `json:"potential_savings"`

	// 风险评估
	RiskLevel   string `gorm:"size:20" json:"risk_level"` // low, medium, high, critical
	ImpactScore float64 `json:"impact_score"` // 0-100

	// 状态
	Status      string     `gorm:"size:20;default:'open'" json:"status"` // open, acknowledged, in_progress, resolved, ignored
	ResolvedAt  *time.Time `json:"resolved_at"`
	ResolvedBy  string     `gorm:"size:100" json:"resolved_by"`
	Resolution  string     `gorm:"type:text" json:"resolution"`

	// 归属
	ProjectID   string `gorm:"size:100;index" json:"project_id"`
	Department  string `gorm:"size:100;index" json:"department"`
	Owner       string `gorm:"size:100;index" json:"owner"`

	// 标签
	Tags string `gorm:"type:text" json:"tags"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WasteRule 浪费检测规则
type WasteRule struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"size:50;not null" json:"type"` // idle, oversized, orphaned

	// 规则配置
	ResourceTypes string `gorm:"type:text" json:"resource_types"` // JSON数组
	Condition     string `gorm:"type:text" json:"condition"`      // JSON条件表达式
	Threshold     float64 `json:"threshold"`
	Duration      int    `json:"duration"` // 持续时间(小时)

	// 排除规则
	Exclusions string `gorm:"type:text" json:"exclusions"` // JSON排除配置

	// 通知配置
	NotifyChannels string `gorm:"type:text" json:"notify_channels"`
	AutoAction     bool   `gorm:"default:false" json:"auto_action"`
	AutoActionType string `gorm:"size:50" json:"auto_action_type"`

	// 优先级
	Priority      int    `gorm:"default:5" json:"priority"`
	Severity      string `gorm:"size:20" json:"severity"` // info, warning, critical

	Enabled       bool   `gorm:"default:true" json:"enabled"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==================== 闲置资源识别 ====================

// IdleResource 闲置资源
type IdleResource struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ResourceID   string    `gorm:"size:100;not null;uniqueIndex" json:"resource_id"`
	ResourceName string    `gorm:"size:200" json:"resource_name"`
	ResourceType string    `gorm:"size:50;not null;index" json:"resource_type"`
	Provider     string    `gorm:"size:50;not null;index" json:"provider"`

	// 资源详情
	InstanceType   string `gorm:"size:50" json:"instance_type"`
	Region         string `gorm:"size:50" json:"region"`
	Zone           string `gorm:"size:50" json:"zone"`
	State          string `gorm:"size:50" json:"state"`

	// 配置
	CPU           int     `json:"cpu"`
	Memory        float64 `json:"memory"` // GB
	Storage       float64 `json:"storage"` // GB
	NetworkSpeed  float64 `json:"network_speed"`

	// 使用指标
	CPUUtilization      float64 `json:"cpu_utilization"`
	MemoryUtilization   float64 `json:"memory_utilization"`
	NetworkThroughput   float64 `json:"network_throughput"`
	DiskIOPS            float64 `json:"disk_iops"`
	ConnectionCount     int     `json:"connection_count"`
	RequestCount        int64   `json:"request_count"`

	// 闲置判断
	IdleStatus      string  `gorm:"size:20;index" json:"idle_status"` // active, low_utilization, idle, abandoned
	IdleDays        int     `json:"idle_days"`
	IdleScore       float64 `json:"idle_score"` // 0-100, 越高越闲置

	// 成本信息
	HourlyCost     float64 `json:"hourly_cost"`
	DailyCost      float64 `json:"daily_cost"`
	MonthlyCost    float64 `json:"monthly_cost"`
	AccumulatedCost float64 `json:"accumulated_cost"` // 累计闲置成本

	// 归属
	ProjectID    string `gorm:"size:100;index" json:"project_id"`
	ProjectName  string `gorm:"size:200" json:"project_name"`
	Department   string `gorm:"size:100;index" json:"department"`
	Owner        string `gorm:"size:100;index" json:"owner"`
	Environment  string `gorm:"size:50;index" json:"environment"`

	// 标签
	Tags string `gorm:"type:text" json:"tags"`

	// 建议
	Recommendation string `gorm:"type:text" json:"recommendation"`

	// 状态
	Status      string     `gorm:"size:20;default:'active'" json:"status"` // active, scheduled, terminated
	ScheduledAction string `gorm:"size:50" json:"scheduled_action"`
	ScheduledAt *time.Time `json:"scheduled_at"`

	FirstDetectedAt time.Time  `json:"first_detected_at"`
	LastActiveAt    *time.Time `json:"last_active_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ==================== 预算管理 ====================

// Budget 预算
type Budget struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`

	// 预算范围
	ScopeType    string `gorm:"size:50;not null" json:"scope_type"` // total, provider, project, department, resource_type
	ScopeValue   string `gorm:"size:200" json:"scope_value"`
	Provider     string `gorm:"size:50" json:"provider"`
	ResourceTypes string `gorm:"type:text" json:"resource_types"` // JSON数组

	// 预算金额
	BudgetAmount   float64 `gorm:"not null" json:"budget_amount"`
	Currency       string  `gorm:"size:10;default:'USD'" json:"currency"`

	// 周期
	PeriodType string    `gorm:"size:20;not null" json:"period_type"` // monthly, quarterly, yearly
	StartDate  time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:date;not null" json:"end_date"`

	// 告警阈值
	AlertThresholds string `gorm:"type:text" json:"alert_thresholds"` // JSON数组, 如[50,80,100]

	// 当前状态
	CurrentSpend float64 `json:"current_spend"`
	UsagePercent float64 `json:"usage_percent"`
	Status       string  `gorm:"size:20" json:"status"` // normal, warning, critical, exceeded

	// 预测
	ForecastedSpend float64 `json:"forecasted_spend"`
	ForecastedUsage float64 `json:"forecasted_usage"`

	// 归属
	Owner       string `gorm:"size:100" json:"owner"`
	Department  string `gorm:"size:100" json:"department"`
	CostCenter  string `gorm:"size:100" json:"cost_center"`

	// 通知配置
	NotifyChannels string `gorm:"type:text" json:"notify_channels"`
	NotifyEmails   string `gorm:"type:text" json:"notify_emails"`

	Enabled      bool `gorm:"default:true" json:"enabled"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BudgetAlert 预算告警
type BudgetAlert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BudgetID  uint      `gorm:"not null;index" json:"budget_id"`

	// 告警信息
	AlertType   string    `gorm:"size:50;not null" json:"alert_type"` // threshold, forecast, anomaly
	Threshold   float64   `json:"threshold"`
	CurrentPercent float64 `json:"current_percent"`

	// 状态
	Status      string    `gorm:"size:20;default:'active'" json:"status"` // active, acknowledged, resolved
	TriggeredAt time.Time `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`

	// 通知
	NotifySent    bool   `gorm:"default:false" json:"notify_sent"`
	NotifyChannels string `gorm:"type:text" json:"notify_channels"`
	NotifySentAt  *time.Time `json:"notify_sent_at"`

	// 详情
	Message string `gorm:"type:text" json:"message"`

	CreatedAt time.Time `json:"created_at"`
}

// ==================== 成本优化建议 ====================

// CostOptimization 成本优化建议
type CostOptimization struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`

	// 资源信息
	Provider     string `gorm:"size:50;not null;index" json:"provider"`
	ResourceType string `gorm:"size:50;not null" json:"resource_type"`
	ResourceID   string `gorm:"size:100;not null;index" json:"resource_id"`
	ResourceName string `gorm:"size:200" json:"resource_name"`

	// 优化类型
	OptimizationType string `gorm:"size:50;not null" json:"optimization_type"` // resize, terminate, reserved, spot, savings_plan
	Category         string `gorm:"size:50" json:"category"` // compute, storage, network, database

	// 当前状态
	CurrentConfig string `gorm:"type:text" json:"current_config"`
	CurrentCost   float64 `json:"current_cost"`

	// 建议配置
	RecommendedConfig string `gorm:"type:text" json:"recommended_config"`
	RecommendedCost   float64 `json:"recommended_cost"`

	// 节省
	MonthlySavings float64 `json:"monthly_savings"`
	AnnualSavings  float64 `json:"annual_savings"`
	SavingsPercent float64 `json:"savings_percent"`

	// 实施信息
	Effort          string `gorm:"size:20" json:"effort"` // low, medium, high
	Complexity      string `gorm:"size:20" json:"complexity"`
	Risk            string `gorm:"size:20" json:"risk"`
	ImplementationSteps string `gorm:"type:text" json:"implementation_steps"`

	// 状态
	Status       string     `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, in_progress, completed, rejected
	ApprovedBy   string     `gorm:"size:100" json:"approved_by"`
	ApprovedAt   *time.Time `json:"approved_at"`
	CompletedAt  *time.Time `json:"completed_at"`

	// ROI
	ROIScore      float64 `json:"roi_score"` // ROI评分 0-100
	Priority      int     `gorm:"default:5" json:"priority"`

	// 归属
	ProjectID   string `gorm:"size:100;index" json:"project_id"`
	Department  string `gorm:"size:100;index" json:"department"`
	Owner       string `gorm:"size:100;index" json:"owner"`
}

// ==================== Kubernetes成本 ====================

// K8sCostRecord K8s成本记录
type K8sCostRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RecordTime time.Time `gorm:"not null;index" json:"record_time"`

	// 集群信息
	ClusterID   string `gorm:"size:100;not null;index" json:"cluster_id"`
	ClusterName string `gorm:"size:200" json:"cluster_name"`
	Provider    string `gorm:"size:50" json:"provider"`
	Region      string `gorm:"size:50" json:"region"`

	// 命名空间
	Namespace string `gorm:"size:100;index" json:"namespace"`

	// 工作负载
	WorkloadType string `gorm:"size:50" json:"workload_type"` // deployment, statefulset, daemonset, job, cronjob
	WorkloadName string `gorm:"size:200;index" json:"workload_name"`
	PodName      string `gorm:"size:200" json:"pod_name"`

	// 资源请求/限制
	CPURequest    float64 `json:"cpu_request"`    // cores
	CPULimit      float64 `json:"cpu_limit"`
	MemoryRequest float64 `json:"memory_request"` // bytes
	MemoryLimit   float64 `json:"memory_limit"`

	// 实际使用
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	NetworkInBytes int64   `json:"network_in_bytes"`
	NetworkOutBytes int64  `json:"network_out_bytes"`

	// 成本
	CPUCost        float64 `json:"cpu_cost"`
	MemoryCost     float64 `json:"memory_cost"`
	NetworkCost    float64 `json:"network_cost"`
	StorageCost    float64 `json:"storage_cost"`
	TotalCost      float64 `json:"total_cost"`

	// 效率
	CPUEfficiency    float64 `json:"cpu_efficiency"`    // usage/request
	MemoryEfficiency float64 `json:"memory_efficiency"`
	TotalEfficiency  float64 `json:"total_efficiency"`

	// 标签
	Labels string `gorm:"type:text" json:"labels"`

	// 归属
	ProjectID  string `gorm:"size:100;index" json:"project_id"`
	Department string `gorm:"size:100;index" json:"department"`
	Owner      string `gorm:"size:100;index" json:"owner"`
	Environment string `gorm:"size:50;index" json:"environment"`

	CreatedAt time.Time `json:"created_at"`
}

// K8sNamespaceCost K8s命名空间成本汇总
type K8sNamespaceCost struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RecordDate time.Time `gorm:"type:date;not null;uniqueIndex:uniq_ns_date" json:"record_date"`

	ClusterID   string `gorm:"size:100;not null;uniqueIndex:uniq_ns_date" json:"cluster_id"`
	ClusterName string `gorm:"size:200" json:"cluster_name"`
	Namespace   string `gorm:"size:100;not null;uniqueIndex:uniq_ns_date" json:"namespace"`

	// 资源汇总
	TotalCPURequest    float64 `json:"total_cpu_request"`
	TotalCPULimit      float64 `json:"total_cpu_limit"`
	TotalMemoryRequest float64 `json:"total_memory_request"`
	TotalMemoryLimit   float64 `json:"total_memory_limit"`
	TotalCPUUsage      float64 `json:"total_cpu_usage"`
	TotalMemoryUsage   float64 `json:"total_memory_usage"`

	// 成本汇总
	TotalCPUCost     float64 `json:"total_cpu_cost"`
	TotalMemoryCost  float64 `json:"total_memory_cost"`
	TotalNetworkCost float64 `json:"total_network_cost"`
	TotalStorageCost float64 `json:"total_storage_cost"`
	TotalCost        float64 `json:"total_cost"`

	// 效率
	AvgCPUEfficiency    float64 `json:"avg_cpu_efficiency"`
	AvgMemoryEfficiency float64 `json:"avg_memory_efficiency"`
	LowEfficiencyPods   int     `json:"low_efficiency_pods"`

	// Pod统计
	TotalPods     int `json:"total_pods"`
	RunningPods   int `json:"running_pods"`
	PendingPods   int `json:"pending_pods"`
	FailedPods    int `json:"failed_pods"`

	// 归属
	ProjectID   string `gorm:"size:100;index" json:"project_id"`
	Department  string `gorm:"size:100;index" json:"department"`
	Owner       string `gorm:"size:100;index" json:"owner"`
	Environment string `gorm:"size:50;index" json:"environment"`

	CreatedAt time.Time `json:"created_at"`
}

// K8sWorkloadCost K8s工作负载成本
type K8sWorkloadCost struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RecordDate time.Time `gorm:"type:date;not null;index" json:"record_date"`

	ClusterID   string `gorm:"size:100;not null;index" json:"cluster_id"`
	ClusterName string `gorm:"size:200" json:"cluster_name"`
	Namespace   string `gorm:"size:100;not null;index" json:"namespace"`

	WorkloadType string `gorm:"size:50;not null" json:"workload_type"`
	WorkloadName string `gorm:"size:200;not null;index" json:"workload_name"`

	// 资源配置
	Replicas      int     `json:"replicas"`
	CPURequest    float64 `json:"cpu_request"`
	CPULimit      float64 `json:"cpu_limit"`
	MemoryRequest float64 `json:"memory_request"`
	MemoryLimit   float64 `json:"memory_limit"`

	// 成本
	DailyCPUCost     float64 `json:"daily_cpu_cost"`
	DailyMemoryCost  float64 `json:"daily_memory_cost"`
	DailyNetworkCost float64 `json:"daily_network_cost"`
	DailyStorageCost float64 `json:"daily_storage_cost"`
	DailyTotalCost   float64 `json:"daily_total_cost"`
	MonthlyTotalCost float64 `json:"monthly_total_cost"`

	// 效率
	CPUEfficiency    float64 `json:"cpu_efficiency"`
	MemoryEfficiency float64 `json:"memory_efficiency"`

	// 优化建议
	OverProvisionedCPU    float64 `json:"over_provisioned_cpu"`
	OverProvisionedMemory float64 `json:"over_provisioned_memory"`
	PotentialSavings      float64 `json:"potential_savings"`

	// 归属
	ProjectID   string `gorm:"size:100;index" json:"project_id"`
	Department  string `gorm:"size:100;index" json:"department"`
	Owner       string `gorm:"size:100;index" json:"owner"`
	Environment string `gorm:"size:50;index" json:"environment"`

	CreatedAt time.Time `json:"created_at"`
}

// ==================== 价格与费率 ====================

// PricingRate 价格费率
type PricingRate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Provider    string    `gorm:"size:50;not null;index" json:"provider"`
	Region      string    `gorm:"size:50;not null;index" json:"region"`

	// 资源类型
	ResourceType string `gorm:"size:50;not null" json:"resource_type"`
	InstanceType string `gorm:"size:50;not null" json:"instance_type"`

	// 计费模式
	BillingMode string `gorm:"size:20;not null" json:"billing_mode"` // on_demand, reserved_1yr, reserved_3yr, spot

	// 价格
	UnitPrice   float64 `gorm:"not null" json:"unit_price"`
	Unit        string  `gorm:"size:20" json:"unit"` // hour, gb, request
	Currency    string  `gorm:"size:10" json:"currency"`

	// 预留实例
	UpfrontCost    float64 `json:"upfront_cost"`
	MonthlyCost    float64 `json:"monthly_cost"`
	EffectiveHourly float64 `json:"effective_hourly"`

	// 有效期
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==================== 报表与导出 ====================

// CostReport 成本报表
type CostReport struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:200;not null" json:"name"`
	ReportType   string    `gorm:"size:50;not null" json:"report_type"` // cost_summary, waste, idle, optimization

	// 范围
	PeriodStart time.Time `gorm:"type:date" json:"period_start"`
	PeriodEnd   time.Time `gorm:"type:date" json:"period_end"`
	Provider    string    `gorm:"size:50" json:"provider"`
	ProjectID   string    `gorm:"size:100" json:"project_id"`
	Department  string    `gorm:"size:100" json:"department"`

	// 配置
	GroupBy   string `gorm:"size:200" json:"group_by"`   // 逗号分隔
	FilterBy  string `gorm:"type:text" json:"filter_by"` // JSON
	SortBy    string `gorm:"size:100" json:"sort_by"`
	SortOrder string `gorm:"size:10" json:"sort_order"`

	// 输出
	Format     string `gorm:"size:20" json:"format"` // json, csv, excel, pdf
	OutputPath string `gorm:"size:500" json:"output_path"`
	FileSize   int64  `json:"file_size"`

	// 调度
	ScheduleType string    `gorm:"size:20" json:"schedule_type"` // once, daily, weekly, monthly
	NextRun      *time.Time `json:"next_run"`
	LastRun      *time.Time `json:"last_run"`

	// 状态
	Status      string     `gorm:"size:20" json:"status"` // pending, running, completed, failed
	CompletedAt *time.Time `json:"completed_at"`
	Error       string     `gorm:"type:text" json:"error"`

	// 归属
	CreatedBy  string `gorm:"size:100" json:"created_by"`
	Owner      string `gorm:"size:100" json:"owner"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
