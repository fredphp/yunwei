package kubernetes

import (
        "encoding/json"
        "fmt"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/model/kubernetes"
        "yunwei/service/ai/llm"
        "yunwei/service/notify"
)

// AutoScaler 自动扩容器
type AutoScaler struct {
        llmClient *llm.GLM5Client
        notifier  notify.Notifier
        executor  K8sExecutor
}

// K8sExecutor K8s 执行器接口
type K8sExecutor interface {
        GetDeploymentStatus(clusterID uint, namespace, deployment string) (*kubernetes.DeploymentStatus, error)
        ScaleDeployment(clusterID uint, namespace, deployment string, replicas int) error
        ApplyHPA(clusterID uint, namespace, deployment string, config kubernetes.HPAConfig) error
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
func (s *AutoScaler) AnalyzeAndScale(cluster *kubernetes.Cluster, namespace, deployment string, metrics map[string]float64) (*kubernetes.ScaleEvent, error) {
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
        event := &kubernetes.ScaleEvent{
                ClusterID:      cluster.ID,
                Namespace:      namespace,
                Deployment:     deployment,
                ScaleType:      kubernetes.ScaleTypeAuto,
                Status:         kubernetes.ScaleStatusPending,
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
func (s *AutoScaler) analyzeWithAI(cluster *kubernetes.Cluster, status *kubernetes.DeploymentStatus, metrics map[string]float64) (*ScaleDecision, error) {
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
func (s *AutoScaler) executeScale(event *kubernetes.ScaleEvent, targetReplicas int) (*kubernetes.ScaleEvent, error) {
        event.Status = kubernetes.ScaleStatusRunning
        now := time.Now()
        event.StartedAt = &now
        global.DB.Save(event)

        // 执行扩容命令
        var cluster kubernetes.Cluster
        if err := global.DB.First(&cluster, event.ClusterID).Error; err != nil {
                event.Status = kubernetes.ScaleStatusFailed
                event.ErrorMessage = "集群不存在"
                global.DB.Save(event)
                return event, err
        }

        err := s.executor.ScaleDeployment(event.ClusterID, event.Namespace, event.Deployment, targetReplicas)
        if err != nil {
                event.Status = kubernetes.ScaleStatusFailed
                event.ErrorMessage = err.Error()
                global.DB.Save(event)
                return event, err
        }

        // 更新状态
        completedAt := time.Now()
        event.Status = kubernetes.ScaleStatusSuccess
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
func (s *AutoScaler) ManualScale(clusterID uint, namespace, deployment string, targetReplicas int, reason string) (*kubernetes.ScaleEvent, error) {
        var cluster kubernetes.Cluster
        if err := global.DB.First(&cluster, clusterID).Error; err != nil {
                return nil, fmt.Errorf("集群不存在")
        }

        // 获取当前状态
        currentStatus, err := s.executor.GetDeploymentStatus(clusterID, namespace, deployment)
        if err != nil {
                return nil, fmt.Errorf("获取状态失败: %w", err)
        }

        event := &kubernetes.ScaleEvent{
                ClusterID:      clusterID,
                Namespace:      namespace,
                Deployment:     deployment,
                ScaleType:      kubernetes.ScaleTypeManual,
                Status:         kubernetes.ScaleStatusPending,
                ReplicasBefore: currentStatus.Replicas,
                ReplicasTarget: targetReplicas,
                TriggerReason:  reason,
        }

        global.DB.Create(event)

        return s.executeScale(event, targetReplicas)
}

// SetupHPA 配置 HPA
func (s *AutoScaler) SetupHPA(clusterID uint, namespace, deployment string, config kubernetes.HPAConfig) error {
        return s.executor.ApplyHPA(clusterID, namespace, deployment, config)
}

// GetScaleHistory 获取扩容历史
func (s *AutoScaler) GetScaleHistory(clusterID uint, namespace, deployment string, limit int) ([]kubernetes.ScaleEvent, error) {
        var events []kubernetes.ScaleEvent
        query := global.DB.Model(&kubernetes.ScaleEvent{}).Order("created_at DESC")
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
                var clusters []kubernetes.Cluster
                global.DB.Where("auto_scale_enabled = ? AND status = ?", true, "connected").Find(&clusters)

                for _, cluster := range clusters {
                        s.checkClusterForScaling(&cluster)
                }
        }
}

// checkClusterForScaling 检查集群是否需要扩容
func (s *AutoScaler) checkClusterForScaling(cluster *kubernetes.Cluster) {
        // 获取所有配置了 HPA 的 Deployment
        var hpaConfigs []kubernetes.HPAConfig
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
func AddCluster(cluster *kubernetes.Cluster) error {
        return global.DB.Create(cluster).Error
}

// GetClusters 获取集群列表
func GetClusters() ([]kubernetes.Cluster, error) {
        var clusters []kubernetes.Cluster
        err := global.DB.Find(&clusters).Error
        return clusters, err
}

// GetCluster 获取集群
func GetCluster(id uint) (*kubernetes.Cluster, error) {
        var cluster kubernetes.Cluster
        err := global.DB.First(&cluster, id).Error
        return &cluster, err
}

// UpdateCluster 更新集群
func UpdateCluster(cluster *kubernetes.Cluster) error {
        return global.DB.Save(cluster).Error
}

// DeleteCluster 删除集群
func DeleteCluster(id uint) error {
        return global.DB.Delete(&kubernetes.Cluster{}, id).Error
}
