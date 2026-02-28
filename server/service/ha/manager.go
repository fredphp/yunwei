package ha

import (
        "context"
        "fmt"
        "os"
        "sync"
        "time"

        "yunwei/global"
        "yunwei/model/ha"
)

// HAManager HA 管理器
type HAManager struct {
        mu             sync.RWMutex
        nodeID         string
        config         *ha.HAClusterConfig
        lockService    *DistributedLockService
        leaderService  *LeaderElectionService
        clusterManager *ClusterManager
        sessionManager *SessionManager
        taskHAManager  *TaskHAManager

        started bool
        stopCh  chan struct{}
}

// HAOption HA 选项
type HAOption func(*HAManager)

// NewHAManager 创建 HA 管理器
func NewHAManager(opts ...HAOption) *HAManager {
        nodeID := getHostname()
        if nodeID == "" {
                nodeID = fmt.Sprintf("node-%d", time.Now().UnixNano())
        }

        mgr := &HAManager{
                nodeID: nodeID,
                stopCh: make(chan struct{}),
        }

        for _, opt := range opts {
                opt(mgr)
        }

        // 初始化服务
        mgr.initServices()

        return mgr
}

// WithNodeID 设置节点 ID
func WithNodeID(nodeID string) HAOption {
        return func(m *HAManager) {
                m.nodeID = nodeID
        }
}

// WithConfig 设置配置
func WithConfig(config *ha.HAClusterConfig) HAOption {
        return func(m *HAManager) {
                m.config = config
        }
}

// initServices 初始化服务
func (m *HAManager) initServices() {
        // 加载配置
        if m.config == nil {
                m.loadConfig()
        }

        // 创建分布式锁服务
        lockBackend := NewDatabaseLockBackend(global.DB)
        m.lockService = NewDistributedLockService(m.nodeID, lockBackend)

        // 创建 Leader 选举服务
        leaseDuration := 15 * time.Second
        if m.config != nil {
                leaseDuration = time.Duration(m.config.LeaderLeaseSeconds) * time.Second
        }
        m.leaderService = NewLeaderElectionService(m.nodeID, "yunwei-leader", leaseDuration)

        // 创建集群管理器
        m.clusterManager = NewClusterManager(m.nodeID)
        m.clusterManager.SetLeaderService(m.leaderService)
        m.clusterManager.SetConfig(m.config)

        // 创建会话管理器
        sessionTTL := 30 * time.Minute
        if m.config != nil {
                sessionTTL = time.Duration(m.config.SessionTTL) * time.Second
        }
        m.sessionManager = NewSessionManager(m.nodeID, sessionTTL)

        // 创建任务 HA 管理器
        m.taskHAManager = NewTaskHAManager(m.nodeID, m.lockService, m.clusterManager)
}

// loadConfig 加载配置
func (m *HAManager) loadConfig() {
        var config ha.HAClusterConfig
        err := global.DB.Where("enabled = ?", true).First(&config).Error
        if err == nil {
                m.config = &config
        } else {
                // 创建默认配置
                m.config = &ha.HAClusterConfig{
                        Name:               "default",
                        ClusterMode:        "active-active",
                        MinNodes:           1,
                        MaxNodes:           10,
                        HeartbeatInterval:  10,
                        HeartbeatTimeout:   30,
                        ElectionTimeout:    30,
                        LeaderLeaseSeconds: 15,
                        FailoverEnabled:    true,
                        FailoverTimeout:    60,
                        LoadBalanceEnabled: true,
                        LockBackend:        "database",
                        SessionMode:        "memory",
                        Enabled:            true,
                }
                global.DB.Create(m.config)
        }
}

// ==================== 启动停止 ====================

// Start 启动 HA 服务
func (m *HAManager) Start(ctx context.Context) error {
        m.mu.Lock()
        defer m.mu.Unlock()

        if m.started {
                return nil
        }

        // 注册本节点
        m.registerSelf()

        // 启动 Leader 选举
        if err := m.leaderService.Start(ctx); err != nil {
                return fmt.Errorf("failed to start leader election: %w", err)
        }

        // 启动心跳监控
        m.clusterManager.StartHeartbeatMonitor(ctx)

        // 启动会话清理
        m.sessionManager.StartCleanup(ctx)

        // 启动任务监控
        m.taskHAManager.StartTaskMonitor(ctx)

        // 启动心跳发送
        go m.heartbeatLoop(ctx)

        m.started = true
        return nil
}

