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

// DatabaseBackupService 数据库备份服务
type DatabaseBackupService struct {
	storageService  *StorageService
	compressService *CompressService
	encryptService  *EncryptService
	notifyService   *NotifyService
}

// NewDatabaseBackupService 创建数据库备份服务
func NewDatabaseBackupService() *DatabaseBackupService {
	return &DatabaseBackupService{
		storageService:  NewStorageService(),
		compressService: NewCompressService(),
		encryptService:  NewEncryptService(),
		notifyService:   NewNotifyService(),
	}
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"type"`     // mysql, postgresql, mongodb, redis
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Charset  string `json:"charset"`
	SSLMode  string `json:"ssl_mode"`
	Options  string `json:"options"` // 额外选项
}

// BackupResult 备份结果
type BackupResult struct {
	Success      bool
	FilePath     string
	FileName     string
	FileSize     int64
	CompressSize int64
	Checksum     string
	Duration     int
	Error        error
	Log          string
}

// Execute 执行数据库备份
func (s *DatabaseBackupService) Execute(ctx context.Context, policy *backup.BackupPolicy, target *backup.BackupTarget) (*BackupResult, error) {
	startTime := time.Now()
	result := &BackupResult{}

	// 解析数据库配置
	var dbConfig DatabaseConfig
	if target != nil && target.DbConfig != "" {
		if err := json.Unmarshal([]byte(target.DbConfig), &dbConfig); err != nil {
			return nil, fmt.Errorf("解析数据库配置失败: %v", err)
		}
	} else if policy.SourceConfig != "" {
		if err := json.Unmarshal([]byte(policy.SourceConfig), &dbConfig); err != nil {
			return nil, fmt.Errorf("解析数据库配置失败: %v", err)
		}
	}

	// 设置默认值
	if dbConfig.Port == 0 {
		dbConfig.Port = s.getDefaultPort(dbConfig.Type)
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "backup_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s_%s.sql", target.Name, dbConfig.Database, timestamp)
	if policy.Compress {
		fileName += ".gz"
	}
	filePath := filepath.Join(tempDir, fileName)

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 开始备份 %s 数据库\n", time.Now().Format("2006-01-02 15:04:05"), dbConfig.Database))

	// 执行预备份脚本
	if policy.PreScript != "" {
		logBuilder.WriteString("[INFO] 执行预备份脚本...\n")
		if output, err := s.executeScript(policy.PreScript, map[string]interface{}{
			"target": target,
			"policy": policy,
			"config": dbConfig,
		}); err != nil {
			logBuilder.WriteString(fmt.Sprintf("[WARN] 预备份脚本执行失败: %v\n", err))
		} else {
			logBuilder.WriteString(fmt.Sprintf("[INFO] 预备份脚本输出: %s\n", output))
		}
	}

	// 执行备份
	var dumpOutput []byte
	switch dbConfig.Type {
	case "mysql", "mariadb":
		dumpOutput, err = s.backupMySQL(ctx, dbConfig, policy.Timeout)
	case "postgresql", "postgres":
		dumpOutput, err = s.backupPostgreSQL(ctx, dbConfig, policy.Timeout)
	case "mongodb", "mongo":
		dumpOutput, err = s.backupMongoDB(ctx, dbConfig, policy.Timeout)
	case "redis":
		dumpOutput, err = s.backupRedis(ctx, dbConfig, policy.Timeout)
	default:
		err = fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
	}

	if err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 备份失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}

	logBuilder.WriteString("[INFO] 数据库导出完成\n")

	// 写入文件
	var finalData []byte
	if policy.Compress {
		compressed, err := s.compressService.Compress(dumpOutput, policy.CompressType)
		if err != nil {
			result.Error = err
			result.Success = false
			logBuilder.WriteString(fmt.Sprintf("[ERROR] 压缩失败: %v\n", err))
			result.Log = logBuilder.String()
			return result, err
		}
		finalData = compressed
		logBuilder.WriteString("[INFO] 数据压缩完成\n")
	} else {
		finalData = dumpOutput
	}

	// 加密
	if policy.Encrypt {
		encrypted, err := s.encryptService.Encrypt(finalData, policy.EncryptKey)
		if err != nil {
			result.Error = err
			result.Success = false
			logBuilder.WriteString(fmt.Sprintf("[ERROR] 加密失败: %v\n", err))
			result.Log = logBuilder.String()
			return result, err
		}
		finalData = encrypted
		fileName += ".enc"
		filePath += ".enc"
		logBuilder.WriteString("[INFO] 数据加密完成\n")
	}

	// 写入文件
	if err := os.WriteFile(filePath, finalData, 0644); err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 写入文件失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}

	// 计算校验和
	checksum := sha256.Sum256(finalData)
	result.Checksum = hex.EncodeToString(checksum[:])

	// 获取文件大小
	fileInfo, _ := os.Stat(filePath)
	result.FileSize = int64(len(dumpOutput))
	result.CompressSize = fileInfo.Size()
	result.FilePath = filePath
	result.FileName = fileName

	logBuilder.WriteString(fmt.Sprintf("[INFO] 文件大小: 原始 %d 字节, 压缩后 %d 字节\n", result.FileSize, result.CompressSize))

	// 上传到存储
	storagePath, err := s.storageService.Upload(ctx, policy.StorageType, policy.StorageConfig, filePath, policy.StoragePath)
	if err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 上传存储失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}
	result.FilePath = storagePath
	logBuilder.WriteString(fmt.Sprintf("[INFO] 上传到存储: %s\n", storagePath))

	// 执行后备份脚本
	if policy.PostScript != "" {
		logBuilder.WriteString("[INFO] 执行后备份脚本...\n")
		if output, err := s.executeScript(policy.PostScript, map[string]interface{}{
			"target":  target,
			"policy":  policy,
			"config":  dbConfig,
			"result":  result,
		}); err != nil {
			logBuilder.WriteString(fmt.Sprintf("[WARN] 后备份脚本执行失败: %v\n", err))
		} else {
			logBuilder.WriteString(fmt.Sprintf("[INFO] 后备份脚本输出: %s\n", output))
		}
	}

	result.Success = true
	result.Duration = int(time.Since(startTime).Seconds())
	logBuilder.WriteString(fmt.Sprintf("[%s] 备份完成, 耗时 %d 秒\n", time.Now().Format("2006-01-02 15:04:05"), result.Duration))
	result.Log = logBuilder.String()

	return result, nil
}

// backupMySQL MySQL备份
func (s *DatabaseBackupService) backupMySQL(ctx context.Context, config DatabaseConfig, timeout int) ([]byte, error) {
	args := []string{
		"-h", config.Host,
		"-P", fmt.Sprintf("%d", config.Port),
		"-u", config.Username,
		fmt.Sprintf("-p%s", config.Password),
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
	}

	if config.Charset != "" {
		args = append(args, "--default-character-set", config.Charset)
	}

	if config.Options != "" {
		args = append(args, strings.Split(config.Options, " ")...)
	}

	args = append(args, config.Database)

	cmd := exec.CommandContext(ctx, "mysqldump", args...)
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, "mysqldump", args...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mysqldump失败: %v, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// backupPostgreSQL PostgreSQL备份
func (s *DatabaseBackupService) backupPostgreSQL(ctx context.Context, config DatabaseConfig, timeout int) ([]byte, error) {
	env := os.Environ()
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", config.Password))

	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
		"-U", config.Username,
		"-d", config.Database,
		"-F", "p", // plain text format
		"-w", // no password prompt
	}

	if config.SSLMode != "" {
		args = append(args, fmt.Sprintf("--set=sslmode=%s", config.SSLMode))
	}

	if config.Options != "" {
		args = append(args, strings.Split(config.Options, " ")...)
	}

	cmd := exec.CommandContext(ctx, "pg_dump", args...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pg_dump失败: %v, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// backupMongoDB MongoDB备份
func (s *DatabaseBackupService) backupMongoDB(ctx context.Context, config DatabaseConfig, timeout int) ([]byte, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "mongo_dump_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", config.Username, config.Password, config.Host, config.Port)
	args := []string{
		"--uri", uri,
		"--db", config.Database,
		"--out", tempDir,
		"--quiet",
	}

	if config.Options != "" {
		args = append(args, strings.Split(config.Options, " ")...)
	}

	cmd := exec.CommandContext(ctx, "mongodump", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mongodump失败: %v, stderr: %s", err, stderr.String())
	}

	// 打包备份文件
	dbDir := filepath.Join(tempDir, config.Database)
	return s.compressService.CompressDir(dbDir)
}

// backupRedis Redis备份
func (s *DatabaseBackupService) backupRedis(ctx context.Context, config DatabaseConfig, timeout int) ([]byte, error) {
	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
	}

	if config.Password != "" {
		args = append(args, "-a", config.Password)
	}

	// 执行 BGSAVE 命令
	cmd := exec.CommandContext(ctx, "redis-cli", append(args, "BGSAVE")...)
	if output, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf("BGSAVE失败: %v, output: %s", err, string(output))
	}

	// 等待备份完成
	time.Sleep(2 * time.Second)

	// 执行 LASTSAVE 检查
	cmd = exec.CommandContext(ctx, "redis-cli", append(args, "LASTSAVE")...)
	lastSaveOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("LASTSAVE失败: %v", err)
	}

	// 获取 RDB 文件路径 (需要通过 CONFIG GET dir 和 dbfilename 获取)
	cmd = exec.CommandContext(ctx, "redis-cli", append(args, "CONFIG", "GET", "dir")...)
	dirOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取Redis目录失败: %v", err)
	}

	cmd = exec.CommandContext(ctx, "redis-cli", append(args, "CONFIG", "GET", "dbfilename")...)
	filenameOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取Redis文件名失败: %v", err)
	}

	// 解析输出
	dirLines := strings.Split(strings.TrimSpace(string(dirOutput)), "\n")
	filenameLines := strings.Split(strings.TrimSpace(string(filenameOutput)), "\n")

	if len(dirLines) < 2 || len(filenameLines) < 2 {
		return nil, fmt.Errorf("解析Redis配置失败")
	}

	rdbPath := filepath.Join(strings.TrimSpace(dirLines[1]), strings.TrimSpace(filenameLines[1]))

	// 读取 RDB 文件
	rdbData, err := os.ReadFile(rdbPath)
	if err != nil {
		return nil, fmt.Errorf("读取RDB文件失败: %v", err)
	}

	_ = lastSaveOutput // 用于日志

	return rdbData, nil
}

// getDefaultPort 获取默认端口
func (s *DatabaseBackupService) getDefaultPort(dbType string) int {
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

// executeScript 执行脚本
func (s *DatabaseBackupService) executeScript(script string, vars map[string]interface{}) (string, error) {
	// 替换变量
	for key, val := range vars {
		jsonVal, _ := json.Marshal(val)
		script = strings.ReplaceAll(script, fmt.Sprintf("{{.%s}}", key), string(jsonVal))
	}

	cmd := exec.Command("bash", "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// QuickBackup 快速备份(简化接口)
func (s *DatabaseBackupService) QuickBackup(ctx context.Context, targetID uint, targetName string, dbConfig DatabaseConfig, storageType, storageConfig, storagePath string) (*BackupResult, error) {
	policy := &backup.BackupPolicy{
		Name:          fmt.Sprintf("quick_backup_%s", targetName),
		Type:          "database",
		StorageType:   storageType,
		StorageConfig: storageConfig,
		StoragePath:   storagePath,
		Compress:      true,
		CompressType:  "gzip",
		Timeout:       3600,
		RetryCount:    3,
	}

	target := &backup.BackupTarget{
		ID:      targetID,
		Name:    targetName,
		Type:    "database",
		DbType:  dbConfig.Type,
		DbConfig: func() string {
			data, _ := json.Marshal(dbConfig)
			return string(data)
		}(),
	}

	return s.Execute(ctx, policy, target)
}

// GetDatabases 获取数据库列表
func (s *DatabaseBackupService) GetDatabases(ctx context.Context, config DatabaseConfig) ([]string, error) {
	switch config.Type {
	case "mysql", "mariadb":
		return s.getMySQLDatabases(ctx, config)
	case "postgresql", "postgres":
		return s.getPostgreSQLDatabases(ctx, config)
	case "mongodb", "mongo":
		return s.getMongoDBDatabases(ctx, config)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}
}

// getMySQLDatabases 获取MySQL数据库列表
func (s *DatabaseBackupService) getMySQLDatabases(ctx context.Context, config DatabaseConfig) ([]string, error) {
	args := []string{
		"-h", config.Host,
		"-P", fmt.Sprintf("%d", config.Port),
		"-u", config.Username,
		fmt.Sprintf("-p%s", config.Password),
		"-e", "SHOW DATABASES",
	}

	cmd := exec.CommandContext(ctx, "mysql", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var databases []string
	for i, line := range lines {
		if i == 0 { // 跳过标题行
			continue
		}
		db := strings.TrimSpace(line)
		if db != "" && db != "information_schema" && db != "mysql" && db != "performance_schema" && db != "sys" {
			databases = append(databases, db)
		}
	}

	return databases, nil
}

// getPostgreSQLDatabases 获取PostgreSQL数据库列表
func (s *DatabaseBackupService) getPostgreSQLDatabases(ctx context.Context, config DatabaseConfig) ([]string, error) {
	env := os.Environ()
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", config.Password))

	cmd := exec.CommandContext(ctx, "psql",
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
		"-U", config.Username,
		"-d", "postgres",
		"-t",
		"-c", "SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'postgres'")
	cmd.Env = env

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var databases []string
	for _, line := range lines {
		db := strings.TrimSpace(line)
		if db != "" {
			databases = append(databases, db)
		}
	}

	return databases, nil
}

// getMongoDBDatabases 获取MongoDB数据库列表
func (s *DatabaseBackupService) getMongoDBDatabases(ctx context.Context, config DatabaseConfig) ([]string, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", config.Username, config.Password, config.Host, config.Port)
	cmd := exec.CommandContext(ctx, "mongosh",
		"--quiet",
		"--eval", "db.getMongo().getDBNames().join('\\n')",
		uri)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var databases []string
	for _, line := range lines {
		db := strings.TrimSpace(line)
		if db != "" && db != "admin" && db != "local" && db != "config" {
			databases = append(databases, db)
		}
	}

	return databases, nil
}

// CompressService 压缩服务
type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (s *CompressService) Compress(data []byte, compressType string) ([]byte, error) {
	// 简化实现，实际应使用压缩库
	switch compressType {
	case "gzip":
		return s.gzipCompress(data)
	case "zstd":
		return s.zstdCompress(data)
	default:
		return data, nil
	}
}

func (s *CompressService) gzipCompress(data []byte) ([]byte, error) {
	// 使用系统 gzip 命令
	cmd := exec.Command("gzip", "-c")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if _, err := io.Copy(stdin, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

func (s *CompressService) zstdCompress(data []byte) ([]byte, error) {
	cmd := exec.Command("zstd", "-c")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if _, err := io.Copy(stdin, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

func (s *CompressService) CompressDir(dir string) ([]byte, error) {
	// 使用 tar + gzip
	cmd := exec.Command("tar", "-czf", "-", "-C", dir, ".")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

func (s *CompressService) Decompress(data []byte) ([]byte, error) {
	cmd := exec.Command("gunzip", "-c")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if _, err := io.Copy(stdin, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

// EncryptService 加密服务
type EncryptService struct{}

func NewEncryptService() *EncryptService {
	return &EncryptService{}
}

func (s *EncryptService) Encrypt(data []byte, key string) ([]byte, error) {
	// 使用 openssl 进行加密
	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-salt", "-pbkdf2", "-pass", "pass:"+key)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if _, err := io.Copy(stdin, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

func (s *EncryptService) Decrypt(data []byte, key string) ([]byte, error) {
	cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2", "-pass", "pass:"+key)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if _, err := io.Copy(stdin, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

// NotifyService 通知服务
type NotifyService struct{}

func NewNotifyService() *NotifyService {
	return &NotifyService{}
}

func (s *NotifyService) NotifyBackupSuccess(policy *backup.BackupPolicy, result *BackupResult) error {
	// 发送备份成功通知
	// 实际实现应该调用通知服务
	return nil
}

func (s *NotifyService) NotifyBackupFailed(policy *backup.BackupPolicy, err error) error {
	// 发送备份失败通知
	return nil
}
