-- 模拟数据初始化脚本（优化版 - 无重复数据）
-- 使用 INSERT IGNORE 避免重复插入

USE `yunwei`;

-- ==================== 系统管理 ====================

-- 系统用户数据
INSERT IGNORE INTO `sys_users` (`id`, `username`, `password`, `nick_name`, `email`, `phone`, `status`, `role_id`) VALUES
(1, 'admin', 'e10adc3949ba59abbe56e057f20f883e', '超级管理员', 'admin@yunwei.com', '13800138000', 1, 1),
(2, 'operator', 'e10adc3949ba59abbe56e057f20f883e', '运维工程师', 'operator@yunwei.com', '13800138001', 1, 2),
(3, 'viewer', 'e10adc3949ba59abbe56e057f20f883e', '只读用户', 'viewer@yunwei.com', '13800138002', 1, 3);

-- 系统角色数据
INSERT IGNORE INTO `sys_roles` (`id`, `name`, `keyword`, `description`, `status`) VALUES
(1, '超级管理员', 'admin', '拥有系统所有权限', 1),
(2, '运维工程师', 'operator', '负责服务器运维管理', 1),
(3, '只读用户', 'viewer', '只能查看数据', 1);

-- 系统菜单数据（完整的菜单树，无重复）
-- 注意：此菜单结构已与 init.sql 保持一致
INSERT IGNORE INTO `sys_menus` (`id`, `parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`) VALUES
-- 一级菜单
(1, 0, '仪表盘', 'Dashboard', '/dashboard', 'views/dashboard/index', 'Odometer', 1, 1, 0),
(2, 0, '服务器管理', 'Servers', '/servers', 'Layout', 'Monitor', 2, 1, 0),
(3, 0, '告警中心', 'Alerts', '/alerts', 'views/alerts/index', 'Bell', 3, 1, 0),
(4, 0, 'Kubernetes', 'Kubernetes', '/kubernetes', 'Layout', 'Grid', 4, 1, 0),
(5, 0, '灰度发布', 'Canary', '/canary', 'Layout', 'Promotion', 5, 1, 0),
(6, 0, '负载均衡', 'LoadBalancer', '/loadbalancer', 'Layout', 'Connection', 6, 1, 0),
(7, 0, '证书管理', 'Certificate', '/certificate', 'Layout', 'DocumentChecked', 7, 1, 0),
(8, 0, 'CDN管理', 'CDN', '/cdn', 'Layout', 'Position', 8, 1, 0),
(9, 0, '智能部署', 'Deploy', '/deploy', 'Layout', 'Upload', 9, 1, 0),
(10, 0, '任务调度', 'Scheduler', '/scheduler', 'Layout', 'Timer', 10, 1, 0),
(11, 0, 'Agent管理', 'Agents', '/agents', 'Layout', 'Cpu', 11, 1, 0),
(12, 0, '高可用', 'HA', '/ha', 'Layout', 'CircleCheck', 12, 1, 0),
(13, 0, '灾备备份', 'Backup', '/backup', 'Layout', 'Files', 13, 1, 0),
(14, 0, '成本控制', 'Cost', '/cost', 'Layout', 'Coin', 14, 1, 0),
(15, 0, '系统管理', 'System', '/system', 'Layout', 'Setting', 15, 1, 0),
-- 二级菜单 - 服务器管理
(101, 2, '服务器列表', 'ServerList', '/servers/list', 'views/servers/index', 'List', 1, 1, 0),
-- 二级菜单 - Kubernetes
(102, 4, '集群管理', 'Clusters', '/kubernetes/clusters', 'views/kubernetes/index', 'Cluster', 1, 1, 0),
-- 二级菜单 - 系统管理
(103, 15, '用户管理', 'UserManage', '/system/user', 'views/system/user/index', 'User', 1, 1, 0),
(104, 15, '角色管理', 'RoleManage', '/system/role', 'views/system/role/index', 'UserFilled', 2, 1, 0),
(105, 15, '菜单管理', 'MenuManage', '/system/menu', 'views/system/menu/index', 'Menu', 3, 1, 0);

