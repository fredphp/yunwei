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
	"sync"
	"time"

	"yunwei/model/backup"
)

// FileBackupService 文件备份服务
type FileBackupService struct {
	storageService  *StorageService
	compressService *CompressService
	encryptService  *EncryptService
	notifyService   *NotifyService
}

// NewFileBackupService 创建文件备份服务
func NewFileBackupService() *FileBackupService {
	return &FileBackupService{
		storageService:  NewStorageService(),
		compressService: NewCompressService(),
		encryptService:  NewEncryptService(),
		notifyService:   NewNotifyService(),
	}
}

// FileConfig 文件备份配置
type FileConfig struct {
	SourcePath    string   `json:"source_path"`
	ExcludePaths  []string `json:"exclude_paths"`
	IncludeHidden bool     `json:"include_hidden"`
	FollowLinks   bool     `json:"follow_links"`
	MaxFileSize   int64    `json:"max_file_size"` // MB
	SplitSize     int64    `json:"split_size"`    // MB, 分卷大小
}

// FileBackupResult 文件备份结果
type FileBackupResult struct {
	Success       bool
	FilePath      string
	FileName      string
	FileSize      int64
	CompressSize  int64
	Checksum      string
	Duration      int
	TotalFiles    int
	TotalDirs     int
	SkippedFiles  int
	SkippedSize   int64
	Error         error
	Log           string
}

// Execute 执行文件备份
func (s *FileBackupService) Execute(ctx context.Context, policy *backup.BackupPolicy, target *backup.BackupTarget) (*FileBackupResult, error) {
	startTime := time.Now()
	result := &FileBackupResult{}

	// 解析文件配置
	var fileConfig FileConfig
	if target != nil && target.DbConfig != "" {
		if err := json.Unmarshal([]byte(target.DbConfig), &fileConfig); err != nil {
			fileConfig.SourcePath = target.RootPath
		}
	} else if policy.SourceConfig != "" {
		if err := json.Unmarshal([]byte(policy.SourceConfig), &fileConfig); err != nil {
			return nil, fmt.Errorf("解析文件配置失败: %v", err)
		}
	}

	// 设置源路径
	if policy.SourcePath != "" {
		fileConfig.SourcePath = policy.SourcePath
	}

	// 解析排除路径
	if policy.ExcludePaths != "" {
		fileConfig.ExcludePaths = strings.Split(policy.ExcludePaths, ",")
	}

	// 验证源路径
	if fileConfig.SourcePath == "" {
		return nil, fmt.Errorf("源路径不能为空")
	}

	sourceInfo, err := os.Stat(fileConfig.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("源路径不存在: %v", err)
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "file_backup_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	sourceName := filepath.Base(fileConfig.SourcePath)
	if sourceName == "." || sourceName == "/" {
		sourceName = "root"
	}
	fileName := fmt.Sprintf("%s_%s.tar.gz", sourceName, timestamp)
	filePath := filepath.Join(tempDir, fileName)

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 开始备份文件: %s\n", time.Now().Format("2006-01-02 15:04:05"), fileConfig.SourcePath))

	// 执行预备份脚本
	if policy.PreScript != "" {
		logBuilder.WriteString("[INFO] 执行预备份脚本...\n")
		if output, err := s.executeScript(policy.PreScript, map[string]interface{}{
			"target": target,
			"policy": policy,
			"config": fileConfig,
		}); err != nil {
			logBuilder.WriteString(fmt.Sprintf("[WARN] 预备份脚本执行失败: %v\n", err))
		} else {
			logBuilder.WriteString(fmt.Sprintf("[INFO] 预备份脚本输出: %s\n", output))
		}
	}

	// 构建排除参数
	var excludeArgs []string
	for _, exclude := range fileConfig.ExcludePaths {
		exclude = strings.TrimSpace(exclude)
		if exclude != "" {
			excludeArgs = append(excludeArgs, "--exclude", exclude)
		}
	}

	// 构建tar命令
	tarArgs := []string{"-czf", filePath}
	if len(excludeArgs) > 0 {
		tarArgs = append(tarArgs, excludeArgs...)
	}
	if fileConfig.IncludeHidden {
		// 包含隐藏文件是默认行为
	}
	if !fileConfig.FollowLinks {
		tarArgs = append(tarArgs, "--no-recursion")
	}

	// 添加源路径
	if sourceInfo.IsDir() {
		tarArgs = append(tarArgs, "-C", fileConfig.SourcePath, ".")
	} else {
		tarArgs = append(tarArgs, fileConfig.SourcePath)
	}

	// 执行打包
	logBuilder.WriteString(fmt.Sprintf("[INFO] 执行打包命令: tar %s\n", strings.Join(tarArgs, " ")))
	cmd := exec.CommandContext(ctx, "tar", tarArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 打包失败: %v, stderr: %s\n", err, stderr.String()))
		result.Log = logBuilder.String()
		return result, err
	}

	logBuilder.WriteString("[INFO] 文件打包完成\n")

	// 统计文件数量
	if sourceInfo.IsDir() {
		result.TotalFiles, result.TotalDirs, result.SkippedFiles, _ = s.countFiles(fileConfig.SourcePath, fileConfig.ExcludePaths)
		logBuilder.WriteString(fmt.Sprintf("[INFO] 统计: 文件 %d, 目录 %d, 跳过 %d\n", result.TotalFiles, result.TotalDirs, result.SkippedFiles))
	}

	// 加密
	var finalData []byte
	if policy.Encrypt {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			result.Error = err
			result.Success = false
			logBuilder.WriteString(fmt.Sprintf("[ERROR] 读取文件失败: %v\n", err))
			result.Log = logBuilder.String()
			return result, err
		}

		encrypted, err := s.encryptService.Encrypt(fileData, policy.EncryptKey)
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
		if err := os.WriteFile(filePath, finalData, 0644); err != nil {
			result.Error = err
			result.Success = false
			logBuilder.WriteString(fmt.Sprintf("[ERROR] 写入加密文件失败: %v\n", err))
			result.Log = logBuilder.String()
			return result, err
		}
		logBuilder.WriteString("[INFO] 文件加密完成\n")
	}

	// 计算校验和
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 读取文件失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}
	checksum := sha256.Sum256(fileData)
	result.Checksum = hex.EncodeToString(checksum[:])

	// 获取文件大小
	fileInfo, _ := os.Stat(filePath)
	result.FileSize = fileInfo.Size()
	result.CompressSize = fileInfo.Size()
	result.FilePath = filePath
	result.FileName = fileName

	logBuilder.WriteString(fmt.Sprintf("[INFO] 备份文件大小: %d 字节\n", result.FileSize))

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
			"target": target,
			"policy": policy,
			"config": fileConfig,
			"result": result,
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

