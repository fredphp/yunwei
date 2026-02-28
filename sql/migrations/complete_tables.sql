-- 完整数据库表迁移脚本
-- 创建所有缺失的表

USE `yunwei`;

-- ==================== 系统管理 ====================

-- 系统用户表
CREATE TABLE IF NOT EXISTS `sys_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码',
  `nick_name` varchar(64) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `email` varchar(128) DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1启用, 0禁用',
  `role_id` bigint unsigned DEFAULT NULL COMMENT '角色ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';

-- 系统角色表
CREATE TABLE IF NOT EXISTS `sys_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `name` varchar(64) NOT NULL COMMENT '角色名称',
  `keyword` varchar(64) NOT NULL COMMENT '角色关键字',
  `description` varchar(255) DEFAULT NULL COMMENT '角色描述',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1启用, 0禁用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_keyword` (`keyword`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色表';

-- 系统菜单表
CREATE TABLE IF NOT EXISTS `sys_menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `parent_id` bigint unsigned DEFAULT 0 COMMENT '父菜单ID',
  `title` varchar(64) NOT NULL COMMENT '菜单标题',
  `name` varchar(64) NOT NULL COMMENT '路由名称',
  `path` varchar(255) DEFAULT NULL COMMENT '路由路径',
  `component` varchar(255) DEFAULT NULL COMMENT '组件路径',
  `icon` varchar(64) DEFAULT NULL COMMENT '菜单图标',
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1启用, 0禁用',
  `hidden` tinyint DEFAULT 0 COMMENT '是否隐藏',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统菜单表';

-- 系统API表
CREATE TABLE IF NOT EXISTS `sys_apis` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `path` varchar(255) NOT NULL COMMENT 'API路径',
  `method` varchar(16) NOT NULL COMMENT '请求方法',
  `group` varchar(64) DEFAULT NULL COMMENT 'API分组',
  `description` varchar(255) DEFAULT NULL COMMENT 'API描述',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统API表';

