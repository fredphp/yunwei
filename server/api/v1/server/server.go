package server

import (
        "strconv"
        "time"

        "yunwei/config"
        "yunwei/global"
        "yunwei/model/common/response"
        "yunwei/model/server"
        "yunwei/service/ai/decision"
        "yunwei/service/ai/llm"
        "yunwei/service/detector"
        "yunwei/service/optimizer"

        "github.com/gin-gonic/gin"
)

// GetServerList 获取服务器列表
func GetServerList(c *gin.Context) {
        var servers []server.Server
        
        query := global.DB.Model(&server.Server{})
        
        // 搜索条件
        if name := c.Query("name"); name != "" {
                query = query.Where("name LIKE ?", "%"+name+"%")
        }
        if status := c.Query("status"); status != "" {
                query = query.Where("status = ?", status)
        }
        if groupId := c.Query("groupId"); groupId != "" {
                query = query.Where("group_id = ?", groupId)
        }

        // 分页
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
        var total int64

        query.Count(&total)
        query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&servers)

        response.OkWithData(gin.H{
                "list":     servers,
                "total":    total,
                "page":     page,
                "pageSize": pageSize,
        }, c)
}

// GetServer 获取服务器详情
func GetServer(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var srv server.Server
        if err := global.DB.First(&srv, id).Error; err != nil {
                response.FailWithMessage("服务器不存在", c)
                return
        }

        response.OkWithData(srv, c)
}

