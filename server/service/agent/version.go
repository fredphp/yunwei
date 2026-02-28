package agent

import (
        "crypto/md5"
        "encoding/hex"
        "encoding/json"
        "fmt"
        "regexp"
        "strconv"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/model/agent"

        "gorm.io/gorm"
)

// VersionManager 版本管理器
type VersionManager struct {
        storagePath string
}

// NewVersionManager 创建版本管理器
func NewVersionManager(storagePath string) *VersionManager {
        return &VersionManager{
                storagePath: storagePath,
        }
}

// ==================== 版本操作 ====================

// CreateVersion 创建新版本
func (vm *VersionManager) CreateVersion(v *agent.AgentVersion) error {
        // 解析版本号生成版本代码
        versionCode, err := vm.ParseVersionCode(v.Version)
        if err != nil {
                return fmt.Errorf("无效的版本号: %w", err)
        }
        v.VersionCode = versionCode

        // 检查版本是否已存在
        var existing agent.AgentVersion
        err = global.DB.Where("version = ? AND platform = ? AND arch = ?", 
                v.Version, v.Platform, v.Arch).First(&existing).Error
        if err == nil {
                return fmt.Errorf("版本 %s (%s/%s) 已存在", v.Version, v.Platform, v.Arch)
        }

        // 设置默认值
        if v.ReleaseType == "" {
                v.ReleaseType = "stable"
        }
        if v.BuildTime.IsZero() {
                v.BuildTime = time.Now()
        }

        return global.DB.Create(v).Error
}

// UpdateVersion 更新版本
func (vm *VersionManager) UpdateVersion(v *agent.AgentVersion) error {
        return global.DB.Save(v).Error
}

// DeleteVersion 删除版本
func (vm *VersionManager) DeleteVersion(id uint) error {
        // 检查是否有正在使用此版本的 Agent
        var count int64
        global.DB.Model(&agent.Agent{}).Where("version = ? AND status != ?", 
                id, agent.AgentStatusDisabled).Count(&count)
        if count > 0 {
                return fmt.Errorf("有 %d 个 Agent 正在使用此版本，无法删除", count)
        }

        return global.DB.Delete(&agent.AgentVersion{}, id).Error
}

// GetVersion 获取版本详情
func (vm *VersionManager) GetVersion(id uint) (*agent.AgentVersion, error) {
        var v agent.AgentVersion
        err := global.DB.First(&v, id).Error
        return &v, err
}

// GetVersionByNumber 根据版本号获取
func (vm *VersionManager) GetVersionByNumber(version, platform, arch string) (*agent.AgentVersion, error) {
        var v agent.AgentVersion
        err := global.DB.Where("version = ? AND platform = ? AND arch = ?", 
                version, platform, arch).First(&v).Error
        return &v, err
}

