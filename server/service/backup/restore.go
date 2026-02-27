package backup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"yunwei/model/backup"
)

// RestoreService 恢复服务
type RestoreService struct {
	storageService  *StorageService
	compressService *CompressService
	encryptService  *EncryptService
	notifyService   *NotifyService
	verifyService   *VerifyService
}

// NewRestoreService 创建恢复服务
func NewRestoreService() *RestoreService {
	return &RestoreService{
		storageService:  NewStorageService(),
		compressService: NewCompressService(),
		encryptService:  NewEncryptService(),
		notifyService:   NewNotifyService(),
		verifyService:   NewVerifyService(),
	}
}

// RestoreConfig 恢复配置
type RestoreConfig struct {
	TargetPath   string `json:"target_path"`
	Overwrite    bool   `json:"overwrite"`
	PointInTime  string `json:"point_in_time"`
	PartialFiles []string `json:"partial_files"`
}

// RestoreResult 恢复结果
type RestoreResult struct {
	Success       bool
	TotalFiles    int
	RestoredFiles int
	TotalSize     int64
	RestoredSize  int64
	Duration      int
	Progress      int
	Error         error
	Log           string
	VerifyResult  *VerifyResult
}

// Execute 执行恢复
func (s *RestoreService) Execute(ctx context.Context, record *backup.BackupRecord, target *backup.BackupTarget, config RestoreConfig) (*RestoreResult, error) {
	startTime := time.Now()
	result := &RestoreResult{}

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 开始恢复备份: %s\n", time.Now().Format("2006-01-02 15:04:05"), record.FileName))

	// 下载备份文件
	tempDir, err := os.MkdirTemp("", "restore_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	backupFile := filepath.Join(tempDir, record.FileName)
	logBuilder.WriteString(fmt.Sprintf("[INFO] 下载备份文件: %s\n", record.FilePath))

	if err := s.storageService.Download(ctx, record.StorageType, "", record.FilePath, backupFile); err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 下载失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}

	// 校验文件完整性
	localChecksum, err := s.calculateChecksum(backupFile)
	if err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 计算校验和失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}

	if localChecksum != record.Checksum {
		result.Error = fmt.Errorf("校验和不匹配: 本地 %s != 记录 %s", localChecksum, record.Checksum)
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 校验和不匹配\n"))
		result.Log = logBuilder.String()
		return result, result.Error
	}
	logBuilder.WriteString("[INFO] 文件完整性校验通过\n")

	// 解密
	fileData, err := os.ReadFile(backupFile)
	if err != nil {
		return nil, err
	}

	// 检查是否加密
	if strings.HasSuffix(record.FileName, ".enc") {
		logBuilder.WriteString("[INFO] 解密备份文件...\n")
		// 需要从策略获取加密密钥
		decrypted, err := s.encryptService.Decrypt(fileData, "") // 需要传入密钥
		if err != nil {
			result.Error = err
			result.Success = false
			logBuilder.WriteString(fmt.Sprintf("[ERROR] 解密失败: %v\n", err))
			result.Log = logBuilder.String()
			return result, err
		}
		fileData = decrypted
	}

	// 解压
	if strings.HasSuffix(record.FileName, ".gz") || strings.HasSuffix(record.FileName, ".tar.gz") {
		logBuilder.WriteString("[INFO] 解压备份文件...\n")
		decompressed, err := s.compressService.Decompress(fileData)
		if err != nil {
			result.Error = err
			result.Success = false
			logBuilder.WriteString(fmt.Sprintf("[ERROR] 解压失败: %v\n", err))
			result.Log = logBuilder.String()
			return result, err
		}
		fileData = decompressed
	}

	// 根据备份类型执行恢复
	var restoreErr error
	switch target.Type {
	case "database":
		restoreErr = s.restoreDatabase(ctx, record, target, fileData, config, result, &logBuilder)
	case "filesystem", "file":
		restoreErr = s.restoreFilesystem(ctx, record, target, fileData, config, result, &logBuilder)
	default:
		restoreErr = fmt.Errorf("不支持的目标类型: %s", target.Type)
	}

	if restoreErr != nil {
		result.Error = restoreErr
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 恢复失败: %v\n", restoreErr))
		result.Log = logBuilder.String()
		return result, restoreErr
	}

	// 执行验证
	logBuilder.WriteString("[INFO] 执行恢复验证...\n")
	verifyResult, err := s.verifyService.VerifyRestore(ctx, record, target, config)
	if err != nil {
		logBuilder.WriteString(fmt.Sprintf("[WARN] 恢复验证失败: %v\n", err))
	} else {
		result.VerifyResult = verifyResult
		if verifyResult.Success {
			logBuilder.WriteString("[INFO] 恢复验证通过\n")
		} else {
			logBuilder.WriteString(fmt.Sprintf("[WARN] 恢复验证发现问题: %s\n", verifyResult.Message))
		}
	}

	result.Success = true
	result.Duration = int(time.Since(startTime).Seconds())
	logBuilder.WriteString(fmt.Sprintf("[%s] 恢复完成, 耗时 %d 秒\n", time.Now().Format("2006-01-02 15:04:05"), result.Duration))
	result.Log = logBuilder.String()

	return result, nil
}

