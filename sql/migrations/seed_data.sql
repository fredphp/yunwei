-- 模拟数据初始化脚本
-- 为所有表添加示例数据

USE `yunwei`;

-- ==================== 系统管理 ====================

-- 系统用户数据
INSERT INTO `sys_users` (`username`, `password`, `nick_name`, `email`, `phone`, `status`, `role_id`) VALUES
('admin', 'e10adc3949ba59abbe56e057f20f883e', '超级管理员', 'admin@yunwei.com', '13800138000', 1, 1),
('operator', 'e10adc3949ba59abbe56e057f20f883e', '运维工程师', 'operator@yunwei.com', '13800138001', 1, 2),
('viewer', 'e10adc3949ba59abbe56e057f20f883e', '只读用户', 'viewer@yunwei.com', '13800138002', 1, 3)
ON DUPLICATE KEY UPDATE `nick_name` = VALUES(`nick_name`);

-- 系统角色数据
INSERT INTO `sys_roles` (`name`, `keyword`, `description`, `status`) VALUES
('超级管理员', 'admin', '拥有系统所有权限', 1),
('运维工程师', 'operator', '负责服务器运维管理', 1),
('只读用户', 'viewer', '只能查看数据', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 系统菜单数据
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`) VALUES
(0, '仪表盘', 'Dashboard', '/dashboard', 'views/dashboard/index.vue', 'Odometer', 1, 1, 0),
(0, '服务器管理', 'Servers', '/servers/list', 'views/servers/index.vue', 'Monitor', 2, 1, 0),
(0, 'Kubernetes', 'Kubernetes', '/kubernetes', 'views/kubernetes/index.vue', 'Grid', 3, 1, 0),
(0, '告警中心', 'Alerts', '/alerts', 'views/alerts/index.vue', 'Bell', 4, 1, 0),
(0, '成本分析', 'Cost', '/cost', 'views/cost/index.vue', 'Money', 5, 1, 0),
(0, '定时任务', 'Scheduler', '/scheduler', 'views/scheduler/index.vue', 'Timer', 6, 1, 0),
(0, 'Agent管理', 'Agents', '/agents', 'views/agents/index.vue', 'Cpu', 7, 1, 0),
(0, '租户管理', 'Tenant', '/tenant', 'views/tenant/index.vue', 'OfficeBuilding', 8, 1, 0),
(0, '高可用', 'HA', '/ha', 'views/ha/index.vue', 'Connection', 9, 1, 0),
(0, '系统设置', 'System', '/system', '', 'Setting', 99, 1, 0)
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`);

-- 系统API数据
INSERT INTO `sys_apis` (`path`, `method`, `group`, `description`) VALUES
('/api/v1/servers', 'GET', '服务器管理', '获取服务器列表'),
('/api/v1/servers', 'POST', '服务器管理', '创建服务器'),
('/api/v1/servers/:id', 'PUT', '服务器管理', '更新服务器'),
('/api/v1/servers/:id', 'DELETE', '服务器管理', '删除服务器'),
('/api/v1/user/menus', 'GET', '用户管理', '获取用户菜单'),
('/api/v1/user/info', 'GET', '用户管理', '获取用户信息')
ON DUPLICATE KEY UPDATE `path` = VALUES(`path`);

-- ==================== Agent管理 ====================

-- Agent版本数据
INSERT INTO `agent_versions` (`version`, `version_code`, `build_time`, `build_commit`, `build_branch`, `file_url`, `file_md5`, `file_size`, `platform`, `arch`, `changelog`, `release_type`, `enabled`, `is_latest`) VALUES
('1.0.0', 100, NOW(), 'a1b2c3d4', 'main', 'https://downloads.yunwei.com/agent/v1.0.0/agent-linux-amd64', 'abc123def456', 15728640, 'linux', 'amd64', '初始版本发布', 'stable', 1, 0),
('1.1.0', 110, NOW(), 'e5f6g7h8', 'main', 'https://downloads.yunwei.com/agent/v1.1.0/agent-linux-amd64', 'def456ghi789', 16777216, 'linux', 'amd64', '新增自动恢复功能', 'stable', 1, 1),
('1.1.0', 110, NOW(), 'e5f6g7h8', 'main', 'https://downloads.yunwei.com/agent/v1.1.0/agent-linux-arm64', 'ghi789jkl012', 15728640, 'linux', 'arm64', '新增自动恢复功能', 'stable', 1, 1)
ON DUPLICATE KEY UPDATE `version` = VALUES(`version`);