// AddServer 添加服务器
func AddServer(c *gin.Context) {
        var req struct {
                Name        string `json:"name" binding:"required"`
                Hostname    string `json:"hostname"`
                Host        string `json:"host" binding:"required"`
                Port        int    `json:"port"`
                User        string `json:"user"`
                Password    string `json:"password"`
                PrivateKey  string `json:"privateKey"`
                GroupID     uint   `json:"groupId"`
                Description string `json:"description"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误: "+err.Error(), c)
                return
        }

        if req.Port == 0 {
                req.Port = 22
        }

        srv := server.Server{
                Name:        req.Name,
                Hostname:    req.Hostname,
                Host:        req.Host,
                Port:        req.Port,
                User:        req.User,
                Password:    req.Password,
                PrivateKey:  req.PrivateKey,
                GroupID:     req.GroupID,
                Description: req.Description,
                Status:      "pending",
        }

        if err := global.DB.Create(&srv).Error; err != nil {
                response.FailWithMessage("创建失败: "+err.Error(), c)
                return
        }

        response.OkWithData(srv, c)
}

// UpdateServer 更新服务器
func UpdateServer(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var srv server.Server
        if err := global.DB.First(&srv, id).Error; err != nil {
                response.FailWithMessage("服务器不存在", c)
                return
        }

        var req map[string]interface{}
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        if err := global.DB.Model(&srv).Updates(req).Error; err != nil {
                response.FailWithMessage("更新失败", c)
                return
        }

        response.OkWithData(srv, c)
}

// DeleteServer 删除服务器
func DeleteServer(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        if err := global.DB.Delete(&server.Server{}, id).Error; err != nil {
                response.FailWithMessage("删除失败", c)
                return
        }

        response.Ok(nil, c)
}

// GetServerMetrics 获取服务器指标
func GetServerMetrics(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var metrics []server.ServerMetric
        global.DB.Where("server_id = ?", id).Order("created_at DESC").Limit(100).Find(&metrics)

        response.OkWithData(metrics, c)
}

// GetServerLogs 获取服务器日志
func GetServerLogs(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var logs []server.ServerLog
        global.DB.Where("server_id = ?", id).Order("created_at DESC").Limit(100).Find(&logs)

        response.OkWithData(logs, c)
}

// ExecuteCommand 执行命令
func ExecuteCommand(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var req struct {
                Command string `json:"command" binding:"required"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        var srv server.Server
        if err := global.DB.First(&srv, id).Error; err != nil {
                response.FailWithMessage("服务器不存在", c)
                return
        }

        // 记录日志
        log := server.ServerLog{
                ServerID: uint(id),
                Type:     "command",
                Content:  req.Command,
        }
        startTime := time.Now()

        // TODO: 实际执行命令（通过SSH或Agent）
        // output, err := ssh.Execute(srv, req.Command)

        log.Duration = time.Since(startTime).Milliseconds()
        // log.Output = output
        global.DB.Create(&log)

        response.OkWithData(gin.H{
                "command":  req.Command,
                "duration": log.Duration,
                "message":  "命令已发送",
        }, c)
}

// RefreshStatus 刷新服务器状态
func RefreshStatus(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var srv server.Server
        if err := global.DB.First(&srv, id).Error; err != nil {
                response.FailWithMessage("服务器不存在", c)
                return
        }

        // TODO: 通过Agent获取最新状态
        // 更新心跳时间
        now := time.Now()
        srv.LastCheck = &now

        response.OkWithData(srv, c)
}

// TestSSH 测试SSH连接
func TestSSH(c *gin.Context) {
        var req struct {
                Host       string `json:"host" binding:"required"`
                Port       int    `json:"port"`
                User       string `json:"user"`
                Password   string `json:"password"`
                PrivateKey string `json:"privateKey"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        if req.Port == 0 {
                req.Port = 22
        }

        // TODO: 实际测试SSH连接
        // success, err := ssh.Test(req)

        response.OkWithData(gin.H{
                "success": true,
                "message": "连接成功",
        }, c)
}

// GetGroups 获取服务器分组
func GetGroups(c *gin.Context) {
        var groups []server.Group
        global.DB.Find(&groups)

        response.OkWithData(groups, c)
}

// CreateGroup 创建分组
func CreateGroup(c *gin.Context) {
        var req struct {
                Name        string `json:"name" binding:"required"`
                ParentID    uint   `json:"parentId"`
                Description string `json:"description"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        group := server.Group{
                Name:        req.Name,
                ParentID:    req.ParentID,
                Description: req.Description,
        }

        if err := global.DB.Create(&group).Error; err != nil {
                response.FailWithMessage("创建失败", c)
                return
        }

        response.OkWithData(group, c)
}

// DeleteGroup 删除分组
func DeleteGroup(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        // 检查是否有服务器使用此分组
        var count int64
        global.DB.Model(&server.Server{}).Where("group_id = ?", id).Count(&count)
        if count > 0 {
                response.FailWithMessage("该分组下存在服务器，无法删除", c)
                return
        }

        if err := global.DB.Delete(&server.Group{}, id).Error; err != nil {
                response.FailWithMessage("删除失败", c)
                return
        }

        response.Ok(nil, c)
}

// GetDockerContainers 获取Docker容器
func GetDockerContainers(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var containers []server.DockerContainer
        global.DB.Where("server_id = ?", id).Find(&containers)

        response.OkWithData(containers, c)
}

// GetPortInfos 获取端口信息
func GetPortInfos(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var ports []server.PortInfo
        global.DB.Where("server_id = ?", id).Find(&ports)

        response.OkWithData(ports, c)
}

// ==================== AI 相关接口 ====================

// AIAnalyze AI分析服务器
func AIAnalyze(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var srv server.Server
        if err := global.DB.First(&srv, id).Error; err != nil {
                response.FailWithMessage("服务器不存在", c)
                return
        }

        // 获取最新指标
        var metric server.ServerMetric
        global.DB.Where("server_id = ?", id).Order("created_at DESC").First(&metric)

        // 创建决策引擎
        llmClient := llm.NewGLM5Client(llm.GLM5Config{
                APIKey: config.CONFIG.AI.APIKey,
                BaseURL: config.CONFIG.AI.BaseURL,
                Model:   config.CONFIG.AI.Model,
        })
        engine := decision.NewEngine(llmClient)

        // 执行分析
        decision, err := engine.QuickAnalyze(&srv, &metric)
        if err != nil {
                response.FailWithMessage("AI分析失败: "+err.Error(), c)
                return
        }

        response.OkWithData(decision, c)
}

// GetAlerts 获取告警列表
func GetAlerts(c *gin.Context) {
        var alerts []detector.Alert

        query := global.DB.Model(&detector.Alert{})

        if status := c.Query("status"); status != "" {
                query = query.Where("status = ?", status)
        }
        if level := c.Query("level"); level != "" {
                query = query.Where("level = ?", level)
        }
        if serverId := c.Query("serverId"); serverId != "" {
                query = query.Where("server_id = ?", serverId)
        }

        query.Order("created_at DESC").Limit(100).Find(&alerts)

        response.OkWithData(alerts, c)
}

// AcknowledgeAlert 确认告警
func AcknowledgeAlert(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var alert detector.Alert
        if err := global.DB.First(&alert, id).Error; err != nil {
                response.FailWithMessage("告警不存在", c)
                return
        }

        alert.Status = "acknowledged"
        // alert.AcknowledgedBy = userID // 从JWT获取

        global.DB.Save(&alert)

        response.OkWithData(alert, c)
}

// GetRules 获取检测规则
func GetRules(c *gin.Context) {
        var rules []detector.DetectRule
        global.DB.Find(&rules)

        response.OkWithData(rules, c)
}

// UpdateRule 更新检测规则
func UpdateRule(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var rule detector.DetectRule
        if err := global.DB.First(&rule, id).Error; err != nil {
                response.FailWithMessage("规则不存在", c)
                return
        }

        var req map[string]interface{}
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage("参数错误", c)
                return
        }

        global.DB.Model(&rule).Updates(req)

        response.OkWithData(rule, c)
}

// GetAutoActions 获取自动操作记录
func GetAutoActions(c *gin.Context) {
        var actions []optimizer.AutoAction

        query := global.DB.Model(&optimizer.AutoAction{})

        if serverId := c.Query("serverId"); serverId != "" {
                query = query.Where("server_id = ?", serverId)
        }

        query.Order("created_at DESC").Limit(100).Find(&actions)

        response.OkWithData(actions, c)
}

// ExecuteAutoAction 执行自动操作
func ExecuteAutoAction(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var action optimizer.AutoAction
        if err := global.DB.First(&action, id).Error; err != nil {
                response.FailWithMessage("操作不存在", c)
                return
        }

        // TODO: 实际执行命令

        response.OkWithData(gin.H{
                "message": "操作已执行",
                "action":  action,
        }, c)
}

// GetDecisions 获取AI决策记录
func GetDecisions(c *gin.Context) {
        var decisions []decision.AIDecision

        query := global.DB.Model(&decision.AIDecision{})

        if serverId := c.Query("serverId"); serverId != "" {
                query = query.Where("server_id = ?", serverId)
        }
        if status := c.Query("status"); status != "" {
                query = query.Where("status = ?", status)
        }

        query.Order("created_at DESC").Limit(100).Find(&decisions)

        response.OkWithData(decisions, c)
}

// ApproveDecision 批准决策
func ApproveDecision(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var dec decision.AIDecision
        if err := global.DB.First(&dec, id).Error; err != nil {
                response.FailWithMessage("决策不存在", c)
                return
        }

        dec.Status = decision.DecisionStatusApproved
        now := time.Now()
        dec.ApprovedAt = &now
        // dec.ApprovedBy = userID

        global.DB.Save(&dec)

        response.OkWithData(dec, c)
}

// RejectDecision 拒绝决策
func RejectDecision(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var req struct {
                Reason string `json:"reason"`
        }
        c.ShouldBindJSON(&req)

        var dec decision.AIDecision
        if err := global.DB.First(&dec, id).Error; err != nil {
                response.FailWithMessage("决策不存在", c)
                return
        }

        dec.Status = decision.DecisionStatusRejected
        now := time.Now()
        dec.RejectedAt = &now
        dec.RejectReason = req.Reason
        // dec.RejectedBy = userID

        global.DB.Save(&dec)

        response.OkWithData(dec, c)
}

// ExecuteDecision 执行决策
func ExecuteDecision(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                response.FailWithMessage("无效的ID", c)
                return
        }

        var dec decision.AIDecision
        if err := global.DB.First(&dec, id).Error; err != nil {
                response.FailWithMessage("决策不存在", c)
                return
        }

        if dec.Status != decision.DecisionStatusApproved {
                response.FailWithMessage("决策未批准，无法执行", c)
                return
        }

        // TODO: 实际执行命令

        dec.Status = decision.DecisionStatusExecuted
        now := time.Now()
        dec.ExecutedAt = &now

        global.DB.Save(&dec)

        response.OkWithData(dec, c)
}
