package loadbalancer

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"yunwei/global"
	"yunwei/service/ai/llm"
	"yunwei/model/notify"
)

// LBType 负载均衡类型
type LBType string

const (
	LBTypeNginx   LBType = "nginx"
	LBTypeHAProxy LBType = "haproxy"
	LBTypeTraefik LBType = "traefik"
	LBTypeALB     LBType = "alb"     // AWS ALB
	LBTypeCLB     LBType = "clb"     // 阿里云 CLB
	LBTypeSLB     LBType = "slb"     // 腾讯云 CLB
)

// AlgorithmType 负载均衡算法
type AlgorithmType string

const (
	AlgorithmRoundRobin    AlgorithmType = "round_robin"
	AlgorithmLeastConn     AlgorithmType = "least_conn"
	AlgorithmIPHash        AlgorithmType = "ip_hash"
	AlgorithmWeighted      AlgorithmType = "weighted"
	AlgorithmRandom        AlgorithmType = "random"
	AlgorithmLeastTime     AlgorithmType = "least_time"
)

// LBStatus 负载均衡状态
type LBStatus string

const (
	LBStatusActive    LBStatus = "active"
	LBStatusDegraded  LBStatus = "degraded"
	LBStatusOffline   LBStatus = "offline"
	LBStatusOptimizing LBStatus = "optimizing"
)

// OptimizeStatus 优化状态
type OptimizeStatus string

const (
	OptimizeStatusPending   OptimizeStatus = "pending"
	OptimizeStatusRunning   OptimizeStatus = "running"
	OptimizeStatusSuccess   OptimizeStatus = "success"
	OptimizeStatusFailed    OptimizeStatus = "failed"
	OptimizeStatusRollback  OptimizeStatus = "rollback"
)

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Name        string `json:"name" gorm:"type:varchar(64)"`
	Type        LBType `json:"type" gorm:"type:varchar(16)"`
	ServerID    uint   `json:"serverId" gorm:"index"` // 所属服务器

	// 连接信息
	Host        string `json:"host" gorm:"type:varchar(64)"`
	Port        int    `json:"port"`
	ConfigPath  string `json:"configPath" gorm:"type:varchar(256)"`

	// 状态
	Status      LBStatus `json:"status" gorm:"type:varchar(16)"`

	// 自动优化
	AutoOptimize    bool    `json:"autoOptimize"`
	HealthCheckInt  int     `json:"healthCheckInt"` // 健康检查间隔(秒)
	HealthThreshold int     `json:"healthThreshold"` // 健康阈值

	// 统计
	TotalRequests   int64   `json:"totalRequests"`
	ActiveConns     int     `json:"activeConns"`
	AvgLatency      int64   `json:"avgLatency"` // ms
	ErrorRate       float64 `json:"errorRate"`

	LastOptimizeAt *time.Time `json:"lastOptimizeAt"`
}

func (LoadBalancer) TableName() string {
	return "load_balancers"
}

// BackendServer 后端服务器
type BackendServer struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	LBID      uint   `json:"lbId" gorm:"index"`
	Name      string `json:"name" gorm:"type:varchar(64)"`
	Host      string `json:"host" gorm:"type:varchar(64)"`
	Port      int    `json:"port"`
	Weight    int    `json:"weight"`
	MaxConns  int    `json:"maxConns"` // 最大连接数

	// 状态
	IsHealthy    bool   `json:"isHealthy"`
	Status       string `json:"status" gorm:"type:varchar(16)"` // up, down, draining
	LastCheckAt  *time.Time `json:"lastCheckAt"`
	ConsecutiveFails int `json:"consecutiveFails"`

	// 统计
	ActiveConns int     `json:"activeConns"`
	TotalReqs   int64   `json:"totalReqs"`
	AvgLatency  int64   `json:"avgLatency"`
	ErrorRate   float64 `json:"errorRate"`

	// 标签
	Region string `json:"region" gorm:"type:varchar(32)"`
	Zone   string `json:"zone" gorm:"type:varchar(32)"`
}

func (BackendServer) TableName() string {
	return "lb_backend_servers"
}

