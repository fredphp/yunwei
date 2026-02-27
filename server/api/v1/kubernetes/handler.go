package kubernetes

import (
        "yunwei/global"
        "yunwei/model/common/response"
        "yunwei/service/kubernetes"

        "github.com/gin-gonic/gin"
)

// GetClusters 获取 K8s 集群列表
func GetClusters(c *gin.Context) {
        clusters, err := kubernetes.GetClusters()
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(clusters, c)
}

// GetCluster 获取集群详情
func GetCluster(c *gin.Context) {
        id := c.Param("id")
        var cluster kubernetes.Cluster
        if err := global.DB.First(&cluster, id).Error; err != nil {
                response.FailWithMessage("集群不存在", c)
                return
        }
        response.OkWithData(cluster, c)
}

// AddCluster 添加集群
func AddCluster(c *gin.Context) {
        var cluster kubernetes.Cluster
        if err := c.ShouldBindJSON(&cluster); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        if err := kubernetes.AddCluster(&cluster); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(cluster, c)
}

// UpdateCluster 更新集群
func UpdateCluster(c *gin.Context) {
        id := c.Param("id")
        var cluster kubernetes.Cluster
        if err := global.DB.First(&cluster, id).Error; err != nil {
                response.FailWithMessage("集群不存在", c)
                return
        }
        if err := c.ShouldBindJSON(&cluster); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        if err := kubernetes.UpdateCluster(&cluster); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(cluster, c)
}

// DeleteCluster 删除集群
func DeleteCluster(c *gin.Context) {
        id := c.Param("id")
        if err := kubernetes.DeleteCluster(parseInt(id)); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.Ok(nil, c)
}

// GetScaleHistory 获取扩容历史
func GetScaleHistory(c *gin.Context) {
        clusterID := c.Query("clusterId")
        namespace := c.Query("namespace")
        deployment := c.Query("deployment")

        scaler := kubernetes.NewAutoScaler()
        events, err := scaler.GetScaleHistory(parseInt(clusterID), namespace, deployment, 50)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(events, c)
}

// ManualScale 手动扩容
func ManualScale(c *gin.Context) {
        var req struct {
                ClusterID   uint   `json:"clusterId" binding:"required"`
                Namespace   string `json:"namespace" binding:"required"`
                Deployment  string `json:"deployment" binding:"required"`
                Replicas    int    `json:"replicas" binding:"required"`
                Reason      string `json:"reason"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }

        scaler := kubernetes.NewAutoScaler()
        event, err := scaler.ManualScale(req.ClusterID, req.Namespace, req.Deployment, req.Replicas, req.Reason)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(event, c)
}

// AnalyzeScale AI 分析扩容
func AnalyzeScale(c *gin.Context) {
        clusterID := c.Param("clusterId")
        namespace := c.Query("namespace")
        deployment := c.Query("deployment")

        var cluster kubernetes.Cluster
        if err := global.DB.First(&cluster, clusterID).Error; err != nil {
                response.FailWithMessage("集群不存在", c)
                return
        }

        scaler := kubernetes.NewAutoScaler()
        // 模拟指标
        metrics := map[string]float64{
                "cpu_usage":    75.5,
                "memory_usage": 82.3,
                "request_rate": 1500,
        }

        event, err := scaler.AnalyzeAndScale(&cluster, namespace, deployment, metrics)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(event, c)
}

// GetHPAConfigs 获取 HPA 配置
func GetHPAConfigs(c *gin.Context) {
        clusterID := c.Query("clusterId")
        var configs []kubernetes.HPAConfig
        query := global.DB.Model(&kubernetes.HPAConfig{})
        if clusterID != "" {
                query = query.Where("cluster_id = ?", clusterID)
        }
        query.Find(&configs)
        response.OkWithData(configs, c)
}

// UpdateHPAConfig 更新 HPA 配置
func UpdateHPAConfig(c *gin.Context) {
        var config kubernetes.HPAConfig
        if err := c.ShouldBindJSON(&config); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        if err := global.DB.Save(&config).Error; err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(config, c)
}

// GetDeploymentStatus 获取 Deployment 状态
func GetDeploymentStatus(c *gin.Context) {
        clusterID := c.Param("clusterId")
        namespace := c.Query("namespace")
        deployment := c.Query("deployment")

        var status kubernetes.DeploymentStatus
        query := global.DB.Model(&kubernetes.DeploymentStatus{}).Where("cluster_id = ?", clusterID)
        if namespace != "" {
                query = query.Where("namespace = ?", namespace)
        }
        if deployment != "" {
                query = query.Where("deployment = ?", deployment)
        }
        query.First(&status)

        response.OkWithData(status, c)
}

func parseInt(s string) uint {
        var result uint
        for _, c := range s {
                if c >= '0' && c <= '9' {
                        result = result*10 + uint(c-'0')
                }
        }
        return result
}
