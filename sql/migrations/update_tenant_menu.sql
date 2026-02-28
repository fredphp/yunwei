-- 更新租户管理菜单结构
-- 1. 将租户管理改为 Layout 类型
UPDATE sys_menus SET component = 'Layout' WHERE name = 'Tenant' AND component != 'Layout';

-- 2. 添加租户管理子菜单（使用 INSERT IGNORE 避免重复）
-- 先获取租户管理菜单的ID
SET @tenant_parent_id = (SELECT id FROM sys_menus WHERE name = 'Tenant' LIMIT 1);

-- 插入子菜单（如果不存在）
INSERT IGNORE INTO sys_menus (parent_id, title, name, path, component, icon, sort, status, hidden, created_at, updated_at)
SELECT @tenant_parent_id, '租户列表', 'TenantList', '/tenant/list', 'views/tenant/list/index', 'List', 1, 1, 0, NOW(), NOW()
FROM DUAL WHERE @tenant_parent_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'TenantList');

INSERT IGNORE INTO sys_menus (parent_id, title, name, path, component, icon, sort, status, hidden, created_at, updated_at)
SELECT @tenant_parent_id, '套餐管理', 'TenantPlan', '/tenant/plan', 'views/tenant/plan/index', 'PriceTag', 2, 1, 0, NOW(), NOW()
FROM DUAL WHERE @tenant_parent_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'TenantPlan');

INSERT IGNORE INTO sys_menus (parent_id, title, name, path, component, icon, sort, status, hidden, created_at, updated_at)
SELECT @tenant_parent_id, '账单管理', 'TenantBilling', '/tenant/billing', 'views/tenant/billing/index', 'Wallet', 3, 1, 0, NOW(), NOW()
FROM DUAL WHERE @tenant_parent_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'TenantBilling');

INSERT IGNORE INTO sys_menus (parent_id, title, name, path, component, icon, sort, status, hidden, created_at, updated_at)
SELECT @tenant_parent_id, '审计日志', 'TenantAudit', '/tenant/audit', 'views/tenant/audit/index', 'Document', 4, 1, 0, NOW(), NOW()
FROM DUAL WHERE @tenant_parent_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'TenantAudit');

-- 3. 更新系统管理的子菜单父ID（动态获取系统管理ID）
SET @system_parent_id = (SELECT id FROM sys_menus WHERE name = 'System' LIMIT 1);
UPDATE sys_menus SET parent_id = @system_parent_id WHERE name IN ('UserManage', 'RoleManage', 'MenuManage') AND @system_parent_id IS NOT NULL;

SELECT '租户菜单更新完成!' AS message;
