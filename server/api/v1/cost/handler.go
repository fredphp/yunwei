package cost

import (
	"net/http"
	"strconv"
	"time"

	costModel "yunwei/model/cost"
	"yunwei/model/common/response"
	costService "yunwei/service/cost"

	"github.com/gin-gonic/gin"
)

// Handler 成本控制API处理器
type Handler struct {
	statisticsSvc *costService.StatisticsService
	forecastSvc   *costService.ForecastService
	wasteSvc      *costService.WasteDetectionService
	idleSvc       *costService.IdleDetectionService
}

// NewHandler 创建处理器
func NewHandler() *Handler {
	return &Handler{
		statisticsSvc: costService.NewStatisticsService(),
		forecastSvc:   costService.NewForecastService(),
		wasteSvc:      costService.NewWasteDetectionService(),
		idleSvc:       costService.NewIdleDetectionService(),
	}
}

// ==================== 成本概览 ====================

// GetOverview 获取成本概览
func (h *Handler) GetOverview(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")
	
	now := time.Now()
	var startTime, endTime time.Time
	switch period {
	case "daily":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		endTime = now
	case "weekly":
		startTime = now.AddDate(0, 0, -7)
		endTime = now
	case "monthly":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		endTime = now
	default:
		startTime = now.AddDate(0, -1, 0)
		endTime = now
	}

	summary, err := h.statisticsSvc.GetCostSummary(c.Request.Context(), startTime, endTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(summary, c)
}

// GetCostSummary 获取成本摘要
func (h *Handler) GetCostSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)

	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, -1, 0)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	summary, err := h.statisticsSvc.GetCostSummary(c.Request.Context(), startTime, endTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(summary, c)
}

// GetCostTrend 获取成本趋势
func (h *Handler) GetCostTrend(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	granularity := c.DefaultQuery("granularity", "daily")

	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)

	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, -1, 0)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	trend, err := h.statisticsSvc.GetCostTrend(c.Request.Context(), startTime, endTime, granularity)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(trend, c)
}

// GetDailyCost 获取每日成本
func (h *Handler) GetDailyCost(c *gin.Context) {
	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	detail, err := h.statisticsSvc.GetDailyCost(c.Request.Context(), date)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(detail, c)
}

// GetMonthlyCost 获取月度成本
func (h *Handler) GetMonthlyCost(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	detail, err := h.statisticsSvc.GetMonthlyCost(c.Request.Context(), year, month)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(detail, c)
}

// GetCostBreakdown 获取成本分解
func (h *Handler) GetCostBreakdown(c *gin.Context) {
	dimension := c.DefaultQuery("dimension", "provider") // provider, type, region, department

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)

	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, -1, 0)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	breakdown, err := h.statisticsSvc.GetCostBreakdown(c.Request.Context(), startTime, endTime, dimension)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(breakdown, c)
}

// GetCostRanking 获取成本排名
func (h *Handler) GetCostRanking(c *gin.Context) {
	topN, _ := strconv.Atoi(c.DefaultQuery("top", "10"))

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)

	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, -1, 0)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	ranking, err := h.statisticsSvc.GetResourceCostRanking(c.Request.Context(), startTime, endTime, topN)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(ranking, c)
}

// ComparePeriods 对比周期
func (h *Handler) ComparePeriods(c *gin.Context) {
	p1Start, _ := time.Parse("2006-01-02", c.Query("p1_start"))
	p1End, _ := time.Parse("2006-01-02", c.Query("p1_end"))
	p2Start, _ := time.Parse("2006-01-02", c.Query("p2_start"))
	p2End, _ := time.Parse("2006-01-02", c.Query("p2_end"))

	comparison, err := h.statisticsSvc.ComparePeriods(c.Request.Context(), p1Start, p1End, p2Start, p2End)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(comparison, c)
}

// ==================== 成本预测 ====================

