-- 更新告警菜单配置
-- 将告警中心作为独立的一级菜单

USE `yunwei`;

-- 1. 删除旧的告警菜单（在服务器管理下面的）
DELETE FROM sys_menus WHERE name = 'Alerts' AND parent_id = 2;

-- 2. 更新或插入告警中心作为一级菜单
INSERT INTO sys_menus (`id`, `parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
VALUES (4, 0, '告警中心', 'Alerts', '/alerts', 'views/alerts/index.vue', 'Bell', 4, 1, 0)
ON DUPLICATE KEY UPDATE 
    `parent_id` = 0,
    `title` = '告警中心',
    `path` = '/alerts',
    `component` = 'views/alerts/index.vue',
    `icon` = 'Bell',
    `sort` = 4;

-- 3. 更新角色菜单关联（确保告警中心菜单有权限）
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) VALUES
(1, 4),
(2, 4),
(3, 4);

-- 4. 更新系统管理的子菜单父ID（如果系统管理ID是10）
UPDATE sys_menus SET parent_id = 10 WHERE name IN ('UserManage', 'RoleManage', 'MenuManage', 'ApiManage') AND parent_id != 10;

-- 完成提示
SELECT '告警菜单更新完成!' AS message;
