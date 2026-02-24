package router

import (
        "yunwei/api/v1/auth"
        "yunwei/api/v1/server"
        "yunwei/api/v1/kubernetes"
        "yunwei/api/v1/canary"
        "yunwei/api/v1/loadbalancer"
        "yunwei/api/v1/cert"
        cdnApi "yunwei/api/v1/cdn"
        deployApi "yunwei/api/v1/deploy"
        schedulerApi "yunwei/api/v1/scheduler"
        agentApi "yunwei/api/v1/agent"
        haApi "yunwei/api/v1/ha"
        backupApi "yunwei/api/v1/backup"
        costApi "yunwei/api/v1/cost"
        tenantApi "yunwei/api/v1/tenant"
        "yunwei/api/v1/system"
        "yunwei/middleware"
        "yunwei/websocket"

        "github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
        // 中间件
        r.Use(middleware.Cors())
        r.Use(middleware.Logger())

        // WebSocket 路由
        wsService := websocket.NewWebSocketService()
        wsService.Start()
        r.GET("/ws", func(c *gin.Context) {
                wsService.HandleWebSocket(c)
        })

        // API v1
        v1 := r.Group("/api/v1")
        {
                // 公开接口
                public := v1.Group("")
                {
                        public.POST("/login", auth.Login)
                        public.POST("/register", auth.Register)
                }

                // 需要认证的接口
                authGroup := v1.Group("")
                authGroup.Use(middleware.JWTAuth())
                {
                        // 用户信息
                        authGroup.GET("/user/info", auth.GetUserInfo)

                        // ==================== 权限管理 ====================
                        // 获取当前用户权限
                        authGroup.GET("/user/permissions", system.GetUserPermissions)
                        authGroup.POST("/user/check-permission", system.CheckPermission)
                        authGroup.POST("/user/check-permissions", system.CheckPermissions)

                        // 权限列表（只读）
                        authGroup.GET("/permissions", middleware.RequirePermission("role:view"), system.GetPermissions)

                        // 角色管理 - 需要 role:* 权限
                        authGroup.GET("/roles", middleware.RequirePermission("role:view"), system.GetRoles)
                        authGroup.GET("/roles/:id", middleware.RequirePermission("role:view"), system.GetRole)
                        authGroup.POST("/roles", middleware.RequirePermission("role:add"), system.CreateRole)
                        authGroup.PUT("/roles/:id", middleware.RequirePermission("role:edit"), system.UpdateRole)
                        authGroup.DELETE("/roles/:id", middleware.RequirePermission("role:delete"), system.DeleteRole)

                        // 用户角色分配 - 需要 user:edit 权限
                        authGroup.POST("/users/assign-role", middleware.RequirePermission("user:edit"), system.AssignRole)
                        authGroup.POST("/users/revoke-role", middleware.RequirePermission("user:edit"), system.RevokeRole)
                        authGroup.GET("/users/:id/roles", middleware.RequirePermission("user:view"), system.GetUserRolePermissions)

                        // ==================== 服务器管理 ====================
                        // 权限说明：
                        // - 查看服务器: server:view (管理员、运维、只读)
                        // - 添加服务器: server:add (管理员)
                        // - 编辑服务器: server:edit (管理员)
                        // - 删除服务器: server:delete (管理员)
                        // - 执行命令: server:execute (管理员、运维)
                        // - SSH连接: server:ssh (管理员、运维)
                        // - AI分析: server:analyze (管理员、运维)
                        servers := authGroup.Group("/servers")
                        {
                                // 查看类操作 - 需要 server:view 权限
                                servers.GET("", middleware.RequirePermission("server:view"), server.GetServerList)
                                servers.GET("/:id", middleware.RequirePermission("server:view"), server.GetServer)
                                servers.GET("/:id/metrics", middleware.RequirePermission("server:view"), server.GetServerMetrics)
                                servers.GET("/:id/logs", middleware.RequirePermission("server:view"), server.GetServerLogs)
                                servers.GET("/:id/containers", middleware.RequirePermission("server:view"), server.GetDockerContainers)
                                servers.GET("/:id/ports", middleware.RequirePermission("server:view"), server.GetPortInfos)
                                servers.POST("/:id/refresh", middleware.RequirePermission("server:view"), server.RefreshStatus)

                                // 添加服务器 - 需要 server:add 权限 (管理员)
                                servers.POST("", middleware.RequirePermission("server:add"), server.AddServer)

                                // 编辑服务器 - 需要 server:edit 权限 (管理员)
                                servers.PUT("/:id", middleware.RequirePermission("server:edit"), server.UpdateServer)

                                // 删除服务器 - 需要 server:delete 权限 (管理员)
                                servers.DELETE("/:id", middleware.RequirePermission("server:delete"), server.DeleteServer)

                                // 执行命令 - 需要 server:execute 权限 (管理员、运维)
                                servers.POST("/:id/command", middleware.RequirePermission("server:execute"), server.ExecuteCommand)

                                // AI分析 - 需要 server:analyze 权限 (管理员、运维)
                                servers.POST("/:id/analyze", middleware.RequirePermission("server:analyze"), server.AIAnalyze)
                        }

                        // SSH 测试 - 需要 server:ssh 权限
                        authGroup.POST("/ssh/test", middleware.RequirePermission("server:ssh"), server.TestSSH)

                        // 服务器分组
                        // 权限说明：
                        // - 查看分组: server_group:view
                        // - 添加分组: server_group:add (管理员)
                        // - 删除分组: server_group:delete (管理员)
                        groups := authGroup.Group("/groups")
                        {
                                groups.GET("", middleware.RequirePermission("server_group:view"), server.GetGroups)
                                groups.POST("", middleware.RequirePermission("server_group:add"), server.CreateGroup)
                                groups.DELETE("/:id", middleware.RequirePermission("server_group:delete"), server.DeleteGroup)
                        }

                        // ==================== 告警管理 ====================
                        alerts := authGroup.Group("/alerts")
                        {
                                alerts.GET("", middleware.RequirePermission("alert:view"), server.GetAlerts)
                                alerts.POST("/:id/acknowledge", middleware.RequirePermission("alert:handle"), server.AcknowledgeAlert)
                        }

                        // ==================== 检测规则 ====================
                        rules := authGroup.Group("/rules")
                        {
                                rules.GET("", middleware.RequirePermission("alert_rule:view"), server.GetRules)
                                rules.PUT("/:id", middleware.RequirePermission("alert_rule:edit"), server.UpdateRule)
                        }

                        // ==================== 自动操作 ====================
                        actions := authGroup.Group("/actions")
                        {
                                actions.GET("", middleware.RequirePermission("alert:view"), server.GetAutoActions)
                                actions.POST("/:id/execute", middleware.RequirePermission("alert:handle"), server.ExecuteAutoAction)
                        }

                        // ==================== AI 决策 ====================
                        decisions := authGroup.Group("/decisions")
                        {
                                decisions.GET("", middleware.RequirePermission("ai:analyze"), server.GetDecisions)
                                decisions.POST("/:id/approve", middleware.RequirePermission("ai:approve"), server.ApproveDecision)
                                decisions.POST("/:id/reject", middleware.RequirePermission("ai:approve"), server.RejectDecision)
                                decisions.POST("/:id/execute", middleware.RequirePermission("ai:execute"), server.ExecuteDecision)
                        }

                        // ==================== Kubernetes 管理 ====================
                        k8s := authGroup.Group("/kubernetes")
                        {
                                // 集群管理 - 查看权限
                                k8s.GET("/clusters", middleware.RequirePermission("k8s:view"), kubernetes.GetClusters)
                                k8s.GET("/clusters/:id", middleware.RequirePermission("k8s:view"), kubernetes.GetCluster)
                                k8s.GET("/clusters/:clusterId/deployments", middleware.RequirePermission("k8s:view"), kubernetes.GetDeploymentStatus)

                                // 集群管理 - 添加权限 (管理员)
                                k8s.POST("/clusters", middleware.RequirePermission("k8s:add"), kubernetes.AddCluster)

                                // 集群管理 - 编辑权限 (管理员)
                                k8s.PUT("/clusters/:id", middleware.RequirePermission("k8s:edit"), kubernetes.UpdateCluster)

                                // 集群管理 - 删除权限 (管理员)
                                k8s.DELETE("/clusters/:id", middleware.RequirePermission("k8s:delete"), kubernetes.DeleteCluster)

                                // 扩容管理
                                k8s.GET("/scale/history", middleware.RequirePermission("k8s:view"), kubernetes.GetScaleHistory)
                                k8s.POST("/scale/manual", middleware.RequirePermission("k8s:scale"), kubernetes.ManualScale)
                                k8s.POST("/clusters/:clusterId/analyze", middleware.RequirePermission("k8s:view"), kubernetes.AnalyzeScale)

                                // HPA 配置
                                k8s.GET("/hpa", middleware.RequirePermission("k8s:view"), kubernetes.GetHPAConfigs)
                                k8s.POST("/hpa", middleware.RequirePermission("k8s:scale"), kubernetes.UpdateHPAConfig)
                        }

                        // ==================== 灰度发布 ====================
                        canaryGroup := authGroup.Group("/canary")
                        {
                                // 查看权限
                                canaryGroup.GET("/releases", middleware.RequirePermission("canary:view"), canary.GetReleases)
                                canaryGroup.GET("/releases/:id", middleware.RequirePermission("canary:view"), canary.GetRelease)
                                canaryGroup.GET("/releases/:id/steps", middleware.RequirePermission("canary:view"), canary.GetReleaseSteps)
                                canaryGroup.GET("/configs", middleware.RequirePermission("canary:view"), canary.GetConfigs)

                                // 创建/配置权限 (管理员)
                                canaryGroup.POST("/releases", middleware.RequirePermission("canary:add"), canary.StartCanary)
                                canaryGroup.POST("/configs", middleware.RequirePermission("canary:config"), canary.UpdateConfig)

                                // 部署操作权限 (管理员)
                                canaryGroup.POST("/releases/:id/promote", middleware.RequirePermission("canary:deploy"), canary.PromoteCanary)
                                canaryGroup.POST("/releases/:id/complete", middleware.RequirePermission("canary:deploy"), canary.CompleteCanary)

                                // 回滚权限 (管理员)
                                canaryGroup.POST("/releases/:id/rollback", middleware.RequirePermission("canary:rollback"), canary.RollbackCanary)
                                canaryGroup.POST("/releases/:id/pause", middleware.RequirePermission("canary:deploy"), canary.PauseCanary)
                                canaryGroup.POST("/releases/:id/abort", middleware.RequirePermission("canary:rollback"), canary.AbortCanary)
                        }

                        // ==================== 负载均衡 ====================
                        lb := authGroup.Group("/loadbalancer")
                        {
                                // 查看权限
                                lb.GET("", middleware.RequirePermission("lb:view"), loadbalancer.GetLBs)
                                lb.GET("/:id", middleware.RequirePermission("lb:view"), loadbalancer.GetLB)
                                lb.GET("/:id/backends", middleware.RequirePermission("lb:view"), loadbalancer.GetBackends)
                                lb.GET("/history", middleware.RequirePermission("lb:view"), loadbalancer.GetOptimizationHistory)
                                lb.GET("/algorithm", middleware.RequirePermission("lb:view"), loadbalancer.GetAlgorithmConfigs)

                                // 添加权限 (管理员)
                                lb.POST("", middleware.RequirePermission("lb:add"), loadbalancer.AddLB)
                                lb.POST("/:id/backends", middleware.RequirePermission("lb:add"), loadbalancer.AddBackend)

                                // 编辑权限 (管理员)
                                lb.PUT("/:id", middleware.RequirePermission("lb:edit"), loadbalancer.UpdateLB)
                                lb.PUT("/backends/:id", middleware.RequirePermission("lb:edit"), loadbalancer.UpdateBackend)

                                // 删除权限 (管理员)
                                lb.DELETE("/:id", middleware.RequirePermission("lb:delete"), loadbalancer.DeleteLB)
                                lb.DELETE("/backends/:id", middleware.RequirePermission("lb:delete"), loadbalancer.DeleteBackend)

                                // 操作权限 (管理员、运维)
                                lb.POST("/:id/optimize", middleware.RequirePermission("lb:optimize"), loadbalancer.OptimizeLB)
                                lb.POST("/:id/autobalance", middleware.RequirePermission("lb:operate"), loadbalancer.AutoBalance)
                                lb.POST("/:id/healthcheck", middleware.RequirePermission("lb:operate"), loadbalancer.HealthCheck)
                                lb.POST("/algorithm", middleware.RequirePermission("lb:optimize"), loadbalancer.UpdateAlgorithmConfig)
                        }

                        // ==================== 证书管理 ====================
                        certGroup := authGroup.Group("/certificates")
                        {
                                // 查看权限
                                certGroup.GET("", middleware.RequirePermission("cert:view"), cert.GetCertificates)
                                certGroup.GET("/:id", middleware.RequirePermission("cert:view"), cert.GetCertificate)
                                certGroup.GET("/history", middleware.RequirePermission("cert:view"), cert.GetRenewalHistory)

                                // 添加权限 (管理员)
                                certGroup.POST("", middleware.RequirePermission("cert:add"), cert.AddCertificate)
                                certGroup.POST("/request", middleware.RequirePermission("cert:add"), cert.RequestNewCert)

                                // 编辑权限 (管理员)
                                certGroup.PUT("/:id", middleware.RequirePermission("cert:edit"), cert.UpdateCertificate)

                                // 删除权限 (管理员)
                                certGroup.DELETE("/:id", middleware.RequirePermission("cert:delete"), cert.DeleteCertificate)

                                // 续签权限 (管理员)
                                certGroup.POST("/:id/renew", middleware.RequirePermission("cert:renew"), cert.RenewCertificate)

                                // 检查权限 (管理员、运维)
                                certGroup.POST("/:id/check", middleware.RequirePermission("cert:check"), cert.CheckCertificate)
                                certGroup.POST("/check-all", middleware.RequirePermission("cert:check"), cert.CheckAllCertificates)
                        }

                        // ==================== CDN 管理 ====================
                        cdnGroup := authGroup.Group("/cdn")
                        {
                                // 查看权限
                                cdnGroup.GET("/domains", middleware.RequirePermission("cdn:view"), cdnApi.GetDomains)
                                cdnGroup.GET("/domains/:id", middleware.RequirePermission("cdn:view"), cdnApi.GetDomain)
                                cdnGroup.GET("/domains/:id/nodes", middleware.RequirePermission("cdn:view"), cdnApi.GetNodeStatus)
                                cdnGroup.GET("/domains/:id/rules", middleware.RequirePermission("cdn:view"), cdnApi.GetCacheRules)
                                cdnGroup.GET("/domains/:id/cost", middleware.RequirePermission("cdn:view"), cdnApi.CalculateCost)
                                cdnGroup.GET("/history", middleware.RequirePermission("cdn:view"), cdnApi.GetOptimizationHistory)

                                // 添加权限 (管理员)
                                cdnGroup.POST("/domains", middleware.RequirePermission("cdn:add"), cdnApi.AddDomain)
                                cdnGroup.POST("/domains/:id/rules", middleware.RequirePermission("cdn:add"), cdnApi.AddCacheRule)

                                // 编辑权限 (管理员)
                                cdnGroup.PUT("/domains/:id", middleware.RequirePermission("cdn:edit"), cdnApi.UpdateDomain)
                                cdnGroup.PUT("/rules/:id", middleware.RequirePermission("cdn:edit"), cdnApi.UpdateCacheRule)

                                // 删除权限 (管理员)
                                cdnGroup.DELETE("/domains/:id", middleware.RequirePermission("cdn:delete"), cdnApi.DeleteDomain)
                                cdnGroup.DELETE("/rules/:id", middleware.RequirePermission("cdn:delete"), cdnApi.DeleteCacheRule)

                                // 优化权限 (管理员)
                                cdnGroup.POST("/domains/:id/optimize", middleware.RequirePermission("cdn:optimize"), cdnApi.OptimizeCDN)
                                cdnGroup.POST("/domains/:id/cost-optimize", middleware.RequirePermission("cdn:optimize"), cdnApi.OptimizeCost)

                                // 操作权限 (管理员、运维)
                                cdnGroup.POST("/domains/:id/purge", middleware.RequirePermission("cdn:operate"), cdnApi.PurgeCache)
                                cdnGroup.POST("/domains/:id/preheat", middleware.RequirePermission("cdn:operate"), cdnApi.PreheatCache)
                        }

                        // ==================== 智能部署 ====================
                        deployGroup := authGroup.Group("/deploy")
                        {
                                // 查看权限
                                deployGroup.GET("/projects", middleware.RequirePermission("deploy:view"), deployApi.GetProjectAnalyses)
                                deployGroup.GET("/projects/:id", middleware.RequirePermission("deploy:view"), deployApi.GetProjectAnalysis)
                                deployGroup.GET("/plans", middleware.RequirePermission("deploy:view"), deployApi.GetDeployPlans)
                                deployGroup.GET("/plans/:id", middleware.RequirePermission("deploy:view"), deployApi.GetDeployPlan)
                                deployGroup.GET("/plans/:id/topology", middleware.RequirePermission("deploy:view"), deployApi.GetServiceTopology)
                                deployGroup.GET("/plans/:id/preview", middleware.RequirePermission("deploy:view"), deployApi.PreviewConfigs)
                                deployGroup.GET("/tasks", middleware.RequirePermission("deploy:view"), deployApi.GetDeployTasks)
                                deployGroup.GET("/tasks/:id", middleware.RequirePermission("deploy:view"), deployApi.GetDeployTask)
                                deployGroup.GET("/tasks/:id/steps", middleware.RequirePermission("deploy:view"), deployApi.GetTaskSteps)

                                // 分析权限 (管理员、运维)
                                deployGroup.POST("/upload", middleware.RequirePermission("deploy:analyze"), deployApi.UploadProject)
                                deployGroup.POST("/analyze", middleware.RequirePermission("deploy:analyze"), deployApi.AnalyzeProject)
                                deployGroup.GET("/servers/analyze", middleware.RequirePermission("deploy:analyze"), deployApi.AnalyzeServers)
                                deployGroup.GET("/servers/capabilities", middleware.RequirePermission("deploy:view"), deployApi.GetServerCapabilities)
                                deployGroup.POST("/servers/find-best", middleware.RequirePermission("deploy:analyze"), deployApi.FindBestServers)

                                // 创建权限 (管理员)
                                deployGroup.POST("/plans", middleware.RequirePermission("deploy:add"), deployApi.GenerateDeployPlan)

                                // 删除权限 (管理员)
                                deployGroup.DELETE("/plans/:id", middleware.RequirePermission("deploy:add"), deployApi.DeleteDeployPlan)

                                // 执行权限 (管理员)
                                deployGroup.POST("/plans/:id/execute", middleware.RequirePermission("deploy:execute"), deployApi.ExecuteDeploy)
                                deployGroup.POST("/tasks/:id/pause", middleware.RequirePermission("deploy:execute"), deployApi.PauseDeploy)
                                deployGroup.POST("/tasks/:id/resume", middleware.RequirePermission("deploy:execute"), deployApi.ResumeDeploy)

                                // 回滚权限 (管理员)
                                deployGroup.POST("/tasks/:id/rollback", middleware.RequirePermission("deploy:rollback"), deployApi.RollbackDeploy)
                        }

                        // ==================== 任务调度中心 ====================
                        schedulerGroup := authGroup.Group("/scheduler")
                        {
                                // 查看权限
                                schedulerGroup.GET("/dashboard", middleware.RequirePermission("scheduler:view"), schedulerApi.GetDashboard)
                                schedulerGroup.GET("/tasks", middleware.RequirePermission("scheduler:view"), schedulerApi.ListTasks)
                                schedulerGroup.GET("/tasks/:id", middleware.RequirePermission("scheduler:view"), schedulerApi.GetTask)
                                schedulerGroup.GET("/tasks/:id/executions", middleware.RequirePermission("scheduler:view"), schedulerApi.GetTaskExecutions)
                                schedulerGroup.GET("/batches", middleware.RequirePermission("scheduler:view"), schedulerApi.ListBatches)
                                schedulerGroup.GET("/batches/:id", middleware.RequirePermission("scheduler:view"), schedulerApi.GetBatch)
                                schedulerGroup.GET("/batches/:id/tasks", middleware.RequirePermission("scheduler:view"), schedulerApi.GetBatchTasks)
                                schedulerGroup.GET("/cron", middleware.RequirePermission("scheduler:view"), schedulerApi.ListCronJobs)
                                schedulerGroup.GET("/cron/:id", middleware.RequirePermission("scheduler:view"), schedulerApi.GetCronJob)
                                schedulerGroup.GET("/cron/:id/executions", middleware.RequirePermission("scheduler:view"), schedulerApi.GetCronExecutions)
                                schedulerGroup.GET("/queues", middleware.RequirePermission("scheduler:view"), schedulerApi.GetQueues)
                                schedulerGroup.GET("/queues/stats", middleware.RequirePermission("scheduler:view"), schedulerApi.GetQueueStats)
                                schedulerGroup.GET("/workers", middleware.RequirePermission("scheduler:view"), schedulerApi.GetWorkers)
                                schedulerGroup.GET("/templates", middleware.RequirePermission("scheduler:view"), schedulerApi.ListTemplates)

                                // 创建权限 (管理员)
                                schedulerGroup.POST("/tasks", middleware.RequirePermission("scheduler:add"), schedulerApi.SubmitTask)
                                schedulerGroup.POST("/tasks/options", middleware.RequirePermission("scheduler:add"), schedulerApi.SubmitTaskWithOptions)
                                schedulerGroup.POST("/batches", middleware.RequirePermission("scheduler:add"), schedulerApi.SubmitBatch)
                                schedulerGroup.POST("/cron", middleware.RequirePermission("scheduler:add"), schedulerApi.CreateCronJob)
                                schedulerGroup.POST("/templates", middleware.RequirePermission("scheduler:add"), schedulerApi.CreateTemplate)
                                schedulerGroup.POST("/templates/submit", middleware.RequirePermission("scheduler:add"), schedulerApi.SubmitFromTemplate)

                                // 操作权限 (管理员、运维)
                                schedulerGroup.POST("/tasks/:id/cancel", middleware.RequirePermission("scheduler:operate"), schedulerApi.CancelTask)
                                schedulerGroup.POST("/tasks/:id/retry", middleware.RequirePermission("scheduler:operate"), schedulerApi.RetryTask)
                                schedulerGroup.POST("/tasks/:id/rollback", middleware.RequirePermission("scheduler:operate"), schedulerApi.RollbackTask)
                                schedulerGroup.PUT("/cron/:id", middleware.RequirePermission("scheduler:operate"), schedulerApi.UpdateCronJob)
                                schedulerGroup.DELETE("/cron/:id", middleware.RequirePermission("scheduler:operate"), schedulerApi.DeleteCronJob)
                                schedulerGroup.POST("/workers/scale", middleware.RequirePermission("scheduler:operate"), schedulerApi.ScaleWorkers)

                                // 触发权限 (管理员、运维)
                                schedulerGroup.POST("/cron/:id/trigger", middleware.RequirePermission("scheduler:trigger"), schedulerApi.TriggerCronJob)
                        }

                        // ==================== Agent 管理 ====================
                        agentGroup := authGroup.Group("/agents")
                        {
                                // 查看权限
                                agentGroup.GET("", middleware.RequirePermission("agent:view"), agentApi.GetAgentList)
                                agentGroup.GET("/stats", middleware.RequirePermission("agent:view"), agentApi.GetAgentStats)
                                agentGroup.GET("/:id", middleware.RequirePermission("agent:view"), agentApi.GetAgent)
                                agentGroup.GET("/:id/config", middleware.RequirePermission("agent:view"), agentApi.GetAgentConfig)
                                agentGroup.GET("/:id/check-upgrade", middleware.RequirePermission("agent:view"), agentApi.CheckUpgrade)
                                agentGroup.GET("/:id/heartbeats", middleware.RequirePermission("agent:view"), agentApi.GetHeartbeatRecords)
                                agentGroup.GET("/:id/recovers", middleware.RequirePermission("agent:view"), agentApi.GetRecoverRecords)
                                agentGroup.GET("/versions", middleware.RequirePermission("agent:view"), agentApi.GetVersionList)
                                agentGroup.GET("/versions/stats", middleware.RequirePermission("agent:view"), agentApi.GetVersionStats)
                                agentGroup.GET("/versions/:id", middleware.RequirePermission("agent:view"), agentApi.GetVersion)
                                agentGroup.GET("/upgrades", middleware.RequirePermission("agent:view"), agentApi.GetUpgradeTaskList)
                                agentGroup.GET("/upgrades/stats", middleware.RequirePermission("agent:view"), agentApi.GetUpgradeStats)
                                agentGroup.GET("/upgrades/:id", middleware.RequirePermission("agent:view"), agentApi.GetUpgradeTask)
                                agentGroup.GET("/gray", middleware.RequirePermission("agent:view"), agentApi.GetGrayStrategyList)
                                agentGroup.GET("/gray/stats", middleware.RequirePermission("agent:view"), agentApi.GetMonitorStats)
                                agentGroup.GET("/gray/:id", middleware.RequirePermission("agent:view"), agentApi.GetGrayStrategy)
                                agentGroup.GET("/gray/:id/progress", middleware.RequirePermission("agent:view"), agentApi.GetGrayStrategyProgress)
                                agentGroup.GET("/monitor/stats", middleware.RequirePermission("agent:view"), agentApi.GetMonitorStats)
                                agentGroup.GET("/monitor/offline", middleware.RequirePermission("agent:view"), agentApi.GetOfflineAgents)

                                // 编辑权限 (管理员)
                                agentGroup.PUT("/:id", middleware.RequirePermission("agent:edit"), agentApi.UpdateAgent)
                                agentGroup.POST("/versions", middleware.RequirePermission("agent:edit"), agentApi.CreateVersion)
                                agentGroup.PUT("/versions/:id", middleware.RequirePermission("agent:edit"), agentApi.UpdateVersion)

                                // 删除权限 (管理员)
                                agentGroup.DELETE("/:id", middleware.RequirePermission("agent:delete"), agentApi.DeleteAgent)
                                agentGroup.DELETE("/versions/:id", middleware.RequirePermission("agent:delete"), agentApi.DeleteVersion)

                                // 操作权限 (管理员、运维)
                                agentGroup.POST("/:id/disable", middleware.RequirePermission("agent:operate"), agentApi.DisableAgent)
                                agentGroup.POST("/:id/enable", middleware.RequirePermission("agent:operate"), agentApi.EnableAgent)
                                agentGroup.POST("/batch", middleware.RequirePermission("agent:operate"), agentApi.BatchOperation)
                                agentGroup.POST("/upgrades", middleware.RequirePermission("agent:upgrade"), agentApi.CreateUpgradeTask)
                                agentGroup.POST("/upgrades/batch", middleware.RequirePermission("agent:upgrade"), agentApi.CreateBatchUpgrade)
                                agentGroup.POST("/upgrades/:id/execute", middleware.RequirePermission("agent:upgrade"), agentApi.ExecuteUpgrade)
                                agentGroup.POST("/upgrades/:id/cancel", middleware.RequirePermission("agent:upgrade"), agentApi.CancelUpgrade)
                                agentGroup.POST("/upgrades/:id/rollback", middleware.RequirePermission("agent:upgrade"), agentApi.RollbackUpgrade)
                                agentGroup.POST("/gray", middleware.RequirePermission("agent:operate"), agentApi.CreateGrayStrategy)
                                agentGroup.POST("/gray/:id/start", middleware.RequirePermission("agent:operate"), agentApi.StartGrayStrategy)
                                agentGroup.POST("/gray/:id/pause", middleware.RequirePermission("agent:operate"), agentApi.PauseGrayStrategy)
                                agentGroup.POST("/gray/:id/resume", middleware.RequirePermission("agent:operate"), agentApi.ResumeGrayStrategy)
                                agentGroup.POST("/gray/:id/cancel", middleware.RequirePermission("agent:operate"), agentApi.CancelGrayStrategy)
                        }

                        // ==================== HA 高可用管理 ====================
                        haGroup := authGroup.Group("/ha")
                        {
                                // 查看权限
                                haGroup.GET("/stats", middleware.RequirePermission("ha:view"), haApi.GetClusterStats)
                                haGroup.GET("/nodes", middleware.RequirePermission("ha:view"), haApi.GetClusterNodes)
                                haGroup.GET("/nodes/:id", middleware.RequirePermission("ha:view"), haApi.GetClusterNode)
                                haGroup.GET("/nodes/:id/metrics", middleware.RequirePermission("ha:view"), haApi.GetNodeMetrics)
                                haGroup.GET("/leader", middleware.RequirePermission("ha:view"), haApi.GetLeaderStatus)
                                haGroup.GET("/leader/records", middleware.RequirePermission("ha:view"), haApi.GetLeaderElectionRecords)
                                haGroup.GET("/locks", middleware.RequirePermission("ha:view"), haApi.GetLocks)
                                haGroup.GET("/locks/records", middleware.RequirePermission("ha:view"), haApi.GetLockDBRecords)
                                haGroup.GET("/locks/:key", middleware.RequirePermission("ha:view"), haApi.GetLock)
                                haGroup.GET("/sessions", middleware.RequirePermission("ha:view"), haApi.GetSessions)
                                haGroup.GET("/sessions/stats", middleware.RequirePermission("ha:view"), haApi.GetSessionStats)
                                haGroup.GET("/configs", middleware.RequirePermission("ha:view"), haApi.ListHAConfigs)
                                haGroup.GET("/config", middleware.RequirePermission("ha:view"), haApi.GetHAConfig)
                                haGroup.GET("/failover", middleware.RequirePermission("ha:view"), haApi.GetFailoverRecords)
                                haGroup.GET("/events", middleware.RequirePermission("ha:view"), haApi.GetClusterEvents)
                                haGroup.GET("/tasks/running", middleware.RequirePermission("ha:view"), haApi.GetRunningTasks)

                                // 操作权限 (管理员)
                                haGroup.POST("/nodes/:id/enable", middleware.RequirePermission("ha:operate"), haApi.EnableNode)
                                haGroup.POST("/nodes/:id/disable", middleware.RequirePermission("ha:operate"), haApi.DisableNode)
                                haGroup.POST("/leader/resign", middleware.RequirePermission("ha:operate"), haApi.ResignLeader)
                                haGroup.POST("/leader/force", middleware.RequirePermission("ha:operate"), haApi.ForceLeader)
                                haGroup.POST("/locks/:key/release", middleware.RequirePermission("ha:operate"), haApi.ForceReleaseLock)
                                haGroup.DELETE("/sessions/:id", middleware.RequirePermission("ha:operate"), haApi.DeleteSession)

                                // 配置权限 (管理员)
                                haGroup.PUT("/config", middleware.RequirePermission("ha:config"), haApi.UpdateHAConfig)
                                haGroup.POST("/configs", middleware.RequirePermission("ha:config"), haApi.CreateHAConfig)
                                haGroup.DELETE("/configs/:id", middleware.RequirePermission("ha:config"), haApi.DeleteHAConfig)

                                // 故障转移权限 (超级管理员)
                                haGroup.POST("/failover/trigger", middleware.RequirePermission("ha:failover"), haApi.TriggerFailover)
                        }

                        // ==================== 灾备与备份管理 ====================
                        backupGroup := authGroup.Group("/backup")
                        {
                                backupApi.RegisterRoutes(backupGroup, backupApi.NewHandler())
                        }

                        // ==================== 成本控制系统 ====================
                        costGroup := authGroup.Group("/cost")
                        {
                                costApi.RegisterRoutes(costGroup, costApi.NewHandler())
                        }

                        // ==================== 多租户系统 ====================
                        tenantGroup := authGroup.Group("")
                        {
                                tenantApi.RegisterRoutes(tenantGroup, tenantApi.NewHandler())
                        }
                }
        }

        // 健康检查
        r.GET("/health", func(c *gin.Context) {
                c.JSON(200, gin.H{
                        "status":  "ok",
                        "message": "Yunwei Server is running",
                })
        })
}