// countFiles 统计文件数量
func (s *FileBackupService) countFiles(rootPath string, excludePaths []string) (files, dirs, skipped int, skippedSize int64) {
	excludeMap := make(map[string]bool)
	for _, p := range excludePaths {
		excludeMap[strings.TrimSpace(p)] = true
	}

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			skipped++
			return nil
		}

		// 检查是否排除
		relPath, _ := filepath.Rel(rootPath, path)
		if excludeMap[relPath] || excludeMap[path] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			skipped++
			skippedSize += info.Size()
			return nil
		}

		if info.IsDir() {
			dirs++
		} else {
			files++
		}
		return nil
	})

	return
}

// executeScript 执行脚本
func (s *FileBackupService) executeScript(script string, vars map[string]interface{}) (string, error) {
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

// IncrementalBackup 增量备份
func (s *FileBackupService) IncrementalBackup(ctx context.Context, policy *backup.BackupPolicy, target *backup.BackupTarget, baseBackupID uint) (*FileBackupResult, error) {
	startTime := time.Now()
	result := &FileBackupResult{}

	// 解析配置
	var fileConfig FileConfig
	if policy.SourceConfig != "" {
		json.Unmarshal([]byte(policy.SourceConfig), &fileConfig)
	}
	if policy.SourcePath != "" {
		fileConfig.SourcePath = policy.SourcePath
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "inc_backup_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// 获取基础备份信息，计算变更文件
	// 这里简化实现，实际应该从数据库获取基础备份
	changedFiles := s.getChangedFiles(fileConfig.SourcePath, time.Now().Add(-24*time.Hour))
	result.TotalFiles = len(changedFiles)

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_inc_%s.tar.gz", filepath.Base(fileConfig.SourcePath), timestamp)
	filePath := filepath.Join(tempDir, fileName)

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 开始增量备份\n", time.Now().Format("2006-01-02 15:04:05")))
	logBuilder.WriteString(fmt.Sprintf("[INFO] 变更文件数: %d\n", len(changedFiles)))

	if len(changedFiles) == 0 {
		logBuilder.WriteString("[INFO] 无变更文件，跳过备份\n")
		result.Success = true
		result.Duration = int(time.Since(startTime).Seconds())
		result.Log = logBuilder.String()
		return result, nil
	}

	// 创建文件列表
	listFile := filepath.Join(tempDir, "filelist.txt")
	listData := strings.Join(changedFiles, "\n")
	if err := os.WriteFile(listFile, []byte(listData), 0644); err != nil {
		return nil, err
	}

	// 打包变更文件
	args := []string{"-czf", filePath, "-T", listFile}
	cmd := exec.CommandContext(ctx, "tar", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 打包失败: %v\n", err))
		result.Log = logBuilder.String()
		return result, err
	}

	// 计算校验和
	fileData, _ := os.ReadFile(filePath)
	checksum := sha256.Sum256(fileData)
	result.Checksum = hex.EncodeToString(checksum[:])

	fileInfo, _ := os.Stat(filePath)
	result.FileSize = fileInfo.Size()
	result.CompressSize = fileInfo.Size()
	result.FilePath = filePath
	result.FileName = fileName

	// 上传
	storagePath, err := s.storageService.Upload(ctx, policy.StorageType, policy.StorageConfig, filePath, policy.StoragePath)
	if err != nil {
		return nil, err
	}
	result.FilePath = storagePath

	result.Success = true
	result.Duration = int(time.Since(startTime).Seconds())
	logBuilder.WriteString(fmt.Sprintf("[%s] 增量备份完成\n", time.Now().Format("2006-01-02 15:04:05")))
	result.Log = logBuilder.String()

	return result, nil
}

// getChangedFiles 获取变更文件
func (s *FileBackupService) getChangedFiles(rootPath string, since time.Time) []string {
	var changedFiles []string

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.ModTime().After(since) {
			changedFiles = append(changedFiles, path)
		}
		return nil
	})

	return changedFiles
}