-- Agent实例数据
INSERT INTO `agents` (`server_id`, `server_name`, `server_ip`, `agent_id`, `version`, `version_code`, `platform`, `arch`, `os`, `status`, `auto_upgrade`, `enabled`) VALUES
(1, '生产服务器-01', '192.168.1.100', 'agent-001-abc', '1.1.0', 110, 'linux', 'amd64', 'CentOS 7.9', 'online', 1, 1),
(2, '测试服务器-01', '192.168.1.101', 'agent-002-def', '1.0.0', 100, 'linux', 'amd64', 'Ubuntu 20.04', 'online', 1, 1)
ON DUPLICATE KEY UPDATE `agent_id` = VALUES(`agent_id`);

-- Agent配置数据
INSERT INTO `agent_configs` (`name`, `description`, `config_json`, `scope`, `is_default`, `enabled`) VALUES
('默认配置', 'Agent默认配置', '{"heartbeat_interval": 10, "report_interval": 60, "log_level": "info"}', 'all', 1, 1),
('高性能配置', '适用于高性能服务器', '{"heartbeat_interval": 5, "report_interval": 30, "log_level": "debug"}', 'group', 0, 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 灰度发布策略数据
INSERT INTO `gray_release_strategies` (`name`, `description`, `version_name`, `strategy_type`, `initial_weight`, `target_weight`, `step_size`, `step_interval`, `status`, `enabled`) VALUES
('v1.1.0灰度升级', '逐步升级到v1.1.0版本', '1.1.0', 'weight', 10, 100, 20, 60, 'pending', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- ==================== 调度器 ====================

-- 调度队列数据
INSERT INTO `scheduler_queues` (`name`, `description`, `max_workers`, `max_pending`, `priority`, `timeout`, `max_retry`, `enabled`) VALUES
('default', '默认任务队列', 10, 1000, 5, 3600, 3, 1),
('high-priority', '高优先级队列', 20, 500, 10, 1800, 5, 1),
('low-priority', '低优先级队列', 5, 2000, 1, 7200, 2, 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 定时任务数据
INSERT INTO `scheduler_cron_jobs` (`name`, `description`, `cron_expr`, `timezone`, `enabled`, `concurrent_policy`, `notify_on_success`, `notify_on_fail`) VALUES
('每日备份', '每天凌晨2点执行备份', '0 2 * * *', 'Asia/Shanghai', 1, 'forbid', 0, 1),
('清理日志', '每小时清理过期日志', '0 * * * *', 'Asia/Shanghai', 1, 'allow', 0, 0),
('健康检查', '每5分钟检查服务器健康状态', '*/5 * * * *', 'Asia/Shanghai', 1, 'allow', 0, 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 任务模板数据
INSERT INTO `scheduler_task_templates` (`name`, `category`, `description`, `task_def`, `enabled`) VALUES
('服务器备份', 'backup', '备份服务器数据', '{"type": "backup", "timeout": 3600}', 1),
('日志清理', 'cleanup', '清理过期日志文件', '{"type": "cleanup", "timeout": 600}', 1),
('服务重启', 'operation', '重启指定服务', '{"type": "command", "timeout": 300}', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 调度任务数据
INSERT INTO `scheduler_tasks` (`name`, `type`, `priority`, `status`, `command`, `server_id`, `server_name`, `created_by`) VALUES
('备份生产数据库', 'backup', 10, 'pending', 'mysqldump -u root -p*** yunwei > backup.sql', 1, '生产服务器-01', 1),
('清理临时文件', 'cleanup', 5, 'completed', 'rm -rf /tmp/old/*', 2, '测试服务器-01', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- ==================== 租户管理 ====================

-- 租户数据
INSERT INTO `tenants` (`id`, `name`, `slug`, `domain`, `description`, `status`, `plan`, `contact_name`, `contact_email`, `contact_phone`) VALUES
(UUID(), '示例公司', 'demo-company', 'demo.yunwei.com', '示例租户公司', 'active', 'pro', '张三', 'zhangsan@demo.com', '13900139000'),
(UUID(), '测试企业', 'test-enterprise', 'test.yunwei.com', '测试用企业租户', 'active', 'starter', '李四', 'lisi@test.com', '13900139001')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 租户角色数据
INSERT INTO `tenant_roles` (`id`, `tenant_id`, `name`, `slug`, `description`, `is_system`, `permissions`, `scope`) VALUES
(UUID(), (SELECT id FROM tenants WHERE slug = 'demo-company' LIMIT 1), '管理员', 'admin', '租户管理员', 1, '["*"]', 'tenant'),
(UUID(), (SELECT id FROM tenants WHERE slug = 'demo-company' LIMIT 1), '运维人员', 'operator', '运维操作人员', 1, '["servers:*", "alerts:read"]', 'tenant')
ON DUPLICATE KEY UPDATE `slug` = VALUES(`slug`);

-- ==================== 高可用 ====================

-- 集群节点数据
INSERT INTO `cluster_nodes` (`node_id`, `node_name`, `hostname`, `internal_ip`, `api_port`, `grpc_port`, `status`, `role`, `is_leader`, `version`, `weight`, `data_center`, `zone`, `enabled`) VALUES
('node-001', '主节点-北京', 'yunwei-node-01', '10.0.1.10', 8080, 9090, 'online', 'leader', 1, '1.0.0', 100, 'beijing', 'zone-a', 1),
('node-002', '从节点-北京', 'yunwei-node-02', '10.0.1.11', 8080, 9090, 'online', 'follower', 0, '1.0.0', 100, 'beijing', 'zone-a', 1),
('node-003', '从节点-上海', 'yunwei-node-03', '10.0.2.10', 8080, 9090, 'online', 'follower', 0, '1.0.0', 80, 'shanghai', 'zone-b', 1)
ON DUPLICATE KEY UPDATE `node_id` = VALUES(`node_id`);

-- HA集群配置数据
INSERT INTO `ha_cluster_configs` (`name`, `description`, `cluster_name`, `cluster_mode`, `min_nodes`, `max_nodes`, `heartbeat_interval`, `heartbeat_timeout`, `failover_enabled`, `enabled`) VALUES
('生产集群配置', '生产环境HA配置', 'prod-cluster', 'active-active', 2, 10, 10, 30, 1, 1),
('测试集群配置', '测试环境HA配置', 'test-cluster', 'active-passive', 1, 3, 15, 60, 1, 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 集群事件数据
INSERT INTO `cluster_events` (`event_type`, `node_id`, `node_name`, `node_ip`, `title`, `detail`, `level`, `source`) VALUES
('leader_election', 'node-001', '主节点-北京', '10.0.1.10', 'Leader选举完成', '节点 node-001 当选为 Leader', 'info', 'election'),
('node_join', 'node-003', '从节点-上海', '10.0.2.10', '节点加入集群', '新节点 node-003 成功加入集群', 'info', 'cluster')
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`);

-- ==================== Kubernetes ====================

-- K8s集群数据
INSERT INTO `k8s_clusters` (`name`, `api_endpoint`, `status`, `version`, `node_count`, `auto_scale_enabled`, `min_replicas`, `max_replicas`, `cpu_threshold`, `mem_threshold`) VALUES
('生产K8s集群', 'https://k8s-prod.example.com:6443', 'connected', 'v1.28.0', 10, 1, 2, 20, 80, 80),
('测试K8s集群', 'https://k8s-test.example.com:6443', 'connected', 'v1.27.0', 3, 0, 1, 5, 70, 70)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- K8s扩容事件数据
INSERT INTO `k8s_scale_events` (`cluster_id`, `namespace`, `deployment`, `scale_type`, `status`, `replicas_before`, `replicas_after`, `replicas_target`, `trigger_reason`, `ai_decision`) VALUES
(1, 'default', 'web-api', 'auto', 'success', 3, 5, 5, 'CPU使用率超过阈值', 'AI分析建议扩容到5个副本'),
(1, 'production', 'payment-service', 'manual', 'success', 2, 4, 4, '业务高峰期手动扩容', '运维人员手动触发扩容')
ON DUPLICATE KEY UPDATE `deployment` = VALUES(`deployment`);

-- K8s HPA配置数据
INSERT INTO `k8s_hpa_configs` (`cluster_id`, `namespace`, `deployment`, `min_replicas`, `max_replicas`, `target_cpu_util`, `target_mem_util`, `enabled`) VALUES
(1, 'default', 'web-api', 2, 10, 70, 75, 1),
(1, 'production', 'payment-service', 3, 15, 65, 70, 1)
ON DUPLICATE KEY UPDATE `deployment` = VALUES(`deployment`);

-- K8s Deployment状态数据
INSERT INTO `k8s_deployment_status` (`cluster_id`, `namespace`, `deployment`, `replicas`, `ready_replicas`, `cpu_usage`, `memory_usage`, `hpa_enabled`) VALUES
(1, 'default', 'web-api', 5, 5, 45.5, 62.3, 1),
(1, 'production', 'payment-service', 4, 4, 38.2, 55.8, 1)
ON DUPLICATE KEY UPDATE `deployment` = VALUES(`deployment`);

-- ==================== 其他服务 ====================

-- 证书数据
INSERT INTO `certificates` (`id`, `created_at`, `domain`, `issuer`, `not_before`, `not_after`, `status`, `auto_renew`) VALUES
(1, NOW(), '*.yunwei.com', "Let's Encrypt", DATE_SUB(NOW(), INTERVAL 30 DAY), DATE_ADD(NOW(), INTERVAL 60 DAY), 'valid', 1),
(2, NOW(), 'api.yunwei.com', "Let's Encrypt", DATE_SUB(NOW(), INTERVAL 60 DAY), DATE_ADD(NOW(), INTERVAL 30 DAY), 'expiring_soon', 1)
ON DUPLICATE KEY UPDATE `domain` = VALUES(`domain`);

-- 负载均衡器数据
INSERT INTO `load_balancers` (`name`, `type`, `address`, `port`, `status`, `algorithm`, `enabled`) VALUES
('生产LB', 'nginx', '10.0.1.100', 80, 'active', 'round-robin', 1),
('测试LB', 'nginx', '10.0.2.100', 80, 'active', 'least-conn', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- CDN域名数据
INSERT INTO `cdn_domains` (`domain`, `origin`, `type`, `status`, `enabled`) VALUES
('static.yunwei.com', 'https://origin.yunwei.com', 'web', 'active', 1),
('img.yunwei.com', 'https://img-origin.yunwei.com', 'image', 'active', 1)
ON DUPLICATE KEY UPDATE `domain` = VALUES(`domain`);

-- 告警规则数据
INSERT INTO `alert_rules` (`name`, `description`, `metric`, `operator`, `threshold`, `duration`, `level`, `status`) VALUES
('CPU使用率告警', 'CPU使用率超过80%', 'cpu_usage', '>', 80, 60, 'warning', 1),
('CPU使用率严重告警', 'CPU使用率超过95%', 'cpu_usage', '>', 95, 30, 'critical', 1),
('内存使用率告警', '内存使用率超过85%', 'memory_usage', '>', 85, 60, 'warning', 1),
('磁盘使用率告警', '磁盘使用率超过90%', 'disk_usage', '>', 90, 60, 'critical', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 告警记录数据
INSERT INTO `alerts` (`rule_id`, `server_id`, `level`, `title`, `message`, `metric_value`, `status`, `fired_at`) VALUES
(1, 1, 'warning', 'CPU使用率告警', '服务器 生产服务器-01 CPU使用率达到 85%', 85.5, 'resolved', DATE_SUB(NOW(), INTERVAL 2 HOUR)),
(3, 2, 'warning', '内存使用率告警', '服务器 测试服务器-01 内存使用率达到 88%', 88.2, 'firing', DATE_SUB(NOW(), INTERVAL 30 MINUTE))
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`);

-- 部署计划数据
INSERT INTO `deploy_plans` (`name`, `description`, `status`, `strategy`, `target_servers`) VALUES
('Web应用部署计划', '部署最新版本Web应用', 'completed', 'rolling', '[1, 2]'),
('API服务更新', '更新API服务到v2.0', 'pending', 'blue-green', '[1]')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 部署任务数据
INSERT INTO `deploy_tasks` (`plan_id`, `server_id`, `status`, `version`, `started_at`, `completed_at`) VALUES
(1, 1, 'success', 'v1.2.0', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 30 MINUTE),
(1, 2, 'success', 'v1.2.0', DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 5 MINUTE, DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 35 MINUTE)
ON DUPLICATE KEY UPDATE `plan_id` = VALUES(`plan_id`);

-- 自愈规则数据
INSERT INTO `heal_rules` (`name`, `description`, `trigger_condition`, `action`, `enabled`) VALUES
('服务重启', '当服务异常退出时自动重启', '{"type": "service_crash"}', '{"type": "restart_service"}', 1),
('磁盘清理', '当磁盘使用率超过阈值时清理日志', '{"type": "disk_high", "threshold": 90}', '{"type": "cleanup_logs"}', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 自愈记录数据
INSERT INTO `heal_records` (`rule_id`, `server_id`, `trigger_type`, `action`, `status`, `started_at`, `completed_at`) VALUES
(1, 1, 'service_crash', 'restart_service', 'success', DATE_SUB(NOW(), INTERVAL 3 HOUR), DATE_SUB(NOW(), INTERVAL 3 HOUR) + INTERVAL 1 MINUTE),
(2, 2, 'disk_high', 'cleanup_logs', 'success', DATE_SUB(NOW(), INTERVAL 5 HOUR), DATE_SUB(NOW(), INTERVAL 5 HOUR) + INTERVAL 5 MINUTE)
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 安全规则数据
INSERT INTO `security_rules` (`name`, `type`, `rule`, `action`, `enabled`) VALUES
('防暴力破解', 'login', '{"max_attempts": 5, "window": 300}', 'block_ip', 1),
('SQL注入防护', 'request', '{"patterns": ["SELECT.*FROM", "UNION.*SELECT"]}', 'block_request', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- IP黑名单数据
INSERT INTO `ip_blacklist` (`ip`, `reason`, `expire_at`, `created_by`) VALUES
('192.168.100.100', '暴力破解尝试', DATE_ADD(NOW(), INTERVAL 24 HOUR), 1),
('10.0.0.50', '恶意请求', DATE_ADD(NOW(), INTERVAL 1 HOUR), 1)
ON DUPLICATE KEY UPDATE `ip` = VALUES(`ip`);

-- IP白名单数据
INSERT INTO `ip_whitelist` (`ip`, `description`, `created_by`) VALUES
('10.0.0.0/8', '内网IP段', 1),
('192.168.0.0/16', '内网IP段', 1)
ON DUPLICATE KEY UPDATE `ip` = VALUES(`ip`);

-- 登录记录数据
INSERT INTO `login_records` (`user_id`, `username`, `ip`, `status`, `user_agent`) VALUES
(1, 'admin', '192.168.1.50', 'success', 'Mozilla/5.0 Chrome/120.0'),
(1, 'admin', '192.168.100.100', 'failed', 'Mozilla/5.0 Chrome/120.0')
ON DUPLICATE KEY UPDATE `user_id` = VALUES(`user_id`);

-- 预测结果数据
INSERT INTO `prediction_results` (`server_id`, `metric_type`, `predicted_value`, `confidence`, `predicted_at`) VALUES
(1, 'cpu_usage', 75.5, 0.85, DATE_ADD(NOW(), INTERVAL 1 HOUR)),
(1, 'memory_usage', 82.3, 0.78, DATE_ADD(NOW(), INTERVAL 1 HOUR))
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- AI决策数据
INSERT INTO `ai_decisions` (`server_id`, `decision_type`, `input_data`, `output_data`, `confidence`, `executed`) VALUES
(1, 'auto_scale', '{"cpu_usage": 85, "memory_usage": 70}', '{"action": "scale_up", "replicas": 3}', 0.92, 1),
(2, 'alert_suppress', '{"alert_type": "cpu_high", "duration": 30}', '{"action": "suppress", "reason": "scheduled_maintenance"}', 0.88, 0)
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 审计日志数据
INSERT INTO `audit_logs` (`user_id`, `username`, `action`, `resource`, `resource_id`, `ip_address`, `status`) VALUES
(1, 'admin', 'create', 'server', 1, '192.168.1.50', 'success'),
(1, 'admin', 'update', 'server', 1, '192.168.1.50', 'success'),
(2, 'operator', 'delete', 'server', 2, '192.168.1.51', 'failed')
ON DUPLICATE KEY UPDATE `user_id` = VALUES(`user_id`);

-- 通知记录数据
INSERT INTO `notification_records` (`type`, `title`, `content`, `recipient`, `status`) VALUES
('email', '服务器告警', '生产服务器-01 CPU使用率超过阈值', 'admin@yunwei.com', 'sent'),
('webhook', '部署完成', 'Web应用部署计划已完成', 'https://hooks.example.com/notify', 'sent')
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`);

-- 巡检记录数据
INSERT INTO `patrol_records` (`server_id`, `type`, `status`, `result`, `started_at`, `completed_at`) VALUES
(1, 'daily', 'completed', '{"cpu": "ok", "memory": "ok", "disk": "warning"}', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 5 MINUTE),
(2, 'weekly', 'completed', '{"cpu": "ok", "memory": "ok", "disk": "ok"}', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 10 MINUTE)
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 检测规则数据
INSERT INTO `detect_rules` (`name`, `type`, `condition`, `severity`, `enabled`) VALUES
('CPU持续高负载', 'metric', '{"metric": "cpu_usage", "operator": ">", "threshold": 80, "duration": 300}', 'high', 1),
('服务端口不可达', 'network', '{"port": 8080, "timeout": 5}', 'critical', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 执行记录数据
INSERT INTO `execution_records` (`task_id`, `server_id`, `status`, `output`, `started_at`, `completed_at`) VALUES
(1, 1, 'success', '备份完成: backup_20240115.sql', DATE_SUB(NOW(), INTERVAL 12 HOUR), DATE_SUB(NOW(), INTERVAL 12 HOUR) + INTERVAL 120 SECOND),
(2, 2, 'success', '清理完成: 删除 50 个文件', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 30 SECOND)
ON DUPLICATE KEY UPDATE `task_id` = VALUES(`task_id`);

-- 自动化操作记录
INSERT INTO `auto_actions` (`server_id`, `type`, `trigger`, `action`, `status`, `executed_at`) VALUES
(1, 'scale', 'cpu_high', '{"replicas": 5}', 'success', DATE_SUB(NOW(), INTERVAL 2 HOUR)),
(2, 'restart', 'service_down', '{"service": "nginx"}', 'success', DATE_SUB(NOW(), INTERVAL 4 HOUR))
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 异常检测数据
INSERT INTO `anomaly_detections` (`server_id`, `metric_type`, `anomaly_score`, `detected_at`, `confirmed`) VALUES
(1, 'cpu_usage', 0.92, DATE_SUB(NOW(), INTERVAL 1 HOUR), 1),
(2, 'network_latency', 0.78, DATE_SUB(NOW(), INTERVAL 3 HOUR), 0)
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 自动扩容建议数据
INSERT INTO `auto_scale_recommendations` (`server_id`, `current_replicas`, `recommended_replicas`, `reason`, `confidence`) VALUES
(1, 3, 5, 'CPU使用率持续高于阈值', 0.85),
(2, 2, 3, '请求量增长趋势', 0.72)
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 服务健康数据
INSERT INTO `service_health` (`server_id`, `service_name`, `status`, `last_check`) VALUES
(1, 'nginx', 'healthy', NOW()),
(1, 'mysql', 'healthy', NOW()),
(2, 'redis', 'healthy', NOW())
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 项目分析数据
INSERT INTO `project_analyses` (`project_name`, `project_type`, `framework`, `language`, `recommendations`) VALUES
('yunwei-web', 'web', 'Vue3', 'TypeScript', '{"build": "vite", "deploy": "docker"}'),
('yunwei-api', 'api', 'Gin', 'Go', '{"build": "go build", "deploy": "docker"}')
ON DUPLICATE KEY UPDATE `project_name` = VALUES(`project_name`);

-- 服务器能力数据
INSERT INTO `server_capabilities` (`server_id`, `cpu_available`, `memory_available`, `disk_available`, `services`) VALUES
(1, 4, 8192, 500, '["nginx", "mysql", "redis"]'),
(2, 2, 4096, 200, '["nginx", "redis"]')
ON DUPLICATE KEY UPDATE `server_id` = VALUES(`server_id`);

-- 资源池数据
INSERT INTO `resource_pools` (`name`, `type`, `total_cpu`, `total_memory`, `used_cpu`, `used_memory`) VALUES
('生产资源池', 'production', 32, 65536, 20, 45000),
('测试资源池', 'test', 16, 32768, 8, 15000)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 金丝雀发布数据
INSERT INTO `canary_releases` (`name`, `service`, `stable_version`, `canary_version`, `canary_weight`, `status`) VALUES
('API金丝雀发布', 'api-service', 'v1.0', 'v2.0', 10, 'running'),
('Web金丝雀发布', 'web-service', 'v1.5', 'v1.6', 5, 'paused')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 金丝雀配置数据
INSERT INTO `canary_configs` (`name`, `strategy`, `steps`, `metrics_threshold`) VALUES
('渐进式发布', 'progressive', '[{"weight": 5, "duration": 300}, {"weight": 20, "duration": 600}, {"weight": 50, "duration": 900}]', '{"error_rate": 1, "latency_p99": 500}'),
('快速发布', 'fast', '[{"weight": 10, "duration": 60}, {"weight": 50, "duration": 120}, {"weight": 100, "duration": 180}]', '{"error_rate": 5, "latency_p99": 1000}')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 备份记录数据
INSERT INTO `backup_records` (`id`, `created_at`, `server_id`, `type`, `path`, `size`, `status`) VALUES
(1, DATE_SUB(NOW(), INTERVAL 1 DAY), 1, 'database', '/backup/mysql/backup_20240115.sql', 524288000, 'completed'),
(2, DATE_SUB(NOW(), INTERVAL 2 DAY), 1, 'file', '/backup/files/backup_20240114.tar.gz', 1073741824, 'completed')
ON DUPLICATE KEY UPDATE `id` = VALUES(`id`);

-- 证书续期记录
INSERT INTO `cert_renewal_records` (`certificate_id`, `status`, `old_expire`, `new_expire`, `renewed_at`) VALUES
(2, 'success', DATE_ADD(NOW(), INTERVAL 30 DAY), DATE_ADD(NOW(), INTERVAL 90 DAY), NOW())
ON DUPLICATE KEY UPDATE `certificate_id` = VALUES(`certificate_id`);

-- 工作流记录
INSERT INTO `workflow_records` (`name`, `type`, `status`, `trigger`, `started_at`, `completed_at`) VALUES
('部署工作流', 'deploy', 'completed', 'manual', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 30 MINUTE),
('备份工作流', 'backup', 'completed', 'scheduled', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 10 MINUTE)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 完成提示
SELECT '模拟数据初始化完成!' AS message;
