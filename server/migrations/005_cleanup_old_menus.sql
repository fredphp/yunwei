-- 清理 FastAdmin/旧系统残留的菜单数据
-- 这些菜单路径包含 .php 或其他非 Vue 路由格式

-- 1. 删除所有旧的菜单数据（包括 FastAdmin 残留）
DELETE FROM sys_role_menus WHERE menu_id IN (SELECT id FROM sys_menus WHERE path LIKE '%.php%' OR path LIKE '%.php%/%' OR component LIKE '%.php%');
DELETE FROM sys_menus WHERE path LIKE '%.php%' OR path LIKE '%.php%/%' OR component LIKE '%.php%';

-- 2. 清空所有菜单数据，准备重新初始化
-- 注意：如果您想保留自定义菜单，请不要执行以下语句
-- DELETE FROM sys_role_menus;
-- DELETE FROM sys_menus;

-- 3. 重新插入正确的 Vue.js 路由格式的菜单数据
-- 检查并插入一级菜单（如果不存在）
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '仪表盘', 'Dashboard', '/dashboard', 'views/dashboard/index', 'Odometer', 1, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Dashboard' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '服务器管理', 'Servers', '/servers', 'Layout', 'Monitor', 2, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Servers' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '告警中心', 'Alerts', '/alerts', 'views/alerts/index', 'Bell', 3, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Alerts' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, 'Kubernetes', 'Kubernetes', '/kubernetes', 'Layout', 'Grid', 4, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Kubernetes' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '灰度发布', 'Canary', '/canary', 'Layout', 'Promotion', 5, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Canary' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '负载均衡', 'LoadBalancer', '/loadbalancer', 'Layout', 'Connection', 6, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'LoadBalancer' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '证书管理', 'Certificate', '/certificate', 'Layout', 'DocumentChecked', 7, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Certificate' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, 'CDN管理', 'CDN', '/cdn', 'Layout', 'Position', 8, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'CDN' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '智能部署', 'Deploy', '/deploy', 'Layout', 'Upload', 9, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Deploy' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '任务调度', 'Scheduler', '/scheduler', 'Layout', 'Timer', 10, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Scheduler' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, 'Agent管理', 'Agents', '/agents', 'Layout', 'Cpu', 11, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Agents' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '高可用', 'HA', '/ha', 'Layout', 'CircleCheck', 12, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'HA' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '灾备备份', 'Backup', '/backup', 'Layout', 'Files', 13, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Backup' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '成本控制', 'Cost', '/cost', 'Layout', 'Coin', 14, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Cost' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '租户管理', 'Tenant', '/tenant', 'Layout', 'OfficeBuilding', 15, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Tenant' AND `parent_id` = 0);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 0, '系统管理', 'System', '/system', 'Layout', 'Setting', 99, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'System' AND `parent_id` = 0);

-- 插入二级菜单（服务器管理子菜单）
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 2, '服务器列表', 'ServerList', '/servers/list', 'views/servers/index', 'List', 1, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'ServerList' AND `parent_id` = 2);

-- 插入二级菜单（Kubernetes子菜单）
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 4, '集群管理', 'Clusters', '/kubernetes/clusters', 'views/kubernetes/index', 'Cluster', 1, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'Clusters' AND `parent_id` = 4);

-- 插入二级菜单（租户管理子菜单）
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 15, '租户列表', 'TenantList', '/tenant/list', 'views/tenant/list/index', 'List', 1, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'TenantList' AND `parent_id` = 15);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 15, '套餐管理', 'TenantPlan', '/tenant/plan', 'views/tenant/plan/index', 'PriceTag', 2, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'TenantPlan' AND `parent_id` = 15);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 15, '账单管理', 'TenantBilling', '/tenant/billing', 'views/tenant/billing/index', 'Wallet', 3, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'TenantBilling' AND `parent_id` = 15);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 15, '审计日志', 'TenantAudit', '/tenant/audit', 'views/tenant/audit/index', 'Document', 4, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'TenantAudit' AND `parent_id` = 15);

-- 插入二级菜单（系统管理子菜单）
INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 16, '用户管理', 'UserManage', '/system/user', 'views/system/user/index', 'User', 1, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'UserManage' AND `parent_id` = 16);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 16, '角色管理', 'RoleManage', '/system/role', 'views/system/role/index', 'UserFilled', 2, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'RoleManage' AND `parent_id` = 16);

INSERT INTO `sys_menus` (`parent_id`, `title`, `name`, `path`, `component`, `icon`, `sort`, `status`, `hidden`)
SELECT 16, '菜单管理', 'MenuManage', '/system/menu', 'views/system/menu/index', 'Menu', 3, 1, 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `name` = 'MenuManage' AND `parent_id` = 16);

-- 4. 为超级管理员角色分配所有菜单权限
-- 获取超级管理员角色ID和所有菜单ID，然后插入关联
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`)
SELECT r.id, m.id FROM `sys_roles` r, `sys_menus` m WHERE r.keyword = 'admin';