// OptimizationRecord 优化记录
type OptimizationRecord struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`

	LBID        uint           `json:"lbId" gorm:"index"`
	LB          *LoadBalancer  `json:"lb" gorm:"foreignKey:LBID"`

	Status      OptimizeStatus `json:"status" gorm:"type:varchar(16)"`
	Type        string         `json:"type" gorm:"type:varchar(32)"` // weight_adjust, algorithm_change, server_toggle

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
	Commands      string `json:"commands" gorm:"type:text"`
	ExecutionLog  string `json:"executionLog" gorm:"type:text"`
	ErrorMessage  string `json:"errorMessage" gorm:"type:text"`

	// 时间
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 毫秒
}

func (OptimizationRecord) TableName() string {
	return "lb_optimization_records"
}

// AlgorithmConfig 算法配置
type AlgorithmConfig struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	LBID       uint           `json:"lbId" gorm:"index"`
	Algorithm  AlgorithmType  `json:"algorithm" gorm:"type:varchar(16)"`

	// 参数
	Weights         string `json:"weights" gorm:"type:text"` // JSON
	HealthCheckPath string `json:"healthCheckPath" gorm:"type:varchar(128)"`
	HealthCheckInt  int    `json:"healthCheckInt"`

	// 会话保持
	SessionPersistence bool   `json:"sessionPersistence"`
	SessionMethod      string `json:"sessionMethod" gorm:"type:varchar(32)"` // cookie, ip_hash

	// 连接限制
	MaxConnsPerServer int `json:"maxConnsPerServer"`
	MaxQueueSize      int `json:"maxQueueSize"`
	QueueTimeout      int `json:"queueTimeout"` // ms

	Enabled bool `json:"enabled"`
}

func (AlgorithmConfig) TableName() string {
	return "lb_algorithm_configs"
}

// LBOptimizer 负载均衡优化器
type LBOptimizer struct {
	llmClient *llm.GLM5Client
	notifier  notify.Notifier
	executor  LBExecutor
}

// LBExecutor 负载均衡执行器接口
type LBExecutor interface {
	GetLBStatus(lbID uint) (*LoadBalancerStatus, error)
	GetBackendStatus(lbID uint, backendID uint) (*BackendStatus, error)
	UpdateWeight(lbID uint, backendID uint, weight int) error
	UpdateAlgorithm(lbID uint, algorithm AlgorithmType) error
	EnableBackend(lbID uint, backendID uint) error
	DisableBackend(lbID uint, backendID uint) error
	ReloadConfig(lbID uint) error
	GetConfig(lbID uint) (string, error)
	ApplyConfig(lbID uint, config string) error
}

// LoadBalancerStatus 负载均衡状态
type LoadBalancerStatus struct {
	TotalRequests int64   `json:"totalRequests"`
	ActiveConns   int     `json:"activeConns"`
	AvgLatency    int64   `json:"avgLatency"`
	ErrorRate     float64 `json:"errorRate"`
	Upstreams     []UpstreamStatus `json:"upstreams"`
}

// UpstreamStatus 上游状态
type UpstreamStatus struct {
	Name        string  `json:"name"`
	Server      string  `json:"server"`
	State       string  `json:"state"`
	ActiveConns int     `json:"activeConns"`
	TotalReqs   int64   `json:"totalReqs"`
	AvgLatency  int64   `json:"avgLatency"`
	ErrorRate   float64 `json:"errorRate"`
}

// BackendStatus 后端状态
type BackendStatus struct {
	IsHealthy    bool    `json:"isHealthy"`
	ActiveConns  int     `json:"activeConns"`
	TotalReqs    int64   `json:"totalReqs"`
	AvgLatency   int64   `json:"avgLatency"`
	ErrorRate    float64 `json:"errorRate"`
	ResponseTime int64   `json:"responseTime"`
}

// NewLBOptimizer 创建负载均衡优化器
func NewLBOptimizer() *LBOptimizer {
	return &LBOptimizer{}
}

// SetLLMClient 设置 LLM 客户端
func (o *LBOptimizer) SetLLMClient(client *llm.GLM5Client) {
	o.llmClient = client
}

// SetNotifier 设置通知器
func (o *LBOptimizer) SetNotifier(notifier notify.Notifier) {
	o.notifier = notifier
}

// SetExecutor 设置执行器
func (o *LBOptimizer) SetExecutor(executor LBExecutor) {
	o.executor = executor
}

// AnalyzeAndOptimize 分析并优化
func (o *LBOptimizer) AnalyzeAndOptimize(lb *LoadBalancer) (*OptimizationRecord, error) {
	// 获取当前状态
	status, err := o.executor.GetLBStatus(lb.ID)
	if err != nil {
		return nil, fmt.Errorf("获取负载均衡状态失败: %w", err)
	}

	// 获取后端服务器
	var backends []BackendServer
	global.DB.Where("lb_id = ?", lb.ID).Find(&backends)

	// AI 分析
	decision, err := o.analyzeWithAI(lb, status, backends)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 获取当前配置
	beforeConfig, _ := o.executor.GetConfig(lb.ID)
	beforeStatsJSON, _ := json.Marshal(status)

	record := &OptimizationRecord{
		LBID:          lb.ID,
		Status:        OptimizeStatusPending,
		Type:          decision.OptimizeType,
		BeforeConfig:  beforeConfig,
		BeforeStats:   string(beforeStatsJSON),
		TriggerReason: decision.Reason,
		AIDecision:    decision.Analysis,
		AIConfidence:  decision.Confidence,
	}

	global.DB.Create(record)

	// 执行优化
	return o.executeOptimization(record, lb, decision)
}

// OptimizeDecision 优化决策
type OptimizeDecision struct {
	OptimizeType string              `json:"optimizeType"`
	Reason       string              `json:"reason"`
	Analysis     string              `json:"analysis"`
	Confidence   float64             `json:"confidence"`
	Actions      []OptimizeAction    `json:"actions"`
}

// OptimizeAction 优化动作
type OptimizeAction struct {
	Type     string `json:"type"` // weight, algorithm, enable, disable
	Target   string `json:"target"`
	Value    interface{} `json:"value"`
}

// analyzeWithAI AI 分析
func (o *LBOptimizer) analyzeWithAI(lb *LoadBalancer, status *LoadBalancerStatus, backends []BackendServer) (*OptimizeDecision, error) {
	backendsJSON, _ := json.Marshal(backends)
	statusJSON, _ := json.Marshal(status)

	prompt := fmt.Sprintf(`你是一个负载均衡优化专家。请分析以下负载均衡状态并给出优化建议。

