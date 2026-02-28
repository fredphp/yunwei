-- 添加 detect_rules 表缺失的列
-- 执行时间: 2026-02-28

-- 添加 action_command 列
ALTER TABLE detect_rules ADD COLUMN action_command TEXT;

-- 添加 auto_action 列
ALTER TABLE detect_rules ADD COLUMN auto_action TINYINT(1) DEFAULT 0;

-- 添加 count 列
ALTER TABLE detect_rules ADD COLUMN count INT DEFAULT 0;

-- 添加 duration 列
ALTER TABLE detect_rules ADD COLUMN duration INT DEFAULT 0;

-- 添加 enabled 列
ALTER TABLE detect_rules ADD COLUMN enabled TINYINT(1) DEFAULT 1;

-- 添加 level 列
ALTER TABLE detect_rules ADD COLUMN level VARCHAR(16);

-- 添加 description 列
ALTER TABLE detect_rules ADD COLUMN description VARCHAR(255);

-- 添加 threshold 列
ALTER TABLE detect_rules ADD COLUMN threshold DOUBLE DEFAULT 0;

-- 添加 type 列
ALTER TABLE detect_rules ADD COLUMN type VARCHAR(32);

-- 添加通知相关列
ALTER TABLE detect_rules ADD COLUMN notify_email TINYINT(1) DEFAULT 0;
ALTER TABLE detect_rules ADD COLUMN notify_sms TINYINT(1) DEFAULT 0;
ALTER TABLE detect_rules ADD COLUMN notify_webhook TINYINT(1) DEFAULT 0;
ALTER TABLE detect_rules ADD COLUMN webhook_url VARCHAR(255);
