package migration

import (
        "crypto/md5"
        "encoding/hex"
        "fmt"
        "io/ioutil"
        "os"
        "path/filepath"
        "sort"
        "strings"
        "time"
        "yunwei/global"
        "yunwei/model/system"

        "gorm.io/gorm"
)

const (
        // MigrationStatusSuccess 执行成功
        MigrationStatusSuccess = "success"
        // MigrationStatusFailed 执行失败
        MigrationStatusFailed = "failed"
)

// MigrationManager 迁移管理器
type MigrationManager struct {
        db             *gorm.DB
        migrationsDirs []string // 支持多个迁移目录
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(migrationsDir string) *MigrationManager {
        return &MigrationManager{
                db:             global.DB,
                migrationsDirs: []string{migrationsDir},
        }
}

// NewMigrationManagerWithDirs 创建支持多目录的迁移管理器
func NewMigrationManagerWithDirs(migrationsDirs []string) *MigrationManager {
        return &MigrationManager{
                db:             global.DB,
                migrationsDirs: migrationsDirs,
        }
}

// Run 执行所有待执行的迁移
func (m *MigrationManager) Run() error {
        fmt.Println("\n===========================================")
        fmt.Println("  开始执行数据库迁移...")
        fmt.Println("===========================================")

        // 1. 确保迁移表存在
        if err := m.ensureMigrationTable(); err != nil {
                return fmt.Errorf("创建迁移表失败: %v", err)
        }

        // 2. 获取已执行的迁移
        executedMigrations, err := m.getExecutedMigrations()
        if err != nil {
                return fmt.Errorf("获取已执行迁移失败: %v", err)
        }

        // 3. 扫描迁移文件
        migrationFiles, err := m.scanMigrationFiles()
        if err != nil {
                return fmt.Errorf("扫描迁移文件失败: %v", err)
        }

        if len(migrationFiles) == 0 {
                fmt.Println("  没有找到迁移文件")
                return nil
        }

        // 4. 过滤出待执行的迁移
        pendingMigrations := m.filterPendingMigrations(migrationFiles, executedMigrations)

        if len(pendingMigrations) == 0 {
                fmt.Println("  所有迁移已执行，无需处理")
                return nil
        }

        fmt.Printf("  发现 %d 个待执行的迁移文件\n", len(pendingMigrations))

        // 5. 依次执行迁移
        successCount := 0
        failCount := 0

        for _, file := range pendingMigrations {
                if err := m.executeMigration(file); err != nil {
                        fmt.Printf("  ❌ 执行失败: %s (%v)\n", file.Name, err)
                        failCount++
                        // 继续执行下一个迁移，不中断
                } else {
                        fmt.Printf("  ✅ 执行成功: %s\n", file.Name)
                        successCount++
                }
        }

        fmt.Println("===========================================")
        fmt.Printf("  迁移完成: 成功 %d, 失败 %d\n", successCount, failCount)
        fmt.Println("===========================================\n")

        return nil
}

// ensureMigrationTable 确保迁移表存在
func (m *MigrationManager) ensureMigrationTable() error {
        return m.db.AutoMigrate(&system.SysMigration{})
}

// getExecutedMigrations 获取已执行的迁移记录
func (m *MigrationManager) getExecutedMigrations() (map[string]system.SysMigration, error) {
        var migrations []system.SysMigration
        if err := m.db.Find(&migrations).Error; err != nil {
                return nil, err
        }

        result := make(map[string]system.SysMigration)
        for _, m := range migrations {
                result[m.Name] = m
        }
        return result, nil
}

// MigrationFile 迁移文件信息
type MigrationFile struct {
        Name     string
        Path     string
        Checksum string
}

// scanMigrationFiles 扫描迁移文件目录
func (m *MigrationManager) scanMigrationFiles() ([]MigrationFile, error) {
        var files []MigrationFile
        fileMap := make(map[string]bool) // 用于去重

        // 遍历所有迁移目录
        for _, migrationsDir := range m.migrationsDirs {
                // 检查目录是否存在
                if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
                        continue
                }

                // 遍历目录
                err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
                        if err != nil {
                                return err
                        }

                        // 只处理 .sql 文件
                        if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sql") {
                                // 去重：同名文件只保留第一个
                                if fileMap[info.Name()] {
                                        return nil
                                }
                                fileMap[info.Name()] = true

                                // 计算文件校验和
                                checksum, err := m.calculateFileChecksum(path)
                                if err != nil {
                                        return fmt.Errorf("计算校验和失败 %s: %v", path, err)
                                }

                                files = append(files, MigrationFile{
                                        Name:     info.Name(),
                                        Path:     path,
                                        Checksum: checksum,
                                })
                        }

                        return nil
                })

                if err != nil {
                        return nil, err
                }
        }

        // 按文件名排序（确保按顺序执行）
        sort.Slice(files, func(i, j int) bool {
                return files[i].Name < files[j].Name
        })

        return files, nil
}

