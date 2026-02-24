package agent

import (
	"fmt"
	"sync"
	"time"

	"yunwei/global"
	"yunwei/model/agent"
	"yunwei/model/server"
	"yunwei/service/notifier"
)

// HeartbeatMonitor 心跳监控器
type HeartbeatMonitor struct {
	heartbeatTimeout   time.Duration // 心跳超时时间
	checkInterval      time.Duration // 检查间隔
	notifier           *notifier.NotifierService
	offlineAgents      map[uint]*OfflineContext // 离线 Agent 上下文
	mu                 sync.RWMutex
	stopCh             chan struct{}
}

// OfflineContext 离线上下文
type OfflineContext struct {
	Agent        *agent.Agent
	OfflineAt    time.Time
	RecoverCount int
	LastAttempt  time.Time
}

// NewHeartbeatMonitor 创建心跳监控器
func NewHeartbeatMonitor() *HeartbeatMonitor {
	return &HeartbeatMonitor{
		heartbeatTimeout: 60 * time.Second, // 默认 60 秒超时
		checkInterval:    10 * time.Second, // 默认 10 秒检查一次
		offlineAgents:    make(map[uint]*OfflineContext),
		stopCh:           make(chan struct{}),
	}
}

// SetTimeout 设置超时时间
func (m *HeartbeatMonitor) SetTimeout(timeout time.Duration) {
	m.heartbeatTimeout = timeout
}

// SetCheckInterval 设置检查间隔
func (m *HeartbeatMonitor) SetCheckInterval(interval time.Duration) {
	m.checkInterval = interval
}

// SetNotifier 设置通知服务
func (m *HeartbeatMonitor) SetNotifier(n *notifier.NotifierService) {
	m.notifier = n
}

// Start 启动监控
func (m *HeartbeatMonitor) Start() {
	go m.run()
}

// Stop 停止监控
func (m *HeartbeatMonitor) Stop() {
	close(m.stopCh)
}

// run 运行监控循环
func (m *HeartbeatMonitor) run() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkAllAgents()
		}
	}
}

// checkAllAgents 检查所有 Agent
func (m *HeartbeatMonitor) checkAllAgents() {
	var agents []agent.Agent
	global.DB.Where("status IN ?", []agent.AgentStatus{
		agent.AgentStatusOnline,
		agent.AgentStatusOffline,
		agent.AgentStatusError,
	}).Find(&agents)

	now := time.Now()

	for _, ag := range agents {
		m.checkAgent(&ag, now)
	}
}

// checkAgent 检查单个 Agent
func (m *HeartbeatMonitor) checkAgent(ag *agent.Agent, now time.Time) {
	// 计算距离上次心跳的时间
	var lastHeartbeat time.Time
	if ag.LastHeartbeat != nil {
		lastHeartbeat = *ag.LastHeartbeat
	} else {
		// 没有心跳记录，使用创建时间
		lastHeartbeat = ag.CreatedAt
	}

	timeSinceLastHeartbeat := now.Sub(lastHeartbeat)

	// 判断是否超时
	if timeSinceLastHeartbeat > m.heartbeatTimeout {
		// Agent 超时
		m.handleAgentOffline(ag, now)
	} else {
		// Agent 正常
		m.handleAgentOnline(ag, now)
	}
}

// handleAgentOnline 处理 Agent 上线
func (m *HeartbeatMonitor) handleAgentOnline(ag *agent.Agent, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果之前是离线状态，现在恢复了
	if ag.Status == agent.AgentStatusOffline || ag.Status == agent.AgentStatusError {
		// 更新状态
		ag.Status = agent.AgentStatusOnline
		ag.StatusMessage = ""
		ag.LastOnlineAt = &now
		ag.OfflineCount = 0 // 重置离线计数
		global.DB.Save(ag)

		// 从离线列表移除
		delete(m.offlineAgents, ag.ID)

		// 发送通知
		if m.notifier != nil {
			m.notifier.Broadcast(
				fmt.Sprintf("🟢 Agent 恢复上线 - %s", ag.ServerName),
				fmt.Sprintf("Agent ID: %s\n版本: %s", ag.AgentID, ag.Version),
			)
		}

		// 记录心跳
		m.recordHeartbeat(ag, "recovered")
	}
}

