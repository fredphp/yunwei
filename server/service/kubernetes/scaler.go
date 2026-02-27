package kubernetes

import (
        "encoding/json"
        "fmt"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/service/ai/llm"
        "yunwei/model/notify"
)

// ScaleStatus 扩容状态
type ScaleStatus string

const (
        ScaleStatusPending   ScaleStatus = "pending"
        ScaleStatusRunning   ScaleStatus = "running"
        ScaleStatusSuccess   ScaleStatus = "success"
        ScaleStatusFailed    ScaleStatus = "failed"
        ScaleStatusRollback  ScaleStatus = "rollback"
)

// ScaleType 扩容类型
type ScaleType string

const (
        ScaleTypeHorizontal ScaleType = "horizontal" // HPA 水平扩容
        ScaleTypeVertical   ScaleType = "vertical"   // VPA 垂直扩容
        ScaleTypeManual     ScaleType = "manual"     // 手动扩容
        ScaleTypeAuto       ScaleType = "auto"       // 自动扩容
)

// ScaleEvent 扩容事件
type ScaleEvent struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`

        ClusterID   uint           `json:"clusterId" gorm:"index"`
        Cluster     *Cluster       `json:"cluster" gorm:"foreignKey:ClusterID"`

        // 扩容信息
        Namespace   string     `json:"namespace" gorm:"type:varchar(64)"`
        Deployment  string     `json:"deployment" gorm:"type:varchar(128)"`
        ScaleType   ScaleType  `json:"scaleType" gorm:"type:varchar(16)"`
        Status      ScaleStatus `json:"status" gorm:"type:varchar(16)"`

        // 副本数
        ReplicasBefore int `json:"replicasBefore"`
        ReplicasAfter  int `json:"replicasAfter"`
        ReplicasTarget int `json:"replicasTarget"`

        // 触发原因
        TriggerReason string `json:"triggerReason" gorm:"type:text"`
        TriggerMetric string `json:"triggerMetric" gorm:"type:text"` // JSON 触发指标

        // AI 决策
        AIDecision    string `json:"aiDecision" gorm:"type:text"`
        AIConfidence  float64 `json:"aiConfidence"`
        AIAutoApprove bool   `json:"aiAutoApprove"`

        // 执行信息
        Commands       string `json:"commands" gorm:"type:text"`
        ExecutionLog   string `json:"executionLog" gorm:"type:text"`
        ErrorMessage   string `json:"errorMessage" gorm:"type:text"`

        // 时间
        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration    int64      `json:"duration"` // 毫秒
}

func (ScaleEvent) TableName() string {
        return "k8s_scale_events"
}

// Cluster Kubernetes 集群
type Cluster struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`

        Name        string `json:"name" gorm:"type:varchar(64)"`
        APIEndpoint string `json:"apiEndpoint" gorm:"type:varchar(256)"`
        Token       string `json:"token" gorm:"type:text"` // ServiceAccount Token
        KubeConfig  string `json:"kubeConfig" gorm:"type:text"` // Kubeconfig 内容

        // 状态
        Status      string `json:"status" gorm:"type:varchar(16)"` // connected, disconnected, error
        Version     string `json:"version" gorm:"type:varchar(32)"`
        NodeCount   int    `json:"nodeCount"`

        // 自动扩容配置
        AutoScaleEnabled bool    `json:"autoScaleEnabled"`
        MinReplicas      int     `json:"minReplicas"`
        MaxReplicas      int     `json:"maxReplicas"`
        CPUThreshold     float64 `json:"cpuThreshold"` // CPU 阈值
        MemThreshold     float64 `json:"memThreshold"` // 内存阈值

        // 最后同步
        LastSyncAt *time.Time `json:"lastSyncAt"`
}

func (Cluster) TableName() string {
        return "k8s_clusters"
}

