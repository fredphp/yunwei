-- 模拟数据初始化脚本（精简版 - 只使用核心字段）
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

-- ==================== Agent管理 ====================

-- Agent版本数据
INSERT IGNORE INTO `agent_versions` (`id`, `version`, `version_code`, `build_time`, `build_commit`, `build_branch`, `file_url`, `file_md5`, `file_size`, `platform`, `arch`, `changelog`, `release_type`, `enabled`, `is_latest`) VALUES
(1, '1.0.0', 100, NOW(), 'a1b2c3d4', 'main', 'https://downloads.yunwei.com/agent/v1.0.0/agent-linux-amd64', 'abc123def456', 15728640, 'linux', 'amd64', '初始版本发布', 'stable', 1, 0),
(2, '1.1.0', 110, NOW(), 'e5f6g7h8', 'main', 'https://downloads.yunwei.com/agent/v1.1.0/agent-linux-amd64', 'def456ghi789', 16777216, 'linux', 'amd64', '新增自动恢复功能', 'stable', 1, 1);

-- Agent实例数据（只使用核心字段）
INSERT IGNORE INTO `agents` (`id`, `server_id`, `server_name`, `server_ip`, `agent_id`, `version`, `version_code`, `platform`, `arch`, `os`, `status`, `auto_upgrade`) VALUES
(1, 1, '生产服务器-01', '192.168.1.100', 'agent-001-abc', '1.1.0', 110, 'linux', 'amd64', 'CentOS 7.9', 'online', 1),
(2, 2, '测试服务器-01', '192.168.1.101', 'agent-002-def', '1.0.0', 100, 'linux', 'amd64', 'Ubuntu 20.04', 'online', 1);

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
(2, 'high-priority', '高优先级队列', 20, 500, 10, 1800, 5, 1);

-- 定时任务数据
INSERT IGNORE INTO `scheduler_cron_jobs` (`id`, `name`, `description`, `cron_expr`, `timezone`, `enabled`, `concurrent_policy`) VALUES
(1, '每日备份', '每天凌晨2点执行备份', '0 2 * * *', 'Asia/Shanghai', 1, 'forbid'),
(2, '清理日志', '每小时清理过期日志', '0 * * * *', 'Asia/Shanghai', 1, 'allow');

-- 任务模板数据
INSERT IGNORE INTO `scheduler_task_templates` (`id`, `name`, `category`, `description`, `task_def`, `enabled`) VALUES
(1, '服务器备份', 'backup', '备份服务器数据', '{"type": "backup", "timeout": 3600}', 1),
(2, '日志清理', 'cleanup', '清理过期日志文件', '{"type": "cleanup", "timeout": 600}', 1);

-- ==================== 租户管理 ====================

-- 租户数据
INSERT IGNORE INTO `tenants` (`id`, `name`, `slug`, `domain`, `description`, `status`, `plan`, `contact_name`, `contact_email`) VALUES
('demo-tenant-001', '示例公司', 'demo-company', 'demo.yunwei.com', '示例租户公司', 'active', 'pro', '张三', 'zhangsan@demo.com'),
('demo-tenant-002', '测试企业', 'test-enterprise', 'test.yunwei.com', '测试用企业租户', 'active', 'starter', '李四', 'lisi@test.com');

-- 租户角色数据
INSERT IGNORE INTO `tenant_roles` (`id`, `tenant_id`, `name`, `slug`, `description`, `is_system`, `scope`) VALUES
('role-001', 'demo-tenant-001', '管理员', 'admin', '租户管理员', 1, 'tenant'),
('role-002', 'demo-tenant-001', '运维人员', 'operator', '运维操作人员', 1, 'tenant');

-- ==================== 高可用 ====================

-- 集群节点数据（只使用核心字段）
INSERT IGNORE INTO `cluster_nodes` (`id`, `node_id`, `node_name`, `hostname`, `internal_ip`, `api_port`, `status`, `role`, `is_leader`, `version`, `weight`, `data_center`, `zone`, `enabled`) VALUES
(1, 'node-001', '主节点-北京', 'yunwei-node-01', '10.0.1.10', 8080, 'online', 'leader', 1, '1.0.0', 100, 'beijing', 'zone-a', 1),
(2, 'node-002', '从节点-北京', 'yunwei-node-02', '10.0.1.11', 8080, 'online', 'follower', 0, '1.0.0', 100, 'beijing', 'zone-a', 1),
(3, 'node-003', '从节点-上海', 'yunwei-node-03', '10.0.2.10', 8080, 'online', 'follower', 0, '1.0.0', 80, 'shanghai', 'zone-b', 1);

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