// QuickBackup 快速文件备份
func (s *FileBackupService) QuickBackup(ctx context.Context, sourcePath, storageType, storageConfig, storagePath string) (*FileBackupResult, error) {
	policy := &backup.BackupPolicy{
		Name:         "quick_file_backup",
		Type:         "file",
		SourcePath:   sourcePath,
		StorageType:  storageType,
		StorageConfig: storageConfig,
		StoragePath:  storagePath,
		Compress:     true,
		Timeout:      3600,
	}

	target := &backup.BackupTarget{
		Type:     "filesystem",
		RootPath: sourcePath,
	}

	return s.Execute(ctx, policy, target)
}

// SyncBackup 同步备份(rsync)
func (s *FileBackupService) SyncBackup(ctx context.Context, sourcePath, destPath string, excludePaths []string) (*FileBackupResult, error) {
	startTime := time.Now()
	result := &FileBackupResult{}

	args := []string{"-avz", "--delete"}
	for _, exclude := range excludePaths {
		args = append(args, "--exclude", exclude)
	}
	args = append(args, sourcePath, destPath)

	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("[%s] 开始同步备份\n", time.Now().Format("2006-01-02 15:04:05")))
	logBuilder.WriteString(fmt.Sprintf("[INFO] 执行: rsync %s\n", strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, "rsync", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = err
		result.Success = false
		logBuilder.WriteString(fmt.Sprintf("[ERROR] 同步失败: %v\n%s\n", err, stderr.String()))
		result.Log = logBuilder.String()
		return result, err
	}

	result.Success = true
	result.Duration = int(time.Since(startTime).Seconds())
	result.Log = logBuilder.String() + stdout.String()

	return result, nil
}

// ==================== 存储服务 ====================

// StorageService 存储服务
type StorageService struct {
	s3Clients sync.Map
}

