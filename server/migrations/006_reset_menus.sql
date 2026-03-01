-- 一键清理脚本：彻底清理所有旧菜单并重新初始化
-- 警告：此脚本会删除所有菜单数据，请谨慎使用

-- 1. 删除角色-菜单关联
DELETE FROM sys_role_menus;

-- 2. 删除所有菜单
DELETE FROM sys_menus;

-- 3. 重置自增ID
ALTER TABLE sys_menus AUTO_INCREMENT = 1;
ALTER TABLE sys_role_menus AUTO_INCREMENT = 1;

-- 4. 重新插入一级菜单
INSERT INTO `sys_menus` (`id`, `parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`) VALUES
(1, 0, '仪表盘', 'Dashboard', '/dashboard', 'views/dashboard/index', 'Odometer', 1, 1, 0),
(2, 0, '服务器管理', 'Servers', '/servers', 'Layout', 'Monitor', 2, 1, 0),
(3, 0, 'Kubernetes', 'Kubernetes', '/kubernetes', 'Layout', 'Grid', 3, 1, 0),
(4, 0, '告警中心', 'Alerts', '/alerts', 'views/alerts/index', 'Bell', 4, 1, 0),
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
(15, 0, '租户管理', 'Tenant', '/tenant', 'Layout', 'OfficeBuilding', 15, 1, 0),
(16, 0, '系统管理', 'System', '/system', 'Layout', 'Setting', 99, 1, 0);

-- 5. 插入二级菜单
INSERT INTO `sys_menus` (`id`, `parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`) VALUES
-- 服务器管理子菜单 (ParentID: 2)
(100, 2, '服务器列表', 'ServerList', '/servers/list', 'views/servers/index', 'List', 1, 1, 0),

-- Kubernetes子菜单 (ParentID: 3)
(200, 3, '集群管理', 'Clusters', '/kubernetes/clusters', 'views/kubernetes/index', 'Cluster', 1, 1, 0),

-- 租户管理子菜单 (ParentID: 15)
(301, 15, '租户列表', 'TenantList', '/tenant/list', 'views/tenant/list/index', 'List', 1, 1, 0),
(302, 15, '套餐管理', 'TenantPlan', '/tenant/plan', 'views/tenant/plan/index', 'PriceTag', 2, 1, 0),
(303, 15, '账单管理', 'TenantBilling', '/tenant/billing', 'views/tenant/billing/index', 'Wallet', 3, 1, 0),
(304, 15, '审计日志', 'TenantAudit', '/tenant/audit', 'views/tenant/audit/index', 'Document', 4, 1, 0),

-- 系统管理子菜单 (ParentID: 16)
(401, 16, '用户管理', 'UserManage', '/system/user', 'views/system/user/index', 'User', 1, 1, 0),
(402, 16, '角色管理', 'RoleManage', '/system/role', 'views/system/role/index', 'UserFilled', 2, 1, 0),
(403, 16, '菜单管理', 'MenuManage', '/system/menu', 'views/system/menu/index', 'Menu', 3, 1, 0);

-- 6. 为超级管理员分配所有菜单权限
INSERT INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT r.id, m.id FROM `sys_roles` r, `sys_menus` m WHERE r.keyword = 'admin';
