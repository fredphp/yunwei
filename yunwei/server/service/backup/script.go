package backup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"yunwei/model/backup"
)

// ScriptEngine 恢复脚本引擎
type ScriptEngine struct {
	mu             sync.Mutex
	runningScripts map[string]*ScriptContext
	scriptDir      string
}

// ScriptContext 脚本执行上下文
type ScriptContext struct {
	ExecID    string
	ScriptID  uint
	Status    string
	StartTime time.Time
	Cancelled bool
	Output    strings.Builder
	Error     error
}

// NewScriptEngine 创建脚本引擎
func NewScriptEngine(scriptDir string) *ScriptEngine {
	if scriptDir == "" {
		scriptDir = "/var/lib/backup/scripts"
	}
	os.MkdirAll(scriptDir, 0755)

	return &ScriptEngine{
		runningScripts: make(map[string]*ScriptContext),
		scriptDir:      scriptDir,
	}
}

// ScriptResult 脚本执行结果
type ScriptResult struct {
	Success   bool
	ExecID    string
	Output    string
	Error     string
	ExitCode  int
	Duration  int
	StartTime time.Time
	EndTime   *time.Time
}

// ExecuteScript 执行脚本
func (e *ScriptEngine) ExecuteScript(ctx context.Context, script *backup.RecoveryScript, params map[string]interface{}) (*ScriptResult, error) {
	execID := fmt.Sprintf("exec_%d_%d", script.ID, time.Now().UnixNano())

	e.mu.Lock()
	scriptCtx := &ScriptContext{
		ExecID:    execID,
		ScriptID:  script.ID,
		Status:    "running",
		StartTime: time.Now(),
	}
	e.runningScripts[execID] = scriptCtx
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.runningScripts, execID)
		e.mu.Unlock()
	}()

	result := &ScriptResult{
		ExecID:    execID,
		StartTime: scriptCtx.StartTime,
	}

	// 准备脚本内容
	scriptContent := script.Script

	// 替换参数
	for key, val := range params {
		jsonVal, _ := json.Marshal(val)
		placeholder := fmt.Sprintf("{{.%s}}", key)
		scriptContent = strings.ReplaceAll(scriptContent, placeholder, string(jsonVal))
		// 也支持不带大括号的格式
		scriptContent = strings.ReplaceAll(scriptContent, fmt.Sprintf("$%s", key), fmt.Sprintf("%v", val))
	}

	// 创建临时脚本文件
	tempScript := filepath.Join(e.scriptDir, fmt.Sprintf("%s.sh", execID))

	// 根据语言选择执行方式
	var cmd *exec.Cmd
	switch script.Language {
	case "bash", "sh":
		// 写入脚本文件
		if err := os.WriteFile(tempScript, []byte(scriptContent), 0755); err != nil {
			result.Error = err.Error()
			result.Success = false
			return result, err
		}
		defer os.Remove(tempScript)

		cmd = exec.CommandContext(ctx, "/bin/bash", tempScript)

	case "python", "python3":
		if err := os.WriteFile(tempScript, []byte(scriptContent), 0644); err != nil {
			result.Error = err.Error()
			result.Success = false
			return result, err
		}
		defer os.Remove(tempScript)

		cmd = exec.CommandContext(ctx, "python3", tempScript)

	default:
		// 默认使用 bash
		if err := os.WriteFile(tempScript, []byte(scriptContent), 0755); err != nil {
			result.Error = err.Error()
			result.Success = false
			return result, err
		}
		defer os.Remove(tempScript)

		cmd = exec.CommandContext(ctx, "/bin/bash", tempScript)
	}

	// 设置超时
	if script.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(script.Timeout)*time.Second)
		defer cancel()
	}

	// 设置环境变量
	env := os.Environ()
	for key, val := range params {
		env = append(env, fmt.Sprintf("%s=%v", strings.ToUpper(key), val))
	}
	cmd.Env = env

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行前置脚本
	if script.PreScript != "" {
		preCmd := exec.CommandContext(ctx, "/bin/bash", "-c", script.PreScript)
		if err := preCmd.Run(); err != nil {
			// 前置脚本失败
			if !script.IgnoreError {
				result.Error = fmt.Sprintf("前置脚本失败: %v", err)
				result.Success = false
				return result, fmt.Errorf(result.Error)
			}
		}
	}

	// 执行主脚本
	err := cmd.Run()
	result.Output = stdout.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = fmt.Sprintf("%s\n%s", err.Error(), stderr.String())
		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	// 执行后置脚本
	if script.PostScript != "" {
		postCmd := exec.CommandContext(ctx, "/bin/bash", "-c", script.PostScript)
		if err := postCmd.Run(); err != nil {
			// 后置脚本失败只记录警告
			result.Output += fmt.Sprintf("\n[WARN] 后置脚本失败: %v", err)
		}
	}

	// 执行验证脚本
	if script.VerifyScript != "" && result.Success {
		verifyResult := e.executeVerifyScript(ctx, script.VerifyScript, params)
		if !verifyResult.Success {
			result.Success = false
			result.Error = fmt.Sprintf("验证失败: %s", verifyResult.Error)
		}
	}

	// 设置结束时间
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = int(endTime.Sub(result.StartTime).Seconds())

	return result, nil
}

