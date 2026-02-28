-- SSH 密钥管理表
CREATE TABLE IF NOT EXISTS ssh_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    name VARCHAR(64) NOT NULL COMMENT '密钥名称',
    filename VARCHAR(128) COMMENT '原始文件名',
    key_content TEXT NOT NULL COMMENT 'PEM私钥内容',
    passphrase VARCHAR(128) COMMENT '密钥密码',
    fingerprint VARCHAR(64) COMMENT '密钥指纹',
    description VARCHAR(255) COMMENT '描述'
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ssh_keys_deleted_at ON ssh_keys(deleted_at);
CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON ssh_keys(fingerprint);

-- 为 servers 表添加 SSH 密钥相关字段
-- SQLite 不支持直接的 ALTER TABLE ADD COLUMN IF NOT EXISTS，需要手动检查

-- 添加 auth_type 字段
ALTER TABLE servers ADD COLUMN auth_type VARCHAR(16) DEFAULT 'password' COMMENT '认证方式';

-- 添加 ssh_key_id 字段
ALTER TABLE servers ADD COLUMN ssh_key_id INTEGER COMMENT 'SSH密钥ID';
