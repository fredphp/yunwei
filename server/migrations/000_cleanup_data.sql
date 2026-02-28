-- 数据清理迁移脚本
-- 清理重复数据，确保数据一致性
-- 执行时间: 2026-02-28

USE `yunwei`;

-- ==================== 1. 清理重复菜单数据 ====================

-- 创建临时表存储需要保留的菜单ID（保留ID最小的）
CREATE TEMPORARY TABLE IF NOT EXISTS menus_to_keep AS
SELECT MIN(id) as id
FROM sys_menus
WHERE deleted_at IS NULL
GROUP BY parent_id, name;

-- 删除重复菜单（保留ID最小的）
DELETE FROM sys_role_menus WHERE menu_id IN (
    SELECT id FROM sys_menus 
    WHERE id NOT IN (SELECT id FROM menus_to_keep) 
    AND deleted_at IS NULL
);

DELETE FROM sys_menus 
WHERE id NOT IN (SELECT id FROM menus_to_keep)
AND deleted_at IS NULL;

-- 删除临时表
DROP TEMPORARY TABLE IF EXISTS menus_to_keep;

-- ==================== 2. 修复租户domain唯一索引问题 ====================

-- 删除旧的唯一索引（如果存在）
SET @exist_idx := (SELECT COUNT(1) FROM information_schema.statistics
    WHERE table_schema = 'yunwei' AND table_name = 'tenants' AND index_name = 'idx_tenants_domain');
SET @sql := IF(@exist_idx > 0,
    'ALTER TABLE tenants DROP INDEX idx_tenants_domain',
    'SELECT "索引不存在，跳过" AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 为空 domain 设置唯一值（使用 slug 作为子域名）
UPDATE tenants SET domain = CONCAT(slug, '.example.com') WHERE domain = '' OR domain IS NULL;

-- ==================== 3. 清理重复的角色-API关联 ====================

DELETE t1 FROM sys_role_apis t1
INNER JOIN sys_role_apis t2
WHERE t1.id > t2.id
AND t1.role_id = t2.role_id
AND t1.api_id = t2.api_id;

-- ==================== 4. 清理重复的角色-菜单关联 ====================

DELETE t1 FROM sys_role_menus t1
INNER JOIN sys_role_menus t2
WHERE t1.id > t2.id
AND t1.role_id = t2.role_id
AND t1.menu_id = t2.menu_id;

-- ==================== 5. 确保菜单结构正确 ====================

-- 更新服务器管理菜单为 Layout 容器
UPDATE sys_menus SET component = 'Layout', path = '/servers' 
WHERE name = 'Servers' AND parent_id = 0;

-- 更新 Kubernetes 菜单为 Layout 容器
UPDATE sys_menus SET component = 'Layout', path = '/kubernetes' 
WHERE name = 'Kubernetes' AND parent_id = 0;

-- 更新系统管理菜单为 Layout 容器
UPDATE sys_menus SET component = 'Layout', path = '/system' 
WHERE name = 'System' AND parent_id = 0;

-- ==================== 6. 清理迁移记录表（重新执行迁移）====================

-- 清空迁移记录，让系统重新执行所有迁移
DELETE FROM sys_migrations WHERE 1=1;

-- 完成提示
SELECT '数据清理完成!' AS message;