// GetForecast 获取成本预测
func (h *Handler) GetForecast(c *gin.Context) {
	forecastType := c.DefaultQuery("type", "monthly")

	var result interface{}
	var err error

	switch forecastType {
	case "daily":
		result, err = h.forecastSvc.ForecastDaily(c.Request.Context(), time.Now().AddDate(0, 0, 1))
	case "monthly":
		result, err = h.forecastSvc.ForecastMonthly(c.Request.Context(), time.Now().Year(), int(time.Now().Month())+1)
	case "quarterly":
		quarter := (int(time.Now().Month())-1)/3 + 1
		result, err = h.forecastSvc.ForecastQuarterly(c.Request.Context(), time.Now().Year(), quarter+1)
	default:
		result, err = h.forecastSvc.ForecastMonthly(c.Request.Context(), time.Now().Year(), int(time.Now().Month())+1)
	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// GetDailyForecast 获取日度预测
func (h *Handler) GetDailyForecast(c *gin.Context) {
	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now().AddDate(0, 0, 1)
	}

	result, err := h.forecastSvc.ForecastDaily(c.Request.Context(), date)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// GetMonthlyForecast 获取月度预测
func (h *Handler) GetMonthlyForecast(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())+1))

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	result, err := h.forecastSvc.ForecastMonthly(c.Request.Context(), year, month)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// GetTrendAnalysis 获取趋势分析
func (h *Handler) GetTrendAnalysis(c *gin.Context) {
	dimension := c.DefaultQuery("dimension", "daily")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)

	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, -1, 0)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	result, err := h.forecastSvc.TrendAnalysis(c.Request.Context(), startTime, endTime, dimension)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// ==================== 资源浪费检测 ====================

// GetWasteSummary 获取浪费摘要
func (h *Handler) GetWasteSummary(c *gin.Context) {
	summary, err := h.wasteSvc.DetectWaste(c.Request.Context())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(summary, c)
}

// GetEC2Waste 获取EC2浪费
func (h *Handler) GetEC2Waste(c *gin.Context) {
	report, err := h.wasteSvc.DetectEC2Waste(c.Request.Context())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(report, c)
}

// GetRDSWaste 获取RDS浪费
func (h *Handler) GetRDSWaste(c *gin.Context) {
	report, err := h.wasteSvc.DetectRDSWaste(c.Request.Context())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(report, c)
}

// GetStorageWaste 获取存储浪费
func (h *Handler) GetStorageWaste(c *gin.Context) {
	report, err := h.wasteSvc.DetectStorageWaste(c.Request.Context())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(report, c)
}

// OptimizeResource 优化资源
func (h *Handler) OptimizeResource(c *gin.Context) {
	resourceID := c.Param("id")
	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := h.wasteSvc.OptimizeResource(c.Request.Context(), resourceID, req.Action)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// ==================== 闲置资源检测 ====================

// GetIdleResources 获取闲置资源
func (h *Handler) GetIdleResources(c *gin.Context) {
	threshold, _ := strconv.ParseFloat(c.DefaultQuery("threshold", "10"), 64)

	summary, err := h.idleSvc.DetectIdleResources(c.Request.Context(), threshold)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(summary, c)
}

// AnalyzeResourceUsage 分析资源使用
func (h *Handler) AnalyzeResourceUsage(c *gin.Context) {
	resourceID := c.Param("id")
	durationStr := c.DefaultQuery("duration", "168h") // 默认7天

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 168 * time.Hour
	}

	analysis, err := h.idleSvc.AnalyzeResourceUsage(c.Request.Context(), resourceID, duration)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(analysis, c)
}

