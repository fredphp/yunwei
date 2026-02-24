-- AI 自动化运维系统 数据库初始化脚本

CREATE DATABASE IF NOT EXISTS `yunwei` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `yunwei`;

-- ==================== 用户认证 ====================

CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码',
  `nick_name` varchar(64) DEFAULT NULL COMMENT '昵称',
  `email` varchar(128) DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `role` varchar(32) DEFAULT 'user' COMMENT '角色',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1启用, 0禁用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ==================== 服务器管理 ====================

CREATE TABLE IF NOT EXISTS `server_groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `name` varchar(64) NOT NULL COMMENT '分组名称',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `parent_id` bigint unsigned DEFAULT 0 COMMENT '父分组ID',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器分组表';

CREATE TABLE IF NOT EXISTS `servers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  
  -- 基本信息
  `name` varchar(64) NOT NULL COMMENT '服务器名称',
  `hostname` varchar(64) DEFAULT NULL COMMENT '主机名',
  `host` varchar(64) NOT NULL COMMENT 'IP地址',
  `port` int DEFAULT 22 COMMENT 'SSH端口',
  `user` varchar(32) DEFAULT NULL COMMENT 'SSH用户',
  `password` varchar(255) DEFAULT NULL COMMENT 'SSH密码(加密)',
  `private_key` text COMMENT 'SSH私钥(加密)',
  
  -- 分组
  `group_id` bigint unsigned DEFAULT NULL COMMENT '分组ID',
  
  -- 系统信息
  `os` varchar(64) DEFAULT NULL COMMENT '操作系统',
  `arch` varchar(32) DEFAULT NULL COMMENT '架构',
  `kernel` varchar(64) DEFAULT NULL COMMENT '内核版本',
  `cpu_cores` int DEFAULT NULL COMMENT 'CPU核心数',
  `memory_total` bigint DEFAULT NULL COMMENT '内存总量(MB)',
  `disk_total` bigint DEFAULT NULL COMMENT '磁盘总量(GB)',
  
  -- 状态
  `status` varchar(16) DEFAULT 'pending' COMMENT '状态: online/offline/pending',
  `ssh_status` varchar(16) DEFAULT NULL COMMENT 'SSH状态: success/failed',
  `ssh_error` varchar(255) DEFAULT NULL COMMENT 'SSH错误',
  
  -- 实时指标
  `cpu_usage` double DEFAULT 0 COMMENT 'CPU使用率',
  `memory_usage` double DEFAULT 0 COMMENT '内存使用率',
  `disk_usage` double DEFAULT 0 COMMENT '磁盘使用率',
  `load1` double DEFAULT 0 COMMENT '1分钟负载',
  `load5` double DEFAULT 0 COMMENT '5分钟负载',
  `load15` double DEFAULT 0 COMMENT '15分钟负载',
  
  -- Agent
  `agent_id` varchar(64) DEFAULT NULL COMMENT 'Agent ID',
  `agent_online` tinyint DEFAULT 0 COMMENT 'Agent在线',
  `last_check` datetime DEFAULT NULL COMMENT '最后检测时间',
  `last_heartbeat` datetime DEFAULT NULL COMMENT '最后心跳时间',
  
  -- 其他
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_agent_id` (`agent_id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器表';

CREATE TABLE IF NOT EXISTS `server_metrics` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned NOT NULL COMMENT '服务器ID',
  
  -- CPU
  `cpu_usage` double DEFAULT 0,
  `cpu_user` double DEFAULT 0,
  `cpu_system` double DEFAULT 0,
  `cpu_idle` double DEFAULT 0,
  
  -- 内存
  `memory_usage` double DEFAULT 0,
  `memory_used` bigint DEFAULT 0,
  `memory_free` bigint DEFAULT 0,
  `memory_cache` bigint DEFAULT 0,
  
  -- 磁盘
  `disk_usage` double DEFAULT 0,
  `disk_used` bigint DEFAULT 0,
  `disk_free` bigint DEFAULT 0,
  `disk_io_read` bigint DEFAULT 0,
  `disk_io_write` bigint DEFAULT 0,
  
  -- 网络
  `net_in` bigint DEFAULT 0,
  `net_out` bigint DEFAULT 0,
  
  -- 负载
  `load1` double DEFAULT 0,
  `load5` double DEFAULT 0,
  `load15` double DEFAULT 0,
  
  -- 进程
  `process_count` int DEFAULT 0,
  
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器指标表';

