-- 添加 ai_decisions 表缺失的列
-- 执行时间: 2026-02-28

-- 添加 alert_id 列
ALTER TABLE ai_decisions ADD COLUMN alert_id INT UNSIGNED DEFAULT 0;

-- 添加索引
ALTER TABLE ai_decisions ADD INDEX idx_alert_id (alert_id);