-- K8s扩容事件数据（只使用核心字段）
INSERT IGNORE INTO `k8s_scale_events` (`id`, `cluster_id`, `namespace`, `deployment`, `scale_type`, `status`, `replicas_before`, `replicas_after`, `replicas_target`, `trigger_reason`) VALUES
(1, 1, 'default', 'web-api', 'auto', 'success', 3, 5, 5, 'CPU使用率超过阈值'),
(2, 1, 'production', 'payment-service', 'manual', 'success', 2, 4, 4, '业务高峰期手动扩容');

-- K8s HPA配置数据
INSERT IGNORE INTO `k8s_hpa_configs` (`id`, `cluster_id`, `namespace`, `deployment`, `min_replicas`, `max_replicas`, `target_cpu_util`, `target_mem_util`, `enabled`) VALUES
(1, 1, 'default', 'web-api', 2, 10, 70, 75, 1),
(2, 1, 'production', 'payment-service', 3, 15, 65, 70, 1);

-- K8s Deployment状态数据
INSERT IGNORE INTO `k8s_deployment_status` (`id`, `cluster_id`, `namespace`, `deployment`, `replicas`, `ready_replicas`, `cpu_usage`, `memory_usage`, `hpa_enabled`) VALUES
(1, 1, 'default', 'web-api', 5, 5, 45.5, 62.3, 1),
(2, 1, 'production', 'payment-service', 4, 4, 38.2, 55.8, 1);

-- ==================== 告警管理 ====================

-- 告警规则数据
INSERT IGNORE INTO `alert_rules` (`id`, `name`, `description`, `metric`, `operator`, `threshold`, `duration`, `level`, `status`) VALUES
(1, 'CPU使用率告警', 'CPU使用率超过80%', 'cpu_usage', '>', 80, 60, 'warning', 1),
(2, 'CPU使用率严重告警', 'CPU使用率超过95%', 'cpu_usage', '>', 95, 30, 'critical', 1),
(3, '内存使用率告警', '内存使用率超过85%', 'memory_usage', '>', 85, 60, 'warning', 1),
(4, '磁盘使用率告警', '磁盘使用率超过90%', 'disk_usage', '>', 90, 60, 'critical', 1);

-- 告警记录数据
INSERT IGNORE INTO `alerts` (`id`, `rule_id`, `server_id`, `level`, `title`, `message`, `metric_value`, `status`, `fired_at`) VALUES
(1, 1, 1, 'warning', 'CPU使用率告警', '服务器 CPU使用率达到 85%', 85.5, 'resolved', DATE_SUB(NOW(), INTERVAL 2 HOUR)),
(2, 3, 2, 'warning', '内存使用率告警', '服务器 内存使用率达到 88%', 88.2, 'firing', DATE_SUB(NOW(), INTERVAL 30 MINUTE));

-- ==================== 证书管理 ====================

-- 证书数据
INSERT IGNORE INTO `certificates` (`id`, `name`, `domain`, `issuer`, `not_before`, `not_after`, `status`, `auto_renew`) VALUES
(1, '主域名证书', '*.yunwei.com', "Let's Encrypt", DATE_SUB(NOW(), INTERVAL 30 DAY), DATE_ADD(NOW(), INTERVAL 60 DAY), 'valid', 1),
(2, 'API证书', 'api.yunwei.com', "Let's Encrypt", DATE_SUB(NOW(), INTERVAL 60 DAY), DATE_ADD(NOW(), INTERVAL 30 DAY), 'expiring_soon', 1);

-- ==================== 负载均衡 ====================

-- 负载均衡器数据
INSERT IGNORE INTO `load_balancers` (`id`, `name`, `type`, `address`, `port`, `status`, `algorithm`, `enabled`) VALUES
(1, '生产LB', 'nginx', '10.0.1.100', 80, 'active', 'round-robin', 1),
(2, '测试LB', 'nginx', '10.0.2.100', 80, 'active', 'least-conn', 1);

-- CDN域名数据
INSERT IGNORE INTO `cdn_domains` (`id`, `domain`, `origin`, `type`, `status`, `enabled`) VALUES
(1, 'static.yunwei.com', 'https://origin.yunwei.com', 'web', 'active', 1),
(2, 'img.yunwei.com', 'https://img-origin.yunwei.com', 'image', 'active', 1);

-- 完成提示
SELECT '模拟数据初始化完成!' AS message;
