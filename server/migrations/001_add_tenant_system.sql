-- 多租户系统数据库迁移脚本
-- 为现有表添加 tenant_id 字段实现数据隔离

-- 1. 租户相关表（使用 CREATE TABLE IF NOT EXISTS 避免重复创建）
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    domain VARCHAR(255),
    logo VARCHAR(500),
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    plan VARCHAR(50) DEFAULT 'free',
    billing_cycle VARCHAR(20),
    contact_name VARCHAR(100),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    address TEXT,
    settings JSON,
    features JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE INDEX idx_name (name),
    UNIQUE INDEX idx_slug (slug)
);

CREATE TABLE IF NOT EXISTS tenant_quotas (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    max_users INT DEFAULT 5,
    max_admins INT DEFAULT 2,
    max_resources INT DEFAULT 100,
    max_servers INT DEFAULT 50,
    max_databases INT DEFAULT 20,
    max_monitors INT DEFAULT 100,
    max_alert_rules INT DEFAULT 50,
    metrics_retention INT DEFAULT 30,
    max_cloud_accounts INT DEFAULT 5,
    budget_limit DECIMAL(15,2) DEFAULT 0,
    max_storage_gb INT DEFAULT 100,
    max_backup_gb INT DEFAULT 500,
    max_api_calls INT DEFAULT 10000,
    max_webhooks INT DEFAULT 10,
    current_users INT DEFAULT 0,
    current_resources INT DEFAULT 0,
    current_storage_gb INT DEFAULT 0,
    current_api_calls INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_tenant_id (tenant_id)
);

CREATE TABLE IF NOT EXISTS tenant_users (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    email VARCHAR(255),
    name VARCHAR(100),
    avatar VARCHAR(500),
    role_id VARCHAR(36),
    role_name VARCHAR(50),
    is_owner BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active',
    invited_by VARCHAR(36),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_email (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS tenant_roles (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(50) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    permissions JSON,
    scope VARCHAR(20) DEFAULT 'tenant',
    parent_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_slug (tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS tenant_invitations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role_id VARCHAR(36),
    role_name VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    token VARCHAR(64),
    invited_by VARCHAR(36),
    inviter_name VARCHAR(100),
    expires_at TIMESTAMP NULL,
    accepted_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_email (tenant_id, email),
    UNIQUE INDEX idx_token (token)
);

CREATE TABLE IF NOT EXISTS tenant_resource_usage (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    date DATE NOT NULL,
    hour INT DEFAULT -1,
    user_count INT DEFAULT 0,
    active_users INT DEFAULT 0,
    resource_count INT DEFAULT 0,
    server_count INT DEFAULT 0,
    database_count INT DEFAULT 0,
    monitor_count INT DEFAULT 0,
    alert_count INT DEFAULT 0,
    metrics_data_mb INT DEFAULT 0,
    total_cost DECIMAL(15,2) DEFAULT 0,
    cloud_cost DECIMAL(15,2) DEFAULT 0,
    storage_used_mb INT DEFAULT 0,
    backup_used_mb INT DEFAULT 0,
    api_calls INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_date (tenant_id, date)
);

CREATE TABLE IF NOT EXISTS tenant_billings (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    billing_period VARCHAR(20) NOT NULL,
    due_date TIMESTAMP NULL,
    base_amount DECIMAL(15,2) DEFAULT 0,
    usage_amount DECIMAL(15,2) DEFAULT 0,
    overage_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    usage_details JSON,
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_id VARCHAR(100),
    paid_at TIMESTAMP NULL,
    invoice_number VARCHAR(50),
    invoice_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_period (tenant_id, billing_period)
);

CREATE TABLE IF NOT EXISTS tenant_audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    user_name VARCHAR(100),
    user_email VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id VARCHAR(36),
    resource_name VARCHAR(255),
    old_value JSON,
    new_value JSON,
    changes JSON,
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    request_id VARCHAR(36),
    status VARCHAR(20) DEFAULT 'success',
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_action (tenant_id, action),
    INDEX idx_tenant_resource (tenant_id, resource, resource_id),
    INDEX idx_created_at (created_at)
);

-- 2. 为现有业务表添加 tenant_id 字段（使用存储过程检查字段是否存在）
-- 注意：以下语句会自动忽略已存在的字段

-- 服务器表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'servers' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE servers ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 告警表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'alerts' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE alerts ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Kubernetes集群表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'k8s_clusters' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE k8s_clusters ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 灰度发布表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'canary_releases' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE canary_releases ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 负载均衡表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'load_balancers' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE load_balancers ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 证书表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'certificates' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE certificates ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- CDN域名表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'cdn_domains' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE cdn_domains ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Agent表
SET @exist := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'agents' AND column_name = 'tenant_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE agents ADD COLUMN tenant_id VARCHAR(36)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. 创建默认租户（用于迁移现有数据）
INSERT IGNORE INTO tenants (id, name, slug, status, plan, contact_email, contact_name)
VALUES ('default', 'Default Tenant', 'default', 'active', 'enterprise', 'admin@example.com', 'Admin');

-- 4. 为现有数据设置默认租户
UPDATE servers SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE alerts SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE k8s_clusters SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE canary_releases SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE load_balancers SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE certificates SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE cdn_domains SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE agents SET tenant_id = 'default' WHERE tenant_id IS NULL;

SELECT '多租户系统迁移完成!' AS message;