// HPAConfig HPA 配置
type HPAConfig struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`
        UpdatedAt time.Time `json:"updatedAt"`

        ClusterID   uint `json:"clusterId" gorm:"index"`
        Namespace   string `json:"namespace" gorm:"type:varchar(64)"`
        Deployment  string `json:"deployment" gorm:"type:varchar(128)"`

        // HPA 配置
        MinReplicas    int     `json:"minReplicas"`
        MaxReplicas    int     `json:"maxReplicas"`
        TargetCPUUtil  float64 `json:"targetCpuUtil"` // 目标 CPU 使用率
        TargetMemUtil  float64 `json:"targetMemUtil"` // 目标内存使用率

        // 自定义指标
        CustomMetrics string `json:"customMetrics" gorm:"type:text"` // JSON

        // 行为配置
        ScaleUpStabilization   int `json:"scaleUpStabilization"`   // 扩容稳定窗口(秒)
        ScaleDownStabilization int `json:"scaleDownStabilization"` // 缩容稳定窗口(秒)

        Enabled bool `json:"enabled"`
}

func (HPAConfig) TableName() string {
        return "k8s_hpa_configs"
}

// DeploymentStatus Deployment 状态
type DeploymentStatus struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`

        ClusterID   uint   `json:"clusterId" gorm:"index"`
        Namespace   string `json:"namespace" gorm:"type:varchar(64)"`
        Deployment  string `json:"deployment" gorm:"type:varchar(128)"`

        // 状态
        Replicas      int `json:"replicas"`
        ReadyReplicas int `json:"readyReplicas"`
        UpdatedReplicas int `json:"updatedReplicas"`

        // 资源使用
        CPUUsage    float64 `json:"cpuUsage"`
        MemoryUsage float64 `json:"memoryUsage"`

        // 请求
        CPURequest    string `json:"cpuRequest"`
        MemoryRequest string `json:"memoryRequest"`
        CPULimit      string `json:"cpuLimit"`
        MemoryLimit   string `json:"memoryLimit"`

        // HPA 状态
        HPAEnabled     bool    `json:"hpaEnabled"`
        HPATargetCPU   float64 `json:"hpaTargetCpu"`
        CurrentReplicas int    `json:"currentReplicas"`
        DesiredReplicas int    `json:"desiredReplicas"`
}

func (DeploymentStatus) TableName() string {
        return "k8s_deployment_status"
}

// AutoScaler 自动扩容器
type AutoScaler struct {
        llmClient *llm.GLM5Client
        notifier  notify.Notifier
        executor  K8sExecutor
}

// K8sExecutor K8s 执行器接口
type K8sExecutor interface {
        GetDeploymentStatus(clusterID uint, namespace, deployment string) (*DeploymentStatus, error)
        ScaleDeployment(clusterID uint, namespace, deployment string, replicas int) error
        ApplyHPA(clusterID uint, namespace, deployment string, config HPAConfig) error
        GetHPAStatus(clusterID uint, namespace, deployment string) (map[string]interface{}, error)
}

// NewAutoScaler 创建自动扩容器
func NewAutoScaler() *AutoScaler {
        return &AutoScaler{}
}

// SetLLMClient 设置 LLM 客户端
func (s *AutoScaler) SetLLMClient(client *llm.GLM5Client) {
        s.llmClient = client
}

// SetNotifier 设置通知器
func (s *AutoScaler) SetNotifier(notifier notify.Notifier) {
        s.notifier = notifier
}

// SetExecutor 设置执行器
func (s *AutoScaler) SetExecutor(executor K8sExecutor) {
        s.executor = executor
}

// AnalyzeAndScale 分析并执行扩容
func (s *AutoScaler) AnalyzeAndScale(cluster *Cluster, namespace, deployment string, metrics map[string]float64) (*ScaleEvent, error) {
        // 获取当前状态
        currentStatus, err := s.executor.GetDeploymentStatus(cluster.ID, namespace, deployment)
        if err != nil {
                return nil, fmt.Errorf("获取 Deployment 状态失败: %w", err)
        }

        // AI 分析
        decision, err := s.analyzeWithAI(cluster, currentStatus, metrics)
        if err != nil {
                return nil, fmt.Errorf("AI 分析失败: %w", err)
        }

        // 创建扩容事件
        event := &ScaleEvent{
                ClusterID:      cluster.ID,
                Namespace:      namespace,
                Deployment:     deployment,
                ScaleType:      ScaleTypeAuto,
                Status:         ScaleStatusPending,
                ReplicasBefore: currentStatus.Replicas,
                ReplicasTarget: decision.TargetReplicas,
                TriggerReason:  decision.Reason,
                TriggerMetric:  s.formatMetrics(metrics),
                AIDecision:     decision.Analysis,
                AIConfidence:   decision.Confidence,
                AIAutoApprove:  decision.AutoApprove,
        }

        metricsJSON, _ := json.Marshal(metrics)
        event.TriggerMetric = string(metricsJSON)

        global.DB.Create(event)

        // 自动批准或等待人工确认
        if decision.AutoApprove && cluster.AutoScaleEnabled {
                return s.executeScale(event, decision.TargetReplicas)
        }

        return event, nil
}

