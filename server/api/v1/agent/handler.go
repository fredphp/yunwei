package agent

import (
        "strconv"
        "time"

        "yunwei/global"
        "yunwei/model/agent"
        "yunwei/model/common/response"
        agentService "yunwei/service/agent"

        "github.com/gin-gonic/gin"
)

var agentManager = agentService.NewAgentManager()

// ==================== Agent 管理 ====================

// GetAgentList 获取 Agent 列表
func GetAgentList(c *gin.Context) {
        filter := &agentService.AgentFilter{
                Status:   c.Query("status"),
                Platform: c.Query("platform"),
                Arch:     c.Query("arch"),
                Version:  c.Query("version"),
                Keyword:  c.Query("keyword"),
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        if serverId := c.Query("serverId"); serverId != "" {
                id, _ := strconv.ParseUint(serverId, 10, 32)
                filter.ServerID = uint(id)
        }

        agents, total, err := agentManager.ListAgents(filter)
        if err != nil {
                response.FailWithMessage("获取列表失败: "+err.Error(), c)
                return
        }

        response.OkWithData(gin.H{
                "list":     agents,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// GetAgent 获取 Agent 详情
func GetAgent(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        ag, err := agentManager.GetAgent(uint(id))
        if err != nil {
                response.FailWithMessage("Agent不存在", c)
                return
        }

        response.OkWithData(ag, c)
}

// UpdateAgent 更新 Agent
func UpdateAgent(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var req map[string]interface{}
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        // 过滤可更新字段
        allowedFields := map[string]bool{
                "auto_upgrade":    true,
                "upgrade_channel": true,
                "auto_recover":    true,
                "gray_group":      true,
                "gray_weight":     true,
        }

        updates := make(map[string]interface{})
        for k, v := range req {
                if allowedFields[k] {
                        updates[k] = v
                }
        }

        if err := agentManager.UpdateAgent(uint(id), updates); err != nil {
                response.FailWithMessage("更新失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// DeleteAgent 删除 Agent
func DeleteAgent(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        if err := agentManager.DeleteAgent(uint(id)); err != nil {
                response.FailWithMessage("删除失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// DisableAgent 禁用 Agent
func DisableAgent(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var req struct {
                Reason string `json:"reason"`
        }
        c.ShouldBindJSON(&req)

        if err := agentManager.DisableAgent(uint(id), req.Reason); err != nil {
                response.FailWithMessage("禁用失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// EnableAgent 启用 Agent
func EnableAgent(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        if err := agentManager.EnableAgent(uint(id)); err != nil {
                response.FailWithMessage("启用失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// GetAgentStats 获取 Agent 统计
func GetAgentStats(c *gin.Context) {
        stats, err := agentManager.GetAgentStats()
        if err != nil {
                response.FailWithMessage("获取统计失败", c)
                return
        }

        response.OkWithData(stats, c)
}

// GetAgentConfig 获取 Agent 配置
func GetAgentConfig(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        config, err := agentManager.GetAgentConfig(uint(id))
        if err != nil {
                response.FailWithMessage("获取配置失败: "+err.Error(), c)
                return
        }

        response.OkWithData(config, c)
}

// BatchOperation 批量操作
func BatchOperation(c *gin.Context) {
        var req agentService.BatchOperationRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        result, err := agentManager.BatchOperation(&req)
        if err != nil {
                response.FailWithMessage("操作失败: "+err.Error(), c)
                return
        }

        response.OkWithData(result, c)
}

// ==================== 版本管理 ====================

// GetVersionList 获取版本列表
func GetVersionList(c *gin.Context) {
        filter := &agentService.VersionFilter{
                Platform:    c.Query("platform"),
                Arch:        c.Query("arch"),
                ReleaseType: c.Query("releaseType"),
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        if enabled := c.Query("enabled"); enabled != "" {
                val := enabled == "true"
                filter.Enabled = &val
        }

        vm := agentManager.GetVersionManager()
        versions, total, err := vm.ListVersions(filter)
        if err != nil {
                response.FailWithMessage("获取列表失败", c)
                return
        }

        response.OkWithData(gin.H{
                "list":     versions,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// GetVersion 获取版本详情
func GetVersion(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        vm := agentManager.GetVersionManager()
        version, err := vm.GetVersion(uint(id))
        if err != nil {
                response.FailWithMessage("版本不存在", c)
                return
        }

        response.OkWithData(version, c)
}

// CreateVersion 创建版本
func CreateVersion(c *gin.Context) {
        var req struct {
                Version     string `json:"version" binding:"required"`
                Platform    string `json:"platform" binding:"required"`
                Arch        string `json:"arch" binding:"required"`
                FileURL     string `json:"fileUrl"`
                FileMD5     string `json:"fileMd5"`
                FileSize    int64  `json:"fileSize"`
                MinVersion  string `json:"minVersion"`
                Changelog   string `json:"changelog"`
                ReleaseType string `json:"releaseType"`
                ForceUpdate bool   `json:"forceUpdate"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误: "+err.Error(), c)
                return
        }

        version := &agent.AgentVersion{
                Version:     req.Version,
                Platform:    req.Platform,
                Arch:        req.Arch,
                FileURL:     req.FileURL,
                FileMD5:     req.FileMD5,
                FileSize:    req.FileSize,
                MinVersion:  req.MinVersion,
                Changelog:   req.Changelog,
                ReleaseType: req.ReleaseType,
                ForceUpdate: req.ForceUpdate,
                Enabled:     true,
                BuildTime:   time.Now(),
        }

        if version.ReleaseType == "" {
                version.ReleaseType = "stable"
        }

        vm := agentManager.GetVersionManager()
        if err := vm.CreateVersion(version); err != nil {
                response.FailWithMessage("创建失败: "+err.Error(), c)
                return
        }

        response.OkWithData(version, c)
}

// UpdateVersion 更新版本
func UpdateVersion(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var req map[string]interface{}
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        vm := agentManager.GetVersionManager()
        version, err := vm.GetVersion(uint(id))
        if err != nil {
                response.FailWithMessage("版本不存在", c)
                return
        }

        // 更新字段
        if v, ok := req["fileUrl"].(string); ok {
                version.FileURL = v
        }
        if v, ok := req["fileMd5"].(string); ok {
                version.FileMD5 = v
        }
        if v, ok := req["fileSize"].(float64); ok {
                version.FileSize = int64(v)
        }
        if v, ok := req["minVersion"].(string); ok {
                version.MinVersion = v
        }
        if v, ok := req["changelog"].(string); ok {
                version.Changelog = v
        }
        if v, ok := req["releaseType"].(string); ok {
                version.ReleaseType = v
        }
        if v, ok := req["enabled"].(bool); ok {
                version.Enabled = v
        }
        if v, ok := req["forceUpdate"].(bool); ok {
                version.ForceUpdate = v
        }

        if err := vm.UpdateVersion(version); err != nil {
                response.FailWithMessage("更新失败: "+err.Error(), c)
                return
        }

        response.OkWithData(version, c)
}

// DeleteVersion 删除版本
func DeleteVersion(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        vm := agentManager.GetVersionManager()
        if err := vm.DeleteVersion(uint(id)); err != nil {
                response.FailWithMessage("删除失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// GetVersionStats 获取版本统计
func GetVersionStats(c *gin.Context) {
        vm := agentManager.GetVersionManager()
        stats, err := vm.GetVersionStats()
        if err != nil {
                response.FailWithMessage("获取统计失败", c)
                return
        }

        response.OkWithData(stats, c)
}

// CheckUpgrade 检查升级
func CheckUpgrade(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        ag, err := agentManager.GetAgent(uint(id))
        if err != nil {
                response.FailWithMessage("Agent不存在", c)
                return
        }

        vm := agentManager.GetVersionManager()
        info, err := vm.CheckUpgrade(ag)
        if err != nil {
                response.FailWithMessage("检查失败: "+err.Error(), c)
                return
        }

        response.OkWithData(info, c)
}

// ==================== 升级任务 ====================

// CreateUpgradeTask 创建升级任务
func CreateUpgradeTask(c *gin.Context) {
        var req agentService.CreateUpgradeRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误: "+err.Error(), c)
                return
        }

        ue := agentManager.GetUpgradeEngine()
        task, err := ue.CreateUpgradeTask(&req)
        if err != nil {
                response.FailWithMessage("创建失败: "+err.Error(), c)
                return
        }

        response.OkWithData(task, c)
}

// CreateBatchUpgrade 创建批量升级
func CreateBatchUpgrade(c *gin.Context) {
        var req agentService.BatchUpgradeRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误: "+err.Error(), c)
                return
        }

        ue := agentManager.GetUpgradeEngine()
        result, err := ue.CreateBatchUpgradeTask(&req)
        if err != nil {
                response.FailWithMessage("创建失败: "+err.Error(), c)
                return
        }

        response.OkWithData(result, c)
}

// GetUpgradeTask 获取升级任务
func GetUpgradeTask(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        ue := agentManager.GetUpgradeEngine()
        task, err := ue.GetUpgradeTask(uint(id))
        if err != nil {
                response.FailWithMessage("任务不存在", c)
                return
        }

        response.OkWithData(task, c)
}

// GetUpgradeTaskList 获取升级任务列表
func GetUpgradeTaskList(c *gin.Context) {
        filter := &agentService.UpgradeTaskFilter{
                Status:   c.Query("status"),
                TaskType: c.Query("taskType"),
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        if agentId := c.Query("agentId"); agentId != "" {
                id, _ := strconv.ParseUint(agentId, 10, 32)
                filter.AgentID = uint(id)
        }

        ue := agentManager.GetUpgradeEngine()
        tasks, total, err := ue.ListUpgradeTasks(filter)
        if err != nil {
                response.FailWithMessage("获取列表失败", c)
                return
        }

        response.OkWithData(gin.H{
                "list":     tasks,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// ExecuteUpgrade 执行升级
func ExecuteUpgrade(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        ue := agentManager.GetUpgradeEngine()
        if err := ue.ExecuteUpgrade(uint(id)); err != nil {
                response.FailWithMessage("执行失败: "+err.Error(), c)
                return
        }

        response.OkWithMessage("升级任务已开始执行", c)
}

// CancelUpgrade 取消升级
func CancelUpgrade(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        ue := agentManager.GetUpgradeEngine()
        if err := ue.CancelUpgradeTask(uint(id)); err != nil {
                response.FailWithMessage("取消失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// RollbackUpgrade 回滚升级
func RollbackUpgrade(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        ue := agentManager.GetUpgradeEngine()
        if err := ue.RollbackUpgrade(uint(id)); err != nil {
                response.FailWithMessage("回滚失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// GetUpgradeStats 获取升级统计
func GetUpgradeStats(c *gin.Context) {
        ue := agentManager.GetUpgradeEngine()
        stats, err := ue.GetUpgradeStats()
        if err != nil {
                response.FailWithMessage("获取统计失败", c)
                return
        }

        response.OkWithData(stats, c)
}

// ==================== 灰度发布 ====================

// GetGrayStrategyList 获取灰度策略列表
func GetGrayStrategyList(c *gin.Context) {
        gr := agentManager.GetGrayReleaseEngine()
        strategies, err := gr.ListStrategies(c.Query("status"))
        if err != nil {
                response.FailWithMessage("获取列表失败", c)
                return
        }

        response.OkWithData(strategies, c)
}

// GetGrayStrategy 获取灰度策略
func GetGrayStrategy(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        strategy, err := gr.GetStrategy(uint(id))
        if err != nil {
                response.FailWithMessage("策略不存在", c)
                return
        }

        response.OkWithData(strategy, c)
}

// CreateGrayStrategy 创建灰度策略
func CreateGrayStrategy(c *gin.Context) {
        var req agentService.CreateStrategyRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误: "+err.Error(), c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        strategy, err := gr.CreateStrategy(&req)
        if err != nil {
                response.FailWithMessage("创建失败: "+err.Error(), c)
                return
        }

        response.OkWithData(strategy, c)
}

// StartGrayStrategy 启动灰度策略
func StartGrayStrategy(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        if err := gr.StartStrategy(uint(id)); err != nil {
                response.FailWithMessage("启动失败: "+err.Error(), c)
                return
        }

        response.OkWithMessage("灰度发布已启动", c)
}

// PauseGrayStrategy 暂停灰度策略
func PauseGrayStrategy(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        if err := gr.PauseStrategy(uint(id)); err != nil {
                response.FailWithMessage("暂停失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// ResumeGrayStrategy 恢复灰度策略
func ResumeGrayStrategy(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        if err := gr.ResumeStrategy(uint(id)); err != nil {
                response.FailWithMessage("恢复失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// CancelGrayStrategy 取消灰度策略
func CancelGrayStrategy(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        if err := gr.CancelStrategy(uint(id)); err != nil {
                response.FailWithMessage("取消失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// GetGrayStrategyProgress 获取灰度进度
func GetGrayStrategyProgress(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        gr := agentManager.GetGrayReleaseEngine()
        progress, err := gr.GetStrategyProgress(uint(id))
        if err != nil {
                response.FailWithMessage("获取进度失败: "+err.Error(), c)
                return
        }

        response.OkWithData(progress, c)
}

// ==================== 心跳监控 ====================

// GetMonitorStats 获取监控统计
func GetMonitorStats(c *gin.Context) {
        hm := agentManager.GetHeartbeatMonitor()
        stats := hm.GetMonitorStats()
        response.OkWithData(stats, c)
}

// GetOfflineAgents 获取离线 Agent
func GetOfflineAgents(c *gin.Context) {
        hm := agentManager.GetHeartbeatMonitor()
        agents := hm.GetOfflineAgents()
        response.OkWithData(agents, c)
}

// ==================== 心跳记录 ====================

// GetHeartbeatRecords 获取心跳记录
func GetHeartbeatRecords(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

        var records []agent.AgentHeartbeatRecord
        query := global.DB.Model(&agent.AgentHeartbeatRecord{}).Where("agent_id = ?", id)

        var total int64
        query.Count(&total)

        query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records)

        response.OkWithData(gin.H{
                "list":     records,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// GetRecoverRecords 获取恢复记录
func GetRecoverRecords(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

        var records []agent.AgentRecoverRecord
        query := global.DB.Model(&agent.AgentRecoverRecord{}).Where("agent_id = ?", id)

        var total int64
        query.Count(&total)

        query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records)

        response.OkWithData(gin.H{
                "list":     records,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}
