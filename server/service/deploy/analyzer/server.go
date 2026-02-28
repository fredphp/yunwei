package analyzer

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"yunwei/global"
	"yunwei/model/server"
)

// ServerResourceType 服务器资源类型
type ServerResourceType string

const (
	ResourceTypeCompute ServerResourceType = "compute" // 计算型
	ResourceTypeMemory  ServerResourceType = "memory"  // 内存型
	ResourceTypeStorage ServerResourceType = "storage" // 存储型
	ResourceTypeBalanced ServerResourceType = "balanced" // 均衡型
	ResourceTypeGPU     ServerResourceType = "gpu"     // GPU型
)

// ServerRole 服务器角色
type ServerRole string

const (
	RoleWeb       ServerRole = "web"       // Web 服务器
	RoleAPI       ServerRole = "api"       // API 服务器
	RoleDatabase  ServerRole = "database"  // 数据库服务器
	RoleCache     ServerRole = "cache"     // 缓存服务器
	RoleMQ        ServerRole = "mq"        // 消息队列服务器
	RoleLB        ServerRole = "lb"        // 负载均衡
	RoleStorage   ServerRole = "storage"   // 存储服务器
	RoleWorker    ServerRole = "worker"    // 工作节点
	RoleMaster    ServerRole = "master"    // 主节点
	RoleSlave     ServerRole = "slave"     // 从节点
	RoleGateway   ServerRole = "gateway"   // 网关
)

// ServerCapability 服务器能力评估
type ServerCapability struct {
	ID              uint      `json:"id" gorm:"primarykey"`
	CreatedAt       time.Time `json:"createdAt"`
	
	ServerID        uint              `json:"serverId" gorm:"index"`
	Server          *server.Server    `json:"server" gorm:"foreignKey:ServerID"`
	
	// 资源评估
	ResourceType    ServerResourceType `json:"resourceType" gorm:"type:varchar(16)"`
	
	// 容量评估
	TotalScore      float64 `json:"totalScore"`     // 综合评分 0-100
	CPUScore        float64 `json:"cpuScore"`       // CPU 评分
	MemoryScore     float64 `json:"memoryScore"`    // 内存评分
	DiskScore       float64 `json:"diskScore"`      // 磁盘评分
	NetworkScore    float64 `json:"networkScore"`   // 网络评分
	
	// 可用资源
	AvailableCPU    int     `json:"availableCpu"`    // 可用 CPU 核心
	AvailableMemory uint64  `json:"availableMemory"` // 可用内存 MB
	AvailableDisk   uint64  `json:"availableDisk"`   // 可用磁盘 GB
	
	// 负载评估
	CPULoad         float64 `json:"cpuLoad"`        // CPU 负载率
	MemoryLoad      float64 `json:"memoryLoad"`     // 内存使用率
	DiskLoad        float64 `json:"diskLoad"`       // 磁盘使用率
	
	// 推荐角色
	RecommendedRoles string `json:"recommendedRoles" gorm:"type:text"` // JSON 数组
	CurrentRole     ServerRole `json:"currentRole" gorm:"type:varchar(16)"`
	
	// 容器支持
	DockerReady     bool    `json:"dockerReady"`
	K8sReady        bool    `json:"k8sReady"`
	ContainerCount  int     `json:"containerCount"`
	
	// 网络
	Bandwidth       int64   `json:"bandwidth"`      // Mbps
	Latency         int64   `json:"latency"`        // ms (到主节点的延迟)
	Region          string  `json:"region" gorm:"type:varchar(32)"`
	Zone            string  `json:"zone" gorm:"type:varchar(32)"`
}

func (ServerCapability) TableName() string {
	return "server_capabilities"
}

