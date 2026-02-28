package cdn

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"yunwei/global"
	"yunwei/service/ai/llm"
	"yunwei/service/notify"
)

// CDNProvider CDN 提供商
type CDNProvider string

const (
	ProviderAliyun   CDNProvider = "aliyun"
	ProviderTencent  CDNProvider = "tencent"
	ProviderAWS      CDNProvider = "aws"
	ProviderCloudflare CDNProvider = "cloudflare"
	ProviderQiniu    CDNProvider = "qiniu"
	ProviderUPCloud  CDNProvider = "upyun"
)

// OptimizeType 优化类型
type OptimizeType string

const (
	OptimizeCache      OptimizeType = "cache"      // 缓存优化
	OptimizeBandwidth  OptimizeType = "bandwidth"  // 带宽优化
	OptimizeCost       OptimizeType = "cost"       // 成本优化
	OptimizePerformance OptimizeType = "performance" // 性能优化
	OptimizeSecurity   OptimizeType = "security"   // 安全优化
)

// CDNStatus CDN 状态
type CDNStatus string

const (
	CDNStatusActive    CDNStatus = "active"
	CDNStatusDegraded  CDNStatus = "degraded"
	CDNStatusOffline   CDNStatus = "offline"
	CDNStatusOptimizing CDNStatus = "optimizing"
)

// CDNDomain CDN 域名配置
type CDNDomain struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string       `json:"name" gorm:"type:varchar(64)"`
	Domain      string       `json:"domain" gorm:"type:varchar(128);index"`
	Provider    CDNProvider  `json:"provider" gorm:"type:varchar(16)"`
	Status      CDNStatus    `json:"status" gorm:"type:varchar(16)"`

	// 源站配置
	OriginType   string `json:"originType" gorm:"type:varchar(16)"` // domain, ip, oss
	OriginHost   string `json:"originHost" gorm:"type:varchar(128)"`
	OriginPort   int    `json:"originPort"`
	OriginProtocol string `json:"originProtocol" gorm:"type:varchar(8)"` // http, https

	// 加速配置
	EnableHTTPS    bool   `json:"enableHttps"`
	HTTP2Enabled   bool   `json:"http2Enabled"`
	ForceHTTPS     bool   `json:"forceHttps"`
	CertID         uint   `json:"certId"`

	// 缓存配置
	CacheConfig    string `json:"cacheConfig" gorm:"type:text"` // JSON
	DefaultTTL     int    `json:"defaultTtl"` // 默认缓存时间(秒)
	MaxTTL         int    `json:"maxTtl"`

	// 统计
	Bandwidth      int64   `json:"bandwidth"`      // 带宽使用(Mbps)
	Traffic        int64   `json:"traffic"`        // 流量使用(GB)
	RequestCount   int64   `json:"requestCount"`   // 请求数
	HitRate        float64 `json:"hitRate"`        // 缓存命中率
	AvgLatency     int64   `json:"avgLatency"`     // 平均延迟(ms)
	ErrorRate      float64 `json:"errorRate"`      // 错误率

	// 成本
	MonthlyCost    float64 `json:"monthlyCost"`    // 月成本

	// 自动优化
	AutoOptimize   bool    `json:"autoOptimize"`
	LastOptimizeAt *time.Time `json:"lastOptimizeAt"`

	// 区域配置
	Regions        string `json:"regions" gorm:"type:text"` // JSON 加速区域列表
}

func (CDNDomain) TableName() string {
	return "cdn_domains"
}

// CDNNode CDN 节点
type CDNNode struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	DomainID  uint   `json:"domainId" gorm:"index"`
	Name      string `json:"name" gorm:"type:varchar(64)"`
	IP        string `json:"ip" gorm:"type:varchar(64)"`
	Region    string `json:"region" gorm:"type:varchar(32)"`
	ISP       string `json:"isp" gorm:"type:varchar(32)"` // 运营商

	// 状态
	Status      string `json:"status" gorm:"type:varchar(16)"` // online, offline, degraded
	IsHealthy   bool   `json:"isHealthy"`

	// 性能指标
	Latency     int64   `json:"latency"` // ms
	Bandwidth   int64   `json:"bandwidth"` // Mbps
	Connections int     `json:"connections"`
	ErrorRate   float64 `json:"errorRate"`

	LastCheckAt *time.Time `json:"lastCheckAt"`
}

func (CDNNode) TableName() string {
	return "cdn_nodes"
}