// handleAgentOffline 处理 Agent 离线
func (m *HeartbeatMonitor) handleAgentOffline(ag *agent.Agent, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已经在处理中
	ctx, exists := m.offlineAgents[ag.ID]

	if !exists {
		// 新的离线事件
		ag.Status = agent.AgentStatusOffline
		ag.StatusMessage = "心跳超时"
		ag.OfflineCount++
		ag.LastOfflineAt = &now
		global.DB.Save(ag)

		// 添加到离线列表
		m.offlineAgents[ag.ID] = &OfflineContext{
			Agent:     ag,
			OfflineAt: now,
		}

		// 发送告警
		if m.notifier != nil {
			m.notifier.Broadcast(
				fmt.Sprintf("🔴 Agent 离线告警 - %s", ag.ServerName),
				fmt.Sprintf("Agent ID: %s\n离线时间: %s\n离线次数: %d",
					ag.AgentID, now.Format("2006-01-02 15:04:05"), ag.OfflineCount),
			)
		}

		// 记录心跳
		m.recordHeartbeat(ag, "offline")

		// 尝试自动恢复
		if ag.AutoRecover {
			go m.attemptRecovery(ag)
		}
	} else {
		// 已在处理中，检查是否需要重试
		if ag.AutoRecover && now.Sub(ctx.LastAttempt) > 5*time.Minute {
			go m.attemptRecovery(ag)
		}
	}
}

// attemptRecovery 尝试恢复
func (m *HeartbeatMonitor) attemptRecovery(ag *agent.Agent) {
	m.mu.Lock()
	ctx, exists := m.offlineAgents[ag.ID]
	if !exists {
		m.mu.Unlock()
		return
	}
	ctx.LastAttempt = time.Now()
	ctx.RecoverCount++
	m.mu.Unlock()

	// 记录恢复尝试
	record := &agent.AgentRecoverRecord{
		AgentID:     ag.ID,
		AgentUUID:   ag.AgentID,
		ServerID:    ag.ServerID,
		ServerName:  ag.ServerName,
		TriggerType: "auto",
		TriggerCause: "heartbeat_timeout",
		TriggerMsg:   fmt.Sprintf("心跳超时 %d 秒，自动尝试恢复", m.heartbeatTimeout/time.Second),
		Action:       "check_status",
		Status:       "running",
	}

	now := time.Now()
	record.StartedAt = &now
	global.DB.Create(record)

	// 获取服务器信息
	var srv server.Server
	if err := global.DB.First(&srv, ag.ServerID).Error; err != nil {
		record.Status = "failed"
		record.Error = fmt.Sprintf("服务器不存在: %v", err)
		record.Success = false
		global.DB.Save(record)
		return
	}

	// 尝试通过 SSH 检查 Agent 状态
	output, err := m.checkAgentStatus(&srv)
	if err != nil {
		record.Status = "failed"
		record.Error = err.Error()
		record.Output = output
		record.Success = false
		global.DB.Save(record)
		return
	}

	// 尝试重启 Agent
	restartOutput, restartErr := m.restartAgent(&srv)
	if restartErr != nil {
		record.Status = "failed"
		record.Error = restartErr.Error()
		record.Output = output + "\n" + restartOutput
		record.Success = false
		global.DB.Save(record)
		return
	}

	// 更新 Agent 状态
	ag.Status = agent.AgentStatusOnline
	ag.StatusMessage = "自动恢复成功"
	ag.RecoverCount++
	recoverTime := time.Now()
	ag.LastRecoverAt = &recoverTime
	global.DB.Save(ag)

	// 从离线列表移除
	m.mu.Lock()
	delete(m.offlineAgents, ag.ID)
	m.mu.Unlock()

	// 更新记录
	record.Status = "success"
	record.Action = "restart"
	record.Output = output + "\n" + restartOutput
	record.Success = true
	endTime := time.Now()
	record.CompletedAt = &endTime
	record.Duration = endTime.Sub(*record.StartedAt).Milliseconds()
	global.DB.Save(record)

	// 发送通知
	if m.notifier != nil {
		m.notifier.Broadcast(
			fmt.Sprintf("✅ Agent 自动恢复成功 - %s", ag.ServerName),
			fmt.Sprintf("Agent ID: %s\n恢复时间: %s", ag.AgentID, endTime.Format("2006-01-02 15:04:05")),
		)
	}
}