// ResourcePool 资源池
type ResourcePool struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	Name        string `json:"name" gorm:"type:varchar(64)"`
	Description string `json:"description" gorm:"type:text"`
	
	// 总资源
	TotalCPU       int     `json:"totalCpu"`
	TotalMemory    uint64  `json:"totalMemory"` // MB
	TotalDisk      uint64  `json:"totalDisk"`   // GB
	TotalBandwidth int64   `json:"totalBandwidth"` // Mbps
	
	// 可用资源
	AvailableCPU       int     `json:"availableCpu"`
	AvailableMemory    uint64  `json:"availableMemory"`
	AvailableDisk      uint64  `json:"availableDisk"`
	AvailableBandwidth int64   `json:"availableBandwidth"`
	
	// 服务器数量
	TotalServers    int `json:"totalServers"`
	OnlineServers   int `json:"onlineServers"`
	OfflineServers  int `json:"offlineServers"`
	
	// 分布
	Regions         string `json:"regions" gorm:"type:text"` // JSON 区域分布
	Zones           string `json:"zones" gorm:"type:text"`   // JSON 可用区分布
	
	// 服务分布
	ServiceDistribution string `json:"serviceDistribution" gorm:"type:text"` // JSON 服务分布
}

func (ResourcePool) TableName() string {
	return "resource_pools"
}

// ServerResourceAnalyzer 服务器资源分析器
type ServerResourceAnalyzer struct{}

// NewServerResourceAnalyzer 创建服务器资源分析器
func NewServerResourceAnalyzer() *ServerResourceAnalyzer {
	return &ServerResourceAnalyzer{}
}

// AnalyzeServer 分析单个服务器
func (a *ServerResourceAnalyzer) AnalyzeServer(srv *server.Server) (*ServerCapability, error) {
	capability := &ServerCapability{
		ServerID: srv.ID,
	}
	
	// 获取最新指标
	var metric server.ServerMetric
	result := global.DB.Where("server_id = ?", srv.ID).Order("created_at DESC").First(&metric)
	if result.Error == nil {
		capability.CPULoad = metric.CPUUsage
		capability.MemoryLoad = metric.MemoryUsage
		capability.DiskLoad = metric.DiskUsage
	}
	
	// 计算可用资源
	capability.AvailableCPU = srv.CPUCores
	capability.AvailableMemory = srv.MemoryTotal
	capability.AvailableDisk = srv.DiskTotal
	
	if capability.CPULoad > 0 {
		capability.AvailableCPU = int(float64(srv.CPUCores) * (1 - capability.CPULoad/100))
	}
	if capability.MemoryLoad > 0 {
		capability.AvailableMemory = uint64(float64(srv.MemoryTotal) * (1 - capability.MemoryLoad/100))
	}
	if capability.DiskLoad > 0 {
		capability.AvailableDisk = uint64(float64(srv.DiskTotal) * (1 - capability.DiskLoad/100))
	}
	
	// 计算评分
	capability.CPUScore = a.calculateCPUScore(srv.CPUCores, capability.CPULoad)
	capability.MemoryScore = a.calculateMemoryScore(srv.MemoryTotal, capability.MemoryLoad)
	capability.DiskScore = a.calculateDiskScore(srv.DiskTotal, capability.DiskLoad)
	capability.NetworkScore = 80 // 默认网络评分
	
	// 综合评分
	capability.TotalScore = (capability.CPUScore*0.3 + capability.MemoryScore*0.3 + 
		capability.DiskScore*0.2 + capability.NetworkScore*0.2)
	
	// 确定资源类型
	capability.ResourceType = a.determineResourceType(srv, capability)
	
	// 推荐角色
	roles := a.recommendRoles(srv, capability)
	rolesJSON, _ := json.Marshal(roles)
	capability.RecommendedRoles = string(rolesJSON)
	
	// 检查容器支持
	capability.DockerReady = a.checkDockerReady(srv)
	capability.K8sReady = a.checkK8sReady(srv)
	
	// 保存
	global.DB.Create(capability)
	
	return capability, nil
}