// CDNCacheRule 缓存规则
type CDNCacheRule struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	DomainID  uint   `json:"domainId" gorm:"index"`
	Name      string `json:"name" gorm:"type:varchar(64)"`
	Enabled   bool   `json:"enabled"`

	// 匹配条件
	PathPattern string `json:"pathPattern" gorm:"type:varchar(256)"` // 支持通配符
	FileType    string `json:"fileType" gorm:"type:varchar(128)"`   // 文件类型
	HeaderMatch string `json:"headerMatch" gorm:"type:varchar(256)"` // Header 匹配

	// 缓存配置
	TTL         int    `json:"ttl"`         // 缓存时间(秒)
	CacheKey    string `json:"cacheKey" gorm:"type:text"` // 缓存 Key 规则
	IgnoreParam bool   `json:"ignoreParam"` // 忽略参数
	IgnoreCase  bool   `json:"ignoreCase"`  // 忽略大小写

	// 优先级
	Priority    int `json:"priority"`
}

func (CDNCacheRule) TableName() string {
	return "cdn_cache_rules"
}

// CDNOptimizationRecord 优化记录
type CDNOptimizationRecord struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	DomainID  uint           `json:"domainId" gorm:"index"`
	Domain    *CDNDomain     `json:"domain" gorm:"foreignKey:DomainID"`

	Type      OptimizeType   `json:"type" gorm:"type:varchar(16)"`
	Status    string         `json:"status" gorm:"type:varchar(16)"` // pending, running, success, failed

	// 优化前状态
	BeforeConfig string `json:"beforeConfig" gorm:"type:text"`
	BeforeStats  string `json:"beforeStats" gorm:"type:text"`

	// 优化后状态
	AfterConfig  string `json:"afterConfig" gorm:"type:text"`
	AfterStats   string `json:"afterStats" gorm:"type:text"`

	// AI 决策
	TriggerReason string  `json:"triggerReason" gorm:"type:text"`
	AIDecision    string  `json:"aiDecision" gorm:"type:text"`
	AIConfidence  float64 `json:"aiConfidence"`

	// 执行信息
	Actions       string `json:"actions" gorm:"type:text"` // JSON
	ExecutionLog  string `json:"executionLog" gorm:"type:text"`
	ErrorMessage  string `json:"errorMessage" gorm:"type:text"`

	// 收益
	BandwidthSaved int64   `json:"bandwidthSaved"` // 节省带宽(Mbps)
	CostSaved      float64 `json:"costSaved"`      // 节省成本

	// 时间
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 毫秒
}

func (CDNOptimizationRecord) TableName() string {
	return "cdn_optimization_records"
}

// CDNManager CDN 管理器
type CDNManager struct {
	llmClient *llm.GLM5Client
	notifier  notify.Notifier
	executor  CDNExecutor
}

// CDNExecutor CDN 执行器接口
type CDNExecutor interface {
	GetDomainStatus(domainID uint) (*DomainStatus, error)
	GetDomainMetrics(domainID uint, startTime, endTime time.Time) (*DomainMetrics, error)
	RefreshCache(domainID uint, paths []string) error
	UpdateCacheRule(domainID uint, rule CDNCacheRule) error
	UpdateOrigin(domainID uint, origin string) error
	GetNodeStatus(domainID uint) ([]CDNNode, error)
	PurgeCache(domainID uint, urls []string) error
	PreheatCache(domainID uint, urls []string) error
}

// DomainStatus 域名状态
type DomainStatus struct {
	Status      string  `json:"status"`
	Bandwidth   int64   `json:"bandwidth"`
	Traffic     int64   `json:"traffic"`
	RequestCount int64  `json:"requestCount"`
	HitRate     float64 `json:"hitRate"`
	AvgLatency  int64   `json:"avgLatency"`
	ErrorRate   float64 `json:"errorRate"`
}

// DomainMetrics 域名指标
type DomainMetrics struct {
	Timestamps    []int64   `json:"timestamps"`
	Bandwidth     []int64   `json:"bandwidth"`
	Traffic       []int64   `json:"traffic"`
	RequestCount  []int64   `json:"requestCount"`
	HitRate       []float64 `json:"hitRate"`
	Latency       []int64   `json:"latency"`
	ErrorRate     []float64 `json:"errorRate"`
	TopURLs       []URLStat `json:"topUrls"`
	TopRegions    []RegionStat `json:"topRegions"`
}

// URLStat URL 统计
type URLStat struct {
	URL       string `json:"url"`
	Requests  int64  `json:"requests"`
	Traffic   int64  `json:"traffic"`
	HitRate   float64 `json:"hitRate"`
}

// RegionStat 区域统计
type RegionStat struct {
	Region    string  `json:"region"`
	Requests  int64   `json:"requests"`
	Traffic   int64   `json:"traffic"`
	Latency   int64   `json:"latency"`
}