// executeVerifyScript 执行验证脚本
func (e *ScriptEngine) executeVerifyScript(ctx context.Context, verifyScript string, params map[string]interface{}) *ScriptResult {
	result := &ScriptResult{}

	// 替换参数
	for key, val := range params {
		verifyScript = strings.ReplaceAll(verifyScript, fmt.Sprintf("{{.%s}}", key), fmt.Sprintf("%v", val))
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", verifyScript)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Success = false
		result.Error = stderr.String()
		return result
	}

	result.Success = true
	result.Output = stdout.String()
	return result
}

// ExecuteBackupScript 执行备份脚本
func (e *ScriptEngine) ExecuteBackupScript(ctx context.Context, target *backup.BackupTarget, policy *backup.BackupPolicy) (*ScriptResult, error) {
	params := map[string]interface{}{
		"target_id":   target.ID,
		"target_name": target.Name,
		"target_type": target.Type,
		"policy_id":   policy.ID,
		"policy_name": policy.Name,
		"source_path": policy.SourcePath,
		"storage_path": policy.StoragePath,
	}

	// 创建默认备份脚本
	script := &backup.RecoveryScript{
		Name:      fmt.Sprintf("backup_%s", target.Name),
		Type:      "backup",
		Language:  "bash",
		Script:    e.generateBackupScript(target, policy),
		Timeout:   policy.Timeout,
	}

	return e.ExecuteScript(ctx, script, params)
}

// ExecuteRestoreScript 执行恢复脚本
func (e *ScriptEngine) ExecuteRestoreScript(ctx context.Context, record *backup.BackupRecord, target *backup.BackupTarget, restorePath string) (*ScriptResult, error) {
	params := map[string]interface{}{
		"backup_id":    record.ID,
		"backup_file":  record.FilePath,
		"target_id":    target.ID,
		"target_name":  target.Name,
		"target_type":  target.Type,
		"restore_path": restorePath,
		"checksum":     record.Checksum,
	}

	script := &backup.RecoveryScript{
		Name:      fmt.Sprintf("restore_%s", target.Name),
		Type:      "restore",
		Language:  "bash",
		Script:    e.generateRestoreScript(record, target, restorePath),
		Timeout:   3600,
	}

	return e.ExecuteScript(ctx, script, params)
}