// AnalyzeAllServers 分析所有服务器
func (a *ServerResourceAnalyzer) AnalyzeAllServers() (*ResourcePool, error) {
	var servers []server.Server
	global.DB.Where("agent_online = ?", true).Find(&servers)
	
	pool := &ResourcePool{
		Name:        "默认资源池",
		Description: "所有在线服务器的资源池",
	}
	
	var capabilities []ServerCapability
	regionMap := make(map[string]int)
	zoneMap := make(map[string]int)
	
	for _, srv := range servers {
		cap, err := a.AnalyzeServer(&srv)
		if err != nil {
			continue
		}
		
		capabilities = append(capabilities, *cap)
		
		// 统计总资源
		pool.TotalCPU += srv.CPUCores
		pool.TotalMemory += srv.MemoryTotal
		pool.TotalDisk += srv.DiskTotal
		
		// 统计可用资源
		pool.AvailableCPU += cap.AvailableCPU
		pool.AvailableMemory += cap.AvailableMemory
		pool.AvailableDisk += cap.AvailableDisk
		
		// 统计区域
		if cap.Region != "" {
			regionMap[cap.Region]++
		}
		if cap.Zone != "" {
			zoneMap[cap.Zone]++
		}
	}
	
	pool.TotalServers = len(servers)
	pool.OnlineServers = len(capabilities)
	pool.OfflineServers = pool.TotalServers - pool.OnlineServers
	
	// 区域分布
	regionsJSON, _ := json.Marshal(regionMap)
	pool.Regions = string(regionsJSON)
	zonesJSON, _ := json.Marshal(zoneMap)
	pool.Zones = string(zonesJSON)
	
	return pool, nil
}

// calculateCPUScore 计算 CPU 评分
func (a *ServerResourceAnalyzer) calculateCPUScore(cores int, load float64) float64 {
	// CPU 核心数评分 (最多 50 分)
	coreScore := math.Min(float64(cores)*5, 50)
	
	// CPU 负载评分 (最多 50 分)
	loadScore := 50 * (1 - load/100)
	
	return coreScore + loadScore
}

// calculateMemoryScore 计算内存评分
func (a *ServerResourceAnalyzer) calculateMemoryScore(total uint64, load float64) float64 {
	// 内存大小评分 (最多 50 分)
	memoryGB := float64(total) / 1024
	sizeScore := math.Min(memoryGB*2, 50)
	
	// 内存使用率评分 (最多 50 分)
	loadScore := 50 * (1 - load/100)
	
	return sizeScore + loadScore
}

// calculateDiskScore 计算磁盘评分
func (a *ServerResourceAnalyzer) calculateDiskScore(total uint64, load float64) float64 {
	// 磁盘大小评分 (最多 50 分)
	sizeScore := math.Min(float64(total)*0.5, 50)
	
	// 磁盘使用率评分 (最多 50 分)
	loadScore := 50 * (1 - load/100)
	
	return sizeScore + loadScore
}

// determineResourceType 确定资源类型
func (a *ServerResourceAnalyzer) determineResourceType(srv *server.Server, cap *ServerCapability) ServerResourceType {
	cpuMemRatio := float64(srv.CPUCores) / (float64(srv.MemoryTotal) / 1024)
	
	// 计算型：CPU/内存比 > 1
	if cpuMemRatio > 1 {
		return ResourceTypeCompute
	}
	
	// 内存型：内存大，CPU/内存比 < 0.5
	if srv.MemoryTotal > 16*1024 && cpuMemRatio < 0.5 {
		return ResourceTypeMemory
	}
	
	// 存储型：磁盘大
	if srv.DiskTotal > 500 {
		return ResourceTypeStorage
	}
	
	return ResourceTypeBalanced
}

// recommendRoles 推荐角色
func (a *ServerResourceAnalyzer) recommendRoles(srv *server.Server, cap *ServerCapability) []ServerRole {
	var roles []ServerRole
	
	// 根据资源类型推荐
	switch cap.ResourceType {
	case ResourceTypeCompute:
		roles = append(roles, RoleWeb, RoleAPI, RoleWorker)
		
	case ResourceTypeMemory:
		roles = append(roles, RoleDatabase, RoleCache, RoleMQ)
		
	case ResourceTypeStorage:
		roles = append(roles, RoleStorage, RoleDatabase)
		
	case ResourceTypeBalanced:
		roles = append(roles, RoleWeb, RoleAPI, RoleWorker, RoleCache)
	}
	
	// 容器支持检查
	if cap.DockerReady {
		roles = append(roles, RoleWorker)
	}
	if cap.K8sReady {
		roles = append(roles, RoleMaster, RoleWorker)
	}
	
	return roles
}