// NewCDNManager 创建 CDN 管理器
func NewCDNManager() *CDNManager {
	return &CDNManager{}
}

// SetLLMClient 设置 LLM 客户端
func (m *CDNManager) SetLLMClient(client *llm.GLM5Client) {
	m.llmClient = client
}

// SetNotifier 设置通知器
func (m *CDNManager) SetNotifier(notifier notify.Notifier) {
	m.notifier = notifier
}

// SetExecutor 设置执行器
func (m *CDNManager) SetExecutor(executor CDNExecutor) {
	m.executor = executor
}

// AnalyzeAndOptimize 分析并优化 CDN
func (m *CDNManager) AnalyzeAndOptimize(domain *CDNDomain) (*CDNOptimizationRecord, error) {
	// 获取当前状态
	status, err := m.executor.GetDomainStatus(domain.ID)
	if err != nil {
		return nil, fmt.Errorf("获取 CDN 状态失败: %w", err)
	}

	// 获取详细指标
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	metrics, _ := m.executor.GetDomainMetrics(domain.ID, startTime, endTime)

	// AI 分析
	decision, err := m.analyzeWithAI(domain, status, metrics)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 获取当前配置
	beforeConfig, _ := json.Marshal(domain)
	beforeStats, _ := json.Marshal(status)

	record := &CDNOptimizationRecord{
		DomainID:      domain.ID,
		Type:          decision.OptimizeType,
		Status:        "pending",
		BeforeConfig:  string(beforeConfig),
		BeforeStats:   string(beforeStats),
		TriggerReason: decision.Reason,
		AIDecision:    decision.Analysis,
		AIConfidence:  decision.Confidence,
	}

	global.DB.Create(record)

	return m.executeOptimization(record, domain, decision)
}

// CDNOptimizeDecision CDN 优化决策
type CDNOptimizeDecision struct {
	OptimizeType OptimizeType    `json:"optimizeType"`
	Reason       string          `json:"reason"`
	Analysis     string          `json:"analysis"`
	Confidence   float64         `json:"confidence"`
	Actions      []CDNAction     `json:"actions"`
}

// CDNAction CDN 动作
type CDNAction struct {
	Type     string      `json:"type"` // cache_rule, purge, preheat, origin
	Target   string      `json:"target"`
	Value    interface{} `json:"value"`
}

// analyzeWithAI AI 分析
func (m *CDNManager) analyzeWithAI(domain *CDNDomain, status *DomainStatus, metrics *DomainMetrics) (*CDNOptimizeDecision, error) {
	statusJSON, _ := json.Marshal(status)

	prompt := fmt.Sprintf(`你是一个 CDN 优化专家。请分析以下 CDN 状态并给出优化建议。

## CDN 域名信息
- 域名: %s
- 提供商: %s
- 状态: %s

## 当前状态
- 带宽: %d Mbps
- 流量: %d GB
- 请求数: %d
- 缓存命中率: %.2f%%
- 平均延迟: %dms
- 错误率: %.2f%%

## 详细状态
%s

请分析并给出优化建议，按以下 JSON 格式回复:
{
  "optimizeType": "cache/bandwidth/cost/performance/security",
  "reason": "优化原因",
  "analysis": "详细分析",
  "confidence": 0.0-1.0,
  "actions": [
    {"type": "cache_rule", "target": "*.jpg", "value": {"ttl": 86400}},
    {"type": "purge", "target": "/api/*"},
    {"type": "preheat", "target": "/static/*"}
  ]
}`,
		domain.Domain, domain.Provider, domain.Status,
		status.Bandwidth, status.Traffic, status.RequestCount,
		status.HitRate, status.AvgLatency, status.ErrorRate,
		string(statusJSON))

	response, err := m.llmClient.QuickChat(prompt)
	if err != nil {
		return nil, err
	}

	decision := &CDNOptimizeDecision{}
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart != -1 && jsonEnd != -1 {
		json.Unmarshal([]byte(response[jsonStart:jsonEnd+1]), decision)
	}

	return decision, nil
}

