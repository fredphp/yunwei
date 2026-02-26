package agent

import (
        "crypto/md5"
        "encoding/hex"
        "encoding/json"
        "fmt"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/model/agent"
        "yunwei/model/server"
        "yunwei/service/notifier"
)

// AgentManager Agent 管理器
type AgentManager struct {
        versionManager   *VersionManager
        upgradeEngine    *UpgradeEngine
        grayRelease      *GrayReleaseEngine
        heartbeatMonitor *HeartbeatMonitor
        notifier         *notifier.NotifierService
}

// NewAgentManager 创建 Agent 管理器
func NewAgentManager() *AgentManager {
        vm := NewVersionManager("")
        ue := NewUpgradeEngine()
        gr := NewGrayReleaseEngine()
        hm := NewHeartbeatMonitor()

        am := &AgentManager{
                versionManager:   vm,
                upgradeEngine:    ue,
                grayRelease:      gr,
                heartbeatMonitor: hm,
                notifier:         notifier.NewNotifierService(),
        }

        // 设置通知服务
        ue.SetNotifier(am.notifier)
        hm.SetNotifier(am.notifier)

        return am
}

// Start 启动所有服务
func (m *AgentManager) Start() {
        // 启动心跳监控
        m.heartbeatMonitor.Start()

        // TODO: 启动其他后台服务
}

// Stop 停止所有服务
func (m *AgentManager) Stop() {
        m.heartbeatMonitor.Stop()
}

// ==================== Agent 注册 ====================

// RegisterAgent 注册 Agent
func (m *AgentManager) RegisterAgent(req *RegisterRequest) (*agent.Agent, error) {
        // 检查是否已注册
        var existing agent.Agent
        err := global.DB.Where("agent_id = ?", req.AgentID).First(&existing).Error
        if err == nil {
                // 已注册，更新信息
                now := time.Now()
                existing.LastHeartbeat = &now
                existing.HeartbeatIP = req.IP
                existing.Platform = req.Platform
                existing.Arch = req.Arch
                if req.Version != "" {
                        existing.Version = req.Version
                        versionCode, _ := m.versionManager.ParseVersionCode(req.Version)
                        existing.VersionCode = versionCode
                }
                existing.Status = agent.AgentStatusOnline
                existing.StatusMessage = ""
                global.DB.Save(&existing)
                return &existing, nil
        }

        // 查找关联的服务器
        var srv server.Server
        err = global.DB.Where("host = ? OR agent_id = ?", req.IP, req.AgentID).First(&srv).Error

        agentSecret := m.generateSecret()

        ag := &agent.Agent{
                AgentID:        req.AgentID,
                AgentSecret:    agentSecret,
                Platform:       req.Platform,
                Arch:           req.Arch,
                Version:        req.Version,
                Status:         agent.AgentStatusOnline,
                AutoRecover:    true,
                AutoUpgrade:    true,
                UpgradeChannel: "stable",
                HeartbeatIP:    req.IP,
        }

        if err == nil {
                // 找到关联服务器
                ag.ServerID = srv.ID
                ag.ServerName = srv.Name

                // 更新服务器的 Agent 信息
                srv.AgentID = req.AgentID
                srv.AgentOnline = true
                now := time.Now()
                srv.LastHeartbeat = &now
                global.DB.Save(&srv)
        }

        // 解析版本号
        if req.Version != "" {
                versionCode, _ := m.versionManager.ParseVersionCode(req.Version)
                ag.VersionCode = versionCode
        }

        // 设置心跳时间
        now := time.Now()
        ag.LastHeartbeat = &now
        ag.LastOnlineAt = &now

        if err := global.DB.Create(ag).Error; err != nil {
                return nil, err
        }

        // 发送通知
        if m.notifier != nil {
                m.notifier.Broadcast(
                        fmt.Sprintf("🆕 Agent 新注册 - %s", ag.ServerName),
                        fmt.Sprintf("Agent ID: %s\n平台: %s/%s\n版本: %s",
                                ag.AgentID, ag.Platform, ag.Arch, ag.Version),
                )
        }

        return ag, nil
}

// RegisterRequest 注册请求
type RegisterRequest struct {
        AgentID  string `json:"agentId" binding:"required"`
        IP       string `json:"ip"`
        Platform string `json:"platform"`
        Arch     string `json:"arch"`
        Version  string `json:"version"`
        Hostname string `json:"hostname"`
}