// Stop 停止 HA 服务
func (m *HAManager) Stop() {
        m.mu.Lock()
        defer m.mu.Unlock()

        if !m.started {
                return
        }

        close(m.stopCh)

        // 停止各服务
        m.leaderService.Stop()
        m.clusterManager.StopHeartbeatMonitor()
        m.sessionManager.StopCleanup()

        // 注销本节点
        m.unregisterSelf()

        m.started = false
}

// registerSelf 注册本节点
func (m *HAManager) registerSelf() error {
        node := &ha.ClusterNode{
                NodeID:     m.nodeID,
                NodeName:   m.nodeID,
                Hostname:   getHostname(),
                Status:     ha.NodeStatusOnline,
                Role:       "follower",
                Enabled:    true,
                Version:    "1.0.0",
                GoVersion:  "1.21",
        }

        return m.clusterManager.RegisterNode(node)
}

// unregisterSelf 注销本节点
func (m *HAManager) unregisterSelf() {
        m.clusterManager.UnregisterNode(m.nodeID)
}

// heartbeatLoop 心跳循环
func (m *HAManager) heartbeatLoop(ctx context.Context) {
        interval := 10 * time.Second
        if m.config != nil {
                interval = time.Duration(m.config.HeartbeatInterval) * time.Second
        }

        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
                select {
                case <-ctx.Done():
                        return
                case <-m.stopCh:
                        return
                case <-ticker.C:
                        m.sendHeartbeat()
                }
        }
}

// sendHeartbeat 发送心跳
func (m *HAManager) sendHeartbeat() {
        // 收集本节点指标
        metrics := &NodeMetrics{
                // TODO: 收集实际指标
        }

        m.clusterManager.UpdateHeartbeat(m.nodeID, metrics)
}

// ==================== 状态查询 ====================

// IsLeader 是否是 Leader
func (m *HAManager) IsLeader() bool {
        return m.leaderService.IsLeader()
}

// GetNodeID 获取节点 ID
func (m *HAManager) GetNodeID() string {
        return m.nodeID
}

// GetLeader 获取 Leader 节点
func (m *HAManager) GetLeader() string {
        return m.leaderService.GetLeader()
}

// GetConfig 获取配置
func (m *HAManager) GetConfig() *ha.HAClusterConfig {
        return m.config
}

// GetClusterStats 获取集群统计
func (m *HAManager) GetClusterStats() *ClusterStats {
        return m.clusterManager.GetClusterStats()
}

// ==================== 服务访问 ====================

// GetLockService 获取分布式锁服务
func (m *HAManager) GetLockService() *DistributedLockService {
        return m.lockService
}

// GetLeaderService 获取 Leader 选举服务
func (m *HAManager) GetLeaderService() *LeaderElectionService {
        return m.leaderService
}

// GetClusterManager 获取集群管理器
func (m *HAManager) GetClusterManager() *ClusterManager {
        return m.clusterManager
}

// GetSessionManager 获取会话管理器
func (m *HAManager) GetSessionManager() *SessionManager {
        return m.sessionManager
}

// GetTaskHAManager 获取任务 HA 管理器
func (m *HAManager) GetTaskHAManager() *TaskHAManager {
        return m.taskHAManager
}

// ==================== 配置管理 ====================

// UpdateConfig 更新配置
func (m *HAManager) UpdateConfig(config *ha.HAClusterConfig) error {
        if err := global.DB.Save(config).Error; err != nil {
                return err
        }

        m.mu.Lock()
        m.config = config
        m.clusterManager.SetConfig(config)
        m.mu.Unlock()

        return nil
}

// ==================== 故障转移 ====================

// TriggerFailover 手动触发故障转移
func (m *HAManager) TriggerFailover(nodeID string) error {
        node, err := m.clusterManager.GetNode(nodeID)
        if err != nil {
                return err
        }

        go m.clusterManager.handleFailover(node)
        return nil
}