// ExecuteVerifyScript 执行验证脚本
func (e *ScriptEngine) ExecuteVerifyScript(ctx context.Context, record *backup.BackupRecord, verifyType string) (*ScriptResult, error) {
	params := map[string]interface{}{
		"backup_id":    record.ID,
		"backup_file":  record.FilePath,
		"verify_type":  verifyType,
		"checksum":     record.Checksum,
	}

	script := &backup.RecoveryScript{
		Name:      fmt.Sprintf("verify_%d", record.ID),
		Type:      "verify",
		Language:  "bash",
		Script:    e.generateVerifyScript(record, verifyType),
		Timeout:   1800,
	}

	return e.ExecuteScript(ctx, script, params)
}

// generateBackupScript 生成备份脚本
func (e *ScriptEngine) generateBackupScript(target *backup.BackupTarget, policy *backup.BackupPolicy) string {
	var script strings.Builder

	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -e\n\n")

	// 变量定义
	script.WriteString(fmt.Sprintf("TARGET_ID=%d\n", target.ID))
	script.WriteString(fmt.Sprintf("TARGET_NAME=\"%s\"\n", target.Name))
	script.WriteString(fmt.Sprintf("SOURCE_PATH=\"%s\"\n", policy.SourcePath))
	script.WriteString(fmt.Sprintf("STORAGE_PATH=\"%s\"\n", policy.StoragePath))
	script.WriteString(fmt.Sprintf("TIMESTAMP=$(date +%%Y%%m%%d_%%H%%M%%S)\n"))
	script.WriteString(fmt.Sprintf("BACKUP_FILE=\"${TARGET_NAME}_${TIMESTAMP}.tar.gz\"\n"))

	// 创建备份目录
	script.WriteString("\nmkdir -p \"${STORAGE_PATH}\"\n")

	// 执行备份
	switch target.Type {
	case "database":
		script.WriteString(e.generateDatabaseBackupScript(target, policy))
	case "filesystem", "file":
		script.WriteString(e.generateFileBackupScript(target, policy))
	default:
		script.WriteString(e.generateFileBackupScript(target, policy))
	}

	// 计算校验和
	script.WriteString("\necho \"计算校验和...\"\n")
	script.WriteString("CHECKSUM=$(sha256sum \"${STORAGE_PATH}/${BACKUP_FILE}\" | cut -d' ' -f1)\n")
	script.WriteString("echo \"校验和: ${CHECKSUM}\"\n")

	// 清理旧备份
	if policy.RetentionDays > 0 {
		script.WriteString(fmt.Sprintf("\necho \"清理 %d 天前的备份...\"\n", policy.RetentionDays))
		script.WriteString(fmt.Sprintf("find \"${STORAGE_PATH}\" -name \"${TARGET_NAME}_*.tar.gz\" -mtime +%d -delete\n", policy.RetentionDays))
	}

	script.WriteString("\necho \"备份完成: ${STORAGE_PATH}/${BACKUP_FILE}\"\n")
	script.WriteString("echo \"BACKUP_FILE=${STORAGE_PATH}/${BACKUP_FILE}\" >> $GITHUB_OUTPUT 2>/dev/null || true\n")

	return script.String()
}

// generateDatabaseBackupScript 生成数据库备份脚本
func (e *ScriptEngine) generateDatabaseBackupScript(target *backup.BackupTarget, policy *backup.BackupPolicy) string {
	var script strings.Builder

	// 解析数据库配置
	var dbConfig DatabaseConfig
	if target.DbConfig != "" {
		json.Unmarshal([]byte(target.DbConfig), &dbConfig)
	}

	script.WriteString("\necho \"开始数据库备份...\"\n")

	switch dbConfig.Type {
	case "mysql", "mariadb":
		script.WriteString(fmt.Sprintf("mysqldump -h%s -P%d -u%s -p%s --single-transaction --routines --triggers %s > /tmp/db_backup.sql\n",
			dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database))
		script.WriteString("tar -czf \"${STORAGE_PATH}/${BACKUP_FILE}\" -C /tmp db_backup.sql\n")
		script.WriteString("rm -f /tmp/db_backup.sql\n")

	case "postgresql":
		script.WriteString(fmt.Sprintf("PGPASSWORD=%s pg_dump -h%s -p%d -U%s %s > /tmp/db_backup.sql\n",
			dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Database))
		script.WriteString("tar -czf \"${STORAGE_PATH}/${BACKUP_FILE}\" -C /tmp db_backup.sql\n")
		script.WriteString("rm -f /tmp/db_backup.sql\n")

	default:
		script.WriteString("echo \"不支持的数据库类型\"\n")
		script.WriteString("exit 1\n")
	}

	return script.String()
}