// generateSecret 生成密钥
func (m *AgentManager) generateSecret() string {
        data := fmt.Sprintf("%d%s", time.Now().UnixNano(), "yunwei-agent")
        hash := md5.Sum([]byte(data))
        return hex.EncodeToString(hash[:])
}

// ==================== Agent 管理 ====================

// GetAgent 获取 Agent
func (m *AgentManager) GetAgent(id uint) (*agent.Agent, error) {
        var ag agent.Agent
        err := global.DB.First(&ag, id).Error
        return &ag, err
}

// GetAgentByUUID 根据 UUID 获取 Agent
func (m *AgentManager) GetAgentByUUID(uuid string) (*agent.Agent, error) {
        var ag agent.Agent
        err := global.DB.Where("agent_id = ?", uuid).First(&ag).Error
        return &ag, err
}

// ListAgents 列出 Agent
func (m *AgentManager) ListAgents(filter *AgentFilter) ([]agent.Agent, int64, error) {
        query := global.DB.Model(&agent.Agent{})

        if filter != nil {
                if filter.Status != "" {
                        query = query.Where("status = ?", filter.Status)
                }
                if filter.Platform != "" {
                        query = query.Where("platform = ?", filter.Platform)
                }
                if filter.Arch != "" {
                        query = query.Where("arch = ?", filter.Arch)
                }
                if filter.Version != "" {
                        query = query.Where("version = ?", filter.Version)
                }
                if filter.ServerID > 0 {
                        query = query.Where("server_id = ?", filter.ServerID)
                }
                if filter.Keyword != "" {
                        query = query.Where("server_name LIKE ? OR agent_id LIKE ?",
                                "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
                }
        }

        var total int64
        query.Count(&total)

        var agents []agent.Agent
        err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&agents).Error
        return agents, total, err
}

// AgentFilter Agent 过滤器
type AgentFilter struct {
        Status   string `json:"status"`
        Platform string `json:"platform"`
        Arch     string `json:"arch"`
        Version  string `json:"version"`
        ServerID uint   `json:"serverId"`
        Keyword  string `json:"keyword"`
        Limit    int    `json:"limit"`
        Offset   int    `json:"offset"`
}