// GetIdleTimeline 获取闲置时间线
func (h *Handler) GetIdleTimeline(c *gin.Context) {
	resourceID := c.Param("id")

	timeline, err := h.idleSvc.GetIdleTimeline(c.Request.Context(), resourceID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(timeline, c)
}

// SetIdlePolicy 设置闲置策略
func (h *Handler) SetIdlePolicy(c *gin.Context) {
	var policy costService.IdlePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := h.idleSvc.SetIdlePolicy(c.Request.Context(), policy); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithMessage("闲置策略已设置", c)
}

// ==================== 预算管理 ====================

// GetBudgets 获取预算列表
func (h *Handler) GetBudgets(c *gin.Context) {
	budgets := []costModel.Budget{
		{BudgetID: "budget_001", Name: "月度总预算", Amount: 50000, UsedAmount: 35000, UsedPercent: 70},
		{BudgetID: "budget_002", Name: "研发部预算", Amount: 20000, UsedAmount: 15000, UsedPercent: 75},
		{BudgetID: "budget_003", Name: "测试环境预算", Amount: 5000, UsedAmount: 4500, UsedPercent: 90},
	}
	response.OkWithData(budgets, c)
}

// GetBudget 获取预算详情
func (h *Handler) GetBudget(c *gin.Context) {
	id := c.Param("id")
	_ = id

	budget := &costModel.Budget{
		BudgetID:    id,
		Name:        "月度总预算",
		Amount:      50000,
		UsedAmount:  35000,
		UsedPercent: 70,
	}
	response.OkWithData(budget, c)
}

// CreateBudget 创建预算
func (h *Handler) CreateBudget(c *gin.Context) {
	var budget costModel.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": "budget_" + strconv.FormatInt(time.Now().Unix(), 10)}, c)
}

// UpdateBudget 更新预算
func (h *Handler) UpdateBudget(c *gin.Context) {
	id := c.Param("id")
	var budget costModel.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = id

	response.Ok(c)
}

// DeleteBudget 删除预算
func (h *Handler) DeleteBudget(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.Ok(c)
}

// GetBudgetForecast 获取预算预测
func (h *Handler) GetBudgetForecast(c *gin.Context) {
	id := c.Param("id")

	budget := &costModel.Budget{
		BudgetID:   id,
		Amount:     50000,
		UsedAmount: 35000,
		StartDate:  time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local),
		EndDate:    time.Date(time.Now().Year(), time.Now().Month()+1, 0, 0, 0, 0, 0, time.Local),
	}

	forecast, err := h.forecastSvc.BudgetForecast(c.Request.Context(), budget)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(forecast, c)
}

// ==================== 优化建议 ====================

// GetOptimizations 获取优化建议
func (h *Handler) GetOptimizations(c *gin.Context) {
	optimizations := []costModel.CostOptimization{
		{OptimizationID: "opt_001", Type: "resize", ResourceType: "ecs", ResourceName: "主应用服务器", MonthlySavings: 1500, Priority: 9},
		{OptimizationID: "opt_002", Type: "terminate", ResourceType: "ecs", ResourceName: "测试服务器A", MonthlySavings: 800, Priority: 8},
		{OptimizationID: "opt_003", Type: "reserve", ResourceType: "rds", ResourceName: "主数据库", MonthlySavings: 2000, Priority: 10},
	}
	response.OkWithData(optimizations, c)
}

// AcceptOptimization 接受优化建议
func (h *Handler) AcceptOptimization(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.OkWithMessage("优化建议已接受", c)
}

// RejectOptimization 拒绝优化建议
func (h *Handler) RejectOptimization(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.OkWithMessage("优化建议已拒绝", c)
}

// ==================== 云账户管理 ====================

// GetCloudAccounts 获取云账户列表
func (h *Handler) GetCloudAccounts(c *gin.Context) {
	accounts := []costModel.CloudAccount{
		{AccountID: "acc_001", Name: "阿里云主账户", Provider: "aliyun", ResourceCount: 50, MonthlyCost: 20000},
		{AccountID: "acc_002", Name: "AWS生产环境", Provider: "aws", ResourceCount: 30, MonthlyCost: 15000},
		{AccountID: "acc_003", Name: "腾讯云测试", Provider: "tencent", ResourceCount: 20, MonthlyCost: 5000},
	}
	response.OkWithData(accounts, c)
}

// CreateCloudAccount 创建云账户
func (h *Handler) CreateCloudAccount(c *gin.Context) {
	var account costModel.CloudAccount
	if err := c.ShouldBindJSON(&account); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": "acc_" + strconv.FormatInt(time.Now().Unix(), 10)}, c)
}

// UpdateCloudAccount 更新云账户
func (h *Handler) UpdateCloudAccount(c *gin.Context) {
	id := c.Param("id")
	var account costModel.CloudAccount
	if err := c.ShouldBindJSON(&account); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = id

	response.Ok(c)
}