// generateFileBackupScript 生成文件备份脚本
func (e *ScriptEngine) generateFileBackupScript(target *backup.BackupTarget, policy *backup.BackupPolicy) string {
	var script strings.Builder

	script.WriteString("\necho \"开始文件备份...\"\n")
	script.WriteString("tar -czf \"${STORAGE_PATH}/${BACKUP_FILE}\"")

	// 添加排除规则
	if policy.ExcludePaths != "" {
		excludes := strings.Split(policy.ExcludePaths, ",")
		for _, exclude := range excludes {
			exclude = strings.TrimSpace(exclude)
			if exclude != "" {
				script.WriteString(fmt.Sprintf(" --exclude \"%s\"", exclude))
			}
		}
	}

	script.WriteString(" \"${SOURCE_PATH}\"\n")

	return script.String()
}

// generateRestoreScript 生成恢复脚本
func (e *ScriptEngine) generateRestoreScript(record *backup.BackupRecord, target *backup.BackupTarget, restorePath string) string {
	var script strings.Builder

	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -e\n\n")

	// 变量定义
	script.WriteString(fmt.Sprintf("BACKUP_FILE=\"%s\"\n", record.FilePath))
	script.WriteString(fmt.Sprintf("RESTORE_PATH=\"%s\"\n", restorePath))
	script.WriteString(fmt.Sprintf("CHECKSUM=\"%s\"\n", record.Checksum))

	// 验证校验和
	script.WriteString("\necho \"验证备份文件完整性...\"\n")
	script.WriteString("ACTUAL_CHECKSUM=$(sha256sum \"${BACKUP_FILE}\" | cut -d' ' -f1)\n")
	script.WriteString("if [ \"${ACTUAL_CHECKSUM}\" != \"${CHECKSUM}\" ]; then\n")
	script.WriteString("  echo \"校验和不匹配!\"\n")
	script.WriteString("  exit 1\n")
	script.WriteString("fi\n")

	// 创建恢复目录
	script.WriteString("\nmkdir -p \"${RESTORE_PATH}\"\n")

	// 执行恢复
	switch target.Type {
	case "database":
		script.WriteString(e.generateDatabaseRestoreScript(target))
	case "filesystem", "file":
		script.WriteString(e.generateFileRestoreScript())
	default:
		script.WriteString(e.generateFileRestoreScript())
	}

	script.WriteString("\necho \"恢复完成\"\n")

	return script.String()
}

// generateDatabaseRestoreScript 生成数据库恢复脚本
func (e *ScriptEngine) generateDatabaseRestoreScript(target *backup.BackupTarget) string {
	var script strings.Builder

	var dbConfig DatabaseConfig
	if target.DbConfig != "" {
		json.Unmarshal([]byte(target.DbConfig), &dbConfig)
	}

	script.WriteString("\necho \"开始数据库恢复...\"\n")
	script.WriteString("tar -xzf \"${BACKUP_FILE}\" -C /tmp\n")

	switch dbConfig.Type {
	case "mysql", "mariadb":
		script.WriteString(fmt.Sprintf("mysql -h%s -P%d -u%s -p%s %s < /tmp/db_backup.sql\n",
			dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database))
	case "postgresql":
		script.WriteString(fmt.Sprintf("PGPASSWORD=%s psql -h%s -p%d -U%s -d%s -f /tmp/db_backup.sql\n",
			dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Database))
	}

	script.WriteString("rm -f /tmp/db_backup.sql\n")

	return script.String()
}