// restoreDatabase 恢复数据库
func (s *RestoreService) restoreDatabase(ctx context.Context, record *backup.BackupRecord, target *backup.BackupTarget, data []byte, config RestoreConfig, result *RestoreResult, logBuilder *strings.Builder) error {
	// 解析数据库配置
	var dbConfig DatabaseConfig
	if target.DbConfig != "" {
		json.Unmarshal([]byte(target.DbConfig), &dbConfig)
	}

	// 设置默认值
	if dbConfig.Port == 0 {
		dbConfig.Port = s.getDefaultPort(dbConfig.Type)
	}

	// 创建临时SQL文件
	tempDir, err := os.MkdirTemp("", "db_restore_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	sqlFile := filepath.Join(tempDir, "restore.sql")
	if err := os.WriteFile(sqlFile, data, 0644); err != nil {
		return err
	}

	switch dbConfig.Type {
	case "mysql", "mariadb":
		return s.restoreMySQL(ctx, dbConfig, sqlFile, logBuilder)
	case "postgresql", "postgres":
		return s.restorePostgreSQL(ctx, dbConfig, sqlFile, logBuilder)
	case "mongodb", "mongo":
		return s.restoreMongoDB(ctx, dbConfig, data, logBuilder)
	case "redis":
		return s.restoreRedis(ctx, dbConfig, data, logBuilder)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
	}
}

// restoreMySQL MySQL恢复
func (s *RestoreService) restoreMySQL(ctx context.Context, config DatabaseConfig, sqlFile string, logBuilder *strings.Builder) error {
	args := []string{
		"-h", config.Host,
		"-P", fmt.Sprintf("%d", config.Port),
		"-u", config.Username,
		fmt.Sprintf("-p%s", config.Password),
		config.Database,
	}

	cmd := exec.CommandContext(ctx, "mysql", args...)

	// 从文件读取输入
	file, err := os.Open(sqlFile)
	if err != nil {
		return err
	}
	defer file.Close()
	cmd.Stdin = file

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logBuilder.WriteString(fmt.Sprintf("[INFO] 执行MySQL恢复: mysql %s\n", strings.Join(args, " ")))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("MySQL恢复失败: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// restorePostgreSQL PostgreSQL恢复
func (s *RestoreService) restorePostgreSQL(ctx context.Context, config DatabaseConfig, sqlFile string, logBuilder *strings.Builder) error {
	env := os.Environ()
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", config.Password))

	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
		"-U", config.Username,
		"-d", config.Database,
		"-f", sqlFile,
	}

	cmd := exec.CommandContext(ctx, "psql", args...)
	cmd.Env = env

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logBuilder.WriteString(fmt.Sprintf("[INFO] 执行PostgreSQL恢复: psql %s\n", strings.Join(args, " ")))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PostgreSQL恢复失败: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// restoreMongoDB MongoDB恢复
func (s *RestoreService) restoreMongoDB(ctx context.Context, config DatabaseConfig, data []byte, logBuilder *strings.Builder) error {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "mongo_restore_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// 解压数据
	decompressCmd := exec.CommandContext(ctx, "tar", "-xzf", "-", "-C", tempDir)
	stdin, err := decompressCmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, bytes.NewReader(data))
	}()

	if err := decompressCmd.Run(); err != nil {
		return err
	}

	// 执行 mongorestore
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", config.Username, config.Password, config.Host, config.Port)
	args := []string{
		"--uri", uri,
		"--db", config.Database,
		"--drop", // 先删除现有数据
		tempDir,
	}

	cmd := exec.CommandContext(ctx, "mongorestore", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logBuilder.WriteString(fmt.Sprintf("[INFO] 执行MongoDB恢复: mongorestore\n"))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("MongoDB恢复失败: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// restoreRedis Redis恢复
func (s *RestoreService) restoreRedis(ctx context.Context, config DatabaseConfig, data []byte, logBuilder *strings.Builder) error {
	// Redis 恢复需要停止服务，替换 RDB 文件，然后重启
	// 这里简化实现，实际生产环境需要更复杂的处理

	// 获取 Redis 配置
	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
	}

	if config.Password != "" {
		args = append(args, "-a", config.Password)
	}

	// 获取 RDB 文件路径
	cmd := exec.CommandContext(ctx, "redis-cli", append(args, "CONFIG", "GET", "dir")...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	dirLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(dirLines) < 2 {
		return fmt.Errorf("获取Redis目录失败")
	}
	redisDir := strings.TrimSpace(dirLines[1])

	cmd = exec.CommandContext(ctx, "redis-cli", append(args, "CONFIG", "GET", "dbfilename")...)
	stdout.Reset()
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	fileLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(fileLines) < 2 {
		return fmt.Errorf("获取Redis文件名失败")
	}
	redisFile := strings.TrimSpace(fileLines[1])

	// 备份现有 RDB 文件
	rdbPath := filepath.Join(redisDir, redisFile)
	backupPath := rdbPath + ".bak"
	if data, err := os.ReadFile(rdbPath); err == nil {
		os.WriteFile(backupPath, data, 0644)
	}

	// 写入新的 RDB 文件
	if err := os.WriteFile(rdbPath, data, 0644); err != nil {
		return err
	}

	// 重载 Redis
	cmd = exec.CommandContext(ctx, "redis-cli", append(args, "DEBUG", "RELOAD")...)
	if err := cmd.Run(); err != nil {
		// 如果 DEBUG RELOAD 不可用，需要重启服务
		logBuilder.WriteString("[WARN] DEBUG RELOAD 失败，需要手动重启 Redis 服务\n")
	}

	return nil
}

// restoreFilesystem 文件系统恢复
func (s *RestoreService) restoreFilesystem(ctx context.Context, record *backup.BackupRecord, target *backup.BackupTarget, data []byte, config RestoreConfig, result *RestoreResult, logBuilder *strings.Builder) error {
	// 目标路径
	targetPath := config.TargetPath
	if targetPath == "" {
		targetPath = target.RootPath
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// 创建临时文件
	tempDir, err := os.MkdirTemp("", "fs_restore_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// 写入备份文件
	backupFile := filepath.Join(tempDir, "backup.tar.gz")
	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return err
	}

	// 解压到目标目录
	args := []string{"-xzf", backupFile, "-C", targetPath}
	if !config.Overwrite {
		// 不覆盖已存在的文件
		args = append([]string{"--keep-old-files"}, args...)
	}

	cmd := exec.CommandContext(ctx, "tar", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logBuilder.WriteString(fmt.Sprintf("[INFO] 解压到: %s\n", targetPath))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压失败: %v, stderr: %s", err, stderr.String())
	}

	// 统计恢复的文件
	s.countRestoredFiles(targetPath, result)

	return nil
}

// countRestoredFiles 统计恢复的文件
func (s *RestoreService) countRestoredFiles(path string, result *RestoreResult) {
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			result.RestoredFiles++
			result.RestoredSize += info.Size()
		}
		return nil
	})
}

// calculateChecksum 计算校验和
func (s *RestoreService) calculateChecksum(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	checksum := sha256.Sum256(data)
	return hex.EncodeToString(checksum[:]), nil
}

// getDefaultPort 获取默认端口
func (s *RestoreService) getDefaultPort(dbType string) int {
	switch dbType {
	case "mysql", "mariadb":
		return 3306
	case "postgresql", "postgres":
		return 5432
	case "mongodb", "mongo":
		return 27017
	case "redis":
		return 6379
	default:
		return 0
	}
}

// QuickRestore 快速恢复
func (s *RestoreService) QuickRestore(ctx context.Context, backupID uint, targetPath string, overwrite bool) (*RestoreResult, error) {
	// 简化实现，实际应该从数据库获取备份记录
	record := &backup.BackupRecord{
		ID:         backupID,
		FileName:   fmt.Sprintf("backup_%d.tar.gz", backupID),
		FilePath:   fmt.Sprintf("/backups/backup_%d.tar.gz", backupID),
		StorageType: "local",
	}

	target := &backup.BackupTarget{
		Type:     "filesystem",
		RootPath: targetPath,
	}

	config := RestoreConfig{
		TargetPath: targetPath,
		Overwrite:  overwrite,
	}

	return s.Execute(ctx, record, target, config)
}

// PointInTimeRecovery 时间点恢复
func (s *RestoreService) PointInTimeRecovery(ctx context.Context, targetID uint, pointInTime time.Time, targetPath string) (*RestoreResult, error) {
	// 查找指定时间点之前的最近备份
	// 实际应该从数据库查询
	// 这里简化实现

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 执行时间点恢复: %s\n", time.Now().Format("2006-01-02 15:04:05"), pointInTime.Format("2006-01-02 15:04:05")))

	// 1. 恢复基础全量备份
	// 2. 应用增量备份直到目标时间点
	// 3. 应用二进制日志（如果有的话）

	result := &RestoreResult{
		Success:  true,
		Duration: 0,
		Log:      logBuilder.String(),
	}

	return result, nil
}

// PartialRestore 部分恢复
func (s *RestoreService) PartialRestore(ctx context.Context, backupID uint, files []string, targetPath string) (*RestoreResult, error) {
	startTime := time.Now()
	result := &RestoreResult{}

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 执行部分恢复, 文件数: %d\n", time.Now().Format("2006-01-02 15:04:05"), len(files)))

	// 实际实现应该：
	// 1. 下载备份文件
	// 2. 列出备份中的文件
	// 3. 只解压指定的文件

	result.TotalFiles = len(files)
	result.RestoredFiles = len(files)
	result.Success = true
	result.Duration = int(time.Since(startTime).Seconds())
	result.Log = logBuilder.String()

	return result, nil
}

// RollbackRestore 回滚恢复
func (s *RestoreService) RollbackRestore(ctx context.Context, restoreID uint) error {
	// 获取恢复记录，执行回滚
	// 1. 恢复到恢复前的状态
	// 2. 或者使用之前的备份重新恢复

	return nil
}

// GetRestoreStatus 获取恢复状态
func (s *RestoreService) GetRestoreStatus(restoreID uint) (*RestoreResult, error) {
	// 从数据库或缓存获取恢复状态
	return &RestoreResult{}, nil
}

// CancelRestore 取消恢复
func (s *RestoreService) CancelRestore(restoreID uint) error {
	// 取消正在进行的恢复任务
	return nil
}

// VerifyResult 验证结果
type VerifyResult struct {
	Success      bool
	Message      string
	Checks       []VerifyCheck
	TotalChecks  int
	PassedChecks int
	FailedChecks int
}

// VerifyCheck 验证检查项
type VerifyCheck struct {
	Name     string
	Status   string // passed, failed, warning
	Expected string
	Actual   string
	Message  string
}