CREATE TABLE IF NOT EXISTS `server_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned NOT NULL,
  `type` varchar(32) DEFAULT NULL COMMENT '类型',
  `content` text COMMENT '内容',
  `output` text COMMENT '输出',
  `error` text COMMENT '错误',
  `duration` bigint DEFAULT 0 COMMENT '耗时(ms)',
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器日志表';

-- ==================== Docker 容器 ====================

CREATE TABLE IF NOT EXISTS `docker_containers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `server_id` bigint unsigned NOT NULL,
  `container_id` varchar(64) DEFAULT NULL,
  `name` varchar(128) DEFAULT NULL,
  `image` varchar(255) DEFAULT NULL,
  `status` varchar(32) DEFAULT NULL,
  `state` varchar(32) DEFAULT NULL,
  `cpu_usage` double DEFAULT 0,
  `memory_usage` double DEFAULT 0,
  `net_io` varchar(32) DEFAULT NULL,
  `block_io` varchar(32) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_container_id` (`container_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Docker容器表';

-- ==================== 端口信息 ====================

CREATE TABLE IF NOT EXISTS `port_infos` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `server_id` bigint unsigned NOT NULL,
  `port` int NOT NULL,
  `protocol` varchar(16) DEFAULT NULL,
  `service` varchar(64) DEFAULT NULL,
  `pid` int DEFAULT NULL,
  `process` varchar(128) DEFAULT NULL,
  `state` varchar(32) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_port` (`port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='端口信息表';

-- ==================== 告警管理 ====================

CREATE TABLE IF NOT EXISTS `alert_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) DEFAULT NULL,
  `metric` varchar(64) NOT NULL COMMENT '指标名称',
  `operator` varchar(16) NOT NULL COMMENT '操作符',
  `threshold` double NOT NULL COMMENT '阈值',
  `duration` int DEFAULT 60 COMMENT '持续时间(秒)',
  `level` varchar(16) DEFAULT 'warning',
  `notify_channels` varchar(255) DEFAULT NULL,
  `status` tinyint DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警规则表';

CREATE TABLE IF NOT EXISTS `alerts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `rule_id` bigint unsigned DEFAULT NULL,
  `server_id` bigint unsigned DEFAULT NULL,
  `level` varchar(16) DEFAULT 'warning',
  `title` varchar(255) NOT NULL,
  `message` text,
  `metric_value` double DEFAULT NULL,
  `status` varchar(16) DEFAULT 'firing',
  `fired_at` datetime DEFAULT NULL,
  `resolved_at` datetime DEFAULT NULL,
  `resolved_by` varchar(64) DEFAULT NULL,
  `remark` text,
  PRIMARY KEY (`id`),
  KEY `idx_server_id` (`server_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';

-- ==================== 初始化数据 ====================

-- 管理员用户 (密码: admin123)
INSERT INTO `users` (`username`, `password`, `nick_name`, `email`, `role`, `status`) VALUES
('admin', 'e10adc3949ba59abbe56e057f20f883e', '系统管理员', 'admin@example.com', 'admin', 1);

-- 服务器分组
INSERT INTO `server_groups` (`name`, `description`) VALUES
('默认分组', '默认服务器分组'),
('生产环境', '生产环境服务器'),
('测试环境', '测试环境服务器');

-- 告警规则
INSERT INTO `alert_rules` (`name`, `description`, `metric`, `operator`, `threshold`, `duration`, `level`, `status`) VALUES
('CPU使用率告警', 'CPU使用率超过80%', 'cpu_usage', '>', 80, 60, 'warning', 1),
('CPU使用率严重告警', 'CPU使用率超过95%', 'cpu_usage', '>', 95, 30, 'critical', 1),
('内存使用率告警', '内存使用率超过85%', 'memory_usage', '>', 85, 60, 'warning', 1),
('磁盘使用率告警', '磁盘使用率超过90%', 'disk_usage', '>', 90, 60, 'critical', 1);