// generateFileRestoreScript 生成文件恢复脚本
func (e *ScriptEngine) generateFileRestoreScript() string {
	var script strings.Builder

	script.WriteString("\necho \"开始文件恢复...\"\n")
	script.WriteString("tar -xzf \"${BACKUP_FILE}\" -C \"${RESTORE_PATH}\"\n")

	return script.String()
}

// generateVerifyScript 生成验证脚本
func (e *ScriptEngine) generateVerifyScript(record *backup.BackupRecord, verifyType string) string {
	var script strings.Builder

	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -e\n\n")

	script.WriteString(fmt.Sprintf("BACKUP_FILE=\"%s\"\n", record.FilePath))
	script.WriteString(fmt.Sprintf("EXPECTED_CHECKSUM=\"%s\"\n", record.Checksum))

	switch verifyType {
	case "integrity":
		script.WriteString("\necho \"执行完整性验证...\"\n")
		script.WriteString("ACTUAL_CHECKSUM=$(sha256sum \"${BACKUP_FILE}\" | cut -d' ' -f1)\n")
		script.WriteString("if [ \"${ACTUAL_CHECKSUM}\" == \"${EXPECTED_CHECKSUM}\" ]; then\n")
		script.WriteString("  echo \"完整性验证通过\"\n")
		script.WriteString("else\n")
		script.WriteString("  echo \"完整性验证失败\"\n")
		script.WriteString("  exit 1\n")
		script.WriteString("fi\n")

	case "structure":
		script.WriteString("\necho \"执行结构验证...\"\n")
		script.WriteString("FILE_COUNT=$(tar -tzf \"${BACKUP_FILE}\" | wc -l)\n")
		script.WriteString("if [ ${FILE_COUNT} -gt 0 ]; then\n")
		script.WriteString("  echo \"结构验证通过，包含 ${FILE_COUNT} 个文件\"\n")
		script.WriteString("else\n")
		script.WriteString("  echo \"备份为空\"\n")
		script.WriteString("  exit 1\n")
		script.WriteString("fi\n")

	case "full":
		script.WriteString("\necho \"执行完整验证...\"\n")
		// 完整性检查
		script.WriteString("ACTUAL_CHECKSUM=$(sha256sum \"${BACKUP_FILE}\" | cut -d' ' -f1)\n")
		script.WriteString("if [ \"${ACTUAL_CHECKSUM}\" != \"${EXPECTED_CHECKSUM}\" ]; then\n")
		script.WriteString("  echo \"完整性验证失败\"\n")
		script.WriteString("  exit 1\n")
		script.WriteString("fi\n")
		// 结构检查
		script.WriteString("FILE_COUNT=$(tar -tzf \"${BACKUP_FILE}\" | wc -l)\n")
		script.WriteString("if [ ${FILE_COUNT} -eq 0 ]; then\n")
		script.WriteString("  echo \"备份为空\"\n")
		script.WriteString("  exit 1\n")
		script.WriteString("fi\n")
		// 恢复测试
		script.WriteString("TEMP_DIR=$(mktemp -d)\n")
		script.WriteString("tar -xzf \"${BACKUP_FILE}\" -C ${TEMP_DIR}\n")
		script.WriteString("rm -rf ${TEMP_DIR}\n")
		script.WriteString("echo \"完整验证通过\"\n")

	default:
		script.WriteString("\necho \"执行基本验证...\"\n")
		script.WriteString("test -f \"${BACKUP_FILE}\"\n")
	}

	return script.String()
}

// CancelScript 取消脚本执行
func (e *ScriptEngine) CancelScript(execID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ctx, exists := e.runningScripts[execID]; exists {
		ctx.Cancelled = true
		ctx.Status = "cancelled"
		return nil
	}

	return fmt.Errorf("脚本执行不存在或已完成")
}