// checkAgentStatus 检查 Agent 状态
func (m *HeartbeatMonitor) checkAgentStatus(srv *server.Server) (string, error) {
	// TODO: 通过 SSH 执行检查命令
	// 这里是模拟实现
	checkCmd := "systemctl status yunwei-agent || ps aux | grep yunwei-agent"
	_ = checkCmd

	// 实际应该通过 SSH 执行
	// output, err := ssh.Execute(srv, checkCmd)
	// return output, err

	return "Agent 进程检查完成", nil
}

// restartAgent 重启 Agent
func (m *HeartbeatMonitor) restartAgent(srv *server.Server) (string, error) {
	// TODO: 通过 SSH 执行重启命令
	// 这里是模拟实现
	restartCmd := "systemctl restart yunwei-agent || (killall yunwei-agent && nohup /usr/local/bin/yunwei-agent &)"
	_ = restartCmd

	// 实际应该通过 SSH 执行
	// output, err := ssh.Execute(srv, restartCmd)
	// return output, err

	return "Agent 重启命令已执行", nil
}

// recordHeartbeat 记录心跳
func (m *HeartbeatMonitor) recordHeartbeat(ag *agent.Agent, status string) {
	record := &agent.AgentHeartbeatRecord{
		AgentID:   ag.ID,
		AgentUUID: ag.AgentID,
		ServerID:  ag.ServerID,
		Version:   ag.Version,
		Status:    status,
	}

	global.DB.Create(record)
}

// ==================== 心跳处理 ====================