-- 角色-API关联表
CREATE TABLE IF NOT EXISTS `sys_role_apis` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `api_id` bigint unsigned NOT NULL COMMENT 'API ID',
  PRIMARY KEY (`id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_api_id` (`api_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色-API关联表';

-- 角色-菜单关联表
CREATE TABLE IF NOT EXISTS `sys_role_menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色-菜单关联表';

-- ==================== Agent管理 ====================

-- Agent版本表
CREATE TABLE IF NOT EXISTS `agent_versions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `version` varchar(32) NOT NULL COMMENT '版本号',
  `version_code` int NOT NULL COMMENT '版本代码',
  `build_time` datetime DEFAULT NULL COMMENT '构建时间',
  `build_commit` varchar(64) DEFAULT NULL COMMENT 'Git Commit',
  `build_branch` varchar(64) DEFAULT NULL COMMENT 'Git Branch',
  `file_url` varchar(512) DEFAULT NULL COMMENT '下载地址',
  `file_md5` varchar(64) DEFAULT NULL COMMENT '文件MD5',
  `file_sha256` varchar(128) DEFAULT NULL COMMENT '文件SHA256',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小',
  `signature_url` varchar(512) DEFAULT NULL COMMENT '签名文件地址',
  `platform` varchar(32) DEFAULT NULL COMMENT '平台',
  `arch` varchar(32) DEFAULT NULL COMMENT '架构',
  `min_version` varchar(32) DEFAULT NULL COMMENT '最低版本',
  `min_version_code` int DEFAULT NULL COMMENT '最低版本代码',
  `breaking_changes` tinyint DEFAULT 0 COMMENT '是否有破坏性变更',
  `changelog` text COMMENT '更新日志',
  `release_notes` text COMMENT '发布说明',
  `release_type` varchar(16) DEFAULT 'stable' COMMENT '类型',
  `force_update` tinyint DEFAULT 0 COMMENT '是否强制更新',
  `rollback_support` tinyint DEFAULT 1 COMMENT '是否支持回滚',
  `grace_period` int DEFAULT 0 COMMENT '宽限期(小时)',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  `is_latest` tinyint DEFAULT 0 COMMENT '是否最新版本',
  `download_count` int DEFAULT 0 COMMENT '下载次数',
  `install_count` int DEFAULT 0 COMMENT '安装次数',
  `success_count` int DEFAULT 0 COMMENT '成功安装数',
  `fail_count` int DEFAULT 0 COMMENT '失败安装数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_version_platform_arch` (`version`, `platform`, `arch`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent版本表';

-- Agent实例表
CREATE TABLE IF NOT EXISTS `agents` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `server_name` varchar(128) DEFAULT NULL COMMENT '服务器名称',
  `server_ip` varchar(64) DEFAULT NULL COMMENT '服务器IP',
  `agent_id` varchar(64) DEFAULT NULL COMMENT 'Agent唯一标识',
  `agent_secret` varchar(128) DEFAULT NULL COMMENT 'Agent密钥',
  `version` varchar(32) DEFAULT NULL COMMENT '当前版本',
  `version_code` int DEFAULT 0 COMMENT '版本代码',
  `target_version` varchar(32) DEFAULT NULL COMMENT '目标版本',
  `platform` varchar(32) DEFAULT NULL COMMENT '平台',
  `arch` varchar(32) DEFAULT NULL COMMENT '架构',
  `os` varchar(64) DEFAULT NULL COMMENT '操作系统',
  `kernel` varchar(128) DEFAULT NULL COMMENT '内核版本',
  `status` varchar(16) DEFAULT 'pending' COMMENT '状态',
  `status_message` varchar(512) DEFAULT NULL COMMENT '状态消息',
  `status_changed_at` datetime DEFAULT NULL COMMENT '状态变更时间',
  `last_heartbeat` datetime DEFAULT NULL COMMENT '最后心跳时间',
  `heartbeat_ip` varchar(64) DEFAULT NULL COMMENT '心跳来源IP',
  `heartbeat_port` int DEFAULT NULL COMMENT '心跳来源端口',
  `heartbeat_latency` int DEFAULT NULL COMMENT '心跳延迟(ms)',
  `offline_count` int DEFAULT 0 COMMENT '离线次数',
  `last_offline_at` datetime DEFAULT NULL COMMENT '最后离线时间',
  `last_online_at` datetime DEFAULT NULL COMMENT '最后在线时间',
  `total_offline_time` bigint DEFAULT 0 COMMENT '累计离线时间(秒)',
  `error_count` int DEFAULT 0 COMMENT '错误次数',
  `last_error_at` datetime DEFAULT NULL COMMENT '最后错误时间',
  `last_error_msg` text COMMENT '最后错误消息',
  `auto_recover` tinyint DEFAULT 1 COMMENT '是否自动恢复',
  `recover_count` int DEFAULT 0 COMMENT '恢复次数',
  `last_recover_at` datetime DEFAULT NULL COMMENT '最后恢复时间',
  `auto_upgrade` tinyint DEFAULT 1 COMMENT '是否自动升级',
  `upgrade_channel` varchar(16) DEFAULT 'stable' COMMENT '升级通道',
  `upgrade_window` varchar(64) DEFAULT NULL COMMENT '升级时间窗口',
  `gray_group` varchar(32) DEFAULT NULL COMMENT '灰度分组',
  `gray_weight` int DEFAULT 0 COMMENT '灰度权重',
  `config_hash` varchar(64) DEFAULT NULL COMMENT '配置Hash',
  `config_at` datetime DEFAULT NULL COMMENT '配置同步时间',
  `config_error` varchar(512) DEFAULT NULL COMMENT '配置错误',
  `uptime_seconds` bigint DEFAULT NULL COMMENT '运行时长(秒)',
  `task_count` bigint DEFAULT 0 COMMENT '执行任务总数',
  `task_success_count` bigint DEFAULT 0 COMMENT '成功任务数',
  `task_fail_count` bigint DEFAULT 0 COMMENT '失败任务数',
  `tags` text COMMENT '标签(JSON)',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_agent_id` (`agent_id`),
  UNIQUE KEY `idx_server_id` (`server_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent实例表';

-- Agent升级任务表
CREATE TABLE IF NOT EXISTS `agent_upgrade_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `agent_id` bigint unsigned NOT NULL COMMENT 'Agent ID',
  `agent_uuid` varchar(64) DEFAULT NULL COMMENT 'Agent UUID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `server_name` varchar(128) DEFAULT NULL COMMENT '服务器名称',
  `server_ip` varchar(64) DEFAULT NULL COMMENT '服务器IP',
  `from_version` varchar(32) DEFAULT NULL COMMENT '原版本',
  `from_version_code` int DEFAULT NULL COMMENT '原版本代码',
  `to_version` varchar(32) DEFAULT NULL COMMENT '目标版本',
  `to_version_code` int DEFAULT NULL COMMENT '目标版本代码',
  `version_id` bigint unsigned DEFAULT NULL COMMENT '版本记录ID',
  `task_type` varchar(16) DEFAULT 'manual' COMMENT '任务类型',
  `priority` int DEFAULT 5 COMMENT '优先级',
  `scheduled_at` datetime DEFAULT NULL COMMENT '计划执行时间',
  `strategy_id` bigint unsigned DEFAULT NULL COMMENT '灰度策略ID',
  `status` varchar(16) DEFAULT 'pending' COMMENT '状态',
  `status_msg` varchar(512) DEFAULT NULL COMMENT '状态消息',
  `progress` int DEFAULT 0 COMMENT '进度',
  `progress_detail` text COMMENT '进度详情(JSON)',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `download_url` varchar(512) DEFAULT NULL COMMENT '下载地址',
  `download_size` bigint DEFAULT NULL COMMENT '文件大小',
  `download_md5` varchar(64) DEFAULT NULL COMMENT '文件MD5',
  `downloaded_size` bigint DEFAULT NULL COMMENT '已下载大小',
  `download_speed` bigint DEFAULT NULL COMMENT '下载速度',
  `result` text COMMENT '结果详情',
  `output` text COMMENT '输出日志',
  `error` text COMMENT '错误信息',
  `rollback_enabled` tinyint DEFAULT 1 COMMENT '允许回滚',
  `rollback_at` datetime DEFAULT NULL COMMENT '回滚时间',
  `rollback_error` text COMMENT '回滚错误',
  `retry_count` int DEFAULT 0 COMMENT '重试次数',
  `max_retry` int DEFAULT 3 COMMENT '最大重试次数',
  `next_retry_at` datetime DEFAULT NULL COMMENT '下次重试时间',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建者ID',
  `created_by_name` varchar(64) DEFAULT NULL COMMENT '创建者名称',
  PRIMARY KEY (`id`),
  KEY `idx_agent_id` (`agent_id`),
  KEY `idx_agent_uuid` (`agent_uuid`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent升级任务表';

-- Agent心跳记录表
CREATE TABLE IF NOT EXISTS `agent_heartbeat_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `agent_id` bigint unsigned DEFAULT NULL COMMENT 'Agent ID',
  `agent_uuid` varchar(64) DEFAULT NULL COMMENT 'Agent UUID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `ip` varchar(64) DEFAULT NULL COMMENT '来源IP',
  `port` int DEFAULT NULL COMMENT '来源端口',
  `version` varchar(32) DEFAULT NULL COMMENT '版本',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `uptime_seconds` bigint DEFAULT NULL COMMENT '运行时长',
  `cpu_usage` double DEFAULT NULL COMMENT 'CPU使用率',
  `memory_usage` double DEFAULT NULL COMMENT '内存使用率',
  `goroutine_count` int DEFAULT NULL COMMENT '协程数',
  `pending_tasks` int DEFAULT NULL COMMENT '待处理任务数',
  `running_tasks` int DEFAULT NULL COMMENT '运行中任务数',
  `completed_tasks` int DEFAULT NULL COMMENT '已完成任务数',
  `failed_tasks` int DEFAULT NULL COMMENT '失败任务数',
  `net_in_bytes` bigint unsigned DEFAULT NULL COMMENT '网络入流量',
  `net_out_bytes` bigint unsigned DEFAULT NULL COMMENT '网络出流量',
  `latency_ms` int DEFAULT NULL COMMENT '心跳延迟(ms)',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_agent_id` (`agent_id`),
  KEY `idx_agent_uuid` (`agent_uuid`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent心跳记录表';

-- Agent恢复记录表
CREATE TABLE IF NOT EXISTS `agent_recover_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `agent_id` bigint unsigned DEFAULT NULL COMMENT 'Agent ID',
  `agent_uuid` varchar(64) DEFAULT NULL COMMENT 'Agent UUID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `server_name` varchar(128) DEFAULT NULL COMMENT '服务器名称',
  `server_ip` varchar(64) DEFAULT NULL COMMENT '服务器IP',
  `trigger_type` varchar(16) DEFAULT NULL COMMENT '触发类型',
  `trigger_cause` varchar(64) DEFAULT NULL COMMENT '触发原因',
  `trigger_msg` text COMMENT '触发消息',
  `action` varchar(32) DEFAULT NULL COMMENT '恢复动作',
  `command` text COMMENT '执行命令',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `output` text COMMENT '输出',
  `error` text COMMENT '错误',
  `success` tinyint DEFAULT NULL COMMENT '是否成功',
  `retry_count` int DEFAULT NULL COMMENT '重试次数',
  PRIMARY KEY (`id`),
  KEY `idx_agent_id` (`agent_id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent恢复记录表';

-- 灰度发布策略表
CREATE TABLE IF NOT EXISTS `gray_release_strategies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) NOT NULL COMMENT '策略名称',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `version_id` bigint unsigned DEFAULT NULL COMMENT '版本ID',
  `version_name` varchar(32) DEFAULT NULL COMMENT '版本号',
  `strategy_type` varchar(16) DEFAULT NULL COMMENT '策略类型',
  `initial_weight` int DEFAULT 5 COMMENT '初始权重(%)',
  `target_weight` int DEFAULT 100 COMMENT '目标权重(%)',
  `step_size` int DEFAULT 10 COMMENT '步进大小(%)',
  `step_interval` int DEFAULT 30 COMMENT '步进间隔(分钟)',
  `group_list` text COMMENT '灰度分组列表(JSON)',
  `label_selector` text COMMENT '标签选择器(JSON)',
  `server_ids` text COMMENT '指定服务器ID列表(JSON)',
  `status` varchar(16) DEFAULT 'pending' COMMENT '状态',
  `current_step` int DEFAULT 0 COMMENT '当前步数',
  `current_weight` int DEFAULT 0 COMMENT '当前权重(%)',
  `scheduled_at` datetime DEFAULT NULL COMMENT '计划开始时间',
  `started_at` datetime DEFAULT NULL COMMENT '实际开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `paused_at` datetime DEFAULT NULL COMMENT '暂停时间',
  `total_agents` int DEFAULT NULL COMMENT '总Agent数',
  `upgraded_agents` int DEFAULT NULL COMMENT '已升级数',
  `success_agents` int DEFAULT NULL COMMENT '成功数',
  `failed_agents` int DEFAULT NULL COMMENT '失败数',
  `pause_on_failure` tinyint DEFAULT 1 COMMENT '失败时暂停',
  `failure_threshold` double DEFAULT 10 COMMENT '失败阈值(%)',
  `auto_rollback` tinyint DEFAULT 1 COMMENT '自动回滚',
  `rollback_threshold` double DEFAULT 30 COMMENT '回滚阈值(%)',
  `health_check_enabled` tinyint DEFAULT 1 COMMENT '启用健康检查',
  `health_check_url` varchar(256) DEFAULT NULL COMMENT '健康检查URL',
  `health_check_timeout` int DEFAULT 30 COMMENT '健康检查超时(秒)',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建者ID',
  `created_by_name` varchar(64) DEFAULT NULL COMMENT '创建者名称',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='灰度发布策略表';

-- Agent配置表
CREATE TABLE IF NOT EXISTS `agent_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) NOT NULL COMMENT '配置名称',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `config_json` longtext COMMENT '配置内容(JSON)',
  `config_hash` varchar(64) DEFAULT NULL COMMENT '配置Hash',
  `scope` varchar(16) DEFAULT NULL COMMENT '范围',
  `scope_value` text COMMENT '范围值(JSON)',
  `version` int DEFAULT 1 COMMENT '配置版本',
  `is_default` tinyint DEFAULT 0 COMMENT '是否默认配置',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  `applied_count` int DEFAULT 0 COMMENT '应用次数',
  PRIMARY KEY (`id`),
  KEY `idx_config_hash` (`config_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent配置表';

-- Agent指标表
CREATE TABLE IF NOT EXISTS `agent_metrics` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `agent_id` bigint unsigned DEFAULT NULL COMMENT 'Agent ID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `cpu_usage` double DEFAULT NULL COMMENT 'CPU使用率',
  `memory_usage` double DEFAULT NULL COMMENT '内存使用率',
  `memory_used` bigint unsigned DEFAULT NULL COMMENT '已用内存(MB)',
  `memory_total` bigint unsigned DEFAULT NULL COMMENT '总内存(MB)',
  `goroutine_count` int DEFAULT NULL COMMENT '协程数',
  `thread_count` int DEFAULT NULL COMMENT '线程数',
  `handle_count` int DEFAULT NULL COMMENT '句柄数',
  `net_in_bytes` bigint unsigned DEFAULT NULL COMMENT '网络入流量',
  `net_out_bytes` bigint unsigned DEFAULT NULL COMMENT '网络出流量',
  `pending_tasks` int DEFAULT NULL COMMENT '待处理任务',
  `running_tasks` int DEFAULT NULL COMMENT '运行中任务',
  `completed_tasks` int DEFAULT NULL COMMENT '已完成任务',
  `failed_tasks` int DEFAULT NULL COMMENT '失败任务',
  `task_avg_latency` double DEFAULT NULL COMMENT '任务平均延迟(ms)',
  `task_max_latency` double DEFAULT NULL COMMENT '任务最大延迟(ms)',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_agent_id` (`agent_id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent指标表';

-- ==================== 调度器 ====================

-- 调度任务表
CREATE TABLE IF NOT EXISTS `scheduler_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(128) NOT NULL COMMENT '任务名称',
  `type` varchar(32) DEFAULT NULL COMMENT '任务类型',
  `priority` int DEFAULT NULL COMMENT '优先级',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `command` text COMMENT '命令',
  `script` text COMMENT '脚本',
  `params` text COMMENT '参数',
  `timeout` int DEFAULT NULL COMMENT '超时(秒)',
  `max_retry` int DEFAULT NULL COMMENT '最大重试次数',
  `retry_count` int DEFAULT NULL COMMENT '重试次数',
  `retry_delay` int DEFAULT NULL COMMENT '重试延迟(秒)',
  `retry_backoff` varchar(32) DEFAULT NULL COMMENT '重试退避策略',
  `executor` varchar(32) DEFAULT NULL COMMENT '执行器',
  `action` text COMMENT '动作',
  `target_type` varchar(32) DEFAULT NULL COMMENT '目标类型',
  `target_ids` text COMMENT '目标IDs(JSON)',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `server_name` varchar(64) DEFAULT NULL COMMENT '服务器名称',
  `queue_name` varchar(64) DEFAULT NULL COMMENT '队列名称',
  `schedule_type` varchar(16) DEFAULT NULL COMMENT '调度类型',
  `scheduled_at` datetime DEFAULT NULL COMMENT '计划执行时间',
  `schedule_time` datetime DEFAULT NULL COMMENT '调度时间',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `output` text COMMENT '输出',
  `result` text COMMENT '结果',
  `error` text COMMENT '错误',
  `stdout` text COMMENT '标准输出',
  `stderr` text COMMENT '标准错误',
  `exit_code` int DEFAULT NULL COMMENT '退出码',
  `error_message` text COMMENT '错误消息',
  `rollback_enabled` tinyint DEFAULT NULL COMMENT '是否启用回滚',
  `rollback_command` text COMMENT '回滚命令',
  `dependencies` text COMMENT '依赖(JSON)',
  `depends_on` text COMMENT '依赖于(JSON)',
  `batch_id` bigint unsigned DEFAULT NULL COMMENT '批次ID',
  `batch_index` int DEFAULT NULL COMMENT '批次索引',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父任务ID',
  `queue_at` datetime DEFAULT NULL COMMENT '入队时间',
  `start_at` datetime DEFAULT NULL COMMENT '开始时间',
  `end_at` datetime DEFAULT NULL COMMENT '结束时间',
  `worker_id` varchar(32) DEFAULT NULL COMMENT 'Worker ID',
  `callback_url` varchar(255) DEFAULT NULL COMMENT '回调URL',
  `callback_data` text COMMENT '回调数据',
  `idempotent_key` varchar(64) DEFAULT NULL COMMENT '幂等键',
  `dedup_window` int DEFAULT NULL COMMENT '去重窗口(秒)',
  `tags` text COMMENT '标签(JSON)',
  `metadata` text COMMENT '元数据(JSON)',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建者ID',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_idempotent_key` (`idempotent_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调度任务表';

-- 定时任务表
CREATE TABLE IF NOT EXISTS `scheduler_cron_jobs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(128) NOT NULL COMMENT '任务名称',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `cron_expr` varchar(64) NOT NULL COMMENT 'Cron表达式',
  `timezone` varchar(32) DEFAULT NULL COMMENT '时区',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  `task_template` text COMMENT '任务模板(JSON)',
  `concurrent_policy` varchar(16) DEFAULT NULL COMMENT '并发策略',
  `run_count` int DEFAULT NULL COMMENT '运行次数',
  `success_count` int DEFAULT NULL COMMENT '成功次数',
  `fail_count` int DEFAULT NULL COMMENT '失败次数',
  `last_run_at` datetime DEFAULT NULL COMMENT '最后运行时间',
  `next_run_at` datetime DEFAULT NULL COMMENT '下次运行时间',
  `last_error` text COMMENT '最后错误',
  `notify_on_success` tinyint DEFAULT NULL COMMENT '成功时通知',
  `notify_on_fail` tinyint DEFAULT NULL COMMENT '失败时通知',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建者ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';

-- Cron执行记录表
CREATE TABLE IF NOT EXISTS `scheduler_cron_executions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `cron_job_id` bigint unsigned DEFAULT NULL COMMENT 'Cron任务ID',
  `task_id` bigint unsigned DEFAULT NULL COMMENT '任务ID',
  `scheduled_at` datetime DEFAULT NULL COMMENT '计划时间',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `error` text COMMENT '错误',
  PRIMARY KEY (`id`),
  KEY `idx_cron_job_id` (`cron_job_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Cron执行记录表';

-- 调度任务事件表
CREATE TABLE IF NOT EXISTS `scheduler_task_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `task_id` bigint unsigned DEFAULT NULL COMMENT '任务ID',
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `source` varchar(32) DEFAULT NULL COMMENT '来源',
  `message` varchar(255) DEFAULT NULL COMMENT '消息',
  `data` text COMMENT '数据(JSON)',
  `event_type` varchar(32) DEFAULT NULL COMMENT '事件类型',
  `event_data` text COMMENT '事件数据(JSON)',
  `operator` varchar(32) DEFAULT NULL COMMENT '操作者',
  `remark` text COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调度任务事件表';

-- 调度队列配置表
CREATE TABLE IF NOT EXISTS `scheduler_queues` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) NOT NULL COMMENT '队列名称',
  `description` text COMMENT '描述',
  `max_workers` int DEFAULT NULL COMMENT '最大Worker数',
  `max_pending` int DEFAULT NULL COMMENT '最大待处理数',
  `priority` int DEFAULT NULL COMMENT '优先级',
  `timeout` int DEFAULT NULL COMMENT '超时(秒)',
  `max_retry` int DEFAULT NULL COMMENT '最大重试次数',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  `pending_count` int DEFAULT NULL COMMENT '待处理数',
  `running_count` int DEFAULT NULL COMMENT '运行中数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调度队列配置表';

-- 任务执行记录表
CREATE TABLE IF NOT EXISTS `scheduler_task_executions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `task_id` bigint unsigned DEFAULT NULL COMMENT '任务ID',
  `attempt` int DEFAULT NULL COMMENT '尝试次数',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `worker_id` varchar(32) DEFAULT NULL COMMENT 'Worker ID',
  `server_id` bigint unsigned DEFAULT NULL COMMENT '服务器ID',
  `execution_id` varchar(64) DEFAULT NULL COMMENT '执行ID',
  `start_at` datetime DEFAULT NULL COMMENT '开始时间',
  `end_at` datetime DEFAULT NULL COMMENT '结束时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `exit_code` int DEFAULT NULL COMMENT '退出码',
  `stdout` text COMMENT '标准输出',
  `stderr` text COMMENT '标准错误',
  `output` text COMMENT '输出',
  `error` text COMMENT '错误',
  `error_message` text COMMENT '错误消息',
  `rollback_at` datetime DEFAULT NULL COMMENT '回滚时间',
  `rollback_result` text COMMENT '回滚结果',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_execution_id` (`execution_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行记录表';

-- 任务批次表
CREATE TABLE IF NOT EXISTS `scheduler_task_batches` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(128) DEFAULT NULL COMMENT '批次名称',
  `description` text COMMENT '描述',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `total_tasks` int DEFAULT NULL COMMENT '总任务数',
  `pending_tasks` int DEFAULT NULL COMMENT '待处理任务数',
  `running_tasks` int DEFAULT NULL COMMENT '运行中任务数',
  `success_tasks` int DEFAULT NULL COMMENT '成功任务数',
  `failed_tasks` int DEFAULT NULL COMMENT '失败任务数',
  `start_at` datetime DEFAULT NULL COMMENT '开始时间',
  `end_at` datetime DEFAULT NULL COMMENT '结束时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `parallelism` int DEFAULT NULL COMMENT '并行度',
  `stop_on_fail` tinyint DEFAULT NULL COMMENT '失败时停止',
  `notify_on_complete` tinyint DEFAULT NULL COMMENT '完成时通知',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建者ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务批次表';

-- 任务模板表
CREATE TABLE IF NOT EXISTS `scheduler_task_templates` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(128) DEFAULT NULL COMMENT '模板名称',
  `category` varchar(32) DEFAULT NULL COMMENT '分类',
  `description` text COMMENT '描述',
  `task_def` text COMMENT '任务定义(JSON)',
  `params` text COMMENT '参数(JSON)',
  `use_count` int DEFAULT NULL COMMENT '使用次数',
  `enabled` tinyint DEFAULT NULL COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务模板表';

-- 任务日志表
CREATE TABLE IF NOT EXISTS `scheduler_task_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `execution_id` varchar(64) DEFAULT NULL COMMENT '执行ID',
  `task_id` bigint unsigned DEFAULT NULL COMMENT '任务ID',
  `level` varchar(16) DEFAULT NULL COMMENT '日志级别',
  `message` text COMMENT '消息',
  `data` text COMMENT '数据(JSON)',
  PRIMARY KEY (`id`),
  KEY `idx_execution_id` (`execution_id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务日志表';

-- ==================== 租户管理 ====================

-- 租户表
CREATE TABLE IF NOT EXISTS `tenants` (
  `id` varchar(36) NOT NULL COMMENT '租户ID',
  `name` varchar(100) NOT NULL COMMENT '租户名称',
  `slug` varchar(50) NOT NULL COMMENT 'URL友好标识',
  `domain` varchar(255) DEFAULT NULL COMMENT '自定义域名',
  `logo` varchar(500) DEFAULT NULL COMMENT 'Logo',
  `description` text COMMENT '描述',
  `status` varchar(20) DEFAULT 'active' COMMENT '状态',
  `plan` varchar(50) DEFAULT 'free' COMMENT '套餐',
  `billing_cycle` varchar(20) DEFAULT NULL COMMENT '计费周期',
  `contact_name` varchar(100) DEFAULT NULL COMMENT '联系人',
  `contact_email` varchar(255) DEFAULT NULL COMMENT '联系邮箱',
  `contact_phone` varchar(50) DEFAULT NULL COMMENT '联系电话',
  `address` text COMMENT '地址',
  `settings` json DEFAULT NULL COMMENT '设置(JSON)',
  `features` json DEFAULT NULL COMMENT '功能列表(JSON)',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  UNIQUE KEY `idx_slug` (`slug`),
  UNIQUE KEY `idx_domain` (`domain`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表';

-- 租户配额表
CREATE TABLE IF NOT EXISTS `tenant_quotas` (
  `id` varchar(36) NOT NULL COMMENT '配额ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `max_users` int DEFAULT 5 COMMENT '最大用户数',
  `max_admins` int DEFAULT 2 COMMENT '最大管理员数',
  `max_resources` int DEFAULT 100 COMMENT '最大资源数',
  `max_servers` int DEFAULT 50 COMMENT '最大服务器数',
  `max_databases` int DEFAULT 20 COMMENT '最大数据库数',
  `max_monitors` int DEFAULT 100 COMMENT '最大监控数',
  `max_alert_rules` int DEFAULT 50 COMMENT '最大告警规则数',
  `metrics_retention` int DEFAULT 30 COMMENT '指标保留天数',
  `max_cloud_accounts` int DEFAULT 5 COMMENT '最大云账号数',
  `budget_limit` double DEFAULT 0 COMMENT '预算限制',
  `max_storage_gb` int DEFAULT 100 COMMENT '最大存储(GB)',
  `max_backup_gb` int DEFAULT 500 COMMENT '最大备份(GB)',
  `max_api_calls` int DEFAULT 10000 COMMENT '最大API调用数',
  `max_webhooks` int DEFAULT 10 COMMENT '最大Webhooks数',
  `current_users` int DEFAULT 0 COMMENT '当前用户数',
  `current_resources` int DEFAULT 0 COMMENT '当前资源数',
  `current_storage_gb` int DEFAULT 0 COMMENT '当前存储(GB)',
  `current_api_calls` int DEFAULT 0 COMMENT '当前API调用数',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户配额表';

-- 租户用户表
CREATE TABLE IF NOT EXISTS `tenant_users` (
  `id` varchar(36) NOT NULL COMMENT 'ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `user_id` varchar(36) NOT NULL COMMENT '用户ID',
  `email` varchar(255) DEFAULT NULL COMMENT '邮箱',
  `name` varchar(100) DEFAULT NULL COMMENT '名称',
  `avatar` varchar(500) DEFAULT NULL COMMENT '头像',
  `role_id` varchar(36) DEFAULT NULL COMMENT '角色ID',
  `role_name` varchar(50) DEFAULT NULL COMMENT '角色名称',
  `is_owner` tinyint DEFAULT 0 COMMENT '是否所有者',
  `is_admin` tinyint DEFAULT 0 COMMENT '是否管理员',
  `status` varchar(20) DEFAULT 'active' COMMENT '状态',
  `invited_by` varchar(36) DEFAULT NULL COMMENT '邀请人ID',
  `joined_at` datetime DEFAULT NULL COMMENT '加入时间',
  `last_active_at` datetime DEFAULT NULL COMMENT '最后活跃时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_email` (`email`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户用户表';

-- 租户角色表
CREATE TABLE IF NOT EXISTS `tenant_roles` (
  `id` varchar(36) NOT NULL COMMENT 'ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `name` varchar(50) NOT NULL COMMENT '角色名称',
  `slug` varchar(50) NOT NULL COMMENT '角色标识',
  `description` text COMMENT '描述',
  `is_system` tinyint DEFAULT 0 COMMENT '是否系统角色',
  `permissions` json DEFAULT NULL COMMENT '权限列表(JSON)',
  `scope` varchar(20) DEFAULT 'tenant' COMMENT '范围',
  `parent_id` varchar(36) DEFAULT NULL COMMENT '父角色ID',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户角色表';

-- 租户邀请表
CREATE TABLE IF NOT EXISTS `tenant_invitations` (
  `id` varchar(36) NOT NULL COMMENT 'ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `email` varchar(255) NOT NULL COMMENT '邮箱',
  `role_id` varchar(36) DEFAULT NULL COMMENT '角色ID',
  `role_name` varchar(50) DEFAULT NULL COMMENT '角色名称',
  `status` varchar(20) DEFAULT 'pending' COMMENT '状态',
  `token` varchar(64) DEFAULT NULL COMMENT '邀请令牌',
  `invited_by` varchar(36) DEFAULT NULL COMMENT '邀请人ID',
  `inviter_name` varchar(100) DEFAULT NULL COMMENT '邀请人名称',
  `expires_at` datetime DEFAULT NULL COMMENT '过期时间',
  `accepted_at` datetime DEFAULT NULL COMMENT '接受时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_token` (`token`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户邀请表';

-- 租户资源使用表
CREATE TABLE IF NOT EXISTS `tenant_resource_usage` (
  `id` varchar(36) NOT NULL COMMENT 'ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `date` date DEFAULT NULL COMMENT '日期',
  `hour` int DEFAULT -1 COMMENT '小时',
  `user_count` int DEFAULT 0 COMMENT '用户数',
  `active_users` int DEFAULT 0 COMMENT '活跃用户数',
  `resource_count` int DEFAULT 0 COMMENT '资源数',
  `server_count` int DEFAULT 0 COMMENT '服务器数',
  `database_count` int DEFAULT 0 COMMENT '数据库数',
  `monitor_count` int DEFAULT 0 COMMENT '监控数',
  `alert_count` int DEFAULT 0 COMMENT '告警数',
  `metrics_data_mb` int DEFAULT 0 COMMENT '指标数据(MB)',
  `total_cost` double DEFAULT 0 COMMENT '总成本',
  `cloud_cost` double DEFAULT 0 COMMENT '云成本',
  `storage_used_mb` int DEFAULT 0 COMMENT '已用存储(MB)',
  `backup_used_mb` int DEFAULT 0 COMMENT '已用备份(MB)',
  `api_calls` int DEFAULT 0 COMMENT 'API调用数',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户资源使用表';

-- 租户账单表
CREATE TABLE IF NOT EXISTS `tenant_billings` (
  `id` varchar(36) NOT NULL COMMENT 'ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `billing_period` varchar(20) NOT NULL COMMENT '账单周期',
  `due_date` datetime DEFAULT NULL COMMENT '到期日',
  `base_amount` double DEFAULT 0 COMMENT '基础费用',
  `usage_amount` double DEFAULT 0 COMMENT '用量费用',
  `overage_amount` double DEFAULT 0 COMMENT '超额费用',
  `discount_amount` double DEFAULT 0 COMMENT '折扣',
  `tax_amount` double DEFAULT 0 COMMENT '税费',
  `total_amount` double DEFAULT 0 COMMENT '总计',
  `usage_details` json DEFAULT NULL COMMENT '用量明细(JSON)',
  `status` varchar(20) DEFAULT 'pending' COMMENT '状态',
  `payment_method` varchar(50) DEFAULT NULL COMMENT '支付方式',
  `payment_id` varchar(100) DEFAULT NULL COMMENT '支付ID',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  `invoice_number` varchar(50) DEFAULT NULL COMMENT '发票号',
  `invoice_url` varchar(500) DEFAULT NULL COMMENT '发票URL',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户账单表';

-- 租户审计日志表
CREATE TABLE IF NOT EXISTS `tenant_audit_logs` (
  `id` varchar(36) NOT NULL COMMENT 'ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `user_id` varchar(36) DEFAULT NULL COMMENT '用户ID',
  `user_name` varchar(100) DEFAULT NULL COMMENT '用户名',
  `user_email` varchar(255) DEFAULT NULL COMMENT '用户邮箱',
  `action` varchar(100) NOT NULL COMMENT '操作',
  `resource` varchar(100) DEFAULT NULL COMMENT '资源',
  `resource_id` varchar(36) DEFAULT NULL COMMENT '资源ID',
  `resource_name` varchar(255) DEFAULT NULL COMMENT '资源名称',
  `old_value` json DEFAULT NULL COMMENT '旧值(JSON)',
  `new_value` json DEFAULT NULL COMMENT '新值(JSON)',
  `changes` json DEFAULT NULL COMMENT '变更(JSON)',
  `ip_address` varchar(50) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) DEFAULT NULL COMMENT '用户代理',
  `request_id` varchar(36) DEFAULT NULL COMMENT '请求ID',
  `status` varchar(20) DEFAULT 'success' COMMENT '状态',
  `error_msg` text COMMENT '错误消息',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_resource_id` (`resource_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户审计日志表';

-- ==================== 高可用 ====================

-- 集群节点表
CREATE TABLE IF NOT EXISTS `cluster_nodes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `node_id` varchar(64) NOT NULL COMMENT '节点ID',
  `node_name` varchar(128) DEFAULT NULL COMMENT '节点名称',
  `hostname` varchar(128) DEFAULT NULL COMMENT '主机名',
  `internal_ip` varchar(64) DEFAULT NULL COMMENT '内网IP',
  `external_ip` varchar(64) DEFAULT NULL COMMENT '外网IP',
  `api_port` int DEFAULT NULL COMMENT 'API端口',
  `grpc_port` int DEFAULT NULL COMMENT 'gRPC端口',
  `status` varchar(16) DEFAULT 'offline' COMMENT '状态',
  `role` varchar(16) DEFAULT 'follower' COMMENT '角色',
  `is_leader` tinyint DEFAULT 0 COMMENT '是否Leader',
  `last_heartbeat` datetime DEFAULT NULL COMMENT '最后心跳时间',
  `heartbeat_ip` varchar(64) DEFAULT NULL COMMENT '心跳来源IP',
  `heartbeat_count` bigint DEFAULT 0 COMMENT '心跳次数',
  `cpu_usage` double DEFAULT NULL COMMENT 'CPU使用率',
  `memory_usage` double DEFAULT NULL COMMENT '内存使用率',
  `disk_usage` double DEFAULT NULL COMMENT '磁盘使用率',
  `goroutine_count` int DEFAULT NULL COMMENT '协程数',
  `request_count` bigint DEFAULT 0 COMMENT '请求数',
  `connection_count` int DEFAULT NULL COMMENT '连接数',
  `version` varchar(32) DEFAULT NULL COMMENT '版本',
  `go_version` varchar(32) DEFAULT NULL COMMENT 'Go版本',
  `weight` int DEFAULT 100 COMMENT '权重',
  `data_center` varchar(64) DEFAULT NULL COMMENT '数据中心',
  `zone` varchar(64) DEFAULT NULL COMMENT '可用区',
  `rack` varchar(64) DEFAULT NULL COMMENT '机架',
  `labels` text COMMENT '标签(JSON)',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_node_id` (`node_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群节点表';

-- 分布式锁表
CREATE TABLE IF NOT EXISTS `distributed_locks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `lock_key` varchar(255) NOT NULL COMMENT '锁键',
  `lock_value` varchar(128) DEFAULT NULL COMMENT '锁值',
  `holder_node_id` varchar(64) DEFAULT NULL COMMENT '持有者节点ID',
  `holder_ip` varchar(64) DEFAULT NULL COMMENT '持有者IP',
  `acquired_at` datetime DEFAULT NULL COMMENT '获取时间',
  `expires_at` datetime DEFAULT NULL COMMENT '过期时间',
  `released_at` datetime DEFAULT NULL COMMENT '释放时间',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `renew_count` int DEFAULT 0 COMMENT '续期次数',
  `wait_count` bigint DEFAULT 0 COMMENT '等待次数',
  `ttl_seconds` int DEFAULT NULL COMMENT 'TTL(秒)',
  `resource_type` varchar(64) DEFAULT NULL COMMENT '资源类型',
  `resource_id` varchar(128) DEFAULT NULL COMMENT '资源ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_lock_key` (`lock_key`),
  KEY `idx_holder_node_id` (`holder_node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';

-- Leader选举表
CREATE TABLE IF NOT EXISTS `leader_elections` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `election_key` varchar(128) NOT NULL COMMENT '选举键',
  `leader_node_id` varchar(64) DEFAULT NULL COMMENT 'Leader节点ID',
  `leader_ip` varchar(64) DEFAULT NULL COMMENT 'Leader IP',
  `term` bigint DEFAULT NULL COMMENT '任期号',
  `acquired_at` datetime DEFAULT NULL COMMENT '获取时间',
  `expires_at` datetime DEFAULT NULL COMMENT '过期时间',
  `renew_count` int DEFAULT 0 COMMENT '续期次数',
  `leader_count` bigint DEFAULT 0 COMMENT '担任Leader次数',
  `candidates` text COMMENT '候选人列表(JSON)',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_election_key` (`election_key`),
  KEY `idx_leader_node_id` (`leader_node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Leader选举表';

-- HA集群配置表
CREATE TABLE IF NOT EXISTS `ha_cluster_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) NOT NULL COMMENT '配置名称',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `cluster_name` varchar(64) DEFAULT NULL COMMENT '集群名称',
  `cluster_mode` varchar(32) DEFAULT 'active-active' COMMENT '集群模式',
  `min_nodes` int DEFAULT 1 COMMENT '最小节点数',
  `max_nodes` int DEFAULT 10 COMMENT '最大节点数',
  `auto_discovery` tinyint DEFAULT 1 COMMENT '自动发现节点',
  `heartbeat_interval` int DEFAULT 10 COMMENT '心跳间隔(秒)',
  `heartbeat_timeout` int DEFAULT 30 COMMENT '心跳超时(秒)',
  `election_timeout` int DEFAULT 30 COMMENT '选举超时(秒)',
  `leader_lease_seconds` int DEFAULT 15 COMMENT 'Leader租约(秒)',
  `failover_enabled` tinyint DEFAULT 1 COMMENT '启用故障转移',
  `failover_timeout` int DEFAULT 60 COMMENT '故障转移超时(秒)',
  `auto_failback` tinyint DEFAULT 1 COMMENT '自动回切',
  `load_balance_enabled` tinyint DEFAULT 1 COMMENT '启用负载均衡',
  `load_balance_strategy` varchar(32) DEFAULT 'round-robin' COMMENT '负载均衡策略',
  `lock_backend` varchar(32) DEFAULT 'redis' COMMENT '锁后端',
  `lock_ttl_seconds` int DEFAULT 30 COMMENT '锁TTL(秒)',
  `redis_mode` varchar(32) DEFAULT 'standalone' COMMENT 'Redis模式',
  `redis_master_name` varchar(64) DEFAULT NULL COMMENT 'Redis主节点名称',
  `db_mode` varchar(32) DEFAULT 'standalone' COMMENT '数据库模式',
  `db_read_from_slave` tinyint DEFAULT 0 COMMENT '从库读取',
  `session_mode` varchar(32) DEFAULT 'memory' COMMENT '会话模式',
  `session_ttl` int DEFAULT 1800 COMMENT '会话TTL(秒)',
  `enabled` tinyint DEFAULT 1 COMMENT '是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HA集群配置表';

-- 故障转移记录表
CREATE TABLE IF NOT EXISTS `failover_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `failover_type` varchar(32) DEFAULT NULL COMMENT '类型',
  `failed_node_id` varchar(64) DEFAULT NULL COMMENT '故障节点ID',
  `failed_node_name` varchar(128) DEFAULT NULL COMMENT '故障节点名称',
  `failed_node_ip` varchar(64) DEFAULT NULL COMMENT '故障节点IP',
  `reason` varchar(255) DEFAULT NULL COMMENT '故障原因',
  `detected_at` datetime DEFAULT NULL COMMENT '检测时间',
  `target_node_id` varchar(64) DEFAULT NULL COMMENT '目标节点ID',
  `target_node_name` varchar(128) DEFAULT NULL COMMENT '目标节点名称',
  `target_node_ip` varchar(64) DEFAULT NULL COMMENT '目标节点IP',
  `status` varchar(16) DEFAULT NULL COMMENT '状态',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT NULL COMMENT '耗时(毫秒)',
  `success` tinyint DEFAULT NULL COMMENT '是否成功',
  `error` text COMMENT '错误信息',
  `trigger_type` varchar(16) DEFAULT NULL COMMENT '触发类型',
  `triggered_by` varchar(64) DEFAULT NULL COMMENT '触发者',
  PRIMARY KEY (`id`),
  KEY `idx_failed_node_id` (`failed_node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='故障转移记录表';

-- HA会话表
CREATE TABLE IF NOT EXISTS `ha_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_id` varchar(128) NOT NULL COMMENT '会话ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `username` varchar(64) DEFAULT NULL COMMENT '用户名',
  `data` text COMMENT '会话数据(JSON)',
  `created_node_id` varchar(64) DEFAULT NULL COMMENT '创建节点ID',
  `last_access_node_id` varchar(64) DEFAULT NULL COMMENT '最后访问节点ID',
  `last_access_at` datetime DEFAULT NULL COMMENT '最后访问时间',
  `expires_at` datetime DEFAULT NULL COMMENT '过期时间',
  `is_active` tinyint DEFAULT 1 COMMENT '是否活跃',
  `client_ip` varchar(64) DEFAULT NULL COMMENT '客户端IP',
  `user_agent` varchar(255) DEFAULT NULL COMMENT '用户代理',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_session_id` (`session_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HA会话表';

-- 集群事件表
CREATE TABLE IF NOT EXISTS `cluster_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `event_type` varchar(32) DEFAULT NULL COMMENT '事件类型',
  `node_id` varchar(64) DEFAULT NULL COMMENT '节点ID',
  `node_name` varchar(128) DEFAULT NULL COMMENT '节点名称',
  `node_ip` varchar(64) DEFAULT NULL COMMENT '节点IP',
  `title` varchar(255) DEFAULT NULL COMMENT '标题',
  `detail` text COMMENT '详情',
  `level` varchar(16) DEFAULT NULL COMMENT '级别',
  `source` varchar(64) DEFAULT NULL COMMENT '来源',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_node_id` (`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群事件表';

-- 节点指标表
CREATE TABLE IF NOT EXISTS `node_metrics` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `node_id` bigint unsigned DEFAULT NULL COMMENT '节点ID',
  `node_uuid` varchar(64) DEFAULT NULL COMMENT '节点UUID',
  `cpu_usage` double DEFAULT NULL COMMENT 'CPU使用率',
  `memory_usage` double DEFAULT NULL COMMENT '内存使用率',
  `memory_used` bigint unsigned DEFAULT NULL COMMENT '已用内存(MB)',
  `memory_total` bigint unsigned DEFAULT NULL COMMENT '总内存(MB)',
  `disk_usage` double DEFAULT NULL COMMENT '磁盘使用率',
  `disk_used` bigint unsigned DEFAULT NULL COMMENT '已用磁盘(GB)',
  `disk_total` bigint unsigned DEFAULT NULL COMMENT '总磁盘(GB)',
  `net_in_bytes` bigint unsigned DEFAULT NULL COMMENT '网络入流量',
  `net_out_bytes` bigint unsigned DEFAULT NULL COMMENT '网络出流量',
  `goroutine_count` int DEFAULT NULL COMMENT '协程数',
  `thread_count` int DEFAULT NULL COMMENT '线程数',
  `handle_count` int DEFAULT NULL COMMENT '句柄数',
  `request_count` bigint DEFAULT NULL COMMENT '请求数',
  `request_latency` double DEFAULT NULL COMMENT '请求延迟(ms)',
  `request_qps` double DEFAULT NULL COMMENT 'QPS',
  `connection_count` int DEFAULT NULL COMMENT '连接数',
  `load1` double DEFAULT NULL COMMENT '1分钟负载',
  `load5` double DEFAULT NULL COMMENT '5分钟负载',
  `load15` double DEFAULT NULL COMMENT '15分钟负载',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_node_id` (`node_id`),
  KEY `idx_node_uuid` (`node_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点指标表';

-- ==================== 初始化数据 ====================

-- 初始化系统角色
INSERT INTO `sys_roles` (`name`, `keyword`, `description`, `status`) VALUES
('超级管理员', 'admin', '拥有系统所有权限', 1),
('普通用户', 'user', '普通用户权限', 1)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 初始化系统用户 (密码: admin123)
INSERT INTO `sys_users` (`username`, `password`, `nick_name`, `email`, `status`, `role_id`) VALUES
('admin', 'e10adc3949ba59abbe56e057f20f883e', '超级管理员', 'admin@example.com', 1, 1)
ON DUPLICATE KEY UPDATE `username` = VALUES(`username`);

-- 初始化菜单
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`) VALUES
(0, '仪表盘', 'Dashboard', '/dashboard', 'views/dashboard/index.vue', 'dashboard', 1, 1),
(0, '服务器管理', 'Servers', '/servers', 'views/servers/index.vue', 'server', 2, 1),
(0, 'Kubernetes', 'Kubernetes', '/kubernetes', 'views/kubernetes/index.vue', 'kubernetes', 3, 1),
(0, '告警管理', 'Alerts', '/alerts', 'views/alerts/index.vue', 'alert', 4, 1),
(0, '系统管理', 'System', '/system', '', 'setting', 5, 1)
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`);

-- 系统管理子菜单
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`) VALUES
((SELECT id FROM (SELECT id FROM sys_menus WHERE name = 'System') AS tmp), '用户管理', 'User', '/system/user', 'views/system/user/index.vue', 'user', 1, 1),
((SELECT id FROM (SELECT id FROM sys_menus WHERE name = 'System') AS tmp), '角色管理', 'Role', '/system/role', 'views/system/role/index.vue', 'role', 2, 1),
((SELECT id FROM (SELECT id FROM sys_menus WHERE name = 'System') AS tmp), '菜单管理', 'Menu', '/system/menu', 'views/system/menu/index.vue', 'menu', 3, 1)
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`);

-- 初始化角色菜单关联
INSERT INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT 1, id FROM `sys_menus`
ON DUPLICATE KEY UPDATE `role_id` = VALUES(`role_id`);