// DeleteCloudAccount 删除云账户
func (h *Handler) DeleteCloudAccount(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.Ok(c)
}

// SyncCloudAccount 同步云账户
func (h *Handler) SyncCloudAccount(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.OkWithData(gin.H{
		"status":     "syncing",
		"started_at": time.Now(),
	}, c)
}

// ==================== K8s成本分析 ====================

// GetK8sCostAnalysis 获取K8s成本分析
func (h *Handler) GetK8sCostAnalysis(c *gin.Context) {
	clusterID := c.Query("cluster_id")
	namespace := c.Query("namespace")
	_ = clusterID
	_ = namespace

	analysis := []costModel.K8sCostAnalysis{
		{ClusterName: "prod-cluster", Namespace: "default", WorkloadName: "api-server", TotalCost: 5000, CPUEfficiency: 65, MemoryEfficiency: 70},
		{ClusterName: "prod-cluster", Namespace: "monitoring", WorkloadName: "prometheus", TotalCost: 2000, CPUEfficiency: 80, MemoryEfficiency: 75},
	}
	response.OkWithData(analysis, c)
}

// GetNamespaceCost 获取命名空间成本
func (h *Handler) GetNamespaceCost(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	_ = clusterID

	costs := map[string]float64{
		"default":     5000,
		"monitoring":  2000,
		"logging":     1500,
		"istio-system": 1000,
	}
	response.OkWithData(costs, c)
}

// GetWorkloadCost 获取工作负载成本
func (h *Handler) GetWorkloadCost(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")
	_ = clusterID
	_ = namespace

	workloads := []map[string]interface{}{
		{"name": "api-server", "type": "deployment", "cost": 3000, "cpu_efficiency": 65},
		{"name": "worker", "type": "deployment", "cost": 1500, "cpu_efficiency": 70},
		{"name": "scheduler", "type": "deployment", "cost": 500, "cpu_efficiency": 80},
	}
	response.OkWithData(workloads, c)
}

// ==================== 报告 ====================

