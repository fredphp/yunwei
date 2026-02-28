package canary

import (
        "encoding/json"
        "fmt"
        "math"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/service/ai/llm"
        "yunwei/service/notify"
)

// DeployStatus 发布状态
type DeployStatus string

const (
        DeployStatusPending    DeployStatus = "pending"
        DeployStatusRunning    DeployStatus = "running"
        DeployStatusSuccess    DeployStatus = "success"
        DeployStatusFailed     DeployStatus = "failed"
        DeployStatusPaused     DeployStatus = "paused"
        DeployStatusRollback   DeployStatus = "rollback"
        DeployStatusAborted    DeployStatus = "aborted"
)

// StrategyType 发布策略类型
type StrategyType string

const (
        StrategyCanary    StrategyType = "canary"    // 金丝雀发布
        StrategyBlueGreen StrategyType = "bluegreen" // 蓝绿发布
        StrategyAB        StrategyType = "ab"        // A/B 测试
)

// CanaryRelease 灰度发布记录
type CanaryRelease struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`

        ClusterID   uint         `json:"clusterId" gorm:"index"`
        Namespace   string       `json:"namespace" gorm:"type:varchar(64)"`
        ServiceName string       `json:"serviceName" gorm:"type:varchar(128)"`
        Strategy    StrategyType `json:"strategy" gorm:"type:varchar(16)"`
        Status      DeployStatus `json:"status" gorm:"type:varchar(16)"`

        // 版本信息
        CurrentVersion string `json:"currentVersion" gorm:"type:varchar(64)"`
        NewVersion     string `json:"newVersion" gorm:"type:varchar(64)"`
        NewImage       string `json:"newImage" gorm:"type:varchar(256)"`

        // 流量控制
        TotalReplicas   int     `json:"totalReplicas"`
        CanaryReplicas  int     `json:"canaryReplicas"`
        CanaryWeight    float64 `json:"canaryWeight"` // 流量权重 0-100
        WeightStep      float64 `json:"weightStep"`   // 每步增加权重
        CurrentStep     int     `json:"currentStep"`
        TotalSteps      int     `json:"totalSteps"`

        // 阈值配置
        ErrorRateThreshold   float64 `json:"errorRateThreshold"`   // 错误率阈值
        LatencyThreshold     int64   `json:"latencyThreshold"`     // 延迟阈值(ms)
        SuccessRateThreshold float64 `json:"successRateThreshold"` // 成功率阈值

        // 监控指标
        CurrentErrorRate   float64 `json:"currentErrorRate"`
        CurrentLatency     int64   `json:"currentLatency"`
        CurrentSuccessRate float64 `json:"currentSuccessRate"`

        // AI 决策
        AIDecision    string  `json:"aiDecision" gorm:"type:text"`
        AIConfidence  float64 `json:"aiConfidence"`
        AIAutoPromote bool    `json:"aiAutoPromote"`

        // 回滚信息
        RollbackReason string     `json:"rollbackReason" gorm:"type:text"`
        RollbackAt     *time.Time `json:"rollbackAt"`

        // 时间
        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration    int64      `json:"duration"` // 秒
}

func (CanaryRelease) TableName() string {
        return "canary_releases"
}

// CanaryStep 灰度步骤
type CanaryStep struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`

        ReleaseID uint `json:"releaseId" gorm:"index"`

        StepNum    int     `json:"stepNum"`
        Weight     float64 `json:"weight"`
        Replicas   int     `json:"replicas"`

        // 监控数据
        ErrorRate    float64 `json:"errorRate"`
        Latency      int64   `json:"latency"`
        SuccessRate  float64 `json:"successRate"`
        RequestCount int64   `json:"requestCount"`

        // 状态
        Status       DeployStatus `json:"status" gorm:"type:varchar(16)"`
        StartedAt    *time.Time   `json:"startedAt"`
        CompletedAt  *time.Time   `json:"completedAt"`
        Duration     int64        `json:"duration"` // 秒
        PassedChecks int          `json:"passedChecks"`
        FailedChecks int          `json:"failedChecks"`
}

func (CanaryStep) TableName() string {
        return "canary_steps"
}

// CanaryConfig 灰度配置
type CanaryConfig struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        UpdatedAt time.Time `json:"updatedAt"`

        ClusterID   uint         `json:"clusterId" gorm:"index"`
        Namespace   string       `json:"namespace" gorm:"type:varchar(64)"`
        ServiceName string       `json:"serviceName" gorm:"type:varchar(128)"`
        Strategy    StrategyType `json:"strategy" gorm:"type:varchar(16)"`

        // 发布配置
        WeightStep      float64 `json:"weightStep"`
        TotalSteps      int     `json:"totalSteps"`
        StepDuration    int     `json:"stepDuration"` // 每步持续时间(秒)

        // 阈值
        ErrorRateThreshold   float64 `json:"errorRateThreshold"`
        LatencyThreshold     int64   `json:"latencyThreshold"`
        SuccessRateThreshold float64 `json:"successRateThreshold"`

        // 自动化
        AutoPromote   bool `json:"autoPromote"`   // 自动推进
        AutoRollback  bool `json:"autoRollback"`  // 自动回滚
        RequireManual bool `json:"requireManual"` // 需要人工确认

        Enabled bool `json:"enabled"`
}

