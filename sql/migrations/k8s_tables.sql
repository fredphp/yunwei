-- Kubernetes 相关表增量迁移脚本
-- 用于创建缺失的 K8s 相关表

USE `yunwei`;

-- K8s 集群表
CREATE TABLE IF NOT EXISTS `k8s_clusters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) DEFAULT NULL COMMENT '集群名称',
  `api_endpoint` varchar(256) DEFAULT NULL COMMENT 'API地址',
  `token` text COMMENT 'ServiceAccount Token',
  `kube_config` text COMMENT 'Kubeconfig内容',
  `status` varchar(16) DEFAULT NULL COMMENT '状态: connected/disconnected/error',
  `version` varchar(32) DEFAULT NULL COMMENT 'K8s版本',
  `node_count` int DEFAULT 0 COMMENT '节点数量',
  `auto_scale_enabled` tinyint DEFAULT 0 COMMENT '启用自动扩容',
  `min_replicas` int DEFAULT 1 COMMENT '最小副本数',
  `max_replicas` int DEFAULT 10 COMMENT '最大副本数',
  `cpu_threshold` double DEFAULT 80 COMMENT 'CPU扩容阈值',
  `mem_threshold` double DEFAULT 80 COMMENT '内存扩容阈值',
  `last_sync_at` datetime DEFAULT NULL COMMENT '最后同步时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s集群表';

-- K8s 扩容事件表
CREATE TABLE IF NOT EXISTS `k8s_scale_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `cluster_id` bigint unsigned DEFAULT NULL COMMENT '集群ID',
  `namespace` varchar(64) DEFAULT NULL COMMENT '命名空间',
  `deployment` varchar(128) DEFAULT NULL COMMENT 'Deployment名称',
  `scale_type` varchar(16) DEFAULT NULL COMMENT '扩容类型: horizontal/vertical/manual/auto',
  `status` varchar(16) DEFAULT NULL COMMENT '状态: pending/running/success/failed/rollback',
  `replicas_before` int DEFAULT 0 COMMENT '扩容前副本数',
  `replicas_after` int DEFAULT 0 COMMENT '扩容后副本数',
  `replicas_target` int DEFAULT 0 COMMENT '目标副本数',
  `trigger_reason` text COMMENT '触发原因',
  `trigger_metric` text COMMENT '触发指标JSON',
  `ai_decision` text COMMENT 'AI决策',
  `ai_confidence` double DEFAULT 0 COMMENT 'AI置信度',
  `ai_auto_approve` tinyint DEFAULT 0 COMMENT 'AI自动批准',
  `commands` text COMMENT '执行命令',
  `execution_log` text COMMENT '执行日志',
  `error_message` text COMMENT '错误信息',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `duration` bigint DEFAULT 0 COMMENT '耗时(毫秒)',
  PRIMARY KEY (`id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s扩容事件表';

-- K8s HPA 配置表
CREATE TABLE IF NOT EXISTS `k8s_hpa_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cluster_id` bigint unsigned DEFAULT NULL COMMENT '集群ID',
  `namespace` varchar(64) DEFAULT NULL COMMENT '命名空间',
  `deployment` varchar(128) DEFAULT NULL COMMENT 'Deployment名称',
  `min_replicas` int DEFAULT 1 COMMENT '最小副本数',
  `max_replicas` int DEFAULT 10 COMMENT '最大副本数',
  `target_cpu_util` double DEFAULT 80 COMMENT '目标CPU使用率',
  `target_mem_util` double DEFAULT 80 COMMENT '目标内存使用率',
  `custom_metrics` text COMMENT '自定义指标JSON',
  `scale_up_stabilization` int DEFAULT 300 COMMENT '扩容稳定窗口(秒)',
  `scale_down_stabilization` int DEFAULT 300 COMMENT '缩容稳定窗口(秒)',
  `enabled` tinyint DEFAULT 0 COMMENT '是否启用',
  PRIMARY KEY (`id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s HPA配置表';

-- K8s Deployment 状态表
CREATE TABLE IF NOT EXISTS `k8s_deployment_status` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `cluster_id` bigint unsigned DEFAULT NULL COMMENT '集群ID',
  `namespace` varchar(64) DEFAULT NULL COMMENT '命名空间',
  `deployment` varchar(128) DEFAULT NULL COMMENT 'Deployment名称',
  `replicas` int DEFAULT 0 COMMENT '副本数',
  `ready_replicas` int DEFAULT 0 COMMENT '就绪副本数',
  `updated_replicas` int DEFAULT 0 COMMENT '更新副本数',
  `cpu_usage` double DEFAULT 0 COMMENT 'CPU使用率',
  `memory_usage` double DEFAULT 0 COMMENT '内存使用率',
  `cpu_request` varchar(32) DEFAULT NULL COMMENT 'CPU请求',
  `memory_request` varchar(32) DEFAULT NULL COMMENT '内存请求',
  `cpu_limit` varchar(32) DEFAULT NULL COMMENT 'CPU限制',
  `memory_limit` varchar(32) DEFAULT NULL COMMENT '内存限制',
  `hpa_enabled` tinyint DEFAULT 0 COMMENT 'HPA启用',
  `hpa_target_cpu` double DEFAULT 0 COMMENT 'HPA目标CPU',
  `current_replicas` int DEFAULT 0 COMMENT '当前副本数',
  `desired_replicas` int DEFAULT 0 COMMENT '期望副本数',
  PRIMARY KEY (`id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s Deployment状态表';