## 负载均衡器信息
- 名称: %s
- 类型: %s
- 状态: %s
- 总请求数: %d
- 活跃连接: %d
- 平均延迟: %dms
- 错误率: %.2f%%

## 后端服务器
%s

## 当前状态
%s

请分析并给出优化建议，按以下 JSON 格式回复:
{
  "optimizeType": "weight_adjust/algorithm_change/server_toggle",
  "reason": "优化原因",
  "analysis": "详细分析",
  "confidence": 0.0-1.0,
  "actions": [
    {"type": "weight", "target": "server_name", "value": 10},
    {"type": "disable", "target": "server_name"},
    {"type": "algorithm", "value": "least_conn"}
  ]
}`,
		lb.Name, lb.Type, lb.Status,
		status.TotalRequests, status.ActiveConns, status.AvgLatency, status.ErrorRate,
		string(backendsJSON), string(statusJSON))

	response, err := o.llmClient.QuickChat(prompt)
	if err != nil {
		return nil, err
	}

	decision := &OptimizeDecision{}
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart != -1 && jsonEnd != -1 {
		json.Unmarshal([]byte(response[jsonStart:jsonEnd+1]), decision)
	}

	return decision, nil
}

// executeOptimization 执行优化
func (o *LBOptimizer) executeOptimization(record *OptimizationRecord, lb *LoadBalancer, decision *OptimizeDecision) (*OptimizationRecord, error) {
	record.Status = OptimizeStatusRunning
	now := time.Now()
	record.StartedAt = &now
	global.DB.Save(record)

	var commands []string
	var allOutput []string
	success := true

	for _, action := range decision.Actions {
		switch action.Type {
		case "weight":
			// 调整权重
			backendName := action.Target
			weight := int(action.Value.(float64))

			var backend BackendServer
			if err := global.DB.Where("lb_id = ? AND name = ?", lb.ID, backendName).First(&backend).Error; err == nil {
				oldWeight := backend.Weight
				backend.Weight = weight
				global.DB.Save(&backend)

				err := o.executor.UpdateWeight(lb.ID, backend.ID, weight)
				if err != nil {
					allOutput = append(allOutput, fmt.Sprintf("调整权重失败 %s: %s", backendName, err.Error()))
					success = false
				} else {
					commands = append(commands, fmt.Sprintf("update_weight(%s, %d->%d)", backendName, oldWeight, weight))
					allOutput = append(allOutput, fmt.Sprintf("调整权重成功 %s: %d -> %d", backendName, oldWeight, weight))
				}
			}

		case "algorithm":
			// 更换算法
			algorithm := AlgorithmType(action.Value.(string))
			err := o.executor.UpdateAlgorithm(lb.ID, algorithm)
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("更新算法失败: %s", err.Error()))
				success = false
			} else {
				commands = append(commands, fmt.Sprintf("update_algorithm(%s)", algorithm))
				allOutput = append(allOutput, fmt.Sprintf("更新算法成功: %s", algorithm))
			}

		case "enable":
			// 启用后端
			backendName := action.Target
			var backend BackendServer
			if err := global.DB.Where("lb_id = ? AND name = ?", lb.ID, backendName).First(&backend).Error; err == nil {
				err := o.executor.EnableBackend(lb.ID, backend.ID)
				if err != nil {
					allOutput = append(allOutput, fmt.Sprintf("启用后端失败 %s: %s", backendName, err.Error()))
					success = false
				} else {
					backend.Status = "up"
					backend.IsHealthy = true
					global.DB.Save(&backend)
					commands = append(commands, fmt.Sprintf("enable_backend(%s)", backendName))
					allOutput = append(allOutput, fmt.Sprintf("启用后端成功: %s", backendName))
				}
			}

		case "disable":
			// 禁用后端
			backendName := action.Target
			var backend BackendServer
			if err := global.DB.Where("lb_id = ? AND name = ?", lb.ID, backendName).First(&backend).Error; err == nil {
				err := o.executor.DisableBackend(lb.ID, backend.ID)
				if err != nil {
					allOutput = append(allOutput, fmt.Sprintf("禁用后端失败 %s: %s", backendName, err.Error()))
					success = false
				} else {
					backend.Status = "down"
					backend.IsHealthy = false
					global.DB.Save(&backend)
					commands = append(commands, fmt.Sprintf("disable_backend(%s)", backendName))
					allOutput = append(allOutput, fmt.Sprintf("禁用后端成功: %s", backendName))
				}
			}
		}
	}

	// 重载配置
	if err := o.executor.ReloadConfig(lb.ID); err != nil {
		allOutput = append(allOutput, fmt.Sprintf("重载配置失败: %s", err.Error()))
		success = false
	} else {
		allOutput = append(allOutput, "重载配置成功")
	}

	// 获取优化后状态
	afterStatus, _ := o.executor.GetLBStatus(lb.ID)
	afterStatsJSON, _ := json.Marshal(afterStatus)
	afterConfig, _ := o.executor.GetConfig(lb.ID)

	// 更新记录
	commandsJSON, _ := json.Marshal(commands)
	record.Commands = string(commandsJSON)
	record.ExecutionLog = strings.Join(allOutput, "\n")
	record.AfterConfig = afterConfig
	record.AfterStats = string(afterStatsJSON)

	completedAt := time.Now()
	record.CompletedAt = &completedAt
	if record.StartedAt != nil {
		record.Duration = completedAt.Sub(*record.StartedAt).Milliseconds()
	}

	if success {
		record.Status = OptimizeStatusSuccess
	} else {
		record.Status = OptimizeStatusFailed
	}

	global.DB.Save(record)

	// 更新负载均衡器
	lb.LastOptimizeAt = &completedAt
	if afterStatus != nil {
		lb.TotalRequests = afterStatus.TotalRequests
		lb.ActiveConns = afterStatus.ActiveConns
		lb.AvgLatency = afterStatus.AvgLatency
		lb.ErrorRate = afterStatus.ErrorRate
	}
	global.DB.Save(lb)

	// 发送通知
	if o.notifier != nil {
		if success {
			o.notifier.SendMessage("负载均衡优化完成",
				fmt.Sprintf("负载均衡器 %s 优化完成，类型: %s", lb.Name, decision.OptimizeType))
		} else {
			o.notifier.SendMessage("负载均衡优化失败",
				fmt.Sprintf("负载均衡器 %s 优化失败", lb.Name))
		}
	}

	return record, nil
}

// AutoBalance 自动负载均衡
func (o *LBOptimizer) AutoBalance(lbID uint) error {
	var lb LoadBalancer
	if err := global.DB.First(&lb, lbID).Error; err != nil {
		return fmt.Errorf("负载均衡器不存在")
	}

	// 获取所有后端
	var backends []BackendServer
	global.DB.Where("lb_id = ?", lbID).Find(&backends)

	if len(backends) == 0 {
		return nil
	}

	// 计算基于响应时间和错误率的动态权重
	for i := range backends {
		backend := &backends[i]

		// 获取后端状态
		status, err := o.executor.GetBackendStatus(lbID, backend.ID)
		if err != nil {
			continue
		}

		// 计算权重分数
		score := o.calculateWeightScore(status)

		// 更新权重
		newWeight := int(math.Max(1, math.Min(100, float64(backend.Weight)*score)))
		if newWeight != backend.Weight {
			o.executor.UpdateWeight(lbID, backend.ID, newWeight)
			backend.Weight = newWeight
			global.DB.Save(backend)
		}
	}

	// 重载配置
	return o.executor.ReloadConfig(lbID)
}

// calculateWeightScore 计算权重分数
func (o *LBOptimizer) calculateWeightScore(status *BackendStatus) float64 {
	score := 1.0

	// 响应时间影响
	if status.ResponseTime > 0 {
		if status.ResponseTime > 1000 {
			score *= 0.5 // 响应时间超过1秒，降低权重
		} else if status.ResponseTime > 500 {
			score *= 0.8
		}
	}

	// 错误率影响
	if status.ErrorRate > 0 {
		score *= (1 - status.ErrorRate/100)
	}

	// 健康状态影响
	if !status.IsHealthy {
		score = 0.1 // 不健康的服务大幅降低权重
	}

	return score
}

// HealthCheck 健康检查
func (o *LBOptimizer) HealthCheck(lbID uint) error {
	var backends []BackendServer
	global.DB.Where("lb_id = ?", lbID).Find(&backends)

	for i := range backends {
		backend := &backends[i]
		status, err := o.executor.GetBackendStatus(lbID, backend.ID)
		if err != nil {
			backend.IsHealthy = false
			backend.ConsecutiveFails++
		} else {
			backend.IsHealthy = status.IsHealthy
			backend.ActiveConns = status.ActiveConns
			backend.AvgLatency = status.AvgLatency
			backend.ErrorRate = status.ErrorRate

			if !status.IsHealthy {
				backend.ConsecutiveFails++
			} else {
				backend.ConsecutiveFails = 0
			}
		}

		now := time.Now()
		backend.LastCheckAt = &now
		global.DB.Save(backend)
	}

	return nil
}

// MonitorLBs 监控负载均衡器
func (o *LBOptimizer) MonitorLBs() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var lbs []LoadBalancer
		global.DB.Where("auto_optimize = ? AND status = ?", true, LBStatusActive).Find(&lbs)

		for _, lb := range lbs {
			// 健康检查
			o.HealthCheck(lb.ID)

			// 获取状态
			status, err := o.executor.GetLBStatus(lb.ID)
			if err != nil {
				continue
			}

			// 检查是否需要优化
			if status.ErrorRate > 5.0 || status.AvgLatency > 500 {
				o.AnalyzeAndOptimize(&lb)
			}
		}
	}
}

// GetLBs 获取负载均衡器列表
func GetLBs() ([]LoadBalancer, error) {
	var lbs []LoadBalancer
	err := global.DB.Find(&lbs).Error
	return lbs, err
}

// GetLB 获取负载均衡器
func GetLB(id uint) (*LoadBalancer, error) {
	var lb LoadBalancer
	err := global.DB.First(&lb, id).Error
	return &lb, err
}

// GetBackends 获取后端服务器列表
func GetBackends(lbID uint) ([]BackendServer, error) {
	var backends []BackendServer
	err := global.DB.Where("lb_id = ?", lbID).Find(&backends).Error
	return backends, err
}

// GetOptimizationHistory 获取优化历史
func GetOptimizationHistory(lbID uint, limit int) ([]OptimizationRecord, error) {
	var records []OptimizationRecord
	query := global.DB.Model(&OptimizationRecord{}).Order("created_at DESC")
	if lbID > 0 {
		query = query.Where("lb_id = ?", lbID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}
