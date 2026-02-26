package planner

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"yunwei/global"
	deployAnalyzer "yunwei/service/deploy/analyzer"
	"yunwei/service/ai/llm"
)

// DeployPlanType 部署方案类型
type DeployPlanType string

const (
	PlanTypeSingle     DeployPlanType = "single"     // 单机部署
	PlanTypeCluster    DeployPlanType = "cluster"    // 集群部署
	PlanTypeDistributed DeployPlanType = "distributed" // 分布式部署
	PlanTypeHA         DeployPlanType = "ha"         // 高可用部署
	PlanTypeMicroservice DeployPlanType = "microservice" // 微服务部署
)

// DeployStatus 部署状态
type DeployStatus string

const (
	StatusDraft     DeployStatus = "draft"
	StatusPending   DeployStatus = "pending"
	StatusRunning   DeployStatus = "running"
	StatusSuccess   DeployStatus = "success"
	StatusFailed    DeployStatus = "failed"
	StatusRollback  DeployStatus = "rollback"
)

// DeployPlan 部署方案
type DeployPlan struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	// 基本信息
	Name        string         `json:"name" gorm:"type:varchar(128)"`
	Description string         `json:"description" gorm:"type:text"`
	PlanType    DeployPlanType `json:"planType" gorm:"type:varchar(16)"`
	Status      DeployStatus   `json:"status" gorm:"type:varchar(16)"`
	
	// 项目关联
	ProjectAnalysisID uint `json:"projectAnalysisId" gorm:"index"`
	
	// 服务器分配
	ServerAssignments string `json:"serverAssignments" gorm:"type:text"` // JSON 服务器分配列表
	
	// 服务配置
	Services       string `json:"services" gorm:"type:text"`       // JSON 服务配置
	ServiceTopology string `json:"serviceTopology" gorm:"type:text"` // JSON 服务拓扑
	
	// 网络配置
	NetworkConfig string `json:"networkConfig" gorm:"type:text"` // JSON 网络配置
	LoadBalancer  string `json:"loadBalancer" gorm:"type:text"`  // JSON 负载均衡配置
	
	// 存储配置
	StorageConfig string `json:"storageConfig" gorm:"type:text"` // JSON 存储配置
	
	// 数据库配置
	DatabaseConfig string `json:"databaseConfig" gorm:"type:text"` // JSON 数据库配置
	
	// 缓存配置
	CacheConfig string `json:"cacheConfig" gorm:"type:text"` // JSON 缓存配置
	
	// 消息队列配置
	MQConfig string `json:"mqConfig" gorm:"type:text"` // JSON MQ 配置
	
	// 环境变量
	EnvironmentVars string `json:"environmentVars" gorm:"type:text"` // JSON 环境变量
	
	// 部署顺序
	DeployOrder string `json:"deployOrder" gorm:"type:text"` // JSON 部署顺序
	
	// 预估成本
	EstimatedCost float64 `json:"estimatedCost"`
	
	// AI 建议
	AISuggestion string  `json:"aiSuggestion" gorm:"type:text"`
	Confidence   float64 `json:"confidence"`
	
	// 执行信息
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int64      `json:"duration"` // 秒
	Progress    int        `json:"progress"` // 0-100
	
	// 回滚信息
	RollbackPlan string `json:"rollbackPlan" gorm:"type:text"` // JSON 回滚方案
}

func (DeployPlan) TableName() string {
	return "deploy_plans"
}

// ServerAssignment 服务器分配
type ServerAssignment struct {
	ServerID   uint   `json:"serverId"`
	ServerName string `json:"serverName"`
	ServerIP   string `json:"serverIp"`
	Role       string `json:"role"`      // master, slave, worker, db, cache, mq
	Services   []string `json:"services"` // 部署的服务
	Resources  Resources `json:"resources"`
	Priority   int    `json:"priority"`  // 部署优先级
}