// ListVersions 列出版本
func (vm *VersionManager) ListVersions(filter *VersionFilter) ([]agent.AgentVersion, int64, error) {
        query := global.DB.Model(&agent.AgentVersion{})

        if filter != nil {
                if filter.Platform != "" {
                        query = query.Where("platform = ?", filter.Platform)
                }
                if filter.Arch != "" {
                        query = query.Where("arch = ?", filter.Arch)
                }
                if filter.ReleaseType != "" {
                        query = query.Where("release_type = ?", filter.ReleaseType)
                }
                if filter.Enabled != nil {
                        query = query.Where("enabled = ?", *filter.Enabled)
                }
        }

        var total int64
        query.Count(&total)

        var versions []agent.AgentVersion
        err := query.Order("version_code DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&versions).Error
        return versions, total, err
}

// VersionFilter 版本过滤器
type VersionFilter struct {
        Platform    string `json:"platform"`
        Arch        string `json:"arch"`
        ReleaseType string `json:"releaseType"`
        Enabled     *bool  `json:"enabled"`
        Limit       int    `json:"limit"`
        Offset      int    `json:"offset"`
}

// GetLatestVersion 获取最新版本
func (vm *VersionManager) GetLatestVersion(platform, arch, releaseType string) (*agent.AgentVersion, error) {
        var v agent.AgentVersion
        query := global.DB.Where("platform = ? AND arch = ? AND enabled = ?", platform, arch, true)
        if releaseType != "" {
                query = query.Where("release_type = ?", releaseType)
        }
        err := query.Order("version_code DESC").First(&v).Error
        return &v, err
}

// ==================== 版本比对 ====================

// ParseVersionCode 解析版本号为数字代码
// 版本格式: v1.2.3 或 1.2.3
// 转换为: 1002003 (主版本*1000000 + 次版本*1000 + 补丁版本)
func (vm *VersionManager) ParseVersionCode(version string) (int, error) {
        // 移除 v 前缀
        version = strings.TrimPrefix(version, "v")
        
        // 解析版本号
        parts := strings.Split(version, ".")
        if len(parts) < 2 {
                return 0, fmt.Errorf("invalid version format: %s", version)
        }

        major, _ := strconv.Atoi(parts[0])
        minor, _ := strconv.Atoi(parts[1])
        patch := 0
        if len(parts) > 2 {
                // 处理可能的后缀 (如 1.2.3-beta)
                patchStr := regexp.MustCompile(`^(\d+)`).FindString(parts[2])
                patch, _ = strconv.Atoi(patchStr)
        }

        return major*1000000 + minor*1000 + patch, nil
}

// CompareVersions 比较两个版本
// 返回: 1 表示 v1 > v2, -1 表示 v1 < v2, 0 表示相等
func (vm *VersionManager) CompareVersions(v1, v2 string) (int, error) {
        code1, err := vm.ParseVersionCode(v1)
        if err != nil {
                return 0, err
        }
        code2, err := vm.ParseVersionCode(v2)
        if err != nil {
                return 0, err
        }

        if code1 > code2 {
                return 1, nil
        } else if code1 < code2 {
                return -1, nil
        }
        return 0, nil
}

// NeedUpgrade 检查是否需要升级
func (vm *VersionManager) NeedUpgrade(currentVersion, targetVersion string) (bool, error) {
        cmp, err := vm.CompareVersions(currentVersion, targetVersion)
        if err != nil {
                return false, err
        }
        return cmp < 0, nil
}

// CanUpgradeTo 检查是否可以升级到目标版本
func (vm *VersionManager) CanUpgradeTo(agentVersion *agent.AgentVersion, targetVersion string) (bool, error) {
        // 检查最低版本要求
        if agentVersion.MinVersion != "" {
                cmp, err := vm.CompareVersions(targetVersion, agentVersion.MinVersion)
                if err != nil {
                        return false, err
                }
                if cmp < 0 {
                        return false, fmt.Errorf("当前版本 %s 低于最低要求 %s", targetVersion, agentVersion.MinVersion)
                }
        }
        return true, nil
}

// ==================== 升级检查 ====================

// CheckUpgrade 检查 Agent 是否需要升级
func (vm *VersionManager) CheckUpgrade(a *agent.Agent) (*UpgradeInfo, error) {
        info := &UpgradeInfo{
                NeedUpgrade:    false,
                CurrentVersion: a.Version,
        }

        // 获取最新版本
        latest, err := vm.GetLatestVersion(a.Platform, a.Arch, a.UpgradeChannel)
        if err != nil {
                if err == gorm.ErrRecordNotFound {
                        return info, nil
                }
                return nil, err
        }

        info.LatestVersion = latest

        // 比较版本
        needUpgrade, err := vm.NeedUpgrade(a.Version, latest.Version)
        if err != nil {
                return nil, err
        }

        if needUpgrade {
                info.NeedUpgrade = true
                info.TargetVersion = latest.Version
                info.ForceUpdate = latest.ForceUpdate
                info.Changelog = latest.Changelog
                info.DownloadURL = latest.FileURL
                info.FileMD5 = latest.FileMD5
                info.FileSize = latest.FileSize
        }

        return info, nil
}

// UpgradeInfo 升级信息
type UpgradeInfo struct {
        NeedUpgrade    bool                 `json:"needUpgrade"`
        CurrentVersion string               `json:"currentVersion"`
        TargetVersion  string               `json:"targetVersion"`
        ForceUpdate    bool                 `json:"forceUpdate"`
        Changelog      string               `json:"changelog"`
        DownloadURL    string               `json:"downloadUrl"`
        FileMD5        string               `json:"fileMd5"`
        FileSize       int64                `json:"fileSize"`
        LatestVersion  *agent.AgentVersion  `json:"latestVersion"`
}

// ==================== 版本下载 ====================

// IncrementDownloadCount 增加下载计数
func (vm *VersionManager) IncrementDownloadCount(id uint) error {
        return global.DB.Model(&agent.AgentVersion{}).
                Where("id = ?", id).
                UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error
}

// IncrementInstallCount 增加安装计数
func (vm *VersionManager) IncrementInstallCount(id uint) error {
        return global.DB.Model(&agent.AgentVersion{}).
                Where("id = ?", id).
                UpdateColumn("install_count", gorm.Expr("install_count + 1")).Error
}

// ==================== 版本校验 ====================

// VerifyFileMD5 校验文件 MD5
func (vm *VersionManager) VerifyFileMD5(fileData []byte, expectedMD5 string) bool {
        hash := md5.Sum(fileData)
        actualMD5 := hex.EncodeToString(hash[:])
        return strings.EqualFold(actualMD5, expectedMD5)
}

// ==================== 批量操作 ====================

// GetVersionStats 获取版本统计
func (vm *VersionManager) GetVersionStats() (*VersionStats, error) {
        stats := &VersionStats{}

        // 总版本数
        global.DB.Model(&agent.AgentVersion{}).Count(&stats.TotalVersions)

        // 各平台版本数
        global.DB.Model(&agent.AgentVersion{}).
                Select("platform, count(*) as count").
                Group("platform").
                Scan(&stats.ByPlatform)

        // 各架构版本数
        global.DB.Model(&agent.AgentVersion{}).
                Select("arch, count(*) as count").
                Group("arch").
                Scan(&stats.ByArch)

        // Agent 版本分布
        global.DB.Model(&agent.Agent{}).
                Select("version, count(*) as count").
                Where("version != ''").
                Group("version").
                Order("count DESC").
                Limit(10).
                Scan(&stats.AgentVersionDistribution)

        // 需要升级的 Agent 数量
        var agents []agent.Agent
        global.DB.Where("status = ?", agent.AgentStatusOnline).Find(&agents)
        for _, a := range agents {
                info, err := vm.CheckUpgrade(&a)
                if err == nil && info.NeedUpgrade {
                        stats.NeedUpgradeCount++
                }
        }

        return stats, nil
}

// VersionStats 版本统计
type VersionStats struct {
        TotalVersions          int64 `json:"totalVersions"`
        NeedUpgradeCount       int64 `json:"needUpgradeCount"`
        ByPlatform             []struct {
                Platform string `json:"platform"`
                Count    int    `json:"count"`
        } `json:"byPlatform"`
        ByArch []struct {
                Arch  string `json:"arch"`
                Count int    `json:"count"`
        } `json:"byArch"`
        AgentVersionDistribution []struct {
                Version string `json:"version"`
                Count   int    `json:"count"`
        } `json:"agentVersionDistribution"`
}

// ==================== 版本模板 ====================

// GenerateVersionFromTemplate 从模板生成版本
func (vm *VersionManager) GenerateVersionFromTemplate(template *VersionTemplate) (*agent.AgentVersion, error) {
        v := &agent.AgentVersion{
                Version:     template.Version,
                Platform:    template.Platform,
                Arch:        template.Arch,
                ReleaseType: template.ReleaseType,
                Changelog:   template.Changelog,
                FileURL:     template.FileURL,
                MinVersion:  template.MinVersion,
                Enabled:     true,
                BuildTime:   time.Now(),
        }

        if template.FileData != nil {
                hash := md5.Sum(template.FileData)
                v.FileMD5 = hex.EncodeToString(hash[:])
                v.FileSize = int64(len(template.FileData))
        }

        if err := vm.CreateVersion(v); err != nil {
                return nil, err
        }

        return v, nil
}

// VersionTemplate 版本模板
type VersionTemplate struct {
        Version     string `json:"version"`
        Platform    string `json:"platform"`
        Arch        string `json:"arch"`
        ReleaseType string `json:"releaseType"`
        Changelog   string `json:"changelog"`
        FileURL     string `json:"fileUrl"`
        FileData    []byte `json:"-"`
        MinVersion  string `json:"minVersion"`
}

// ==================== 兼容性检查 ====================

// CheckCompatibility 检查兼容性
func (vm *VersionManager) CheckCompatibility(platform, arch string) (*CompatibilityInfo, error) {
        info := &CompatibilityInfo{
                Platform: platform,
                Arch:     arch,
        }

        // 检查是否有可用版本
        var count int64
        global.DB.Model(&agent.AgentVersion{}).
                Where("platform = ? AND arch = ? AND enabled = ?", platform, arch, true).
                Count(&count)
        
        info.Supported = count > 0
        info.AvailableVersions = int(count)

        // 获取最新版本
        latest, err := vm.GetLatestVersion(platform, arch, "stable")
        if err == nil {
                info.LatestStableVersion = latest.Version
        }

        latestBeta, err := vm.GetLatestVersion(platform, arch, "beta")
        if err == nil {
                info.LatestBetaVersion = latestBeta.Version
        }

        return info, nil
}

// CompatibilityInfo 兼容性信息
type CompatibilityInfo struct {
        Platform           string `json:"platform"`
        Arch               string `json:"arch"`
        Supported          bool   `json:"supported"`
        AvailableVersions  int    `json:"availableVersions"`
        LatestStableVersion string `json:"latestStableVersion"`
        LatestBetaVersion  string `json:"latestBetaVersion"`
}

// ==================== 版本导出 ====================

// ExportVersions 导出版本列表
func (vm *VersionManager) ExportVersions(platform, arch string) ([]agent.AgentVersion, error) {
        var versions []agent.AgentVersion
        query := global.DB.Model(&agent.AgentVersion{})
        if platform != "" {
                query = query.Where("platform = ?", platform)
        }
        if arch != "" {
                query = query.Where("arch = ?", arch)
        }
        err := query.Order("version_code DESC").Find(&versions).Error
        return versions, err
}

// ImportVersions 导入版本列表
func (vm *VersionManager) ImportVersions(versions []agent.AgentVersion) ([]ImportResult, error) {
        results := make([]ImportResult, len(versions))
        
        for i, v := range versions {
                err := vm.CreateVersion(&v)
                if err != nil {
                        results[i] = ImportResult{
                                Version: v.Version,
                                Success: false,
                                Error:   err.Error(),
                        }
                } else {
                        results[i] = ImportResult{
                                Version: v.Version,
                                Success: true,
                                ID:      v.ID,
                        }
                }
        }
        
        return results, nil
}

// ImportResult 导入结果
type ImportResult struct {
        Version string `json:"version"`
        Success bool   `json:"success"`
        ID      uint   `json:"id"`
        Error   string `json:"error"`
}

// ToJSON 转换为JSON
func ToJSON(v *agent.AgentVersion) string {
        data, _ := json.Marshal(v)
        return string(data)
}