// ScaleDecision 扩容决策
type ScaleDecision struct {
        TargetReplicas int     `json:"targetReplicas"`
        Reason         string  `json:"reason"`
        Analysis       string  `json:"analysis"`
        Confidence     float64 `json:"confidence"`
        AutoApprove    bool    `json:"autoApprove"`
}

// analyzeWithAI AI 分析
func (s *AutoScaler) analyzeWithAI(cluster *Cluster, status *DeploymentStatus, metrics map[string]float64) (*ScaleDecision, error) {
        prompt := fmt.Sprintf(`你是一个 Kubernetes 运维专家。请分析以下 Deployment 状态并给出扩容建议。

## 当前状态
- Namespace: %s
- Deployment: %s
- 当前副本数: %d
- 就绪副本数: %d
- CPU 使用率: %.2f%%
- 内存使用率: %.2f%%

## 集群配置
- 最小副本数: %d
- 最大副本数: %d
- CPU 扩容阈值: %.2f%%
- 内存扩容阈值: %.2f%%

## 当前指标
%s

请按以下 JSON 格式回复:
{
  "targetReplicas": 目标副本数,
  "reason": "扩容/缩容原因",
  "analysis": "详细分析",
  "confidence": 0.0-1.0,
  "autoApprove": true/false
}`,
                status.Namespace, status.Deployment, status.Replicas, status.ReadyReplicas,
                status.CPUUsage, status.MemoryUsage,
                cluster.MinReplicas, cluster.MaxReplicas,
                cluster.CPUThreshold, cluster.MemThreshold,
                s.formatMetrics(metrics))

        response, err := s.llmClient.QuickChat(prompt)
        if err != nil {
                return nil, err
        }

        decision := &ScaleDecision{}
        jsonStart := strings.Index(response, "{")
        jsonEnd := strings.LastIndex(response, "}")
        if jsonStart != -1 && jsonEnd != -1 {
                json.Unmarshal([]byte(response[jsonStart:jsonEnd+1]), decision)
        }

        // 验证边界
        if decision.TargetReplicas < cluster.MinReplicas {
                decision.TargetReplicas = cluster.MinReplicas
        }
        if decision.TargetReplicas > cluster.MaxReplicas {
                decision.TargetReplicas = cluster.MaxReplicas
        }

        return decision, nil
}

// executeScale 执行扩容
func (s *AutoScaler) executeScale(event *ScaleEvent, targetReplicas int) (*ScaleEvent, error) {
        event.Status = ScaleStatusRunning
        now := time.Now()
        event.StartedAt = &now
        global.DB.Save(event)

        // 执行扩容命令
        var cluster Cluster
        if err := global.DB.First(&cluster, event.ClusterID).Error; err != nil {
                event.Status = ScaleStatusFailed
                event.ErrorMessage = "集群不存在"
                global.DB.Save(event)
                return event, err
        }

        err := s.executor.ScaleDeployment(event.ClusterID, event.Namespace, event.Deployment, targetReplicas)
        if err != nil {
                event.Status = ScaleStatusFailed
                event.ErrorMessage = err.Error()
                global.DB.Save(event)
                return event, err
        }

        // 更新状态
        completedAt := time.Now()
        event.Status = ScaleStatusSuccess
        event.ReplicasAfter = targetReplicas
        event.CompletedAt = &completedAt
        event.Duration = completedAt.Sub(*event.StartedAt).Milliseconds()
        global.DB.Save(event)

        // 发送通知
        if s.notifier != nil {
                s.notifier.SendMessage("K8s 自动扩容",
                        fmt.Sprintf("%s/%s 已从 %d 扩容到 %d 副本",
                                event.Namespace, event.Deployment, event.ReplicasBefore, targetReplicas))
        }

        return event, nil
}