// Resources 资源分配
type Resources struct {
	CPU     int    `json:"cpu"`
	Memory  int    `json:"memory"` // MB
	Disk    int    `json:"disk"`   // GB
	Ports   []int  `json:"ports"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Image        string            `json:"image"`
	Replicas     int               `json:"replicas"`
	Ports        []PortMapping     `json:"ports"`
	Env          map[string]string `json:"env"`
	Resources    Resources         `json:"resources"`
	HealthCheck  HealthCheck       `json:"healthCheck"`
	Dependencies []string          `json:"dependencies"`
	Volumes      []Volume          `json:"volumes"`
}

// PortMapping 端口映射
type PortMapping struct {
	ContainerPort int    `json:"containerPort"`
	HostPort      int    `json:"hostPort"`
	Protocol      string `json:"protocol"`
}

// HealthCheck 健康检查
type HealthCheck struct {
	Type     string `json:"type"` // http, tcp, cmd
	Port     int    `json:"port"`
	Path     string `json:"path"`
	Interval int    `json:"interval"` // 秒
	Timeout  int    `json:"timeout"`  // 秒
	Retries  int    `json:"retries"`
}

// Volume 存储卷
type Volume struct {
	Name      string `json:"name"`
	HostPath  string `json:"hostPath"`
	MountPath string `json:"mountPath"`
	Size      string `json:"size"`
}

// ServiceTopology 服务拓扑
type ServiceTopology struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

// TopologyNode 拓扑节点
type TopologyNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Server string `json:"server"`
}

// TopologyEdge 拓扑边
type TopologyEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // depends, connects, replicates
}

// DeployPlanner 部署规划器
type DeployPlanner struct {
	llmClient *llm.GLM5Client
}

// NewDeployPlanner 创建部署规划器
func NewDeployPlanner() *DeployPlanner {
	return &DeployPlanner{}
}

// SetLLMClient 设置 LLM 客户端
func (p *DeployPlanner) SetLLMClient(client *llm.GLM5Client) {
	p.llmClient = client
}

// GeneratePlan 生成部署方案
func (p *DeployPlanner) GeneratePlan(analysis *deployAnalyzer.ProjectAnalysis, serverMatches []deployAnalyzer.ServerMatch) (*DeployPlan, error) {
	plan := &DeployPlan{
		Name:        fmt.Sprintf("%s-部署方案", analysis.ProjectName),
		Status:      StatusDraft,
	}
	
	// 确定部署类型
	plan.PlanType = p.determinePlanType(analysis, serverMatches)
	
	// 分配服务器
	assignments := p.assignServers(analysis, serverMatches)
	assignmentsJSON, _ := json.Marshal(assignments)
	plan.ServerAssignments = string(assignmentsJSON)
	
	// 生成服务配置
	services := p.generateServiceConfigs(analysis, assignments)
	servicesJSON, _ := json.Marshal(services)
	plan.Services = string(servicesJSON)
	
	// 生成服务拓扑
	topology := p.generateTopology(services, assignments)
	topologyJSON, _ := json.Marshal(topology)
	plan.ServiceTopology = string(topologyJSON)
	
	// 生成网络配置
	networkConfig := p.generateNetworkConfig(services, assignments)
	networkJSON, _ := json.Marshal(networkConfig)
	plan.NetworkConfig = string(networkJSON)
	
	// 生成负载均衡配置
	if analysis.NeedLB {
		lbConfig := p.generateLBConfig(services, assignments)
		lbJSON, _ := json.Marshal(lbConfig)
		plan.LoadBalancer = string(lbJSON)
	}
	
	// 生成数据库配置
	if analysis.NeedDBCluster {
		dbConfig := p.generateDatabaseConfig(analysis, assignments)
		dbJSON, _ := json.Marshal(dbConfig)
		plan.DatabaseConfig = string(dbJSON)
	}
	
	// 生成缓存配置
	if analysis.NeedCache {
		cacheConfig := p.generateCacheConfig(analysis, assignments)
		cacheJSON, _ := json.Marshal(cacheConfig)
		plan.CacheConfig = string(cacheJSON)
	}
	
	// 生成消息队列配置
	if analysis.NeedMQ {
		mqConfig := p.generateMQConfig(analysis, assignments)
		mqJSON, _ := json.Marshal(mqConfig)
		plan.MQConfig = string(mqJSON)
	}
	
	// 生成部署顺序
	deployOrder := p.generateDeployOrder(services)
	orderJSON, _ := json.Marshal(deployOrder)
	plan.DeployOrder = string(orderJSON)
	
	// AI 优化建议
	if p.llmClient != nil {
		suggestion := p.getAISuggestion(analysis, plan)
		plan.AISuggestion = suggestion.Analysis
		plan.Confidence = suggestion.Confidence
	}
	
	// 预估成本
	plan.EstimatedCost = p.estimateCost(assignments, services)
	
	// 保存方案
	global.DB.Create(plan)
	
	return plan, nil
}

// determinePlanType 确定部署类型
func (p *DeployPlanner) determinePlanType(analysis *deployAnalyzer.ProjectAnalysis, matches []deployAnalyzer.ServerMatch) DeployPlanType {
	// 根据项目类型和可用服务器数量确定
	switch analysis.ProjectType {
	case deployAnalyzer.ProjectTypeMicroservice:
		return PlanTypeMicroservice
		
	case deployAnalyzer.ProjectTypeDatabase:
		if len(matches) >= 3 {
			return PlanTypeHA
		}
		return PlanTypeSingle
		
	default:
		if analysis.NeedCluster && len(matches) >= 2 {
			if len(matches) >= 3 {
				return PlanTypeDistributed
			}
			return PlanTypeCluster
		}
		return PlanTypeSingle
	}
}

// assignServers 分配服务器
func (p *DeployPlanner) assignServers(analysis *deployAnalyzer.ProjectAnalysis, matches []deployAnalyzer.ServerMatch) []ServerAssignment {
	var assignments []ServerAssignment
	
	// 解析服务列表
	var projectServices []deployAnalyzer.ServiceInfo
	if analysis.Services != "" {
		json.Unmarshal([]byte(analysis.Services), &projectServices)
	}
	
	// 按分数排序服务器
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	
	serverIdx := 0
	
	// 分配主服务
	for i, svc := range projectServices {
		if serverIdx >= len(matches) {
			break
		}
		
		match := matches[serverIdx]
		assignment := ServerAssignment{
			ServerID:   match.ServerID,
			Role:       string(match.Role),
			Services:   []string{svc.Name},
			Priority:   10 - i, // 高优先级先部署
		}
		
		// 分配资源
		assignment.Resources = Resources{
			CPU:    analysis.RecCPU / len(projectServices),
			Memory: analysis.RecMemory / len(projectServices),
			Disk:   analysis.RecDisk / len(projectServices),
		}
		
		assignments = append(assignments, assignment)
		serverIdx++
	}
	
	// 如果需要负载均衡，分配 LB 服务器
	if analysis.NeedLB && serverIdx < len(matches) {
		match := matches[serverIdx]
		assignments = append(assignments, ServerAssignment{
			ServerID: match.ServerID,
			Role:     "lb",
			Services: []string{"load-balancer"},
			Priority: 100, // 最先部署
			Resources: Resources{
				CPU:    2,
				Memory: 1024,
				Disk:   10,
			},
		})
		serverIdx++
	}
	
	// 如果需要数据库集群，分配数据库服务器
	if analysis.NeedDBCluster && serverIdx < len(matches) {
		// 主库
		match := matches[serverIdx]
		assignments = append(assignments, ServerAssignment{
			ServerID: match.ServerID,
			Role:     "db-master",
			Services: []string{"database"},
			Priority: 90,
			Resources: Resources{
				CPU:    4,
				Memory: 4096,
				Disk:   100,
			},
		})
		serverIdx++
		
		// 从库
		for serverIdx < len(matches) && len(assignments) < 5 {
			match := matches[serverIdx]
			assignments = append(assignments, ServerAssignment{
				ServerID: match.ServerID,
				Role:     "db-slave",
				Services: []string{"database"},
				Priority: 80,
				Resources: Resources{
					CPU:    2,
					Memory: 2048,
					Disk:   100,
				},
			})
			serverIdx++
		}
	}
	
	// 如果需要缓存，分配缓存服务器
	if analysis.NeedCache && serverIdx < len(matches) {
		match := matches[serverIdx]
		assignments = append(assignments, ServerAssignment{
			ServerID: match.ServerID,
			Role:     "cache",
			Services: []string{"redis"},
			Priority: 70,
			Resources: Resources{
				CPU:    2,
				Memory: 4096,
				Disk:   20,
			},
		})
		serverIdx++
	}
	
	// 如果需要消息队列，分配 MQ 服务器
	if analysis.NeedMQ && serverIdx < len(matches) {
		match := matches[serverIdx]
		assignments = append(assignments, ServerAssignment{
			ServerID: match.ServerID,
			Role:     "mq",
			Services: []string{"rabbitmq"},
			Priority: 60,
			Resources: Resources{
				CPU:    2,
				Memory: 2048,
				Disk:   50,
			},
		})
	}
	
	return assignments
}

// generateServiceConfigs 生成服务配置
func (p *DeployPlanner) generateServiceConfigs(analysis *deployAnalyzer.ProjectAnalysis, assignments []ServerAssignment) []ServiceConfig {
	var services []ServiceConfig
	
	// 解析项目服务
	var projectServices []deployAnalyzer.ServiceInfo
	if analysis.Services != "" {
		json.Unmarshal([]byte(analysis.Services), &projectServices)
	}
	
	for _, svc := range projectServices {
		config := ServiceConfig{
			Name:     svc.Name,
			Type:     svc.Type,
			Replicas: svc.Replicas,
			Env:      svc.Env,
			Ports: []PortMapping{
				{ContainerPort: svc.Port, Protocol: "tcp"},
			},
			HealthCheck: HealthCheck{
				Type:     "http",
				Port:     svc.Port,
				Path:     svc.HealthCheck,
				Interval: 30,
				Timeout:  5,
				Retries:  3,
			},
		}
		
		// 根据技术栈设置镜像
		if analysis.TechStacks != "" {
			var stacks []deployAnalyzer.TechStack
			json.Unmarshal([]byte(analysis.TechStacks), &stacks)
			
			for _, stack := range stacks {
				switch stack {
				case deployAnalyzer.TechGo:
					config.Image = "golang:1.21"
				case deployAnalyzer.TechJava:
					config.Image = "openjdk:17"
				case deployAnalyzer.TechPython:
					config.Image = "python:3.11"
				case deployAnalyzer.TechNodeJS:
					config.Image = "node:20"
				case deployAnalyzer.TechReact, deployAnalyzer.TechVue:
					config.Image = "nginx:alpine"
				}
			}
		}
		
		services = append(services, config)
	}
	
	return services
}

// generateTopology 生成服务拓扑
func (p *DeployPlanner) generateTopology(services []ServiceConfig, assignments []ServerAssignment) ServiceTopology {
	topology := ServiceTopology{}
	
	// 创建节点
	for _, svc := range services {
		topology.Nodes = append(topology.Nodes, TopologyNode{
			ID:   svc.Name,
			Name: svc.Name,
			Type: svc.Type,
		})
	}
	
	// 分配服务器
	for _, assignment := range assignments {
		for _, svcName := range assignment.Services {
			for i := range topology.Nodes {
				if topology.Nodes[i].Name == svcName {
					topology.Nodes[i].Server = assignment.ServerName
				}
			}
		}
	}
	
	// 创建依赖关系边
	for _, svc := range services {
		for _, dep := range svc.Dependencies {
			topology.Edges = append(topology.Edges, TopologyEdge{
				Source: svc.Name,
				Target: dep,
				Type:   "depends",
			})
		}
	}
	
	return topology
}

// generateNetworkConfig 生成网络配置
func (p *DeployPlanner) generateNetworkConfig(services []ServiceConfig, assignments []ServerAssignment) map[string]interface{} {
	config := map[string]interface{}{
		"networks": []map[string]interface{}{
			{
				"name":   "app-network",
				"driver": "bridge",
				"subnet": "172.20.0.0/16",
			},
		},
		"firewall": map[string]interface{}{
			"enabled": true,
			"rules": []map[string]interface{}{
				{"port": 80, "protocol": "tcp", "action": "allow"},
				{"port": 443, "protocol": "tcp", "action": "allow"},
				{"port": 22, "protocol": "tcp", "action": "allow", "source": "any"},
			},
		},
	}
	
	return config
}

// generateLBConfig 生成负载均衡配置
func (p *DeployPlanner) generateLBConfig(services []ServiceConfig, assignments []ServerAssignment) map[string]interface{} {
	var backends []map[string]interface{}
	
	for _, assignment := range assignments {
		if strings.Contains(assignment.Role, "web") || strings.Contains(assignment.Role, "api") {
			backends = append(backends, map[string]interface{}{
				"server": assignment.ServerIP,
				"port":   8080,
				"weight": 10,
			})
		}
	}
	
	return map[string]interface{}{
		"type":     "nginx",
		"strategy": "round_robin",
		"backends": backends,
		"health_check": map[string]interface{}{
			"interval": 10,
			"timeout":  5,
			"retries":  3,
		},
		"ssl": map[string]interface{}{
			"enabled": true,
			"port":    443,
		},
	}
}

// generateDatabaseConfig 生成数据库配置
func (p *DeployPlanner) generateDatabaseConfig(analysis *deployAnalyzer.ProjectAnalysis, assignments []ServerAssignment) map[string]interface{} {
	var master map[string]interface{}
	var slaves []map[string]interface{}
	
	for _, assignment := range assignments {
		if assignment.Role == "db-master" {
			master = map[string]interface{}{
				"server": assignment.ServerIP,
				"port":   3306,
				"role":   "master",
			}
		} else if assignment.Role == "db-slave" {
			slaves = append(slaves, map[string]interface{}{
				"server": assignment.ServerIP,
				"port":   3306,
				"role":   "slave",
			})
		}
	}
	
	return map[string]interface{}{
		"type":      "mysql",
		"version":   "8.0",
		"master":    master,
		"slaves":    slaves,
		"replication": map[string]interface{}{
			"enabled":  true,
			"mode":     "async",
			"database": "app_db",
		},
		"backup": map[string]interface{}{
			"enabled":  true,
			"interval": 3600,
			"retain":   7,
		},
	}
}

// generateCacheConfig 生成缓存配置
func (p *DeployPlanner) generateCacheConfig(analysis *deployAnalyzer.ProjectAnalysis, assignments []ServerAssignment) map[string]interface{} {
	var nodes []map[string]interface{}
	
	for _, assignment := range assignments {
		if assignment.Role == "cache" {
			nodes = append(nodes, map[string]interface{}{
				"server": assignment.ServerIP,
				"port":   6379,
			})
		}
	}
	
	return map[string]interface{}{
		"type":    "redis",
		"version": "7.0",
		"mode":    "cluster",
		"nodes":   nodes,
		"memory":  "4gb",
		"persistence": map[string]interface{}{
			"enabled": true,
			"mode":    "rdb+aof",
		},
	}
}

// generateMQConfig 生成消息队列配置
func (p *DeployPlanner) generateMQConfig(analysis *deployAnalyzer.ProjectAnalysis, assignments []ServerAssignment) map[string]interface{} {
	var nodes []map[string]interface{}
	
	for _, assignment := range assignments {
		if assignment.Role == "mq" {
			nodes = append(nodes, map[string]interface{}{
				"server": assignment.ServerIP,
				"port":   5672,
			})
		}
	}
	
	return map[string]interface{}{
		"type":    "rabbitmq",
		"version": "3.12",
		"nodes":   nodes,
		"cluster": len(nodes) > 1,
		"queues": map[string]interface{}{
			"persistent": true,
			"ha":         len(nodes) > 1,
		},
	}
}

// generateDeployOrder 生成部署顺序
func (p *DeployPlanner) generateDeployOrder(services []ServiceConfig) []string {
	// 拓扑排序确定部署顺序
	visited := make(map[string]bool)
	order := []string{}
	
	// 构建依赖图
	graph := make(map[string][]string)
	for _, svc := range services {
		graph[svc.Name] = svc.Dependencies
	}
	
	// DFS 拓扑排序
	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		
		for _, dep := range graph[name] {
			visit(dep)
		}
		
		order = append(order, name)
	}
	
	for _, svc := range services {
		visit(svc.Name)
	}
	
	return order
}

// AISuggestion AI 建议
type AISuggestion struct {
	Analysis   string  `json:"analysis"`
	Confidence float64 `json:"confidence"`
}

// getAISuggestion 获取 AI 建议
func (p *DeployPlanner) getAISuggestion(analysis *deployAnalyzer.ProjectAnalysis, plan *DeployPlan) *AISuggestion {
	suggestion := &AISuggestion{}
	
	if p.llmClient == nil {
		return suggestion
	}
	
	prompt := fmt.Sprintf(`你是一个专业的运维架构师。请分析以下部署方案并给出优化建议。

项目信息:
- 名称: %s
- 类型: %s
- 技术栈: %s
- 资源需求: CPU %d核, 内存 %dMB, 磁盘 %dGB

部署方案:
- 类型: %s
- 服务数: 需分析

请给出:
1. 部署方案合理性评估
2. 潜在风险点
3. 优化建议

请按以下 JSON 格式回复:
{
  "analysis": "详细分析和建议",
  "confidence": 0.0-1.0
}`,
		analysis.ProjectName, analysis.ProjectType, analysis.TechStacks,
		analysis.RecCPU, analysis.RecMemory, analysis.RecDisk,
		plan.PlanType)
	
	response, err := p.llmClient.QuickChat(prompt)
	if err != nil {
		return suggestion
	}
	
	// 解析响应
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart != -1 && jsonEnd != -1 {
		json.Unmarshal([]byte(response[jsonStart:jsonEnd+1]), suggestion)
	}
	
	return suggestion
}

// estimateCost 预估成本
func (p *DeployPlanner) estimateCost(assignments []ServerAssignment, services []ServiceConfig) float64 {
	// 简单的成本计算
	// 实际可以根据云厂商价格计算
	cost := 0.0
	
	for _, assignment := range assignments {
		// 按资源估算月成本
		cpuCost := float64(assignment.Resources.CPU) * 20 // $20/核心/月
		memCost := float64(assignment.Resources.Memory) / 1024 * 10 // $10/GB/月
		diskCost := float64(assignment.Resources.Disk) * 0.1 // $0.1/GB/月
		
		cost += cpuCost + memCost + diskCost
	}
	
	return cost
}

// GetDeployPlan 获取部署方案
func GetDeployPlan(id uint) (*DeployPlan, error) {
	var plan DeployPlan
	err := global.DB.First(&plan, id).Error
	return &plan, err
}

// GetDeployPlans 获取部署方案列表
func GetDeployPlans() ([]DeployPlan, error) {
	var plans []DeployPlan
	err := global.DB.Order("created_at DESC").Find(&plans).Error
	return plans, err
}

// DeleteDeployPlan 删除部署方案
func DeleteDeployPlan(id uint) error {
	return global.DB.Delete(&DeployPlan{}, id).Error
}

// UpdateDeployPlan 更新部署方案
func UpdateDeployPlan(plan *DeployPlan) error {
	return global.DB.Save(plan).Error
}