// NewStorageService 创建存储服务
func NewStorageService() *StorageService {
	return &StorageService{}
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string `json:"type"`
	Endpoint   string `json:"endpoint"`
	Region     string `json:"region"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	Path       string `json:"path"`
	LocalPath  string `json:"local_path"`
	NfsServer  string `json:"nfs_server"`
	NfsPath    string `json:"nfs_path"`
	FtpHost    string `json:"ftp_host"`
	FtpPort    int    `json:"ftp_port"`
	FtpUser    string `json:"ftp_user"`
	FtpPass    string `json:"ftp_pass"`
}

// Upload 上传文件
func (s *StorageService) Upload(ctx context.Context, storageType, configStr, filePath, destPath string) (string, error) {
	var config StorageConfig
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return "", fmt.Errorf("解析存储配置失败: %v", err)
		}
	}
	config.Type = storageType
	if destPath != "" {
		config.Path = destPath
	}

	switch config.Type {
	case "local":
		return s.uploadLocal(filePath, config)
	case "s3":
		return s.uploadS3(ctx, filePath, config)
	case "oss":
		return s.uploadOSS(ctx, filePath, config)
	case "nfs":
		return s.uploadNFS(filePath, config)
	case "ftp", "sftp":
		return s.uploadFTP(ctx, filePath, config)
	default:
		return s.uploadLocal(filePath, config)
	}
}

// uploadLocal 本地存储
func (s *StorageService) uploadLocal(filePath string, config StorageConfig) (string, error) {
	destPath := config.Path
	if destPath == "" {
		destPath = "/var/backups"
	}

	// 确保目录存在
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 复制文件
	fileName := filepath.Base(filePath)
	destFile := filepath.Join(destPath, fileName)

	srcData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destFile, srcData, 0644); err != nil {
		return "", err
	}

	return destFile, nil
}

// uploadS3 上传到S3
func (s *StorageService) uploadS3(ctx context.Context, filePath string, config StorageConfig) (string, error) {
	fileName := filepath.Base(filePath)
	key := filepath.Join(config.Path, fileName)

	// 使用 aws cli 上传
	args := []string{"s3", "cp", filePath, fmt.Sprintf("s3://%s/%s", config.Bucket, key)}
	if config.Endpoint != "" {
		args = append([]string{"--endpoint-url", config.Endpoint}, args...)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("S3上传失败: %v, %s", err, stderr.String())
	}

	return fmt.Sprintf("s3://%s/%s", config.Bucket, key), nil
}

// uploadOSS 上传到阿里云OSS
func (s *StorageService) uploadOSS(ctx context.Context, filePath string, config StorageConfig) (string, error) {
	fileName := filepath.Base(filePath)
	key := filepath.Join(config.Path, fileName)

	// 使用 ossutil 上传
	args := []string{"cp", filePath, fmt.Sprintf("oss://%s/%s", config.Bucket, key)}
	if config.Endpoint != "" {
		args = append(args, "-e", config.Endpoint)
	}
	if config.AccessKey != "" && config.SecretKey != "" {
		args = append(args, "-i", config.AccessKey, "-k", config.SecretKey)
	}

	cmd := exec.CommandContext(ctx, "ossutil", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("OSS上传失败: %v, %s", err, stderr.String())
	}

	return fmt.Sprintf("oss://%s/%s", config.Bucket, key), nil
}

// uploadNFS 上传到NFS
func (s *StorageService) uploadNFS(filePath string, config StorageConfig) (string, error) {
	// 假设NFS已经挂载
	destPath := config.Path
	if destPath == "" {
		destPath = "/mnt/nfs/backups"
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return "", err
	}

	fileName := filepath.Base(filePath)
	destFile := filepath.Join(destPath, fileName)

	srcData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destFile, srcData, 0644); err != nil {
		return "", err
	}

	return destFile, nil
}

// uploadFTP 上传到FTP/SFTP
func (s *StorageService) uploadFTP(ctx context.Context, filePath string, config StorageConfig) (string, error) {
	fileName := filepath.Base(filePath)
	destPath := config.Path
	if destPath == "" {
		destPath = "/backups"
	}
	destFile := filepath.Join(destPath, fileName)

	// 使用 lftp 上传
	ftpURL := fmt.Sprintf("ftp://%s:%s@%s:%d", config.FtpUser, config.FtpPass, config.FtpHost, config.FtpPort)
	cmd := exec.CommandContext(ctx, "lftp", "-c",
		fmt.Sprintf("open %s; put %s -o %s", ftpURL, filePath, destFile))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("FTP上传失败: %v, %s", err, stderr.String())
	}

	return destFile, nil
}

// Download 下载文件
func (s *StorageService) Download(ctx context.Context, storageType, configStr, remotePath, localPath string) error {
	var config StorageConfig
	if configStr != "" {
		json.Unmarshal([]byte(configStr), &config)
	}
	config.Type = storageType

	switch config.Type {
	case "local":
		return s.downloadLocal(remotePath, localPath)
	case "s3":
		return s.downloadS3(ctx, remotePath, localPath, config)
	case "oss":
		return s.downloadOSS(ctx, remotePath, localPath, config)
	default:
		return s.downloadLocal(remotePath, localPath)
	}
}

// downloadLocal 本地下载
func (s *StorageService) downloadLocal(remotePath, localPath string) error {
	data, err := os.ReadFile(remotePath)
	if err != nil {
		return err
	}
	return os.WriteFile(localPath, data, 0644)
}

// downloadS3 S3下载
func (s *StorageService) downloadS3(ctx context.Context, remotePath, localPath string, config StorageConfig) error {
	args := []string{"s3", "cp", remotePath, localPath}
	if config.Endpoint != "" {
		args = append([]string{"--endpoint-url", config.Endpoint}, args...)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	return cmd.Run()
}

// downloadOSS OSS下载
func (s *StorageService) downloadOSS(ctx context.Context, remotePath, localPath string, config StorageConfig) error {
	args := []string{"cp", remotePath, localPath}
	if config.Endpoint != "" {
		args = append(args, "-e", config.Endpoint)
	}
	if config.AccessKey != "" && config.SecretKey != "" {
		args = append(args, "-i", config.AccessKey, "-k", config.SecretKey)
	}

	cmd := exec.CommandContext(ctx, "ossutil", args...)
	return cmd.Run()
}

// Delete 删除文件
func (s *StorageService) Delete(ctx context.Context, storageType, configStr, remotePath string) error {
	var config StorageConfig
	if configStr != "" {
		json.Unmarshal([]byte(configStr), &config)
	}
	config.Type = storageType

	switch config.Type {
	case "local":
		return os.Remove(remotePath)
	case "s3":
		args := []string{"s3", "rm", remotePath}
		if config.Endpoint != "" {
			args = append([]string{"--endpoint-url", config.Endpoint}, args...)
		}
		cmd := exec.CommandContext(ctx, "aws", args...)
		return cmd.Run()
	case "oss":
		args := []string{"rm", remotePath}
		cmd := exec.CommandContext(ctx, "ossutil", args...)
		return cmd.Run()
	}

	return nil
}

// List 列出文件
func (s *StorageService) List(ctx context.Context, storageType, configStr, prefix string) ([]string, error) {
	var config StorageConfig
	if configStr != "" {
		json.Unmarshal([]byte(configStr), &config)
	}
	config.Type = storageType

	switch config.Type {
	case "local":
		return s.listLocal(prefix)
	case "s3":
		return s.listS3(ctx, prefix, config)
	default:
		return s.listLocal(prefix)
	}
}

// listLocal 本地文件列表
func (s *StorageService) listLocal(prefix string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(prefix)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return files, nil
}

// listS3 S3文件列表
func (s *StorageService) listS3(ctx context.Context, prefix string, config StorageConfig) ([]string, error) {
	args := []string{"s3", "ls", prefix}
	if config.Endpoint != "" {
		args = append([]string{"--endpoint-url", config.Endpoint}, args...)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var files []string
	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				files = append(files, parts[3])
			}
		}
	}

	return files, nil
}

// GetStorageUsage 获取存储使用情况
func (s *StorageService) GetStorageUsage(ctx context.Context, storageType, configStr string) (int64, int64, error) {
	var config StorageConfig
	if configStr != "" {
		json.Unmarshal([]byte(configStr), &config)
	}
	config.Type = storageType

	switch config.Type {
	case "local":
		return s.getLocalUsage(config.Path)
	default:
		return 0, 0, nil
	}
}

// getLocalUsage 本地存储使用情况
func (s *StorageService) getLocalUsage(path string) (int64, int64, error) {
	var totalSize int64
	var fileCount int64

	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	return totalSize, fileCount, nil
}