// executeOptimization 执行优化
func (m *CDNManager) executeOptimization(record *CDNOptimizationRecord, domain *CDNDomain, decision *CDNOptimizeDecision) (*CDNOptimizationRecord, error) {
	record.Status = "running"
	now := time.Now()
	record.StartedAt = &now
	global.DB.Save(record)

	var allOutput []string
	var actions []string
	success := true

	for _, action := range decision.Actions {
		switch action.Type {
		case "cache_rule":
			// 更新缓存规则
			rule := CDNCacheRule{
				DomainID: domain.ID,
				PathPattern: action.Target,
				Enabled: true,
			}
			if val, ok := action.Value.(map[string]interface{}); ok {
				if ttl, ok := val["ttl"].(float64); ok {
					rule.TTL = int(ttl)
				}
			}
			err := m.executor.UpdateCacheRule(domain.ID, rule)
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("更新缓存规则失败: %s", err.Error()))
				success = false
			} else {
				actions = append(actions, fmt.Sprintf("cache_rule(%s, ttl=%d)", action.Target, rule.TTL))
				allOutput = append(allOutput, fmt.Sprintf("更新缓存规则成功: %s", action.Target))
			}

		case "purge":
			// 刷新缓存
			paths := strings.Split(action.Target, ",")
			err := m.executor.PurgeCache(domain.ID, paths)
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("刷新缓存失败: %s", err.Error()))
				success = false
			} else {
				actions = append(actions, fmt.Sprintf("purge(%s)", action.Target))
				allOutput = append(allOutput, fmt.Sprintf("刷新缓存成功: %s", action.Target))
			}

		case "preheat":
			// 预热缓存
			urls := strings.Split(action.Target, ",")
			err := m.executor.PreheatCache(domain.ID, urls)
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("预热缓存失败: %s", err.Error()))
				success = false
			} else {
				actions = append(actions, fmt.Sprintf("preheat(%s)", action.Target))
				allOutput = append(allOutput, fmt.Sprintf("预热缓存成功: %s", action.Target))
			}

		case "origin":
			// 更新源站
			origin := action.Value.(string)
			err := m.executor.UpdateOrigin(domain.ID, origin)
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("更新源站失败: %s", err.Error()))
				success = false
			} else {
				actions = append(actions, fmt.Sprintf("origin(%s)", origin))
				allOutput = append(allOutput, fmt.Sprintf("更新源站成功: %s", origin))
			}
		}
	}

	// 更新记录
	actionsJSON, _ := json.Marshal(actions)
	record.Actions = string(actionsJSON)
	record.ExecutionLog = strings.Join(allOutput, "\n")

	completedAt := time.Now()
	record.CompletedAt = &completedAt
	if record.StartedAt != nil {
		record.Duration = completedAt.Sub(*record.StartedAt).Milliseconds()
	}

	if success {
		record.Status = "success"
	} else {
		record.Status = "failed"
	}

	global.DB.Save(record)

	// 更新域名
	domain.LastOptimizeAt = &completedAt
	global.DB.Save(domain)

	// 发送通知
	if m.notifier != nil {
		if success {
			m.notifier.SendMessage("CDN 优化完成",
				fmt.Sprintf("域名 %s CDN 优化完成，类型: %s", domain.Domain, decision.OptimizeType))
		} else {
			m.notifier.SendMessage("CDN 优化失败",
				fmt.Sprintf("域名 %s CDN 优化失败", domain.Domain))
		}
	}

	return record, nil
}

// AutoOptimizeCache 自动优化缓存
func (m *CDNManager) AutoOptimizeCache(domainID uint) error {
	var domain CDNDomain
	if err := global.DB.First(&domain, domainID).Error; err != nil {
		return fmt.Errorf("域名不存在")
	}

	// 获取统计信息
	status, _ := m.executor.GetDomainStatus(domainID)

	// 检查缓存命中率
	if status != nil && status.HitRate < 80 {
		// 命中率低，分析热点内容
		endTime := time.Now()
		startTime := endTime.Add(-1 * time.Hour)
		metrics, _ := m.executor.GetDomainMetrics(domainID, startTime, endTime)

		if metrics != nil && len(metrics.TopURLs) > 0 {
			// 预热热点内容
			var hotURLs []string
			for _, url := range metrics.TopURLs[:10] { // Top 10
				hotURLs = append(hotURLs, url.URL)
			}
			m.executor.PreheatCache(domainID, hotURLs)
		}
	}

	return nil
}

// OptimizeCost 成本优化
func (m *CDNManager) OptimizeCost(domainID uint) (*CDNOptimizationRecord, error) {
	var domain CDNDomain
	if err := global.DB.First(&domain, domainID).Error; err != nil {
		return nil, fmt.Errorf("域名不存在")
	}

	// 获取流量分布
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour)
	metrics, _ := m.executor.GetDomainMetrics(domainID, startTime, endTime)

	record := &CDNOptimizationRecord{
		DomainID:      domainID,
		Type:          OptimizeCost,
		Status:        "pending",
		TriggerReason: "成本优化分析",
	}

	global.DB.Create(record)

	// 分析并优化
	// 1. 识别低价值流量（低命中率的静态资源）
	// 2. 优化缓存策略
	// 3. 考虑回源优化

	var actions []CDNAction
	if metrics != nil {
		// 增加静态资源缓存时间
		actions = append(actions, CDNAction{
			Type:   "cache_rule",
			Target: "*.js,*.css,*.jpg,*.png,*.gif",
			Value:  map[string]interface{}{"ttl": 2592000}, // 30天
		})
	}

	decision := &CDNOptimizeDecision{
		OptimizeType: OptimizeCost,
		Actions:      actions,
	}

	return m.executeOptimization(record, &domain, decision)
}