// filterPendingMigrations 过滤出待执行的迁移
func (m *MigrationManager) filterPendingMigrations(files []MigrationFile, executed map[string]system.SysMigration) []MigrationFile {
        var pending []MigrationFile

        for _, file := range files {
                executedRecord, exists := executed[file.Name]
                // 如果从未执行过，或者校验和变了（文件被修改），则需要执行
                if !exists || executedRecord.Checksum != file.Checksum {
                        pending = append(pending, file)
                }
        }

        return pending
}

// executeMigration 执行单个迁移
func (m *MigrationManager) executeMigration(file MigrationFile) error {
        // 读取文件内容
        content, err := ioutil.ReadFile(file.Path)
        if err != nil {
                return fmt.Errorf("读取文件失败: %v", err)
        }

        startTime := time.Now()

        // 执行 SQL
        sqlContent := string(content)
        err = m.executeSQL(sqlContent)

        executionMs := time.Since(startTime).Milliseconds()

        // 记录迁移结果
        status := MigrationStatusSuccess
        errorMsg := ""
        if err != nil {
                status = MigrationStatusFailed
                errorMsg = err.Error()
        }

        // 保存或更新迁移记录
        migration := system.SysMigration{
                Name:        file.Name,
                Checksum:    file.Checksum,
                ExecutionMs: executionMs,
                Status:      status,
                ErrorMsg:    errorMsg,
        }

        // 先尝试更新已存在的记录
        result := m.db.Model(&system.SysMigration{}).
                Where("name = ?", file.Name).
                Updates(map[string]interface{}{
                        "checksum":     file.Checksum,
                        "execution_ms": executionMs,
                        "status":       status,
                        "error_msg":    errorMsg,
                        "created_at":   time.Now(),
                })

        if result.RowsAffected == 0 {
                // 不存在则创建新记录
                m.db.Create(&migration)
        }

        return err
}

// executeSQL 执行 SQL 语句
func (m *MigrationManager) executeSQL(sql string) error {
        // 分割 SQL 语句（以分号结尾的语句）
        statements := m.splitSQLStatements(sql)

        for i, stmt := range statements {
                stmt = strings.TrimSpace(stmt)
                if stmt == "" {
                        continue
                }

                // 执行单条 SQL
                if err := m.db.Exec(stmt).Error; err != nil {
                        // 忽略某些可以接受的错误
                        if !m.isAcceptableError(err) {
                                preview := stmt
                                if len(preview) > 200 {
                                        preview = preview[:200] + "..."
                                }
                                return fmt.Errorf("语句 %d 执行失败: %v\nSQL: %s", i+1, err, preview)
                        }
                }
        }

        return nil
}

// splitSQLStatements 分割 SQL 语句
func (m *MigrationManager) splitSQLStatements(sql string) []string {
        // 简单的分号分割，处理基本的 SQL 文件
        // 对于复杂的存储过程等，可能需要更复杂的解析
        var statements []string
        var current strings.Builder
        inQuote := false
        quoteChar := byte(0)

        for i := 0; i < len(sql); i++ {
                ch := sql[i]

                // 处理引号
                if (ch == '\'' || ch == '`') && (i == 0 || sql[i-1] != '\\') {
                        if !inQuote {
                                inQuote = true
                                quoteChar = ch
                        } else if ch == quoteChar {
                                inQuote = false
                                quoteChar = 0
                        }
                }

                // 分号分割（不在引号内时）
                if ch == ';' && !inQuote {
                        stmt := strings.TrimSpace(current.String())
                        if stmt != "" {
                                statements = append(statements, stmt)
                        }
                        current.Reset()
                } else {
                        current.WriteByte(ch)
                }
        }

        // 添加最后一条语句
        stmt := strings.TrimSpace(current.String())
        if stmt != "" {
                statements = append(statements, stmt)
        }

        return statements
}

// isAcceptableError 判断是否是可以接受的错误
func (m *MigrationManager) isAcceptableError(err error) bool {
        errMsg := err.Error()
        acceptablePatterns := []string{
                "already exists",
                "duplicate",
                "Duplicate entry",
                "表已存在",
                "索引已存在",
        }

        for _, pattern := range acceptablePatterns {
                if strings.Contains(errMsg, pattern) {
                        return true
                }
        }

        return false
}

// calculateFileChecksum 计算文件校验和
func (m *MigrationManager) calculateFileChecksum(filePath string) (string, error) {
        content, err := ioutil.ReadFile(filePath)
        if err != nil {
                return "", err
        }

        hash := md5.Sum(content)
        return hex.EncodeToString(hash[:]), nil
}

// GetMigrationStatus 获取迁移状态
func (m *MigrationManager) GetMigrationStatus() ([]system.SysMigration, error) {
        var migrations []system.SysMigration
        if err := m.db.Order("created_at desc").Find(&migrations).Error; err != nil {
                return nil, err
        }
        return migrations, nil
}

// ResetMigration 重置指定迁移记录（使其可以重新执行）
func (m *MigrationManager) ResetMigration(name string) error {
        return m.db.Where("name = ?", name).Delete(&system.SysMigration{}).Error
}

// ResetAllMigrations 重置所有迁移记录
func (m *MigrationManager) ResetAllMigrations() error {
        return m.db.Where("1 = 1").Delete(&system.SysMigration{}).Error
}
