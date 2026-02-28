-- 补充缺失的数据库表
-- 包含 seed_data.sql 中引用但在 complete_tables.sql 中缺失的表

USE `yunwei`;

-- ==================== 证书管理 ====================

CREATE TABLE IF NOT EXISTS `certificates` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '证书名称',
  `domain` varchar(256) DEFAULT NULL COMMENT '域名',
  `sans` text COMMENT 'SANs(JSON)',
  `provider` varchar(16) DEFAULT NULL COMMENT '提供商',
  `cert_path` varchar(256) DEFAULT NULL COMMENT '证书路径',
  `key_path` varchar(256) DEFAULT NULL COMMENT '密钥路径',
  `chain_path` varchar(256) DEFAULT NULL COMMENT '链路径',
  `full_chain_path` varchar(256) DEFAULT NULL COMMENT '完整链路径',
  `serial_number` varchar(64) DEFAULT NULL COMMENT '序列号',
  `issuer` varchar(128) DEFAULT NULL COMMENT '颁发者',
  `not_before` datetime DEFAULT NULL COMMENT '生效时间',
  `not_after` datetime DEFAULT NULL COMMENT '过期时间',
  `days_left` int DEFAULT NULL COMMENT '剩余天数',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `auto_renew` tinyint DEFAULT 0 COMMENT '自动续期',
  `renew_before` int DEFAULT 30 COMMENT '提前续期天数',
  `renewal_count` int DEFAULT 0 COMMENT '续期次数',
  `last_renew_at` datetime DEFAULT NULL COMMENT '最后续期时间',
  `next_renew_at` datetime DEFAULT NULL COMMENT '下次续期时间',
  `acme_email` varchar(64) DEFAULT NULL COMMENT 'ACME邮箱',
  `acme_server` varchar(256) DEFAULT NULL COMMENT 'ACME服务器',
  `dns_provider` varchar(32) DEFAULT NULL COMMENT 'DNS提供商',
  `dns_credentials` text COMMENT 'DNS凭证',
  `deploy_target` varchar(32) DEFAULT NULL COMMENT '部署目标',
  `deploy_config` text COMMENT '部署配置(JSON)',
  `last_deploy_at` datetime DEFAULT NULL COMMENT '最后部署时间',
  PRIMARY KEY (`id`),
  KEY `idx_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证书表';

CREATE TABLE IF NOT EXISTS `cert_renewal_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `cert_id` bigint unsigned DEFAULT NULL COMMENT '证书ID',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `old_serial_number` varchar(64) DEFAULT NULL COMMENT '旧序列号',
  `old_not_after` datetime DEFAULT NULL COMMENT '旧过期时间',
  `new_serial_number` varchar(64) DEFAULT NULL COMMENT '新序列号',
  `new_not_after` datetime DEFAULT NULL COMMENT '新过期时间',
  `method` varchar(32) DEFAULT NULL COMMENT '方法',
  `commands` text COMMENT '命令',
  `execution_log` text COMMENT '执行日志',
  `error_message` text COMMENT '错误信息',
  `ai_decision` text COMMENT 'AI决策',
  `ai_confidence` double DEFAULT NULL COMMENT 'AI置信度',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  PRIMARY KEY (`id`),
  KEY `idx_cert_id` (`cert_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证书续期记录表';

-- ==================== CDN管理 ====================

CREATE TABLE IF NOT EXISTS `cdn_domains` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `domain` varchar(256) DEFAULT NULL COMMENT '域名',
  `origin` varchar(256) DEFAULT NULL COMMENT '源站',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`),
  KEY `idx_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CDN域名表';

CREATE TABLE IF NOT EXISTS `cdn_nodes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `domain_id` bigint unsigned DEFAULT NULL COMMENT '域名ID',
  `name` varchar(64) DEFAULT NULL COMMENT '节点名称',
  `ip` varchar(64) DEFAULT NULL COMMENT 'IP地址',
  `region` varchar(64) DEFAULT NULL COMMENT '区域',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  PRIMARY KEY (`id`),
  KEY `idx_domain_id` (`domain_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CDN节点表';

CREATE TABLE IF NOT EXISTS `cdn_cache_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `domain_id` bigint unsigned DEFAULT NULL COMMENT '域名ID',
  `path_pattern` varchar(256) DEFAULT NULL COMMENT '路径模式',
  `cache_ttl` int DEFAULT NULL COMMENT '缓存时间(秒)',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`),
  KEY `idx_domain_id` (`domain_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CDN缓存规则表';

CREATE TABLE IF NOT EXISTS `cdn_optimization_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `domain_id` bigint unsigned DEFAULT NULL COMMENT '域名ID',
  `type` varchar(32) DEFAULT NULL COMMENT '优化类型',
  `before_value` text COMMENT '优化前值',
  `after_value` text COMMENT '优化后值',
  `improvement` double DEFAULT NULL COMMENT '提升比例',
  PRIMARY KEY (`id`),
  KEY `idx_domain_id` (`domain_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CDN优化记录表';

-- ==================== 负载均衡 ====================

CREATE TABLE IF NOT EXISTS `load_balancers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '名称',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `address` varchar(64) DEFAULT NULL COMMENT '地址',
  `port` int DEFAULT NULL COMMENT '端口',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `algorithm` varchar(32) DEFAULT NULL COMMENT '算法',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='负载均衡器表';

CREATE TABLE IF NOT EXISTS `lb_backend_servers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `lb_id` bigint unsigned DEFAULT NULL COMMENT 'LB ID',
  `server_ip` varchar(64) DEFAULT NULL COMMENT '服务器IP',
  `server_port` int DEFAULT NULL COMMENT '服务器端口',
  `weight` int DEFAULT 1 COMMENT '权重',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `health_check_url` varchar(256) DEFAULT NULL COMMENT '健康检查URL',
  PRIMARY KEY (`id`),
  KEY `idx_lb_id` (`lb_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='LB后端服务器表';

CREATE TABLE IF NOT EXISTS `lb_optimization_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `lb_id` bigint unsigned DEFAULT NULL COMMENT 'LB ID',
  `type` varchar(32) DEFAULT NULL COMMENT '优化类型',
  `before_value` text COMMENT '优化前值',
  `after_value` text COMMENT '优化后值',
  `improvement` double DEFAULT NULL COMMENT '提升比例',
  PRIMARY KEY (`id`),
  KEY `idx_lb_id` (`lb_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='LB优化记录表';

CREATE TABLE IF NOT EXISTS `lb_algorithm_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `lb_id` bigint unsigned DEFAULT NULL COMMENT 'LB ID',
  `algorithm` varchar(32) DEFAULT NULL COMMENT '算法',
  `params` text COMMENT '参数(JSON)',
  PRIMARY KEY (`id`),
  KEY `idx_lb_id` (`lb_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='LB算法配置表';

-- ==================== 金丝雀发布 ====================

CREATE TABLE IF NOT EXISTS `canary_releases` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cluster_id` bigint unsigned DEFAULT NULL COMMENT '集群ID',
  `namespace` varchar(64) DEFAULT NULL COMMENT '命名空间',
  `service_name` varchar(128) DEFAULT NULL COMMENT '服务名称',
  `strategy` varchar(16) DEFAULT NULL COMMENT '策略',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `current_version` varchar(64) DEFAULT NULL COMMENT '当前版本',
  `new_version` varchar(64) DEFAULT NULL COMMENT '新版本',
  `new_image` varchar(256) DEFAULT NULL COMMENT '新镜像',
  `total_replicas` int DEFAULT NULL COMMENT '总副本数',
  `canary_replicas` int DEFAULT NULL COMMENT '金丝雀副本数',
  `canary_weight` double DEFAULT NULL COMMENT '金丝雀权重',
  `weight_step` double DEFAULT NULL COMMENT '权重步长',
  `current_step` int DEFAULT NULL COMMENT '当前步骤',
  `total_steps` int DEFAULT NULL COMMENT '总步骤',
  `error_rate_threshold` double DEFAULT NULL COMMENT '错误率阈值',
  `latency_threshold` bigint DEFAULT NULL COMMENT '延迟阈值',
  `success_rate_threshold` double DEFAULT NULL COMMENT '成功率阈值',
  `current_error_rate` double DEFAULT NULL COMMENT '当前错误率',
  `current_latency` bigint DEFAULT NULL COMMENT '当前延迟',
  `current_success_rate` double DEFAULT NULL COMMENT '当前成功率',
  `ai_decision` text COMMENT 'AI决策',
  `ai_confidence` double DEFAULT NULL COMMENT 'AI置信度',
  `ai_auto_promote` tinyint DEFAULT NULL COMMENT 'AI自动推进',
  `rollback_reason` text COMMENT '回滚原因',
  `rollback_at` datetime DEFAULT NULL COMMENT '回滚时间',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(秒)',
  PRIMARY KEY (`id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='金丝雀发布表';

CREATE TABLE IF NOT EXISTS `canary_steps` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `release_id` bigint unsigned DEFAULT NULL COMMENT '发布ID',
  `step_num` int DEFAULT NULL COMMENT '步骤编号',
  `weight` double DEFAULT NULL COMMENT '权重',
  `replicas` int DEFAULT NULL COMMENT '副本数',
  `error_rate` double DEFAULT NULL COMMENT '错误率',
  `latency` bigint DEFAULT NULL COMMENT '延迟',
  `success_rate` double DEFAULT NULL COMMENT '成功率',
  `request_count` bigint DEFAULT NULL COMMENT '请求数',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(秒)',
  `passed_checks` int DEFAULT NULL COMMENT '通过检查数',
  `failed_checks` int DEFAULT NULL COMMENT '失败检查数',
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='金丝雀步骤表';

CREATE TABLE IF NOT EXISTS `canary_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cluster_id` bigint unsigned DEFAULT NULL COMMENT '集群ID',
  `namespace` varchar(64) DEFAULT NULL COMMENT '命名空间',
  `service_name` varchar(128) DEFAULT NULL COMMENT '服务名称',
  `strategy` varchar(16) DEFAULT NULL COMMENT '策略',
  `weight_step` double DEFAULT NULL COMMENT '权重步长',
  `total_steps` int DEFAULT NULL COMMENT '总步骤',
  `step_duration` int DEFAULT NULL COMMENT '步骤持续时间(秒)',
  `error_rate_threshold` double DEFAULT NULL COMMENT '错误率阈值',
  `latency_threshold` bigint DEFAULT NULL COMMENT '延迟阈值',
  `success_rate_threshold` double DEFAULT NULL COMMENT '成功率阈值',
  `auto_promote` tinyint DEFAULT NULL COMMENT '自动推进',
  `auto_rollback` tinyint DEFAULT NULL COMMENT '自动回滚',
  `require_manual` tinyint DEFAULT NULL COMMENT '需要人工确认',
  `enabled` tinyint DEFAULT NULL COMMENT '是否启用',
  PRIMARY KEY (`id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='金丝雀配置表';

-- ==================== 部署管理 ====================

CREATE TABLE IF NOT EXISTS `deploy_plans` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(128) DEFAULT NULL COMMENT '名称',
  `description` text COMMENT '描述',
  `plan_type` varchar(16) DEFAULT NULL COMMENT '方案类型',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `project_analysis_id` bigint unsigned DEFAULT NULL COMMENT '项目分析ID',
  `server_assignments` text COMMENT '服务器分配(JSON)',
  `services` text COMMENT '服务配置(JSON)',
  `service_topology` text COMMENT '服务拓扑(JSON)',
  `network_config` text COMMENT '网络配置(JSON)',
  `load_balancer` text COMMENT '负载均衡配置(JSON)',
  `storage_config` text COMMENT '存储配置(JSON)',
  `database_config` text COMMENT '数据库配置(JSON)',
  `cache_config` text COMMENT '缓存配置(JSON)',
  `mq_config` text COMMENT '消息队列配置(JSON)',
  `environment_vars` text COMMENT '环境变量(JSON)',
  `deploy_order` text COMMENT '部署顺序(JSON)',
  `estimated_cost` double DEFAULT NULL COMMENT '预估成本',
  `ai_suggestion` text COMMENT 'AI建议',
  `confidence` double DEFAULT NULL COMMENT '置信度',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(秒)',
  `progress` int DEFAULT NULL COMMENT '进度',
  `rollback_plan` text COMMENT '回滚方案(JSON)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部署方案表';

CREATE TABLE IF NOT EXISTS `deploy_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `plan_id` bigint unsigned DEFAULT NULL COMMENT '方案ID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `version` varchar(64) DEFAULT NULL COMMENT '版本',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_plan_id` (`plan_id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部署任务表';

CREATE TABLE IF NOT EXISTS `deploy_task_steps` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `task_id` bigint unsigned DEFAULT NULL COMMENT '任务ID',
  `step_name` varchar(64) DEFAULT NULL COMMENT '步骤名称',
  `step_order` int DEFAULT NULL COMMENT '步骤顺序',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `output` text COMMENT '输出',
  `error` text COMMENT '错误',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部署任务步骤表';

-- ==================== 项目分析 ====================

CREATE TABLE IF NOT EXISTS `project_analyses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `project_name` varchar(128) DEFAULT NULL COMMENT '项目名称',
  `project_type` varchar(32) DEFAULT NULL COMMENT '项目类型',
  `framework` varchar(64) DEFAULT NULL COMMENT '框架',
  `language` varchar(32) DEFAULT NULL COMMENT '语言',
  `recommendations` text COMMENT '推荐配置(JSON)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目分析表';

CREATE TABLE IF NOT EXISTS `server_capabilities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `cpu_available` int DEFAULT NULL COMMENT '可用CPU',
  `memory_available` int DEFAULT NULL COMMENT '可用内存(MB)',
  `disk_available` int DEFAULT NULL COMMENT '可用磁盘(GB)',
  `services` text COMMENT '服务列表(JSON)',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器能力表';

CREATE TABLE IF NOT EXISTS `resource_pools` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '名称',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `total_cpu` int DEFAULT NULL COMMENT '总CPU',
  `total_memory` int DEFAULT NULL COMMENT '总内存(MB)',
  `used_cpu` int DEFAULT NULL COMMENT '已用CPU',
  `used_memory` int DEFAULT NULL COMMENT '已用内存(MB)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源池表';

-- ==================== 自愈管理 ====================

CREATE TABLE IF NOT EXISTS `heal_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '规则名称',
  `type` varchar(32) DEFAULT NULL COMMENT '规则类型',
  `condition` text COMMENT '触发条件(JSON)',
  `action` varchar(32) DEFAULT NULL COMMENT '动作',
  `severity` varchar(16) DEFAULT NULL COMMENT '严重级别',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自愈规则表';

CREATE TABLE IF NOT EXISTS `heal_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `rule_id` bigint unsigned DEFAULT NULL COMMENT '规则ID',
  `trigger_type` varchar(32) DEFAULT NULL COMMENT '触发类型',
  `action` varchar(32) DEFAULT NULL COMMENT '动作',
  `command` text COMMENT '命令',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `output` text COMMENT '输出',
  `error` text COMMENT '错误',
  `success` tinyint DEFAULT NULL COMMENT '是否成功',
  `retry_count` int DEFAULT NULL COMMENT '重试次数',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_rule_id` (`rule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自愈记录表';

CREATE TABLE IF NOT EXISTS `service_health` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `service_name` varchar(64) DEFAULT NULL COMMENT '服务名称',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `last_check` datetime DEFAULT NULL COMMENT '最后检查时间',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务健康表';

-- ==================== 安全管理 ====================

CREATE TABLE IF NOT EXISTS `security_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `server_name` varchar(64) DEFAULT NULL COMMENT '服务器名称',
  `event_type` varchar(32) DEFAULT NULL COMMENT '事件类型',
  `level` varchar(16) DEFAULT NULL COMMENT '级别',
  `source_ip` varchar(45) DEFAULT NULL COMMENT '来源IP',
  `source_port` int DEFAULT NULL COMMENT '来源端口',
  `geo_location` varchar(128) DEFAULT NULL COMMENT '地理位置',
  `target_user` varchar(64) DEFAULT NULL COMMENT '目标用户',
  `target_port` int DEFAULT NULL COMMENT '目标端口',
  `target_service` varchar(64) DEFAULT NULL COMMENT '目标服务',
  `description` text COMMENT '描述',
  `raw_log` text COMMENT '原始日志',
  `status` varchar(16) DEFAULT 'new' COMMENT '状态',
  `handled_by` bigint unsigned DEFAULT NULL COMMENT '处理人',
  `handled_at` datetime DEFAULT NULL COMMENT '处理时间',
  `action` varchar(64) DEFAULT NULL COMMENT '动作',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_source_ip` (`source_ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='安全事件表';

CREATE TABLE IF NOT EXISTS `ip_blacklist` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `ip` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `cidr` varchar(64) DEFAULT NULL COMMENT 'CIDR',
  `reason` varchar(255) DEFAULT NULL COMMENT '原因',
  `event_type` varchar(32) DEFAULT NULL COMMENT '事件类型',
  `event_id` bigint unsigned DEFAULT NULL COMMENT '事件ID',
  `auto_banned` tinyint DEFAULT 0 COMMENT '自动封禁',
  `banned_by` bigint unsigned DEFAULT NULL COMMENT '封禁人',
  `permanent` tinyint DEFAULT 0 COMMENT '永久封禁',
  `expires_at` datetime DEFAULT NULL COMMENT '过期时间',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  `ban_count` int DEFAULT 0 COMMENT '封禁次数',
  `attack_count` int DEFAULT 0 COMMENT '攻击次数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP黑名单表';

CREATE TABLE IF NOT EXISTS `ip_whitelist` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `ip` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `cidr` varchar(64) DEFAULT NULL COMMENT 'CIDR',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `added_by` bigint unsigned DEFAULT NULL COMMENT '添加人',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP白名单表';

CREATE TABLE IF NOT EXISTS `login_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `username` varchar(64) DEFAULT NULL COMMENT '用户名',
  `ip` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `user_agent` varchar(500) DEFAULT NULL COMMENT '用户代理',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录记录表';

CREATE TABLE IF NOT EXISTS `security_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '规则名称',
  `type` varchar(32) DEFAULT NULL COMMENT '规则类型',
  `rule` text COMMENT '规则(JSON)',
  `action` varchar(32) DEFAULT NULL COMMENT '动作',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='安全规则表';

CREATE TABLE IF NOT EXISTS `audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `username` varchar(64) DEFAULT NULL COMMENT '用户名',
  `action` varchar(100) DEFAULT NULL COMMENT '操作',
  `resource` varchar(100) DEFAULT NULL COMMENT '资源',
  `resource_id` varchar(36) DEFAULT NULL COMMENT '资源ID',
  `resource_name` varchar(255) DEFAULT NULL COMMENT '资源名称',
  `old_value` text COMMENT '旧值(JSON)',
  `new_value` text COMMENT '新值(JSON)',
  `changes` text COMMENT '变更(JSON)',
  `ip_address` varchar(50) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) DEFAULT NULL COMMENT '用户代理',
  `request_id` varchar(36) DEFAULT NULL COMMENT '请求ID',
  `status` varchar(20) DEFAULT 'success' COMMENT '状态',
  `error_msg` text COMMENT '错误消息',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_resource_id` (`resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审计日志表';

-- ==================== AI决策 ====================

CREATE TABLE IF NOT EXISTS `ai_decisions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `decision_type` varchar(32) DEFAULT NULL COMMENT '决策类型',
  `input_data` text COMMENT '输入数据(JSON)',
  `output_data` text COMMENT '输出数据(JSON)',
  `confidence` double DEFAULT NULL COMMENT '置信度',
  `executed` tinyint DEFAULT 0 COMMENT '是否执行',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI决策表';

CREATE TABLE IF NOT EXISTS `prediction_results` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `metric_type` varchar(32) DEFAULT NULL COMMENT '指标类型',
  `predicted_value` double DEFAULT NULL COMMENT '预测值',
  `confidence` double DEFAULT NULL COMMENT '置信度',
  `predicted_at` datetime DEFAULT NULL COMMENT '预测时间',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预测结果表';

CREATE TABLE IF NOT EXISTS `anomaly_detections` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `metric_type` varchar(32) DEFAULT NULL COMMENT '指标类型',
  `anomaly_score` double DEFAULT NULL COMMENT '异常分数',
  `detected_at` datetime DEFAULT NULL COMMENT '检测时间',
  `confirmed` tinyint DEFAULT 0 COMMENT '是否确认',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异常检测表';

CREATE TABLE IF NOT EXISTS `auto_scale_recommendations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `current_replicas` int DEFAULT NULL COMMENT '当前副本数',
  `recommended_replicas` int DEFAULT NULL COMMENT '推荐副本数',
  `reason` varchar(255) DEFAULT NULL COMMENT '原因',
  `confidence` double DEFAULT NULL COMMENT '置信度',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动扩容推荐表';

-- ==================== 自动操作 ====================

CREATE TABLE IF NOT EXISTS `auto_actions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `trigger` varchar(64) DEFAULT NULL COMMENT '触发条件',
  `action` text COMMENT '动作(JSON)',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `executed_at` datetime DEFAULT NULL COMMENT '执行时间',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动操作表';

-- ==================== 检测规则 ====================

CREATE TABLE IF NOT EXISTS `detect_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '规则名称',
  `type` varchar(32) DEFAULT NULL COMMENT '规则类型',
  `condition` text COMMENT '条件(JSON)',
  `severity` varchar(16) DEFAULT NULL COMMENT '严重级别',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='检测规则表';

-- ==================== 执行记录 ====================

CREATE TABLE IF NOT EXISTS `execution_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `task_id` bigint unsigned DEFAULT NULL COMMENT '任务ID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `output` text COMMENT '输出',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='执行记录表';

-- ==================== 工作流 ====================

CREATE TABLE IF NOT EXISTS `workflow_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `trigger_source` varchar(64) DEFAULT NULL COMMENT '触发源',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `steps` text COMMENT '步骤(JSON)',
  `result` text COMMENT '结果',
  `ai_analysis` text COMMENT 'AI分析',
  `commands` text COMMENT '命令',
  `output` text COMMENT '输出',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `auto_mode` tinyint DEFAULT 0 COMMENT '自动模式',
  `need_approve` tinyint DEFAULT 0 COMMENT '需要审批',
  `approved_by` bigint unsigned DEFAULT NULL COMMENT '审批人',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作流记录表';

-- ==================== 通知记录 ====================

CREATE TABLE IF NOT EXISTS `notification_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `title` varchar(255) DEFAULT NULL COMMENT '标题',
  `content` text COMMENT '内容',
  `recipient` varchar(255) DEFAULT NULL COMMENT '接收者',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `sent_at` datetime DEFAULT NULL COMMENT '发送时间',
  `error` text COMMENT '错误信息',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';

CREATE TABLE IF NOT EXISTS `notify_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `title` varchar(255) DEFAULT NULL COMMENT '标题',
  `content` text COMMENT '内容',
  `recipient` varchar(255) DEFAULT NULL COMMENT '接收者',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `sent_at` datetime DEFAULT NULL COMMENT '发送时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';

-- ==================== 巡检记录 ====================

CREATE TABLE IF NOT EXISTS `patrol_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `result` text COMMENT '结果(JSON)',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检记录表';

-- ==================== 检查报告 ====================

CREATE TABLE IF NOT EXISTS `inspection_reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `report` text COMMENT '报告(JSON)',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='检查报告表';

-- ==================== 备份记录 ====================

CREATE TABLE IF NOT EXISTS `backup_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `path` varchar(255) DEFAULT NULL COMMENT '路径',
  `size` bigint DEFAULT NULL COMMENT '大小(字节)',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `error` text COMMENT '错误信息',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='备份记录表';

-- 完成提示
SELECT '补充表创建完成!' AS message;
