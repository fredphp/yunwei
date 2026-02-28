-- 清理重复数据脚本
-- 1. 删除重复的菜单数据
-- 2. 添加唯一索引
-- 3. 清理其他重复数据

USE `yunwei`;

-- ==================== 1. 清理重复菜单数据 ====================

-- 创建临时表存储需要保留的菜单ID
CREATE TEMPORARY TABLE menus_to_keep AS
SELECT MIN(id) as id
FROM sys_menus
WHERE deleted_at IS NULL
GROUP BY parent_id, name, path;

-- 删除重复菜单（保留ID最小的）
DELETE FROM sys_menus
WHERE id NOT IN (SELECT id FROM menus_to_keep)
AND deleted_at IS NULL;

-- 删除临时表
DROP TEMPORARY TABLE IF EXISTS menus_to_keep;

-- ==================== 2. 添加唯一索引 ====================

-- 检查并添加菜单表的唯一索引 (name字段唯一)
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'sys_menus' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE sys_menus ADD UNIQUE INDEX idx_name_unique (name, parent_id)',
    'SELECT "索引 idx_name_unique 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加API表的唯一索引 (path + method)
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'sys_apis' AND index_name = 'idx_path_method');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE sys_apis ADD UNIQUE INDEX idx_path_method (path, method)',
    'SELECT "索引 idx_path_method 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ==================== 3. 清理角色-API关联表重复 ====================

-- 删除重复的角色-API关联
DELETE t1 FROM sys_role_apis t1
INNER JOIN sys_role_apis t2
WHERE t1.id > t2.id
AND t1.role_id = t2.role_id
AND t1.api_id = t2.api_id;

-- 添加唯一索引
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'sys_role_apis' AND index_name = 'idx_role_api_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE sys_role_apis ADD UNIQUE INDEX idx_role_api_unique (role_id, api_id)',
    'SELECT "索引 idx_role_api_unique 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ==================== 4. 清理角色-菜单关联表重复 ====================

-- 删除重复的角色-菜单关联
DELETE t1 FROM sys_role_menus t1
INNER JOIN sys_role_menus t2
WHERE t1.id > t2.id
AND t1.role_id = t2.role_id
AND t1.menu_id = t2.menu_id;

-- 添加唯一索引
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'sys_role_menus' AND index_name = 'idx_role_menu_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE sys_role_menus ADD UNIQUE INDEX idx_role_menu_unique (role_id, menu_id)',
    'SELECT "索引 idx_role_menu_unique 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ==================== 5. 清理告警规则重复 ====================

-- 删除重复的告警规则
DELETE t1 FROM alert_rules t1
INNER JOIN alert_rules t2
WHERE t1.id > t2.id
AND t1.name = t2.name;

-- 添加唯一索引
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'alert_rules' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE alert_rules ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 alert_rules.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ==================== 6. 清理其他关键表的重复 ====================

-- 负载均衡器名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'load_balancers' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE load_balancers ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 load_balancers.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- CDN域名唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'cdn_domains' AND index_name = 'idx_domain_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE cdn_domains ADD UNIQUE INDEX idx_domain_unique (domain)',
    'SELECT "索引 cdn_domains.domain 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 证书域名唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'certificates' AND index_name = 'idx_domain_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE certificates ADD UNIQUE INDEX idx_domain_unique (domain)',
    'SELECT "索引 certificates.domain 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 部署计划名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'deploy_plans' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE deploy_plans ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 deploy_plans.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 自愈规则名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'heal_rules' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE heal_rules ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 heal_rules.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 安全规则名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'security_rules' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE security_rules ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 security_rules.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检测规则名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'detect_rules' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE detect_rules ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 detect_rules.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 资源池名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'resource_pools' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE resource_pools ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 resource_pools.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Agent配置名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'agent_configs' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE agent_configs ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 agent_configs.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 灰度发布策略名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'gray_release_strategies' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE gray_release_strategies ADD UNIQUE INDEX idx_name_unique (name)',
    'SELECT "索引 gray_release_strategies.name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 金丝雀发布名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'canary_releases' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE canary_releases ADD UNIQUE INDEX idx_name_unique (service_name, namespace)',
    'SELECT "索引 canary_releases.service_name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 金丝雀配置名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'canary_configs' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE canary_configs ADD UNIQUE INDEX idx_name_unique (service_name, namespace)',
    'SELECT "索引 canary_configs.service_name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 项目分析名称唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'project_analyses' AND index_name = 'idx_name_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE project_analyses ADD UNIQUE INDEX idx_name_unique (project_name)',
    'SELECT "索引 project_analyses.project_name 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 服务器能力server_id唯一
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'server_capabilities' AND index_name = 'idx_server_unique');
SET @sql := IF(@exist_idx = 0,
    'ALTER TABLE server_capabilities ADD UNIQUE INDEX idx_server_unique (server_id)',
    'SELECT "索引 server_capabilities.server_id 已存在" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 完成提示
SELECT '重复数据清理完成，唯一索引添加完成!' AS message;