// UpdateAgent 更新 Agent
func (m *AgentManager) UpdateAgent(id uint, updates map[string]interface{}) error {
        return global.DB.Model(&agent.Agent{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAgent 删除 Agent
func (m *AgentManager) DeleteAgent(id uint) error {
        // 检查状态
        var ag agent.Agent
        if err := global.DB.First(&ag, id).Error; err != nil {
                return err
        }

        if ag.Status == agent.AgentStatusOnline {
                return fmt.Errorf("Agent 在线中，无法删除")
        }

        // 删除相关记录
        global.DB.Where("agent_id = ?", id).Delete(&agent.AgentHeartbeatRecord{})
        global.DB.Where("agent_id = ?", id).Delete(&agent.AgentMetric{})
        global.DB.Where("agent_id = ?", id).Delete(&agent.AgentRecoverRecord{})

        return global.DB.Delete(&agent.Agent{}, id).Error
}

// DisableAgent 禁用 Agent
func (m *AgentManager) DisableAgent(id uint, reason string) error {
        updates := map[string]interface{}{
                "status":        agent.AgentStatusDisabled,
                "status_message": reason,
        }
        return m.UpdateAgent(id, updates)
}

// EnableAgent 启用 Agent
func (m *AgentManager) EnableAgent(id uint) error {
        updates := map[string]interface{}{
                "status":        agent.AgentStatusOffline,
                "status_message": "",
        }
        return m.UpdateAgent(id, updates)
}

// ==================== Agent 配置 ====================

// GetAgentConfig 获取 Agent 配置
func (m *AgentManager) GetAgentConfig(agentID uint) (*AgentConfigResponse, error) {
        var ag agent.Agent
        if err := global.DB.First(&ag, agentID).Error; err != nil {
                return nil, err
        }

        response := &AgentConfigResponse{
                AgentID:        ag.AgentID,
                Version:        ag.Version,
                AutoUpgrade:    ag.AutoUpgrade,
                UpgradeChannel: ag.UpgradeChannel,
                AutoRecover:    ag.AutoRecover,
                GrayGroup:      ag.GrayGroup,
                GrayWeight:     ag.GrayWeight,
        }

        // 获取配置模板
        var configs []agent.AgentConfig
        global.DB.Where("enabled = ?", true).Find(&configs)

        // 合并配置
        mergedConfig := make(map[string]interface{})
        for _, cfg := range configs {
                if cfg.Scope == "all" {
                        var cfgData map[string]interface{}
                        json.Unmarshal([]byte(cfg.ConfigJSON), &cfgData)
                        for k, v := range cfgData {
                                mergedConfig[k] = v
                        }
                }
        }
        response.Config = mergedConfig

        // 计算配置 Hash
        configJSON, _ := json.Marshal(response.Config)
        response.ConfigHash = m.hashConfig(string(configJSON))

        // 检查是否有升级任务
        upgradeInfo, err := m.versionManager.CheckUpgrade(&ag)
        if err == nil {
                response.UpgradeInfo = upgradeInfo
        }

        return response, nil
}

// AgentConfigResponse Agent 配置响应
type AgentConfigResponse struct {
        AgentID        string                 `json:"agentId"`
        Version        string                 `json:"version"`
        AutoUpgrade    bool                   `json:"autoUpgrade"`
        UpgradeChannel string                 `json:"upgradeChannel"`
        AutoRecover    bool                   `json:"autoRecover"`
        GrayGroup      string                 `json:"grayGroup"`
        GrayWeight     int                    `json:"grayWeight"`
        Config         map[string]interface{} `json:"config"`
        ConfigHash     string                 `json:"configHash"`
        UpgradeInfo    *UpgradeInfo           `json:"upgradeInfo"`
}

// hashConfig 计算配置 Hash
func (m *AgentManager) hashConfig(config string) string {
        hash := md5.Sum([]byte(config))
        return hex.EncodeToString(hash[:])
}

// ==================== Agent 统计 ====================

// GetAgentStats 获取 Agent 统计
func (m *AgentManager) GetAgentStats() (*AgentStats, error) {
        stats := &AgentStats{}

        // 总数
        global.DB.Model(&agent.Agent{}).Count(&stats.Total)

        // 各状态数量
        global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusOnline).Count(&stats.Online)
        global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusOffline).Count(&stats.Offline)
        global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusUpgrading).Count(&stats.Upgrading)
        global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusError).Count(&stats.Error)
        global.DB.Model(&agent.Agent{}).Where("status = ?", agent.AgentStatusDisabled).Count(&stats.Disabled)

        // 版本分布
        global.DB.Model(&agent.Agent{}).
                Select("version, count(*) as count").
                Where("version != ''").
                Group("version").
                Order("count DESC").
                Limit(10).
                Scan(&stats.VersionDistribution)

        // 平台分布
        global.DB.Model(&agent.Agent{}).
                Select("platform, count(*) as count").
                Where("platform != ''").
                Group("platform").
                Scan(&stats.PlatformDistribution)

        // 架构分布
        global.DB.Model(&agent.Agent{}).
                Select("arch, count(*) as count").
                Where("arch != ''").
                Group("arch").
                Scan(&stats.ArchDistribution)

        // 今日统计
        today := time.Now().Truncate(24 * time.Hour)
        global.DB.Model(&agent.AgentHeartbeatRecord{}).
                Where("created_at >= ?", today).
                Count(&stats.TodayHeartbeats)

        global.DB.Model(&agent.AgentRecoverRecord{}).
                Where("created_at >= ? AND success = ?", today, true).
                Count(&stats.TodayRecoveries)

        global.DB.Model(&agent.AgentUpgradeTask{}).
                Where("created_at >= ?", today).
                Count(&stats.TodayUpgrades)

        // 需要升级的数量
        var agents []agent.Agent
        global.DB.Where("status = ? AND auto_upgrade = ?", agent.AgentStatusOnline, true).Find(&agents)
        for _, ag := range agents {
                info, err := m.versionManager.CheckUpgrade(&ag)
                if err == nil && info.NeedUpgrade {
                        stats.NeedUpgrade++
                }
        }

        return stats, nil
}