func (CanaryConfig) TableName() string {
        return "canary_configs"
}

// CanaryManager 灰度发布管理器
type CanaryManager struct {
        llmClient *llm.GLM5Client
        notifier  notify.Notifier
        executor  CanaryExecutor
}

// CanaryExecutor 灰度执行器接口
type CanaryExecutor interface {
        UpdateServiceWeight(clusterID uint, namespace, serviceName string, canaryWeight float64) error
        GetCanaryMetrics(clusterID uint, namespace, serviceName string) (*CanaryMetrics, error)
        PromoteCanary(clusterID uint, namespace, serviceName string) error
        RollbackCanary(clusterID uint, namespace, serviceName string) error
        GetCurrentVersion(clusterID uint, namespace, serviceName string) (string, error)
}

// CanaryMetrics 灰度指标
type CanaryMetrics struct {
        ErrorRate    float64 `json:"errorRate"`
        Latency      int64   `json:"latency"`
        SuccessRate  float64 `json:"successRate"`
        RequestCount int64   `json:"requestCount"`
}

// NewCanaryManager 创建灰度发布管理器
func NewCanaryManager() *CanaryManager {
        return &CanaryManager{}
}

// SetLLMClient 设置 LLM 客户端
func (m *CanaryManager) SetLLMClient(client *llm.GLM5Client) {
        m.llmClient = client
}

// SetNotifier 设置通知器
func (m *CanaryManager) SetNotifier(notifier notify.Notifier) {
        m.notifier = notifier
}

// SetExecutor 设置执行器
func (m *CanaryManager) SetExecutor(executor CanaryExecutor) {
        m.executor = executor
}

// StartCanary 开始灰度发布
func (m *CanaryManager) StartCanary(clusterID uint, namespace, serviceName, newImage string, config CanaryConfig) (*CanaryRelease, error) {
        // 获取当前版本
        currentVersion, err := m.executor.GetCurrentVersion(clusterID, namespace, serviceName)
        if err != nil {
                return nil, fmt.Errorf("获取当前版本失败: %w", err)
        }

        release := &CanaryRelease{
                ClusterID:           clusterID,
                Namespace:           namespace,
                ServiceName:         serviceName,
                Strategy:            config.Strategy,
                Status:              DeployStatusPending,
                CurrentVersion:      currentVersion,
                NewVersion:          generateVersion(),
                NewImage:            newImage,
                CanaryWeight:        config.WeightStep,
                WeightStep:          config.WeightStep,
                CurrentStep:         1,
                TotalSteps:          config.TotalSteps,
                ErrorRateThreshold:  config.ErrorRateThreshold,
                LatencyThreshold:    config.LatencyThreshold,
                SuccessRateThreshold: config.SuccessRateThreshold,
        }

        global.DB.Create(release)

        // 执行第一步
        return m.executeStep(release, config)
}

// executeStep 执行灰度步骤
func (m *CanaryManager) executeStep(release *CanaryRelease, config CanaryConfig) (*CanaryRelease, error) {
        release.Status = DeployStatusRunning
        now := time.Now()
        release.StartedAt = &now
        global.DB.Save(release)

        // 更新流量权重
        err := m.executor.UpdateServiceWeight(release.ClusterID, release.Namespace, release.ServiceName, release.CanaryWeight)
        if err != nil {
                release.Status = DeployStatusFailed
                global.DB.Save(release)
                return release, fmt.Errorf("更新流量权重失败: %w", err)
        }

        // 创建步骤记录
        step := &CanaryStep{
                ReleaseID: release.ID,
                StepNum:   release.CurrentStep,
                Weight:    release.CanaryWeight,
                Status:    DeployStatusRunning,
        }
        step.StartedAt = &now
        global.DB.Create(step)

        // 等待指标稳定
        time.Sleep(time.Duration(config.StepDuration) * time.Second)

        // 获取监控指标
        metrics, err := m.executor.GetCanaryMetrics(release.ClusterID, release.Namespace, release.ServiceName)
        if err == nil {
                step.ErrorRate = metrics.ErrorRate
                step.Latency = metrics.Latency
                step.SuccessRate = metrics.SuccessRate
                step.RequestCount = metrics.RequestCount
        }

        // 检查是否通过阈值检查
        passed, reason := m.checkThresholds(release, metrics)
        if !passed {
                step.Status = DeployStatusFailed
                completedAt := time.Now()
                step.CompletedAt = &completedAt
                global.DB.Save(step)

                // 自动回滚
                if config.AutoRollback {
                        return m.Rollback(release.ID, reason)
                }

                release.Status = DeployStatusPaused
                global.DB.Save(release)
                return release, fmt.Errorf("阈值检查失败: %s", reason)
        }

        // AI 分析
        if m.llmClient != nil {
                decision := m.analyzeWithAI(release, metrics)
                release.AIDecision = decision.Analysis
                release.AIConfidence = decision.Confidence
                release.AIAutoPromote = decision.ShouldPromote
                global.DB.Save(release)
        }

        // 记录步骤完成
        step.Status = DeployStatusSuccess
        completedAt := time.Now()
        step.CompletedAt = &completedAt
        global.DB.Save(step)

        // 发送通知
        if m.notifier != nil {
                m.notifier.SendMessage("灰度发布进度",
                        fmt.Sprintf("%s/%s 第 %d 步完成，当前权重 %.1f%%",
                                release.Namespace, release.ServiceName, release.CurrentStep, release.CanaryWeight))
        }

        return release, nil
}