// ManualScale 手动扩容
func (s *AutoScaler) ManualScale(clusterID uint, namespace, deployment string, targetReplicas int, reason string) (*ScaleEvent, error) {
        var cluster Cluster
        if err := global.DB.First(&cluster, clusterID).Error; err != nil {
                return nil, fmt.Errorf("集群不存在")
        }

        // 获取当前状态
        currentStatus, err := s.executor.GetDeploymentStatus(clusterID, namespace, deployment)
        if err != nil {
                return nil, fmt.Errorf("获取状态失败: %w", err)
        }

        event := &ScaleEvent{
                ClusterID:      clusterID,
                Namespace:      namespace,
                Deployment:     deployment,
                ScaleType:      ScaleTypeManual,
                Status:         ScaleStatusPending,
                ReplicasBefore: currentStatus.Replicas,
                ReplicasTarget: targetReplicas,
                TriggerReason:  reason,
        }

        global.DB.Create(event)

        return s.executeScale(event, targetReplicas)
}

// SetupHPA 配置 HPA
func (s *AutoScaler) SetupHPA(clusterID uint, namespace, deployment string, config HPAConfig) error {
        return s.executor.ApplyHPA(clusterID, namespace, deployment, config)
}

// GetScaleHistory 获取扩容历史
func (s *AutoScaler) GetScaleHistory(clusterID uint, namespace, deployment string, limit int) ([]ScaleEvent, error) {
        var events []ScaleEvent
        query := global.DB.Model(&ScaleEvent{}).Order("created_at DESC")
        if clusterID > 0 {
                query = query.Where("cluster_id = ?", clusterID)
        }
        if namespace != "" {
                query = query.Where("namespace = ?", namespace)
        }
        if deployment != "" {
                query = query.Where("deployment = ?", deployment)
        }
        if limit > 0 {
                query = query.Limit(limit)
        }
        err := query.Find(&events).Error
        return events, err
}

// MonitorClusters 监控集群
func (s *AutoScaler) MonitorClusters() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
                var clusters []Cluster
                global.DB.Where("auto_scale_enabled = ? AND status = ?", true, "connected").Find(&clusters)

                for _, cluster := range clusters {
                        s.checkClusterForScaling(&cluster)
                }
        }
}

// checkClusterForScaling 检查集群是否需要扩容
func (s *AutoScaler) checkClusterForScaling(cluster *Cluster) {
        // 获取所有配置了 HPA 的 Deployment
        var hpaConfigs []HPAConfig
        global.DB.Where("cluster_id = ? AND enabled = ?", cluster.ID, true).Find(&hpaConfigs)

        for _, config := range hpaConfigs {
                status, err := s.executor.GetDeploymentStatus(cluster.ID, config.Namespace, config.Deployment)
                if err != nil {
                        continue
                }

                // 检查是否需要扩容
                metrics := map[string]float64{
                        "cpu_usage":    status.CPUUsage,
                        "memory_usage": status.MemoryUsage,
                }

                if status.CPUUsage > cluster.CPUThreshold || status.MemoryUsage > cluster.MemThreshold {
                        s.AnalyzeAndScale(cluster, config.Namespace, config.Deployment, metrics)
                }
        }
}

// formatMetrics 格式化指标
func (s *AutoScaler) formatMetrics(metrics map[string]float64) string {
        var parts []string
        for k, v := range metrics {
                parts = append(parts, fmt.Sprintf("- %s: %.2f", k, v))
        }
        return strings.Join(parts, "\n")
}

// AddCluster 添加集群
func AddCluster(cluster *Cluster) error {
        return global.DB.Create(cluster).Error
}

// GetClusters 获取集群列表
func GetClusters() ([]Cluster, error) {
        var clusters []Cluster
        err := global.DB.Find(&clusters).Error
        return clusters, err
}

// GetCluster 获取集群
func GetCluster(id uint) (*Cluster, error) {
        var cluster Cluster
        err := global.DB.First(&cluster, id).Error
        return &cluster, err
}

// UpdateCluster 更新集群
func UpdateCluster(cluster *Cluster) error {
        return global.DB.Save(cluster).Error
}

// DeleteCluster 删除集群
func DeleteCluster(id uint) error {
        return global.DB.Delete(&Cluster{}, id).Error
}