// GetScriptStatus 获取脚本执行状态
func (e *ScriptEngine) GetScriptStatus(execID string) (*ScriptContext, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ctx, exists := e.runningScripts[execID]; exists {
		return ctx, nil
	}

	return nil, fmt.Errorf("脚本执行不存在")
}

// ListRunningScripts 列出正在运行的脚本
func (e *ScriptEngine) ListRunningScripts() []*ScriptContext {
	e.mu.Lock()
	defer e.mu.Unlock()

	scripts := make([]*ScriptContext, 0, len(e.runningScripts))
	for _, ctx := range e.runningScripts {
		scripts = append(scripts, ctx)
	}

	return scripts
}

// SaveScript 保存脚本
func (e *ScriptEngine) SaveScript(script *backup.RecoveryScript) error {
	fileName := fmt.Sprintf("%d_%s.sh", script.ID, script.Name)
	filePath := filepath.Join(e.scriptDir, fileName)

	// 计算脚本哈希
	hash := sha256.Sum256([]byte(script.Script))
	checksum := hex.EncodeToString(hash[:])

	// 添加元数据头
	content := fmt.Sprintf("#!/bin/bash\n# Script ID: %d\n# Name: %s\n# Type: %s\n# Checksum: %s\n# Created: %s\n\n%s",
		script.ID, script.Name, script.Type, checksum, time.Now().Format("2006-01-02 15:04:05"), script.Script)

	return os.WriteFile(filePath, []byte(content), 0755)
}

// LoadScript 加载脚本
func (e *ScriptEngine) LoadScript(scriptID uint) (*backup.RecoveryScript, error) {
	files, err := filepath.Glob(filepath.Join(e.scriptDir, fmt.Sprintf("%d_*.sh", scriptID)))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("脚本不存在")
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		return nil, err
	}

	// 解析脚本
	script := &backup.RecoveryScript{
		ID:       scriptID,
		Script:   string(data),
		Language: "bash",
	}

	return script, nil
}

// DeleteScript 删除脚本
func (e *ScriptEngine) DeleteScript(scriptID uint) error {
	files, err := filepath.Glob(filepath.Join(e.scriptDir, fmt.Sprintf("%d_*.sh", scriptID)))
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	return nil
}