-- 系统 API 数据
INSERT IGNORE INTO `sys_apis` (`id`, `path`, `method`, `group`, `description`) VALUES
(1, '/api/v1/servers', 'GET', '服务器管理', '获取服务器列表'),
(2, '/api/v1/servers', 'POST', '服务器管理', '创建服务器'),
(3, '/api/v1/servers/:id', 'PUT', '服务器管理', '更新服务器'),
(4, '/api/v1/servers/:id', 'DELETE', '服务器管理', '删除服务器'),
(5, '/api/v1/user/menus', 'GET', '用户管理', '获取用户菜单'),
(6, '/api/v1/user/info', 'GET', '用户管理', '获取用户信息');

-- 角色-API 关联
INSERT IGNORE INTO `sys_role_apis` (`role_id`, `api_id`) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6),
(2, 1), (2, 3), (2, 5), (2, 6),
(3, 1), (3, 5), (3, 6);

-- 角色-菜单 关联（菜单ID已更新，与init.sql结构一致）
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`) VALUES
-- 超级管理员：拥有所有菜单权限
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9), (1, 10),
(1, 11), (1, 12), (1, 13), (1, 14), (1, 15),
(1, 101), (1, 102), (1, 103), (1, 104), (1, 105),
-- 运维工程师：拥有核心功能权限
(2, 1), (2, 2), (2, 3), (2, 4), (2, 6), (2, 7), (2, 9), (2, 11), (2, 12),
(2, 101), (2, 102),
-- 只读用户：只有查看权限
(3, 1), (3, 2), (3, 3), (3, 4),
(3, 101), (3, 102);

-- ==================== Agent管理 ====================

-- Agent版本数据
INSERT IGNORE INTO `agent_versions` (`id`, `version`, `version_code`, `build_time`, `build_commit`, `build_branch`, `file_url`, `file_md5`, `file_size`, `platform`, `arch`, `changelog`, `release_type`, `enabled`, `is_latest`) VALUES
(1, '1.0.0', 100, NOW(), 'a1b2c3d4', 'main', 'https://downloads.yunwei.com/agent/v1.0.0/agent-linux-amd64', 'abc123def456', 15728640, 'linux', 'amd64', '初始版本发布', 'stable', 1, 0),
(2, '1.1.0', 110, NOW(), 'e5f6g7h8', 'main', 'https://downloads.yunwei.com/agent/v1.1.0/agent-linux-amd64', 'def456ghi789', 16777216, 'linux', 'amd64', '新增自动恢复功能', 'stable', 1, 1),
(3, '1.1.0', 110, NOW(), 'e5f6g7h8', 'main', 'https://downloads.yunwei.com/agent/v1.1.0/agent-linux-arm64', 'ghi789jkl012', 15728640, 'linux', 'arm64', '新增自动恢复功能', 'stable', 1, 1);

-- Agent实例数据
INSERT IGNORE INTO `agents` (`id`, `server_id`, `server_name`, `server_ip`, `agent_id`, `version`, `version_code`, `platform`, `arch`, `os`, `status`, `auto_upgrade`, `enabled`) VALUES
(1, 1, '生产服务器-01', '192.168.1.100', 'agent-001-abc', '1.1.0', 110, 'linux', 'amd64', 'CentOS 7.9', 'online', 1, 1),
(2, 2, '测试服务器-01', '192.168.1.101', 'agent-002-def', '1.0.0', 100, 'linux', 'amd64', 'Ubuntu 20.04', 'online', 1, 1);

-- Agent配置数据
INSERT IGNORE INTO `agent_configs` (`id`, `name`, `description`, `config_json`, `scope`, `is_default`, `enabled`) VALUES
(1, '默认配置', 'Agent默认配置', '{"heartbeat_interval": 10, "report_interval": 60, "log_level": "info"}', 'all', 1, 1),
(2, '高性能配置', '适用于高性能服务器', '{"heartbeat_interval": 5, "report_interval": 30, "log_level": "debug"}', 'group', 0, 1);

-- 灰度发布策略数据
INSERT IGNORE INTO `gray_release_strategies` (`id`, `name`, `description`, `version_name`, `strategy_type`, `initial_weight`, `target_weight`, `step_size`, `step_interval`, `status`, `enabled`) VALUES
(1, 'v1.1.0灰度升级', '逐步升级到v1.1.0版本', '1.1.0', 'weight', 10, 100, 20, 60, 'pending', 1);

-- ==================== 调度器 ====================

-- 调度队列数据
INSERT IGNORE INTO `scheduler_queues` (`id`, `name`, `description`, `max_workers`, `max_pending`, `priority`, `timeout`, `max_retry`, `enabled`) VALUES
(1, 'default', '默认任务队列', 10, 1000, 5, 3600, 3, 1),
(2, 'high-priority', '高优先级队列', 20, 500, 10, 1800, 5, 1),
(3, 'low-priority', '低优先级队列', 5, 2000, 1, 7200, 2, 1);

-- 定时任务数据
INSERT IGNORE INTO `scheduler_cron_jobs` (`id`, `name`, `description`, `cron_expr`, `timezone`, `enabled`, `concurrent_policy`, `notify_on_success`, `notify_on_fail`) VALUES
(1, '每日备份', '每天凌晨2点执行备份', '0 2 * * *', 'Asia/Shanghai', 1, 'forbid', 0, 1),
(2, '清理日志', '每小时清理过期日志', '0 * * * *', 'Asia/Shanghai', 1, 'allow', 0, 0),
(3, '健康检查', '每5分钟检查服务器健康状态', '*/5 * * * *', 'Asia/Shanghai', 1, 'allow', 0, 1);

-- 任务模板数据
INSERT IGNORE INTO `scheduler_task_templates` (`id`, `name`, `category`, `description`, `task_def`, `enabled`) VALUES
(1, '服务器备份', 'backup', '备份服务器数据', '{"type": "backup", "timeout": 3600}', 1),
(2, '日志清理', 'cleanup', '清理过期日志文件', '{"type": "cleanup", "timeout": 600}', 1),
(3, '服务重启', 'operation', '重启指定服务', '{"type": "command", "timeout": 300}', 1);

-- 调度任务数据
INSERT IGNORE INTO `scheduler_tasks` (`id`, `name`, `type`, `priority`, `status`, `command`, `server_id`, `server_name`, `created_by`) VALUES
(1, '备份生产数据库', 'backup', 10, 'pending', 'mysqldump -u root -p*** yunwei > backup.sql', 1, '生产服务器-01', 1),
(2, '清理临时文件', 'cleanup', 5, 'completed', 'rm -rf /tmp/old/*', 2, '测试服务器-01', 1);

-- ==================== 租户管理 ====================

-- 租户数据
INSERT IGNORE INTO `tenants` (`id`, `name`, `slug`, `domain`, `description`, `status`, `plan`, `contact_name`, `contact_email`, `contact_phone`) VALUES
('demo-tenant-001', '示例公司', 'demo-company', 'demo.yunwei.com', '示例租户公司', 'active', 'pro', '张三', 'zhangsan@demo.com', '13900139000'),
('demo-tenant-002', '测试企业', 'test-enterprise', 'test.yunwei.com', '测试用企业租户', 'active', 'starter', '李四', 'lisi@test.com', '13900139001');

-- 租户角色数据
INSERT IGNORE INTO `tenant_roles` (`id`, `tenant_id`, `name`, `slug`, `description`, `is_system`, `permissions`, `scope`) VALUES
('role-001', 'demo-tenant-001', '管理员', 'admin', '租户管理员', 1, '["*"]', 'tenant'),
('role-002', 'demo-tenant-001', '运维人员', 'operator', '运维操作人员', 1, '["servers:*", "alerts:read"]', 'tenant');

-- ==================== 高可用 ====================

-- 集群节点数据
INSERT IGNORE INTO `cluster_nodes` (`id`, `node_id`, `node_name`, `hostname`, `internal_ip`, `api_port`, `grpc_port`, `status`, `role`, `is_leader`, `version`, `weight`, `data_center`, `zone`, `enabled`) VALUES
(1, 'node-001', '主节点-北京', 'yunwei-node-01', '10.0.1.10', 8080, 9090, 'online', 'leader', 1, '1.0.0', 100, 'beijing', 'zone-a', 1),
(2, 'node-002', '从节点-北京', 'yunwei-node-02', '10.0.1.11', 8080, 9090, 'online', 'follower', 0, '1.0.0', 100, 'beijing', 'zone-a', 1),
(3, 'node-003', '从节点-上海', 'yunwei-node-03', '10.0.2.10', 8080, 9090, 'online', 'follower', 0, '1.0.0', 80, 'shanghai', 'zone-b', 1);

-- HA集群配置数据
INSERT IGNORE INTO `ha_cluster_configs` (`id`, `name`, `description`, `cluster_name`, `cluster_mode`, `min_nodes`, `max_nodes`, `heartbeat_interval`, `heartbeat_timeout`, `failover_enabled`, `enabled`) VALUES
(1, '生产集群配置', '生产环境HA配置', 'prod-cluster', 'active-active', 2, 10, 10, 30, 1, 1),
(2, '测试集群配置', '测试环境HA配置', 'test-cluster', 'active-passive', 1, 3, 15, 60, 1, 1);

-- 集群事件数据
INSERT IGNORE INTO `cluster_events` (`id`, `event_type`, `node_id`, `node_name`, `node_ip`, `title`, `detail`, `level`, `source`) VALUES
(1, 'leader_election', 'node-001', '主节点-北京', '10.0.1.10', 'Leader选举完成', '节点 node-001 当选为 Leader', 'info', 'election'),
(2, 'node_join', 'node-003', '从节点-上海', '10.0.2.10', '节点加入集群', '新节点 node-003 成功加入集群', 'info', 'cluster');

-- ==================== Kubernetes ====================

-- K8s集群数据
INSERT IGNORE INTO `k8s_clusters` (`id`, `name`, `api_endpoint`, `status`, `version`, `node_count`, `auto_scale_enabled`, `min_replicas`, `max_replicas`, `cpu_threshold`, `mem_threshold`) VALUES
(1, '生产K8s集群', 'https://k8s-prod.example.com:6443', 'connected', 'v1.28.0', 10, 1, 2, 20, 80, 80),
(2, '测试K8s集群', 'https://k8s-test.example.com:6443', 'connected', 'v1.27.0', 3, 0, 1, 5, 70, 70);

-- K8s扩容事件数据
INSERT IGNORE INTO `k8s_scale_events` (`id`, `cluster_id`, `namespace`, `deployment`, `scale_type`, `status`, `replicas_before`, `replicas_after`, `replicas_target`, `trigger_reason`, `ai_decision`) VALUES
(1, 1, 'default', 'web-api', 'auto', 'success', 3, 5, 5, 'CPU使用率超过阈值', 'AI分析建议扩容到5个副本'),
(2, 1, 'production', 'payment-service', 'manual', 'success', 2, 4, 4, '业务高峰期手动扩容', '运维人员手动触发扩容');

-- K8s HPA配置数据
INSERT IGNORE INTO `k8s_hpa_configs` (`id`, `cluster_id`, `namespace`, `deployment`, `min_replicas`, `max_replicas`, `target_cpu_util`, `target_mem_util`, `enabled`) VALUES
(1, 1, 'default', 'web-api', 2, 10, 70, 75, 1),
(2, 1, 'production', 'payment-service', 3, 15, 65, 70, 1);

-- K8s Deployment状态数据
INSERT IGNORE INTO `k8s_deployment_status` (`id`, `cluster_id`, `namespace`, `deployment`, `replicas`, `ready_replicas`, `cpu_usage`, `memory_usage`, `hpa_enabled`) VALUES
(1, 1, 'default', 'web-api', 5, 5, 45.5, 62.3, 1),
(2, 1, 'production', 'payment-service', 4, 4, 38.2, 55.8, 1);

-- ==================== 其他服务 ====================

-- 证书数据
INSERT IGNORE INTO `certificates` (`id`, `name`, `domain`, `issuer`, `not_before`, `not_after`, `status`, `auto_renew`) VALUES
(1, '主域名证书', '*.yunwei.com', "Let's Encrypt", DATE_SUB(NOW(), INTERVAL 30 DAY), DATE_ADD(NOW(), INTERVAL 60 DAY), 'valid', 1),
(2, 'API证书', 'api.yunwei.com', "Let's Encrypt", DATE_SUB(NOW(), INTERVAL 60 DAY), DATE_ADD(NOW(), INTERVAL 30 DAY), 'expiring_soon', 1);

-- 负载均衡器数据
INSERT IGNORE INTO `load_balancers` (`id`, `name`, `type`, `address`, `port`, `status`, `algorithm`, `enabled`) VALUES
(1, '生产LB', 'nginx', '10.0.1.100', 80, 'active', 'round-robin', 1),
(2, '测试LB', 'nginx', '10.0.2.100', 80, 'active', 'least-conn', 1);

-- CDN域名数据
INSERT IGNORE INTO `cdn_domains` (`id`, `domain`, `origin`, `type`, `status`, `enabled`) VALUES
(1, 'static.yunwei.com', 'https://origin.yunwei.com', 'web', 'active', 1),
(2, 'img.yunwei.com', 'https://img-origin.yunwei.com', 'image', 'active', 1);

-- 告警规则数据
INSERT IGNORE INTO `alert_rules` (`id`, `name`, `description`, `metric`, `operator`, `threshold`, `duration`, `level`, `status`) VALUES
(1, 'CPU使用率告警', 'CPU使用率超过80%', 'cpu_usage', '>', 80, 60, 'warning', 1),
(2, 'CPU使用率严重告警', 'CPU使用率超过95%', 'cpu_usage', '>', 95, 30, 'critical', 1),
(3, '内存使用率告警', '内存使用率超过85%', 'memory_usage', '>', 85, 60, 'warning', 1),
(4, '磁盘使用率告警', '磁盘使用率超过90%', 'disk_usage', '>', 90, 60, 'critical', 1);

-- 告警记录数据
INSERT IGNORE INTO `alerts` (`id`, `rule_id`, `server_id`, `level`, `title`, `message`, `metric_value`, `status`, `fired_at`) VALUES
(1, 1, 1, 'warning', 'CPU使用率告警', '服务器 生产服务器-01 CPU使用率达到 85%', 85.5, 'resolved', DATE_SUB(NOW(), INTERVAL 2 HOUR)),
(2, 3, 2, 'warning', '内存使用率告警', '服务器 测试服务器-01 内存使用率达到 88%', 88.2, 'firing', DATE_SUB(NOW(), INTERVAL 30 MINUTE));

-- 部署计划数据
INSERT IGNORE INTO `deploy_plans` (`id`, `name`, `description`, `status`, `strategy`, `target_servers`) VALUES
(1, 'Web应用部署计划', '部署最新版本Web应用', 'completed', 'rolling', '[1, 2]'),
(2, 'API服务更新', '更新API服务到v2.0', 'pending', 'blue-green', '[1]');

-- 部署任务数据
INSERT IGNORE INTO `deploy_tasks` (`id`, `plan_id`, `server_id`, `status`, `version`, `started_at`, `completed_at`) VALUES
(1, 1, 1, 'success', 'v1.2.0', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 30 MINUTE),
(2, 1, 2, 'success', 'v1.2.0', DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 5 MINUTE, DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 35 MINUTE);

-- 自愈规则数据
INSERT IGNORE INTO `heal_rules` (`id`, `name`, `description`, `trigger_condition`, `action`, `enabled`) VALUES
(1, '服务重启', '当服务异常退出时自动重启', '{"type": "service_crash"}', '{"type": "restart_service"}', 1),
(2, '磁盘清理', '当磁盘使用率超过阈值时清理日志', '{"type": "disk_high", "threshold": 90}', '{"type": "cleanup_logs"}', 1);

-- 自愈记录数据
INSERT IGNORE INTO `heal_records` (`id`, `rule_id`, `server_id`, `trigger_type`, `action`, `status`, `started_at`, `completed_at`) VALUES
(1, 1, 1, 'service_crash', 'restart_service', 'success', DATE_SUB(NOW(), INTERVAL 3 HOUR), DATE_SUB(NOW(), INTERVAL 3 HOUR) + INTERVAL 1 MINUTE),
(2, 2, 2, 'disk_high', 'cleanup_logs', 'success', DATE_SUB(NOW(), INTERVAL 5 HOUR), DATE_SUB(NOW(), INTERVAL 5 HOUR) + INTERVAL 5 MINUTE);

-- 安全规则数据
INSERT IGNORE INTO `security_rules` (`id`, `name`, `type`, `rule`, `action`, `enabled`) VALUES
(1, '防暴力破解', 'login', '{"max_attempts": 5, "window": 300}', 'block_ip', 1),
(2, 'SQL注入防护', 'request', '{"patterns": ["SELECT.*FROM", "UNION.*SELECT"]}', 'block_request', 1);

-- IP黑名单数据
INSERT IGNORE INTO `ip_blacklist` (`id`, `ip`, `reason`, `expires_at`, `created_by`) VALUES
(1, '192.168.100.100', '暴力破解尝试', DATE_ADD(NOW(), INTERVAL 24 HOUR), 1),
(2, '10.0.0.50', '恶意请求', DATE_ADD(NOW(), INTERVAL 1 HOUR), 1);

-- IP白名单数据
INSERT IGNORE INTO `ip_whitelist` (`id`, `ip`, `description`, `created_by`) VALUES
(1, '10.0.0.0/8', '内网IP段', 1),
(2, '192.168.0.0/16', '内网IP段', 1);

-- 登录记录数据
INSERT IGNORE INTO `login_records` (`id`, `user_id`, `username`, `ip`, `status`, `user_agent`) VALUES
(1, 1, 'admin', '192.168.1.50', 'success', 'Mozilla/5.0 Chrome/120.0'),
(2, 1, 'admin', '192.168.100.100', 'failed', 'Mozilla/5.0 Chrome/120.0');

-- 预测结果数据
INSERT IGNORE INTO `prediction_results` (`id`, `server_id`, `metric_type`, `predicted_value`, `confidence`, `predicted_at`) VALUES
(1, 1, 'cpu_usage', 75.5, 0.85, DATE_ADD(NOW(), INTERVAL 1 HOUR)),
(2, 1, 'memory_usage', 82.3, 0.78, DATE_ADD(NOW(), INTERVAL 1 HOUR));

-- AI决策数据
INSERT IGNORE INTO `ai_decisions` (`id`, `server_id`, `decision_type`, `input_data`, `output_data`, `confidence`, `executed`) VALUES
(1, 1, 'auto_scale', '{"cpu_usage": 85, "memory_usage": 70}', '{"action": "scale_up", "replicas": 3}', 0.92, 1),
(2, 2, 'alert_suppress', '{"alert_type": "cpu_high", "duration": 30}', '{"action": "suppress", "reason": "scheduled_maintenance"}', 0.88, 0);

-- 审计日志数据
INSERT IGNORE INTO `audit_logs` (`id`, `user_id`, `username`, `action`, `resource`, `resource_id`, `ip_address`, `status`) VALUES
(1, 1, 'admin', 'create', 'server', '1', '192.168.1.50', 'success'),
(2, 1, 'admin', 'update', 'server', '1', '192.168.1.50', 'success'),
(3, 2, 'operator', 'delete', 'server', '2', '192.168.1.51', 'failed');

-- 通知记录数据
INSERT IGNORE INTO `notification_records` (`id`, `type`, `title`, `content`, `recipient`, `status`) VALUES
(1, 'email', '服务器告警', '生产服务器-01 CPU使用率超过阈值', 'admin@yunwei.com', 'sent'),
(2, 'webhook', '部署完成', 'Web应用部署计划已完成', 'https://hooks.example.com/notify', 'sent');

-- 巡检记录数据
INSERT IGNORE INTO `patrol_records` (`id`, `server_id`, `type`, `status`, `result`, `started_at`, `completed_at`) VALUES
(1, 1, 'daily', 'completed', '{"cpu": "ok", "memory": "ok", "disk": "warning"}', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 5 MINUTE),
(2, 2, 'weekly', 'completed', '{"cpu": "ok", "memory": "ok", "disk": "ok"}', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 10 MINUTE);

-- 检测规则数据
INSERT IGNORE INTO `detect_rules` (`id`, `name`, `type`, `condition`, `severity`, `enabled`) VALUES
(1, 'CPU持续高负载', 'metric', '{"metric": "cpu_usage", "operator": ">", "threshold": 80, "duration": 300}', 'high', 1),
(2, '服务端口不可达', 'network', '{"port": 8080, "timeout": 5}', 'critical', 1);

-- 执行记录数据
INSERT IGNORE INTO `execution_records` (`id`, `task_id`, `server_id`, `status`, `output`, `started_at`, `completed_at`) VALUES
(1, 1, 1, 'success', '备份完成: backup_20240115.sql', DATE_SUB(NOW(), INTERVAL 12 HOUR), DATE_SUB(NOW(), INTERVAL 12 HOUR) + INTERVAL 120 SECOND),
(2, 2, 2, 'success', '清理完成: 删除 50 个文件', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 30 SECOND);

-- 自动化操作记录
INSERT IGNORE INTO `auto_actions` (`id`, `server_id`, `type`, `trigger`, `action`, `status`, `executed_at`) VALUES
(1, 1, 'scale', 'cpu_high', '{"replicas": 5}', 'success', DATE_SUB(NOW(), INTERVAL 2 HOUR)),
(2, 2, 'restart', 'service_down', '{"service": "nginx"}', 'success', DATE_SUB(NOW(), INTERVAL 4 HOUR));

-- 异常检测数据
INSERT IGNORE INTO `anomaly_detections` (`id`, `server_id`, `metric_type`, `anomaly_score`, `detected_at`, `confirmed`) VALUES
(1, 1, 'cpu_usage', 0.92, DATE_SUB(NOW(), INTERVAL 1 HOUR), 1),
(2, 2, 'network_latency', 0.78, DATE_SUB(NOW(), INTERVAL 3 HOUR), 0);

-- 自动扩容建议数据
INSERT IGNORE INTO `auto_scale_recommendations` (`id`, `server_id`, `current_replicas`, `recommended_replicas`, `reason`, `confidence`) VALUES
(1, 1, 3, 5, 'CPU使用率持续高于阈值', 0.85),
(2, 2, 2, 3, '请求量增长趋势', 0.72);

-- 服务健康数据
INSERT IGNORE INTO `service_health` (`id`, `server_id`, `service_name`, `status`, `last_check`) VALUES
(1, 1, 'nginx', 'healthy', NOW()),
(2, 1, 'mysql', 'healthy', NOW()),
(3, 2, 'redis', 'healthy', NOW());

-- 项目分析数据
INSERT IGNORE INTO `project_analyses` (`id`, `project_name`, `project_type`, `framework`, `language`, `recommendations`) VALUES
(1, 'yunwei-web', 'web', 'Vue3', 'TypeScript', '{"build": "vite", "deploy": "docker"}'),
(2, 'yunwei-api', 'api', 'Gin', 'Go', '{"build": "go build", "deploy": "docker"}');

-- 服务器能力数据
INSERT IGNORE INTO `server_capabilities` (`id`, `server_id`, `cpu_available`, `memory_available`, `disk_available`, `services`) VALUES
(1, 1, 4, 8192, 500, '["nginx", "mysql", "redis"]'),
(2, 2, 2, 4096, 200, '["nginx", "redis"]');

-- 资源池数据
INSERT IGNORE INTO `resource_pools` (`id`, `name`, `type`, `total_cpu`, `total_memory`, `used_cpu`, `used_memory`) VALUES
(1, '生产资源池', 'production', 32, 65536, 20, 45000),
(2, '测试资源池', 'test', 16, 32768, 8, 15000);

-- 金丝雀发布数据
INSERT IGNORE INTO `canary_releases` (`id`, `cluster_id`, `namespace`, `service_name`, `current_version`, `new_version`, `canary_weight`, `status`, `strategy`) VALUES
(1, 1, 'default', 'api-service', 'v1.0', 'v2.0', 10, 'running', 'progressive'),
(2, 1, 'production', 'web-service', 'v1.5', 'v1.6', 5, 'paused', 'progressive');

-- 金丝雀配置数据
INSERT IGNORE INTO `canary_configs` (`id`, `cluster_id`, `namespace`, `service_name`, `strategy`, `step_duration`, `error_rate_threshold`, `auto_promote`, `auto_rollback`, `enabled`) VALUES
(1, 1, 'default', 'api-service', 'progressive', 300, 1, 1, 1, 1),
(2, 1, 'production', 'web-service', 'progressive', 600, 5, 0, 1, 1);

-- 备份记录数据
INSERT IGNORE INTO `backup_records` (`id`, `server_id`, `type`, `path`, `size`, `status`, `started_at`, `completed_at`) VALUES
(1, 1, 'database', '/backup/mysql/backup_20240115.sql', 524288000, 'completed', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 300 SECOND),
(2, 1, 'file', '/backup/files/backup_20240114.tar.gz', 1073741824, 'completed', DATE_SUB(NOW(), INTERVAL 2 DAY), DATE_SUB(NOW(), INTERVAL 2 DAY) + INTERVAL 600 SECOND);

-- 证书续期记录
INSERT IGNORE INTO `cert_renewal_records` (`id`, `cert_id`, `status`, `old_not_after`, `new_not_after`, `started_at`, `completed_at`) VALUES
(1, 2, 'success', DATE_ADD(NOW(), INTERVAL 30 DAY), DATE_ADD(NOW(), INTERVAL 90 DAY), NOW(), NOW() + INTERVAL 60 SECOND);

-- 工作流记录
INSERT IGNORE INTO `workflow_records` (`id`, `type`, `status`, `trigger_source`, `started_at`, `completed_at`) VALUES
(1, 'deploy', 'completed', 'manual', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY) + INTERVAL 30 MINUTE),
(2, 'backup', 'completed', 'scheduled', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 10 MINUTE);

-- 完成提示
SELECT '模拟数据初始化完成（无重复数据版本）!' AS message;
