-- 更新租户管理菜单结构
-- 1. 将租户管理改为 Layout 类型
UPDATE sys_menus SET component = 'Layout' WHERE name = 'Tenant';

-- 2. 添加租户管理子菜单
-- 先获取租户管理菜单的ID
SET @tenant_parent_id = (SELECT id FROM sys_menus WHERE name = 'Tenant' LIMIT 1);

-- 插入子菜单（如果不存在）
INSERT IGNORE INTO sys_menus (parent_id, title, name, path, component, icon, sort, status, hidden, created_at, updated_at)
VALUES 
(@tenant_parent_id, '租户列表', 'TenantList', '/tenant/list', 'views/tenant/list/index', 'List', 1, 1, 0, NOW(), NOW()),
(@tenant_parent_id, '套餐管理', 'TenantPlan', '/tenant/plan', 'views/tenant/plan/index', 'PriceTag', 2, 1, 0, NOW(), NOW()),
(@tenant_parent_id, '账单管理', 'TenantBilling', '/tenant/billing', 'views/tenant/billing/index', 'Wallet', 3, 1, 0, NOW(), NOW()),
(@tenant_parent_id, '审计日志', 'TenantAudit', '/tenant/audit', 'views/tenant/audit/index', 'Document', 4, 1, 0, NOW(), NOW());