// Promote 推进发布
func (m *CanaryManager) Promote(releaseID uint) (*CanaryRelease, error) {
        var release CanaryRelease
        if err := global.DB.First(&release, releaseID).Error; err != nil {
                return nil, fmt.Errorf("发布记录不存在")
        }

        if release.Status != DeployStatusRunning && release.Status != DeployStatusPaused {
                return nil, fmt.Errorf("当前状态不允许推进")
        }

        // 获取配置
        var config CanaryConfig
        global.DB.Where("cluster_id = ? AND namespace = ? AND service_name = ?",
                release.ClusterID, release.Namespace, release.ServiceName).First(&config)

        // 更新权重
        release.CurrentStep++
        release.CanaryWeight = math.Min(release.CanaryWeight+release.WeightStep, 100)

        // 检查是否完成
        if release.CurrentStep >= release.TotalSteps || release.CanaryWeight >= 100 {
                return m.Complete(releaseID)
        }

        global.DB.Save(&release)

        return m.executeStep(&release, config)
}

// Complete 完成发布
func (m *CanaryManager) Complete(releaseID uint) (*CanaryRelease, error) {
        var release CanaryRelease
        if err := global.DB.First(&release, releaseID).Error; err != nil {
                return nil, fmt.Errorf("发布记录不存在")
        }

        // 提升为正式版本
        err := m.executor.PromoteCanary(release.ClusterID, release.Namespace, release.ServiceName)
        if err != nil {
                release.Status = DeployStatusFailed
                global.DB.Save(&release)
                return &release, fmt.Errorf("提升版本失败: %w", err)
        }

        // 更新状态
        completedAt := time.Now()
        release.Status = DeployStatusSuccess
        release.CanaryWeight = 100
        release.CompletedAt = &completedAt
        if release.StartedAt != nil {
                release.Duration = int64(completedAt.Sub(*release.StartedAt).Seconds())
        }
        global.DB.Save(&release)

        // 发送通知
        if m.notifier != nil {
                m.notifier.SendMessage("灰度发布完成",
                        fmt.Sprintf("%s/%s 已成功升级到 %s",
                                release.Namespace, release.ServiceName, release.NewVersion))
        }

        return &release, nil
}

// Rollback 回滚发布
func (m *CanaryManager) Rollback(releaseID uint, reason string) (*CanaryRelease, error) {
        var release CanaryRelease
        if err := global.DB.First(&release, releaseID).Error; err != nil {
                return nil, fmt.Errorf("发布记录不存在")
        }

        // 执行回滚
        err := m.executor.RollbackCanary(release.ClusterID, release.Namespace, release.ServiceName)
        if err != nil {
                release.Status = DeployStatusFailed
                global.DB.Save(&release)
                return &release, fmt.Errorf("回滚失败: %w", err)
        }

        // 更新状态
        rollbackAt := time.Now()
        release.Status = DeployStatusRollback
        release.RollbackReason = reason
        release.RollbackAt = &rollbackAt
        global.DB.Save(&release)

        // 发送通知
        if m.notifier != nil {
                m.notifier.SendMessage("灰度发布回滚",
                        fmt.Sprintf("%s/%s 已回滚，原因: %s",
                                release.Namespace, release.ServiceName, reason))
        }

        return &release, nil
}

