package ha

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/model/ha"
)

// ClusterManager 集群管理器
type ClusterManager struct {
	mu              sync.RWMutex
	localNode       *ha.ClusterNode
	nodes           map[string]*ha.ClusterNode
	leaderService   *LeaderElectionService
	config          *ha.HAClusterConfig
	heartbeatCancel context.CancelFunc
}

// NewClusterManager 创建集群管理器
func NewClusterManager(nodeID string) *ClusterManager {
	return &ClusterManager{
		nodes: make(map[string]*ha.ClusterNode),
	}
}

// SetLeaderService 设置 Leader 选举服务
func (m *ClusterManager) SetLeaderService(ls *LeaderElectionService) {
	m.leaderService = ls
}

// SetConfig 设置配置
func (m *ClusterManager) SetConfig(config *ha.HAClusterConfig) {
	m.config = config
}

// ==================== 节点管理 ====================

// RegisterNode 注册节点
func (m *ClusterManager) RegisterNode(node *ha.ClusterNode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 设置默认值
	if node.Status == "" {
		node.Status = ha.NodeStatusStarting
	}
	if node.Role == "" {
		node.Role = "follower"
	}

	// 保存到数据库
	var existing ha.ClusterNode
	err := global.DB.Where("node_id = ?", node.NodeID).First(&existing).Error
	if err == nil {
		// 更新
		now := time.Now()
		updates := map[string]interface{}{
			"hostname":       node.Hostname,
			"internal_ip":    node.InternalIP,
			"external_ip":    node.ExternalIP,
			"api_port":       node.APIPort,
			"grpc_port":      node.GRPCPort,
			"version":        node.Version,
			"go_version":     node.GoVersion,
			"data_center":    node.DataCenter,
			"zone":           node.Zone,
			"rack":           node.Rack,
			"weight":         node.Weight,
			"status":         ha.NodeStatusOnline,
			"last_heartbeat": now,
		}
		global.DB.Model(&existing).Updates(updates)
		m.nodes[node.NodeID] = &existing
	} else {
		// 创建
		global.DB.Create(node)
		m.nodes[node.NodeID] = node
	}

	// 记录事件
	m.recordEvent("node_registered", node.NodeID, node.NodeName, 
		fmt.Sprintf("Node %s registered from %s", node.NodeName, node.InternalIP))

	return nil
}

// UnregisterNode 注销节点
func (m *ClusterManager) UnregisterNode(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return nil
	}

	// 更新状态
	global.DB.Model(&ha.ClusterNode{}).Where("node_id = ?", nodeID).
		Update("status", ha.NodeStatusOffline)
	delete(m.nodes, nodeID)

	// 记录事件
	m.recordEvent("node_unregistered", nodeID, node.NodeName,
		fmt.Sprintf("Node %s unregistered", node.NodeName))

	return nil
}

// UpdateHeartbeat 更新心跳
func (m *ClusterManager) UpdateHeartbeat(nodeID string, metrics *NodeMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		// 尝试从数据库加载
		var dbNode ha.ClusterNode
		if err := global.DB.Where("node_id = ?", nodeID).First(&dbNode).Error; err == nil {
			m.nodes[nodeID] = &dbNode
			node = &dbNode
		} else {
			return fmt.Errorf("node not found: %s", nodeID)
		}
	}

	now := time.Now()

	// 更新内存
	node.LastHeartbeat = &now
	node.HeartbeatCount++
	node.Status = ha.NodeStatusOnline

	if metrics != nil {
		node.CPUUsage = metrics.CPUUsage
		node.MemoryUsage = metrics.MemoryUsage
		node.DiskUsage = metrics.DiskUsage
		node.GoroutineCount = metrics.GoroutineCount
		node.ConnectionCount = metrics.ConnectionCount
		node.RequestCount += metrics.RequestCount
	}

	// 更新数据库
	updates := map[string]interface{}{
		"last_heartbeat":  now,
		"heartbeat_count": node.HeartbeatCount,
		"status":          ha.NodeStatusOnline,
		"cpu_usage":       node.CPUUsage,
		"memory_usage":    node.MemoryUsage,
		"disk_usage":      node.DiskUsage,
		"goroutine_count": node.GoroutineCount,
		"connection_count": node.ConnectionCount,
		"request_count":   node.RequestCount,
	}
	global.DB.Model(&ha.ClusterNode{}).Where("node_id = ?", nodeID).Updates(updates)

	// 记录指标
	if metrics != nil {
		metric := &ha.NodeMetric{
			NodeID:          node.ID,
			NodeUUID:        node.NodeID,
			CPUUsage:        metrics.CPUUsage,
			MemoryUsage:     metrics.MemoryUsage,
			MemoryUsed:      metrics.MemoryUsed,
			MemoryTotal:     metrics.MemoryTotal,
			DiskUsage:       metrics.DiskUsage,
			GoroutineCount:  metrics.GoroutineCount,
			RequestCount:    metrics.RequestCount,
			RequestLatency:  metrics.RequestLatency,
			RequestQPS:      metrics.RequestQPS,
			ConnectionCount: metrics.ConnectionCount,
			Load1:           metrics.Load1,
			Load5:           metrics.Load5,
			Load15:          metrics.Load15,
		}
		global.DB.Create(metric)
	}

	return nil
}