// MonitorCDNs 监控 CDN
func (m *CDNManager) MonitorCDNs() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		var domains []CDNDomain
		global.DB.Where("auto_optimize = ? AND status = ?", true, CDNStatusActive).Find(&domains)

		for _, domain := range domains {
			status, err := m.executor.GetDomainStatus(domain.ID)
			if err != nil {
				continue
			}

			// 更新统计
			domain.Bandwidth = status.Bandwidth
			domain.Traffic = status.Traffic
			domain.RequestCount = status.RequestCount
			domain.HitRate = status.HitRate
			domain.AvgLatency = status.AvgLatency
			domain.ErrorRate = status.ErrorRate
			global.DB.Save(&domain)

			// 检查是否需要优化
			if status.HitRate < 70 || status.ErrorRate > 1 || status.AvgLatency > 500 {
				m.AnalyzeAndOptimize(&domain)
			}
		}
	}
}

// PurgeCache 刷新缓存
func (m *CDNManager) PurgeCache(domainID uint, urls []string) error {
	return m.executor.PurgeCache(domainID, urls)
}

// PreheatCache 预热缓存
func (m *CDNManager) PreheatCache(domainID uint, urls []string) error {
	return m.executor.PreheatCache(domainID, urls)
}

// GetNodeStatus 获取节点状态
func (m *CDNManager) GetNodeStatus(domainID uint) ([]CDNNode, error) {
	nodes, err := m.executor.GetNodeStatus(domainID)
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	for i := range nodes {
		nodes[i].DomainID = domainID
		nodes[i].LastCheckAt = new(time.Time)
		*nodes[i].LastCheckAt = time.Now()
		global.DB.Save(&nodes[i])
	}

	return nodes, nil
}

// CalculateCost 计算成本
func (m *CDNManager) CalculateCost(domain *CDNDomain) float64 {
	// 基础流量费用 + 带宽峰值费用
	var baseCost float64

	switch domain.Provider {
	case ProviderAliyun:
		baseCost = float64(domain.Traffic) * 0.24 // 每GB
	case ProviderTencent:
		baseCost = float64(domain.Traffic) * 0.21
	case ProviderAWS:
		baseCost = float64(domain.Traffic) * 0.085
	case ProviderCloudflare:
		baseCost = 0 // Cloudflare 按请求计费
	default:
		baseCost = float64(domain.Traffic) * 0.25
	}

	// HTTPS 请求费用
	if domain.EnableHTTPS {
		baseCost += float64(domain.RequestCount) * 0.00001 // 每万次请求
	}

	// 四舍五入到两位小数
	return math.Round(baseCost*100) / 100
}

// AddDomain 添加域名
func AddDomain(domain *CDNDomain) error {
	return global.DB.Create(domain).Error
}

// GetDomains 获取域名列表
func GetDomains() ([]CDNDomain, error) {
	var domains []CDNDomain
	err := global.DB.Find(&domains).Error
	return domains, err
}

// GetDomain 获取域名
func GetDomain(id uint) (*CDNDomain, error) {
	var domain CDNDomain
	err := global.DB.First(&domain, id).Error
	return &domain, err
}

// UpdateDomain 更新域名
func UpdateDomain(domain *CDNDomain) error {
	return global.DB.Save(domain).Error
}

// DeleteDomain 删除域名
func DeleteDomain(id uint) error {
	return global.DB.Delete(&CDNDomain{}, id).Error
}

// GetCacheRules 获取缓存规则
func GetCacheRules(domainID uint) ([]CDNCacheRule, error) {
	var rules []CDNCacheRule
	err := global.DB.Where("domain_id = ?", domainID).Order("priority DESC").Find(&rules).Error
	return rules, err
}

// GetOptimizationHistory 获取优化历史
func GetOptimizationHistory(domainID uint, limit int) ([]CDNOptimizationRecord, error) {
	var records []CDNOptimizationRecord
	query := global.DB.Model(&CDNOptimizationRecord{}).Order("created_at DESC")
	if domainID > 0 {
		query = query.Where("domain_id = ?", domainID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}