// AgentStats Agent 统计
type AgentStats struct {
        Total              int64  `json:"total"`
        Online             int64  `json:"online"`
        Offline            int64  `json:"offline"`
        Upgrading          int64  `json:"upgrading"`
        Error              int64  `json:"error"`
        Disabled           int64  `json:"disabled"`
        NeedUpgrade        int64  `json:"needUpgrade"`
        TodayHeartbeats    int64  `json:"todayHeartbeats"`
        TodayRecoveries    int64  `json:"todayRecoveries"`
        TodayUpgrades      int64  `json:"todayUpgrades"`
        VersionDistribution []struct {
                Version string `json:"version"`
                Count   int    `json:"count"`
        } `json:"versionDistribution"`
        PlatformDistribution []struct {
                Platform string `json:"platform"`
                Count    int    `json:"count"`
        } `json:"platformDistribution"`
        ArchDistribution []struct {
                Arch  string `json:"arch"`
                Count int    `json:"count"`
        } `json:"archDistribution"`
}

// ==================== Agent 批量操作 ====================

// BatchOperation 批量操作
func (m *AgentManager) BatchOperation(req *BatchOperationRequest) (*BatchOperationResult, error) {
        result := &BatchOperationResult{
                SuccessAgents: make([]uint, 0),
                FailedAgents:  make([]FailedAgentOperation, 0),
        }

        var agents []agent.Agent
        if len(req.AgentIDs) > 0 {
                global.DB.Where("id IN ?", req.AgentIDs).Find(&agents)
        } else {
                // 根据 filter 查询
                agents, _, _ = m.ListAgents(&req.Filter)
        }

        for _, ag := range agents {
                err := m.executeOperation(&ag, req.Operation, req.Params)
                if err != nil {
                        result.FailedAgents = append(result.FailedAgents, FailedAgentOperation{
                                AgentID: ag.ID,
                                Name:    ag.ServerName,
                                Reason:  err.Error(),
                        })
                } else {
                        result.SuccessAgents = append(result.SuccessAgents, ag.ID)
                }
        }

        result.Total = int64(len(agents))
        result.SuccessCount = len(result.SuccessAgents)
        result.FailedCount = len(result.FailedAgents)

        return result, nil
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
        AgentIDs  []uint      `json:"agentIds"`
        Filter    AgentFilter `json:"filter"`
        Operation string      `json:"operation"` // enable/disable/upgrade/restart
        Params    map[string]interface{} `json:"params"`
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
        Total         int64                   `json:"total"`
        SuccessCount  int                     `json:"successCount"`
        FailedCount   int                     `json:"failedCount"`
        SuccessAgents []uint                  `json:"successAgents"`
        FailedAgents  []FailedAgentOperation  `json:"failedAgents"`
}

// FailedAgentOperation 失败的 Agent 操作
type FailedAgentOperation struct {
        AgentID uint   `json:"agentId"`
        Name    string `json:"name"`
        Reason  string `json:"reason"`
}

// executeOperation 执行操作
func (m *AgentManager) executeOperation(ag *agent.Agent, operation string, params map[string]interface{}) error {
        switch strings.ToLower(operation) {
        case "enable":
                return m.EnableAgent(ag.ID)
        case "disable":
                reason, _ := params["reason"].(string)
                return m.DisableAgent(ag.ID, reason)
        case "upgrade":
                targetVersion, _ := params["version"].(string)
                req := &CreateUpgradeRequest{
                        AgentID:       ag.ID,
                        TargetVersion: targetVersion,
                        TaskType:      "manual",
                        Priority:      5,
                }
                _, err := m.upgradeEngine.CreateUpgradeTask(req)
                return err
        case "restart":
                // TODO: 发送重启命令给 Agent
                return fmt.Errorf("not implemented")
        default:
                return fmt.Errorf("unknown operation: %s", operation)
        }
}

// ==================== 获取子服务 ====================

// GetVersionManager 获取版本管理器
func (m *AgentManager) GetVersionManager() *VersionManager {
        return m.versionManager
}

// GetUpgradeEngine 获取升级引擎
func (m *AgentManager) GetUpgradeEngine() *UpgradeEngine {
        return m.upgradeEngine
}

// GetGrayReleaseEngine 获取灰度发布引擎
func (m *AgentManager) GetGrayReleaseEngine() *GrayReleaseEngine {
        return m.grayRelease
}

// GetHeartbeatMonitor 获取心跳监控器
func (m *AgentManager) GetHeartbeatMonitor() *HeartbeatMonitor {
        return m.heartbeatMonitor
}