// ProcessHeartbeat 处理心跳
func (m *HeartbeatMonitor) ProcessHeartbeat(req *HeartbeatRequest) (*HeartbeatResponse, error) {
	// 查找或创建 Agent
	var ag agent.Agent
	err := global.DB.Where("agent_id = ?", req.AgentID).First(&ag).Error

	if err != nil {
		// Agent 不存在，尝试通过 Server 关联查找
		var srv server.Server
		if err := global.DB.Where("agent_id = ?", req.AgentID).First(&srv).Error; err == nil {
			// 找到关联的服务器，创建 Agent 记录
			ag = agent.Agent{
				ServerID:       srv.ID,
				ServerName:     srv.Name,
				AgentID:        req.AgentID,
				Version:        req.Version,
				Platform:       req.Platform,
				Arch:           req.Arch,
				Status:         agent.AgentStatusOnline,
				AutoRecover:    true,
				AutoUpgrade:    true,
				UpgradeChannel: "stable",
			}
			global.DB.Create(&ag)
		} else {
			return nil, fmt.Errorf("未注册的 Agent: %s", req.AgentID)
		}
	}

	// 更新心跳时间
	now := time.Now()
	ag.LastHeartbeat = &now
	ag.HeartbeatIP = req.IP
	ag.HeartbeatPort = req.Port
	ag.UptimeSeconds = req.UptimeSeconds

	// 如果之前是离线状态，现在恢复了
	if ag.Status == agent.AgentStatusOffline {
		ag.Status = agent.AgentStatusOnline
		ag.StatusMessage = ""

		// 发送通知
		if m.notifier != nil {
			m.notifier.Broadcast(
				fmt.Sprintf("🟢 Agent 恢复上线 - %s", ag.ServerName),
				fmt.Sprintf("Agent ID: %s\n版本: %s", ag.AgentID, ag.Version),
			)
		}

		// 从离线列表移除
		m.mu.Lock()
		delete(m.offlineAgents, ag.ID)
		m.mu.Unlock()
	}

	// 更新版本（如果变化）
	if req.Version != "" && req.Version != ag.Version {
		ag.Version = req.Version
		versionCode, _ := NewVersionManager("").ParseVersionCode(req.Version)
		ag.VersionCode = versionCode
	}

	global.DB.Save(&ag)

	// 记录心跳
	heartbeatRecord := &agent.AgentHeartbeatRecord{
		AgentID:        ag.ID,
		AgentUUID:      ag.AgentID,
		ServerID:       ag.ServerID,
		IP:             req.IP,
		Port:           req.Port,
		Version:        req.Version,
		Status:         "online",
		UptimeSeconds:  req.UptimeSeconds,
		CPUUsage:       req.CPUUsage,
		MemoryUsage:    req.MemoryUsage,
		GoroutineCount: req.GoroutineCount,
		PendingTasks:   req.PendingTasks,
		RunningTasks:   req.RunningTasks,
		CompletedTasks: req.CompletedTasks,
	}
	global.DB.Create(heartbeatRecord)

	// 检查是否有待执行的升级任务
	var pendingUpgrade agent.AgentUpgradeTask
	err = global.DB.Where("agent_id = ? AND status = ?", ag.ID, "pending").
		Order("priority DESC, created_at ASC").
		First(&pendingUpgrade).Error

	response := &HeartbeatResponse{
		Success: true,
		Message: "OK",
	}

	if err == nil {
		// 有待执行的升级任务
		response.NeedUpgrade = true
		response.UpgradeTaskID = pendingUpgrade.ID
		response.TargetVersion = pendingUpgrade.ToVersion
	}

	return response, nil
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	AgentID        string  `json:"agentId"`
	IP             string  `json:"ip"`
	Port           int     `json:"port"`
	Version        string  `json:"version"`
	Platform       string  `json:"platform"`
	Arch           string  `json:"arch"`
	UptimeSeconds  int64   `json:"uptimeSeconds"`
	CPUUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	GoroutineCount int     `json:"goroutineCount"`
	PendingTasks   int     `json:"pendingTasks"`
	RunningTasks   int     `json:"runningTasks"`
	CompletedTasks int     `json:"completedTasks"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	NeedUpgrade   bool   `json:"needUpgrade"`
	UpgradeTaskID uint   `json:"upgradeTaskId"`
	TargetVersion string `json:"targetVersion"`
}

// ==================== 统计 ====================

// GetMonitorStats 获取监控统计
func (m *HeartbeatMonitor) GetMonitorStats() *MonitorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &MonitorStats{}

	// 在线/离线数量
	global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusOnline).Count(&stats.OnlineCount)
	global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusOffline).Count(&stats.OfflineCount)
	global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusError).Count(&stats.ErrorCount)

	// 今日心跳总数
	today := time.Now().Truncate(24 * time.Hour)
	global.DB.Model(&agent.AgentHeartbeatRecord{}).Where("created_at >= ?", today).Count(&stats.TodayHeartbeats)

	// 今日恢复次数
	global.DB.Model(&agent.AgentRecoverRecord{}).
		Where("created_at >= ? AND success = ?", today, true).
		Count(&stats.TodayRecoveries)

	// 离线 Agent 列表
	stats.OfflineAgents = len(m.offlineAgents)

	return stats
}

// MonitorStats 监控统计
type MonitorStats struct {
	OnlineCount     int64 `json:"onlineCount"`
	OfflineCount    int64 `json:"offlineCount"`
	ErrorCount      int64 `json:"errorCount"`
	TodayHeartbeats int64 `json:"todayHeartbeats"`
	TodayRecoveries int64 `json:"todayRecoveries"`
	OfflineAgents   int   `json:"offlineAgents"`
}

// GetOfflineAgents 获取离线 Agent 列表
func (m *HeartbeatMonitor) GetOfflineAgents() []OfflineAgentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []OfflineAgentInfo
	for id, ctx := range m.offlineAgents {
		result = append(result, OfflineAgentInfo{
			AgentID:     id,
			AgentUUID:   ctx.Agent.AgentID,
			ServerName:  ctx.Agent.ServerName,
			OfflineAt:   ctx.OfflineAt,
			RecoverCount: ctx.RecoverCount,
		})
	}

	return result
}

// OfflineAgentInfo 离线 Agent 信息
type OfflineAgentInfo struct {
	AgentID      uint      `json:"agentId"`
	AgentUUID    string    `json:"agentUuid"`
	ServerName   string    `json:"serverName"`
	OfflineAt    time.Time `json:"offlineAt"`
	RecoverCount int       `json:"recoverCount"`
}