// CreateBuiltInScripts 创建内置脚本
func (e *ScriptEngine) CreateBuiltInScripts() []backup.RecoveryScript {
	return []backup.RecoveryScript{
		{
			Name:        "mysql_backup",
			Description: "MySQL数据库自动备份脚本",
			Type:        "backup",
			Language:    "bash",
			Script: `#!/bin/bash
# MySQL 自动备份脚本
BACKUP_DIR="{{.backup_dir}}"
DB_HOST="{{.db_host}}"
DB_PORT="{{.db_port}}"
DB_USER="{{.db_user}}"
DB_PASS="{{.db_pass}}"
DB_NAME="{{.db_name}}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p ${BACKUP_DIR}
mysqldump -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASS} \
  --single-transaction --routines --triggers ${DB_NAME} | gzip > ${BACKUP_FILE}

echo "Backup completed: ${BACKUP_FILE}"
`,
			TargetTypes: `["database","mysql"]`,
			Enabled:     true,
		},
		{
			Name:        "file_backup",
			Description: "文件系统自动备份脚本",
			Type:        "backup",
			Language:    "bash",
			Script: `#!/bin/bash
# 文件系统自动备份脚本
SOURCE_PATH="{{.source_path}}"
BACKUP_DIR="{{.backup_dir}}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/backup_${TIMESTAMP}.tar.gz"

mkdir -p ${BACKUP_DIR}
tar -czf ${BACKUP_FILE} ${SOURCE_PATH}

echo "Backup completed: ${BACKUP_FILE}"
`,
			TargetTypes: `["filesystem","file"]`,
			Enabled:     true,
		},
		{
			Name:        "auto_restore",
			Description: "自动恢复脚本",
			Type:        "restore",
			Language:    "bash",
			Script: `#!/bin/bash
# 自动恢复脚本
BACKUP_FILE="{{.backup_file}}"
RESTORE_PATH="{{.restore_path}}"
VERIFY_CHECKSUM="{{.verify_checksum}}"

# 验证校验和
if [ "${VERIFY_CHECKSUM}" = "true" ]; then
  EXPECTED_CHECKSUM="{{.expected_checksum}}"
  ACTUAL_CHECKSUM=$(sha256sum ${BACKUP_FILE} | cut -d' ' -f1)
  if [ "${ACTUAL_CHECKSUM}" != "${EXPECTED_CHECKSUM}" ]; then
    echo "Checksum verification failed!"
    exit 1
  fi
fi

# 创建恢复目录
mkdir -p ${RESTORE_PATH}

# 执行恢复
tar -xzf ${BACKUP_FILE} -C ${RESTORE_PATH}

echo "Restore completed to: ${RESTORE_PATH}"
`,
			TargetTypes: `["filesystem","file","database"]`,
			Enabled:     true,
		},
		{
			Name:        "verify_backup",
			Description: "备份验证脚本",
			Type:        "verify",
			Language:    "bash",
			Script: `#!/bin/bash
# 备份验证脚本
BACKUP_FILE="{{.backup_file}}"
EXPECTED_CHECKSUM="{{.expected_checksum}}"

# 完整性检查
ACTUAL_CHECKSUM=$(sha256sum ${BACKUP_FILE} | cut -d' ' -f1)
if [ "${ACTUAL_CHECKSUM}" != "${EXPECTED_CHECKSUM}" ]; then
  echo "FAIL: Checksum mismatch"
  exit 1
fi

# 结构检查
FILE_COUNT=$(tar -tzf ${BACKUP_FILE} | wc -l)
if [ ${FILE_COUNT} -eq 0 ]; then
  echo "FAIL: Backup is empty"
  exit 1
fi

# 恢复测试
TEMP_DIR=$(mktemp -d)
tar -xzf ${BACKUP_FILE} -C ${TEMP_DIR}
RESTORED_COUNT=$(find ${TEMP_DIR} -type f | wc -l)
rm -rf ${TEMP_DIR}

echo "PASS: Verification completed"
echo "Files in backup: ${FILE_COUNT}"
echo "Restorable files: ${RESTORED_COUNT}"
`,
			TargetTypes: `["filesystem","file","database"]`,
			Enabled:     true,
		},
		{
			Name:        "drill_recovery",
			Description: "灾备演练恢复脚本",
			Type:        "drill",
			Language:    "bash",
			Script: `#!/bin/bash
# 灾备演练恢复脚本
DRILL_ID="{{.drill_id}}"
TARGET_SYSTEM="{{.target_system}}"
BACKUP_FILE="{{.backup_file}}"
DRILL_ENV="{{.drill_env}}"

echo "Starting disaster recovery drill: ${DRILL_ID}"
echo "Target: ${TARGET_SYSTEM}"
echo "Environment: ${DRILL_ENV}"

# 停止目标服务
echo "Stopping target services..."
systemctl stop {{.services}} 2>/dev/null || true

# 执行恢复
echo "Executing recovery..."
{{.restore_command}}

# 启动服务
echo "Starting services..."
systemctl start {{.services}} 2>/dev/null || true

# 验证
echo "Verifying recovery..."
{{.verify_command}}

echo "Drill completed: ${DRILL_ID}"
`,
			TargetTypes: `["filesystem","database","application"]`,
			Enabled:     true,
		},
	}
}

// RegisterBuiltInScripts 注册内置脚本
func (e *ScriptEngine) RegisterBuiltInScripts() error {
	scripts := e.CreateBuiltInScripts()
	for i := range scripts {
		scripts[i].ID = uint(i + 1)
		if err := e.SaveScript(&scripts[i]); err != nil {
			return err
		}
	}
	return nil
}