// ==================== 统计信息 ====================

// GetHAStats 获取 HA 统计
func (m *HAManager) GetHAStats() *HAStats {
        stats := &HAStats{
                NodeID:      m.nodeID,
                IsLeader:    m.IsLeader(),
                LeaderNode:  m.GetLeader(),
                Started:     m.started,
        }

        // 集群统计
        clusterStats := m.clusterManager.GetClusterStats()
        stats.TotalNodes = clusterStats.TotalNodes
        stats.OnlineNodes = clusterStats.OnlineNodes
        stats.OfflineNodes = clusterStats.OfflineNodes

        // 会话统计
        sessionStats := m.sessionManager.GetSessionStats()
        stats.ActiveSessions = sessionStats.ActiveSessions

        // 锁统计
        locks, _, _ := m.lockService.ListLocks(nil)
        stats.ActiveLocks = int64(len(locks))

        // 任务统计
        stats.RunningTasks = int64(len(m.taskHAManager.GetRunningTasks()))

        return stats
}

// HAStats HA 统计
type HAStats struct {
        NodeID        string    `json:"nodeId"`
        IsLeader      bool      `json:"isLeader"`
        LeaderNode    string    `json:"leaderNode"`
        Started       bool      `json:"started"`
        TotalNodes    int64     `json:"totalNodes"`
        OnlineNodes   int64     `json:"onlineNodes"`
        OfflineNodes  int64     `json:"offlineNodes"`
        ActiveLocks   int64     `json:"activeLocks"`
        ActiveSessions int      `json:"activeSessions"`
        RunningTasks  int64     `json:"runningTasks"`
        Timestamp     time.Time `json:"timestamp"`
}

// ==================== 事件查询 ====================

// GetEvents 获取事件列表
func (m *HAManager) GetEvents(filter *EventFilter) ([]ha.ClusterEvent, int64, error) {
        query := global.DB.Model(&ha.ClusterEvent{})

        if filter != nil {
                if filter.EventType != "" {
                        query = query.Where("event_type = ?", filter.EventType)
                }
                if filter.NodeID != "" {
                        query = query.Where("node_id = ?", filter.NodeID)
                }
                if filter.Level != "" {
                        query = query.Where("level = ?", filter.Level)
                }
                if !filter.StartTime.IsZero() {
                        query = query.Where("created_at >= ?", filter.StartTime)
                }
                if !filter.EndTime.IsZero() {
                        query = query.Where("created_at <= ?", filter.EndTime)
                }
        }

        var total int64
        query.Count(&total)

        var events []ha.ClusterEvent
        err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&events).Error
        return events, total, err
}

// EventFilter 事件过滤器
type EventFilter struct {
        EventType string    `json:"eventType"`
        NodeID    string    `json:"nodeId"`
        Level     string    `json:"level"`
        StartTime time.Time `json:"startTime"`
        EndTime   time.Time `json:"endTime"`
        Limit     int       `json:"limit"`
        Offset    int       `json:"offset"`
}

// GetFailoverRecords 获取故障转移记录
func (m *HAManager) GetFailoverRecords(filter *FailoverFilter) ([]ha.FailoverRecord, int64, error) {
        query := global.DB.Model(&ha.FailoverRecord{})

        if filter != nil {
                if filter.Status != "" {
                        query = query.Where("status = ?", filter.Status)
                }
                if filter.FailoverType != "" {
                        query = query.Where("failover_type = ?", filter.FailoverType)
                }
        }

        var total int64
        query.Count(&total)

        var records []ha.FailoverRecord
        err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&records).Error
        return records, total, err
}

// FailoverFilter 故障转移过滤器
type FailoverFilter struct {
        Status       string `json:"status"`
        FailoverType string `json:"failoverType"`
        Limit        int    `json:"limit"`
        Offset       int    `json:"offset"`
}

// ==================== 工具函数 ====================

// getHostname 获取主机名
func getHostname() string {
        hostname, _ := os.Hostname()
        return hostname
}