// Pause 暂停发布
func (m *CanaryManager) Pause(releaseID uint) (*CanaryRelease, error) {
        var release CanaryRelease
        if err := global.DB.First(&release, releaseID).Error; err != nil {
                return nil, fmt.Errorf("发布记录不存在")
        }

        release.Status = DeployStatusPaused
        global.DB.Save(&release)

        if m.notifier != nil {
                m.notifier.SendMessage("灰度发布暂停",
                        fmt.Sprintf("%s/%s 发布已暂停", release.Namespace, release.ServiceName))
        }

        return &release, nil
}

// Abort 中止发布
func (m *CanaryManager) Abort(releaseID uint) (*CanaryRelease, error) {
        var release CanaryRelease
        if err := global.DB.First(&release, releaseID).Error; err != nil {
                return nil, fmt.Errorf("发布记录不存在")
        }

        // 执行回滚
        return m.Rollback(releaseID, "用户主动中止发布")
}

// CanaryDecision 灰度决策
type CanaryDecision struct {
        ShouldPromote bool    `json:"shouldPromote"`
        ShouldRollback bool   `json:"shouldRollback"`
        Analysis       string  `json:"analysis"`
        Confidence     float64 `json:"confidence"`
}

// analyzeWithAI AI 分析
func (m *CanaryManager) analyzeWithAI(release *CanaryRelease, metrics *CanaryMetrics) *CanaryDecision {
        decision := &CanaryDecision{}

        if m.llmClient == nil {
                return decision
        }

        prompt := fmt.Sprintf(`你是一个 Kubernetes 发布专家。请分析以下灰度发布状态：

## 发布信息
- 服务: %s/%s
- 当前版本: %s
- 新版本: %s
- 当前步骤: %d/%d
- 当前权重: %.1f%%

## 监控指标
- 错误率: %.2f%%
- 平均延迟: %dms
- 成功率: %.2f%%
- 请求量: %d

## 阈值配置
- 错误率阈值: %.2f%%
- 延迟阈值: %dms
- 成功率阈值: %.2f%%

请按以下 JSON 格式回复:
{
  "shouldPromote": true/false,
  "shouldRollback": true/false,
  "analysis": "分析说明",
  "confidence": 0.0-1.0
}`,
                release.Namespace, release.ServiceName,
                release.CurrentVersion, release.NewVersion,
                release.CurrentStep, release.TotalSteps,
                release.CanaryWeight,
                metrics.ErrorRate, metrics.Latency,
                metrics.SuccessRate, metrics.RequestCount,
                release.ErrorRateThreshold, release.LatencyThreshold,
                release.SuccessRateThreshold)

        response, err := m.llmClient.QuickChat(prompt)
        if err != nil {
                return decision
        }

        jsonStart := strings.Index(response, "{")
        jsonEnd := strings.LastIndex(response, "}")
        if jsonStart != -1 && jsonEnd != -1 {
                json.Unmarshal([]byte(response[jsonStart:jsonEnd+1]), decision)
        }

        return decision
}

// checkThresholds 检查阈值
func (m *CanaryManager) checkThresholds(release *CanaryRelease, metrics *CanaryMetrics) (bool, string) {
        if metrics == nil {
                return true, ""
        }

        if metrics.ErrorRate > release.ErrorRateThreshold {
                return false, fmt.Sprintf("错误率 %.2f%% 超过阈值 %.2f%%", metrics.ErrorRate, release.ErrorRateThreshold)
        }

        if metrics.Latency > release.LatencyThreshold {
                return false, fmt.Sprintf("延迟 %dms 超过阈值 %dms", metrics.Latency, release.LatencyThreshold)
        }

        if metrics.SuccessRate < release.SuccessRateThreshold {
                return false, fmt.Sprintf("成功率 %.2f%% 低于阈值 %.2f%%", metrics.SuccessRate, release.SuccessRateThreshold)
        }

        return true, ""
}

// generateVersion 生成版本号
func generateVersion() string {
        return fmt.Sprintf("v%d", time.Now().Unix())
}

// GetReleases 获取发布列表
func GetReleases(clusterID uint, namespace string) ([]CanaryRelease, error) {
        var releases []CanaryRelease
        query := global.DB.Model(&CanaryRelease{}).Order("created_at DESC")
        if clusterID > 0 {
                query = query.Where("cluster_id = ?", clusterID)
        }
        if namespace != "" {
                query = query.Where("namespace = ?", namespace)
        }
        err := query.Find(&releases).Error
        return releases, err
}

// GetRelease 获取发布详情
func GetRelease(id uint) (*CanaryRelease, error) {
        var release CanaryRelease
        err := global.DB.First(&release, id).Error
        return &release, err
}

// GetReleaseSteps 获取发布步骤
func GetReleaseSteps(releaseID uint) ([]CanaryStep, error) {
        var steps []CanaryStep
        err := global.DB.Where("release_id = ?", releaseID).Order("step_num ASC").Find(&steps).Error
        return steps, err
}
