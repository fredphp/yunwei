package kubernetes

import "time"

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