// checkDockerReady 检查 Docker 是否就绪
func (a *ServerResourceAnalyzer) checkDockerReady(srv *server.Server) bool {
	// 检查服务器是否有 Docker 容器运行
	var containers []server.DockerContainer
	result := global.DB.Where("server_id = ?", srv.ID).Limit(1).Find(&containers)
	return result.RowsAffected > 0
}

// checkK8sReady 检查 K8s 是否就绪
func (a *ServerResourceAnalyzer) checkK8sReady(srv *server.Server) bool {
	// 检查是否有 kubectl 或 k8s 相关进程
	// 这里简化处理，实际需要检查进程
	return false
}

// ServerMatch 服务器匹配结果
type ServerMatch struct {
	ServerID    uint      `json:"serverId"`
	Score       float64   `json:"score"`
	Role        ServerRole `json:"role"`
	Reason      string    `json:"reason"`
}

// FindBestServers 查找最适合的服务器
func (a *ServerResourceAnalyzer) FindBestServers(requirements *ResourceRequirements, limit int) ([]ServerMatch, error) {
	var capabilities []ServerCapability
	global.DB.Where("available_cpu >= ? AND available_memory >= ? AND available_disk >= ?",
		requirements.MinCPU, requirements.MinMemory, requirements.MinDisk).
		Order("total_score DESC").
		Find(&capabilities)
	
	var matches []ServerMatch
	for _, cap := range capabilities {
		// 计算匹配分数
		score := a.calculateMatchScore(cap, requirements)
		
		// 确定角色
		role := a.determineRole(cap, requirements)
		
		matches = append(matches, ServerMatch{
			ServerID: cap.ServerID,
			Score:    score,
			Role:     role,
			Reason:   fmt.Sprintf("CPU: %d核, 内存: %dMB, 磁盘: %dGB", cap.AvailableCPU, cap.AvailableMemory, cap.AvailableDisk),
		})
	}
	
	// 按分数排序
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	
	return matches, nil
}

// ResourceRequirements 资源需求
type ResourceRequirements struct {
	MinCPU       int     `json:"minCpu"`
	MinMemory    uint64  `json:"minMemory"`
	MinDisk      uint64  `json:"minDisk"`
	PreferType   ServerResourceType `json:"preferType"`
	NeedDocker   bool    `json:"needDocker"`
	NeedK8s      bool    `json:"needK8s"`
	Region       string  `json:"region"`
	Zone         string  `json:"zone"`
}

// calculateMatchScore 计算匹配分数
func (a *ServerResourceAnalyzer) calculateMatchScore(cap ServerCapability, req *ResourceRequirements) float64 {
	score := cap.TotalScore
	
	// 资源类型加分
	if req.PreferType != "" && cap.ResourceType == req.PreferType {
		score += 20
	}
	
	// Docker 支持加分
	if req.NeedDocker && cap.DockerReady {
		score += 10
	}
	
	// K8s 支持加分
	if req.NeedK8s && cap.K8sReady {
		score += 10
	}
	
	// 区域加分
	if req.Region != "" && cap.Region == req.Region {
		score += 10
	}
	if req.Zone != "" && cap.Zone == req.Zone {
		score += 5
	}
	
	return score
}

// determineRole 确定角色
func (a *ServerResourceAnalyzer) determineRole(cap ServerCapability, req *ResourceRequirements) ServerRole {
	// 根据资源类型和需求确定角色
	var roles []ServerRole
	json.Unmarshal([]byte(cap.RecommendedRoles), &roles)
	
	if len(roles) > 0 {
		return roles[0]
	}
	
	return RoleWorker
}

// GetServerCapabilities 获取服务器能力列表
func GetServerCapabilities() ([]ServerCapability, error) {
	var capabilities []ServerCapability
	err := global.DB.Preload("Server").Find(&capabilities).Error
	return capabilities, err
}

// GetServerCapability 获取单个服务器能力
func GetServerCapability(serverID uint) (*ServerCapability, error) {
	var capability ServerCapability
	err := global.DB.Where("server_id = ?", serverID).First(&capability).Error
	return &capability, err
}
