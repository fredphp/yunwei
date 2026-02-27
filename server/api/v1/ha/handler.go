package ha

import (
        "strconv"
        "time"

        "yunwei/global"
        "yunwei/model/ha"
        "yunwei/model/common/response"
        haService "yunwei/service/ha"

        "github.com/gin-gonic/gin"
)

var haManager = haService.NewHAManager()

// ==================== 集群状态 ====================

// GetClusterStats 获取集群统计
func GetClusterStats(c *gin.Context) {
        stats := haManager.GetHAStats()
        response.OkWithData(stats, c)
}

// GetClusterNodes 获取集群节点列表
func GetClusterNodes(c *gin.Context) {
        filter := &haService.NodeFilter{
                Status:     c.Query("status"),
                Role:       c.Query("role"),
                DataCenter: c.Query("dataCenter"),
                Zone:       c.Query("zone"),
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        if enabled := c.Query("enabled"); enabled != "" {
                val := enabled == "true"
                filter.Enabled = &val
        }

        nodes, total, err := haManager.GetClusterManager().ListNodes(filter)
        if err != nil {
                response.FailWithMessage("获取节点列表失败: "+err.Error(), c)
                return
        }

        response.OkWithData(gin.H{
                "list":     nodes,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// GetClusterNode 获取节点详情
func GetClusterNode(c *gin.Context) {
        nodeID := c.Param("id")

        node, err := haManager.GetClusterManager().GetNode(nodeID)
        if err != nil {
                response.FailWithMessage("节点不存在", c)
                return
        }

        response.OkWithData(node, c)
}

// EnableNode 启用节点
func EnableNode(c *gin.Context) {
        nodeID := c.Param("id")

        if err := haManager.GetClusterManager().EnableNode(nodeID); err != nil {
                response.FailWithMessage("启用失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// DisableNode 禁用节点
func DisableNode(c *gin.Context) {
        nodeID := c.Param("id")

        if err := haManager.GetClusterManager().DisableNode(nodeID); err != nil {
                response.FailWithMessage("禁用失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// GetNodeMetrics 获取节点指标
func GetNodeMetrics(c *gin.Context) {
        nodeID := c.Param("id")
        hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

        metrics, err := haManager.GetClusterManager().GetNodeMetrics(nodeID, hours)
        if err != nil {
                response.FailWithMessage("获取指标失败: "+err.Error(), c)
                return
        }

        response.OkWithData(metrics, c)
}

// ==================== Leader 选举 ====================

// GetLeaderStatus 获取 Leader 状态
func GetLeaderStatus(c *gin.Context) {
        status := gin.H{
                "nodeId":    haManager.GetNodeID(),
                "isLeader":  haManager.IsLeader(),
                "leader":    haManager.GetLeader(),
                "term":      haManager.GetLeaderService().GetTerm(),
                "timestamp": time.Now(),
        }

        response.OkWithData(status, c)
}

// ResignLeader 辞职 Leader
func ResignLeader(c *gin.Context) {
        if !haManager.IsLeader() {
                response.FailWithMessage("当前不是 Leader", c)
                return
        }

        if err := haManager.GetLeaderService().Resign(); err != nil {
                response.FailWithMessage("辞职失败: "+err.Error(), c)
                return
        }

        response.OkWithMessage("已辞去 Leader", c)
}

// ForceLeader 强制指定 Leader
func ForceLeader(c *gin.Context) {
        var req struct {
                NodeID string `json:"nodeId" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        if err := haManager.GetLeaderService().ForceLeader(req.NodeID); err != nil {
                response.FailWithMessage("设置失败: "+err.Error(), c)
                return
        }

        response.OkWithMessage("已强制指定 Leader", c)
}

// ==================== 分布式锁 ====================

// GetLocks 获取锁列表
func GetLocks(c *gin.Context) {
        filter := &haService.LockFilter{
                HolderNode: c.Query("holderNode"),
        }
        filter.OnlyActive = c.Query("active") == "true"

        locks, total, err := haManager.GetLockService().ListLocks(filter)
        if err != nil {
                response.FailWithMessage("获取锁列表失败", c)
                return
        }

        response.OkWithData(gin.H{
                "list":  locks,
                "total": total,
        }, c)
}

// GetLock 获取锁详情
func GetLock(c *gin.Context) {
        key := c.Param("key")

        info, err := haManager.GetLockService().GetLockInfo(key)
        if err != nil {
                response.FailWithMessage("锁不存在", c)
                return
        }

        response.OkWithData(info, c)
}

// ForceReleaseLock 强制释放锁
func ForceReleaseLock(c *gin.Context) {
        key := c.Param("key")

        var req struct {
                Reason string `json:"reason"`
        }
        c.ShouldBindJSON(&req)

        if err := haManager.GetLockService().ForceRelease(nil, key, req.Reason); err != nil {
                response.FailWithMessage("释放失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// ==================== 会话管理 ====================

// GetSessions 获取会话列表
func GetSessions(c *gin.Context) {
        filter := &haService.SessionFilter{
                Username: c.Query("username"),
        }

        if userId := c.Query("userId"); userId != "" {
                id, _ := strconv.ParseUint(userId, 10, 32)
                filter.UserID = uint(id)
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        sessions, total, err := haManager.GetSessionManager().ListSessions(filter)
        if err != nil {
                response.FailWithMessage("获取会话列表失败", c)
                return
        }

        response.OkWithData(gin.H{
                "list":     sessions,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// DeleteSession 删除会话
func DeleteSession(c *gin.Context) {
        sessionID := c.Param("id")

        if err := haManager.GetSessionManager().DeleteSession(nil, sessionID); err != nil {
                response.FailWithMessage("删除失败: "+err.Error(), c)
                return
        }

        response.Ok(nil, c)
}

// GetSessionStats 获取会话统计
func GetSessionStats(c *gin.Context) {
        stats := haManager.GetSessionManager().GetSessionStats()
        response.OkWithData(stats, c)
}

// ==================== 配置管理 ====================

// GetHAConfig 获取 HA 配置
func GetHAConfig(c *gin.Context) {
        config := haManager.GetConfig()
        response.OkWithData(config, c)
}

// UpdateHAConfig 更新 HA 配置
func UpdateHAConfig(c *gin.Context) {
        var config ha.HAClusterConfig
        if err := c.ShouldBindJSON(&config); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        if err := haManager.UpdateConfig(&config); err != nil {
                response.FailWithMessage("更新失败: "+err.Error(), c)
                return
        }

        response.OkWithData(config, c)
}

// ListHAConfigs 获取配置列表
func ListHAConfigs(c *gin.Context) {
        var configs []ha.HAClusterConfig
        global.DB.Find(&configs)
        response.OkWithData(configs, c)
}

// CreateHAConfig 创建配置
func CreateHAConfig(c *gin.Context) {
        var config ha.HAClusterConfig
        if err := c.ShouldBindJSON(&config); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        if err := global.DB.Create(&config).Error; err != nil {
                response.FailWithMessage("创建失败: "+err.Error(), c)
                return
        }

        response.OkWithData(config, c)
}

// DeleteHAConfig 删除配置
func DeleteHAConfig(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

        if err := global.DB.Delete(&ha.HAClusterConfig{}, id).Error; err != nil {
                response.FailWithMessage("删除失败", c)
                return
        }

        response.Ok(nil, c)
}

// ==================== 故障转移 ====================

// GetFailoverRecords 获取故障转移记录
func GetFailoverRecords(c *gin.Context) {
        filter := &haService.FailoverFilter{
                Status:       c.Query("status"),
                FailoverType: c.Query("type"),
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        records, total, err := haManager.GetFailoverRecords(filter)
        if err != nil {
                response.FailWithMessage("获取记录失败", c)
                return
        }

        response.OkWithData(gin.H{
                "list":     records,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// TriggerFailover 触发故障转移
func TriggerFailover(c *gin.Context) {
        var req struct {
                NodeID string `json:"nodeId" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        if err := haManager.TriggerFailover(req.NodeID); err != nil {
                response.FailWithMessage("触发失败: "+err.Error(), c)
                return
        }

        response.OkWithMessage("故障转移已触发", c)
}

// ==================== 事件 ====================

// GetClusterEvents 获取集群事件
func GetClusterEvents(c *gin.Context) {
        filter := &haService.EventFilter{
                EventType: c.Query("type"),
                NodeID:    c.Query("nodeId"),
                Level:     c.Query("level"),
        }

        if startTime := c.Query("startTime"); startTime != "" {
                t, _ := time.Parse(time.RFC3339, startTime)
                filter.StartTime = t
        }
        if endTime := c.Query("endTime"); endTime != "" {
                t, _ := time.Parse(time.RFC3339, endTime)
                filter.EndTime = t
        }

        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
        filter.Limit = pageSize
        filter.Offset = (page - 1) * pageSize

        events, total, err := haManager.GetEvents(filter)
        if err != nil {
                response.FailWithMessage("获取事件失败", c)
                return
        }

        response.OkWithData(gin.H{
                "list":     events,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// ==================== 任务 HA ====================

// GetRunningTasks 获取运行中任务
func GetRunningTasks(c *gin.Context) {
        tasks := haManager.GetTaskHAManager().GetRunningTasks()
        response.OkWithData(tasks, c)
}

// GetLockDBRecords 从数据库获取锁记录
func GetLockDBRecords(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

        var locks []ha.DistributedLock
        var total int64

        global.DB.Model(&ha.DistributedLock{}).Count(&total)
        global.DB.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&locks)

        response.OkWithData(gin.H{
                "list":     locks,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// GetLeaderElectionRecords 获取选举记录
func GetLeaderElectionRecords(c *gin.Context) {
        var records []ha.LeaderElection
        global.DB.Order("created_at DESC").Limit(10).Find(&records)
        response.OkWithData(records, c)
}