// GetNode 获取节点
func (m *ClusterManager) GetNode(nodeID string) (*ha.ClusterNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	return node, nil
}

// ListNodes 列出节点
func (m *ClusterManager) ListNodes(filter *NodeFilter) ([]ha.ClusterNode, int64, error) {
	query := global.DB.Model(&ha.ClusterNode{})

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Role != "" {
			query = query.Where("role = ?", filter.Role)
		}
		if filter.DataCenter != "" {
			query = query.Where("data_center = ?", filter.DataCenter)
		}
		if filter.Zone != "" {
			query = query.Where("zone = ?", filter.Zone)
		}
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
	}

	var total int64
	query.Count(&total)

	var nodes []ha.ClusterNode
	err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&nodes).Error
	return nodes, total, err
}

// NodeFilter 节点过滤器
type NodeFilter struct {
	Status     string `json:"status"`
	Role       string `json:"role"`
	DataCenter string `json:"dataCenter"`
	Zone       string `json:"zone"`
	Enabled    *bool  `json:"enabled"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	CPUUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	MemoryUsed     uint64  `json:"memoryUsed"`
	MemoryTotal    uint64  `json:"memoryTotal"`
	DiskUsage      float64 `json:"diskUsage"`
	GoroutineCount int     `json:"goroutineCount"`
	RequestCount   int64   `json:"requestCount"`
	RequestLatency float64 `json:"requestLatency"`
	RequestQPS     float64 `json:"requestQps"`
	ConnectionCount int    `json:"connectionCount"`
	Load1          float64 `json:"load1"`
	Load5          float64 `json:"load5"`
	Load15         float64 `json:"load15"`
}

// ==================== 心跳监控 ====================

// StartHeartbeatMonitor 启动心跳监控
func (m *ClusterManager) StartHeartbeatMonitor(ctx context.Context) {
	if m.config == nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	m.heartbeatCancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(m.config.HeartbeatTimeout) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.checkHeartbeats()
			}
		}
	}()
}

// StopHeartbeatMonitor 停止心跳监控
func (m *ClusterManager) StopHeartbeatMonitor() {
	if m.heartbeatCancel != nil {
		m.heartbeatCancel()
	}
}

// checkHeartbeats 检查心跳
func (m *ClusterManager) checkHeartbeats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return
	}

	timeout := time.Duration(m.config.HeartbeatTimeout) * time.Second
	now := time.Now()

	for nodeID, node := range m.nodes {
		if node.LastHeartbeat == nil {
			continue
		}

		// 检查是否超时
		if now.Sub(*node.LastHeartbeat) > timeout {
			// 节点离线
			oldStatus := node.Status
			node.Status = ha.NodeStatusOffline

			// 更新数据库
			global.DB.Model(&ha.ClusterNode{}).Where("node_id = ?", nodeID).
				Update("status", ha.NodeStatusOffline)

			// 记录事件
			m.recordEvent("node_offline", nodeID, node.NodeName,
				fmt.Sprintf("Node %s offline (heartbeat timeout)", node.NodeName))

			// 如果是 Leader，触发故障转移
			if node.IsLeader && m.config.FailoverEnabled {
				go m.handleFailover(node)
			}

			_ = oldStatus
		}
	}
}

// ==================== 故障转移 ====================

// handleFailover 处理故障转移
func (m *ClusterManager) handleFailover(failedNode *ha.ClusterNode) {
	if m.config == nil || !m.config.FailoverEnabled {
		return
	}

	m.recordEvent("failover_started", failedNode.NodeID, failedNode.NodeName,
		fmt.Sprintf("Failover started for node %s", failedNode.NodeName))

	now := time.Now()
	record := &ha.FailoverRecord{
		FailoverType:   "node",
		FailedNodeID:   failedNode.NodeID,
		FailedNodeName: failedNode.NodeName,
		FailedNodeIP:   failedNode.InternalIP,
		Reason:         "Heartbeat timeout",
		DetectedAt:     &now,
		Status:         "running",
		TriggerType:    "auto",
	}
	global.DB.Create(record)

	// 选择新节点
	newLeader, err := m.selectNewLeader(failedNode)
	if err != nil {
		record.Status = "failed"
		record.Error = err.Error()
		global.DB.Save(record)
		return
	}

	record.TargetNodeID = newLeader.NodeID
	record.TargetNodeName = newLeader.NodeName
	record.TargetNodeIP = newLeader.InternalIP

	// 执行故障转移
	startTime := time.Now()
	record.StartedAt = &startTime

	// 更新 Leader
	if m.leaderService != nil {
		if err := m.leaderService.ForceLeader(newLeader.NodeID); err != nil {
			record.Status = "failed"
			record.Error = err.Error()
			global.DB.Save(record)
			return
		}
	}

	// 完成
	endTime := time.Now()
	record.Status = "completed"
	record.CompletedAt = &endTime
	record.Duration = endTime.Sub(startTime).Milliseconds()
	record.Success = true
	global.DB.Save(record)

	// 记录事件
	m.recordEvent("failover_completed", newLeader.NodeID, newLeader.NodeName,
		fmt.Sprintf("Failover completed, new leader: %s", newLeader.NodeName))
}

// selectNewLeader 选择新 Leader
func (m *ClusterManager) selectNewLeader(excludeNode *ha.ClusterNode) (*ha.ClusterNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var candidates []*ha.ClusterNode
	for _, node := range m.nodes {
		if node.NodeID == excludeNode.NodeID {
			continue
		}
		if node.Status != ha.NodeStatusOnline {
			continue
		}
		if !node.Enabled {
			continue
		}
		candidates = append(candidates, node)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available nodes for failover")
	}

	// 按权重排序，选择权重最高的
	bestNode := candidates[0]
	for _, node := range candidates[1:] {
		if node.Weight > bestNode.Weight {
			bestNode = node
		}
	}

	return bestNode, nil
}

// ==================== 负载均衡 ====================

// SelectNode 选择节点（负载均衡）
func (m *ClusterManager) SelectNode(strategy string) (*ha.ClusterNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var onlineNodes []*ha.ClusterNode
	for _, node := range m.nodes {
		if node.Status == ha.NodeStatusOnline && node.Enabled {
			onlineNodes = append(onlineNodes, node)
		}
	}

	if len(onlineNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	switch strategy {
	case "round-robin":
		return m.selectRoundRobin(onlineNodes)
	case "least-connections":
		return m.selectLeastConnections(onlineNodes)
	case "weighted":
		return m.selectWeighted(onlineNodes)
	case "random":
		return m.selectRandom(onlineNodes)
	default:
		return onlineNodes[0], nil
	}
}

// selectRoundRobin 轮询选择
func (m *ClusterManager) selectRoundRobin(nodes []*ha.ClusterNode) (*ha.ClusterNode, error) {
	// 简单实现：选择心跳次数最少的（近似轮询）
	minCount := nodes[0].HeartbeatCount
	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.HeartbeatCount < minCount {
			minCount = node.HeartbeatCount
			selected = node
		}
	}
	return selected, nil
}

// selectLeastConnections 最少连接选择
func (m *ClusterManager) selectLeastConnections(nodes []*ha.ClusterNode) (*ha.ClusterNode, error) {
	minConn := nodes[0].ConnectionCount
	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.ConnectionCount < minConn {
			minConn = node.ConnectionCount
			selected = node
		}
	}
	return selected, nil
}

// selectWeighted 加权选择
func (m *ClusterManager) selectWeighted(nodes []*ha.ClusterNode) (*ha.ClusterNode, error) {
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += node.Weight
	}

	// 简单实现：选择权重占比最高的
	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.Weight > selected.Weight {
			selected = node
		}
	}
	return selected, nil
}

// selectRandom 随机选择
func (m *ClusterManager) selectRandom(nodes []*ha.ClusterNode) (*ha.ClusterNode, error) {
	// 简单实现：选择第一个
	return nodes[0], nil
}

// ==================== 集群状态 ====================

// GetClusterStats 获取集群统计
func (m *ClusterManager) GetClusterStats() *ClusterStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &ClusterStats{}
	stats.TotalNodes = int64(len(m.nodes))

	for _, node := range m.nodes {
		switch node.Status {
		case ha.NodeStatusOnline:
			stats.OnlineNodes++
		case ha.NodeStatusOffline:
			stats.OfflineNodes++
		case ha.NodeStatusStarting:
			stats.StartingNodes++
		case ha.NodeStatusStopping:
			stats.StoppingNodes++
		}

		if node.IsLeader {
			stats.LeaderNodeID = node.NodeID
			stats.LeaderNodeName = node.NodeName
		}

		stats.TotalCPUUsage += node.CPUUsage
		stats.TotalMemoryUsage += node.MemoryUsage
		stats.TotalConnections += node.ConnectionCount
		stats.TotalRequests += node.RequestCount
	}

	if stats.TotalNodes > 0 {
		stats.AvgCPUUsage = stats.TotalCPUUsage / float64(stats.TotalNodes)
		stats.AvgMemoryUsage = stats.TotalMemoryUsage / float64(stats.TotalNodes)
	}

	return stats
}

// ClusterStats 集群统计
type ClusterStats struct {
	TotalNodes       int64   `json:"totalNodes"`
	OnlineNodes      int64   `json:"onlineNodes"`
	OfflineNodes     int64   `json:"offlineNodes"`
	StartingNodes    int64   `json:"startingNodes"`
	StoppingNodes    int64   `json:"stoppingNodes"`
	LeaderNodeID     string  `json:"leaderNodeId"`
	LeaderNodeName   string  `json:"leaderNodeName"`
	TotalCPUUsage    float64 `json:"totalCpuUsage"`
	TotalMemoryUsage float64 `json:"totalMemoryUsage"`
	TotalConnections int     `json:"totalConnections"`
	TotalRequests    int64   `json:"totalRequests"`
	AvgCPUUsage      float64 `json:"avgCpuUsage"`
	AvgMemoryUsage   float64 `json:"avgMemoryUsage"`
}

// ==================== 工具方法 ====================

// recordEvent 记录事件
func (m *ClusterManager) recordEvent(eventType, nodeID, nodeName, detail string) {
	event := &ha.ClusterEvent{
		EventType: eventType,
		NodeID:    nodeID,
		NodeName:  nodeName,
		Detail:    detail,
		Level:     "info",
		Source:    "cluster_manager",
	}
	global.DB.Create(event)
}

// EnableNode 启用节点
func (m *ClusterManager) EnableNode(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	global.DB.Model(&ha.ClusterNode{}).Where("node_id = ?", nodeID).
		Update("enabled", true)

	if node, exists := m.nodes[nodeID]; exists {
		node.Enabled = true
	}

	return nil
}

// DisableNode 禁用节点
func (m *ClusterManager) DisableNode(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	global.DB.Model(&ha.ClusterNode{}).Where("node_id = ?", nodeID).
		Update("enabled", false)

	if node, exists := m.nodes[nodeID]; exists {
		node.Enabled = false
	}

	return nil
}

// GetNodeMetrics 获取节点指标历史
func (m *ClusterManager) GetNodeMetrics(nodeID string, hours int) ([]ha.NodeMetric, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var node ha.ClusterNode
	if err := global.DB.Where("node_id = ?", nodeID).First(&node).Error; err != nil {
		return nil, err
	}

	var metrics []ha.NodeMetric
	err := global.DB.Where("node_id = ? AND created_at > ?", node.ID, startTime).
		Order("created_at ASC").Find(&metrics).Error
	return metrics, err
}

// ExportClusterState 导出集群状态
func (m *ClusterManager) ExportClusterState() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := map[string]interface{}{
		"nodes":   m.nodes,
		"stats":   m.GetClusterStats(),
		"config":  m.config,
		"exportedAt": time.Now(),
	}

	return json.Marshal(state)
}