// GenerateReport 生成报告
func (h *Handler) GenerateReport(c *gin.Context) {
	var req struct {
		ReportType string `json:"report_type" binding:"required"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
		Format     string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"report_id":  "report_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status":     "generating",
		"started_at": time.Now(),
	}, c)
}

// GetReport 获取报告
func (h *Handler) GetReport(c *gin.Context) {
	id := c.Param("id")
	_ = id

	report := map[string]interface{}{
		"report_id":     id,
		"status":        "completed",
		"total_cost":    50000,
		"potential_savings": 8000,
		"generated_at":  time.Now(),
	}
	response.OkWithData(report, c)
}

// ExportCostReport 导出成本报告
func (h *Handler) ExportCostReport(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)

	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, -1, 0)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	report, err := h.statisticsSvc.ExportCostReport(c.Request.Context(), startTime, endTime, format)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", []byte(report))
}

// ==================== 告警 ====================

// GetAlertRules 获取告警规则
func (h *Handler) GetAlertRules(c *gin.Context) {
	rules := []costModel.CostAlertRule{
		{RuleID: "rule_001", Name: "月度预算告警", RuleType: "budget", Threshold: 80},
		{RuleID: "rule_002", Name: "异常成本检测", RuleType: "anomaly", Threshold: 50},
	}
	response.OkWithData(rules, c)
}

// CreateAlertRule 创建告警规则
func (h *Handler) CreateAlertRule(c *gin.Context) {
	var rule costModel.CostAlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"id": "rule_" + strconv.FormatInt(time.Now().Unix(), 10)}, c)
}

// GetAlerts 获取告警列表
func (h *Handler) GetAlerts(c *gin.Context) {
	alerts := []costModel.CostAlert{
		{AlertID: "alert_001", AlertLevel: "warning", Message: "研发部预算已使用80%"},
		{AlertID: "alert_002", AlertLevel: "critical", Message: "测试环境预算即将超支"},
	}
	response.OkWithData(alerts, c)
}

// AckAlert 确认告警
func (h *Handler) AckAlert(c *gin.Context) {
	id := c.Param("id")
	_ = id

	response.OkWithMessage("告警已确认", c)
}

// ==================== 统计仪表盘 ====================

// GetDashboard 获取仪表盘数据
func (h *Handler) GetDashboard(c *gin.Context) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	summary, _ := h.statisticsSvc.GetCostSummary(c.Request.Context(), monthStart, now)
	waste, _ := h.wasteSvc.DetectWaste(c.Request.Context())
	idle, _ := h.idleSvc.DetectIdleResources(c.Request.Context(), 10)

	dashboard := map[string]interface{}{
		"total_cost":        summary.TotalCost,
		"cost_change":       summary.CostChangePercent,
		"by_provider":       summary.ByProvider,
		"by_type":           summary.ByType,
		"wasted_cost":       waste.TotalWastedCost,
		"idle_resources":    idle.TotalIdleResources,
		"potential_savings": waste.PotentialSavings,
		"top_expensive":     summary.TopResources[:5],
		"trend":             summary.Trend[:7],
	}

	response.OkWithData(dashboard, c)
}

// 注册路由
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	// 成本概览
	r.GET("/overview", h.GetOverview)
	r.GET("/summary", h.GetCostSummary)
	r.GET("/trend", h.GetCostTrend)
	r.GET("/daily", h.GetDailyCost)
	r.GET("/monthly", h.GetMonthlyCost)
	r.GET("/breakdown", h.GetCostBreakdown)
	r.GET("/ranking", h.GetCostRanking)
	r.GET("/compare", h.ComparePeriods)

	// 成本预测
	r.GET("/forecast", h.GetForecast)
	r.GET("/forecast/daily", h.GetDailyForecast)
	r.GET("/forecast/monthly", h.GetMonthlyForecast)
	r.GET("/forecast/trend", h.GetTrendAnalysis)

	// 资源浪费
	r.GET("/waste", h.GetWasteSummary)
	r.GET("/waste/ec2", h.GetEC2Waste)
	r.GET("/waste/rds", h.GetRDSWaste)
	r.GET("/waste/storage", h.GetStorageWaste)
	r.POST("/optimize/:id", h.OptimizeResource)

	// 闲置资源
	r.GET("/idle", h.GetIdleResources)
	r.GET("/idle/:id/analysis", h.AnalyzeResourceUsage)
	r.GET("/idle/:id/timeline", h.GetIdleTimeline)
	r.POST("/idle/policy", h.SetIdlePolicy)

	// 预算管理
	r.GET("/budgets", h.GetBudgets)
	r.GET("/budgets/:id", h.GetBudget)
	r.POST("/budgets", h.CreateBudget)
	r.PUT("/budgets/:id", h.UpdateBudget)
	r.DELETE("/budgets/:id", h.DeleteBudget)
	r.GET("/budgets/:id/forecast", h.GetBudgetForecast)

	// 优化建议
	r.GET("/optimizations", h.GetOptimizations)
	r.POST("/optimizations/:id/accept", h.AcceptOptimization)
	r.POST("/optimizations/:id/reject", h.RejectOptimization)

	// 云账户
	r.GET("/accounts", h.GetCloudAccounts)
	r.POST("/accounts", h.CreateCloudAccount)
	r.PUT("/accounts/:id", h.UpdateCloudAccount)
	r.DELETE("/accounts/:id", h.DeleteCloudAccount)
	r.POST("/accounts/:id/sync", h.SyncCloudAccount)

	// K8s成本
	r.GET("/k8s/analysis", h.GetK8sCostAnalysis)
	r.GET("/k8s/:cluster_id/namespaces", h.GetNamespaceCost)
	r.GET("/k8s/:cluster_id/namespaces/:namespace/workloads", h.GetWorkloadCost)

	// 报告
	r.POST("/reports", h.GenerateReport)
	r.GET("/reports/:id", h.GetReport)
	r.GET("/export", h.ExportCostReport)

	// 告警
	r.GET("/alerts/rules", h.GetAlertRules)
	r.POST("/alerts/rules", h.CreateAlertRule)
	r.GET("/alerts", h.GetAlerts)
	r.POST("/alerts/:id/ack", h.AckAlert)

	// 仪表盘
	r.GET("/dashboard", h.GetDashboard)
}
