package canary

import (
        "yunwei/global"
        "yunwei/model/common/response"
        "yunwei/service/canary"

        "github.com/gin-gonic/gin"
)

// GetReleases 获取灰度发布列表
func GetReleases(c *gin.Context) {
        clusterID := c.Query("clusterId")
        namespace := c.Query("namespace")

        releases, err := canary.GetReleases(parseInt(clusterID), namespace)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(releases, c)
}

// GetRelease 获取发布详情
func GetRelease(c *gin.Context) {
        id := c.Param("id")
        release, err := canary.GetRelease(parseInt(id))
        if err != nil {
                response.FailWithMessage("发布记录不存在", c)
                return
        }
        response.OkWithData(release, c)
}

// GetReleaseSteps 获取发布步骤
func GetReleaseSteps(c *gin.Context) {
        id := c.Param("id")
        steps, err := canary.GetReleaseSteps(parseInt(id))
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(steps, c)
}

// StartCanary 开始灰度发布
func StartCanary(c *gin.Context) {
        var req struct {
                ClusterID   uint                `json:"clusterId" binding:"required"`
                Namespace   string              `json:"namespace" binding:"required"`
                ServiceName string              `json:"serviceName" binding:"required"`
                NewImage    string              `json:"newImage" binding:"required"`
                Config      canary.CanaryConfig `json:"config"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }

        manager := canary.NewCanaryManager()
        release, err := manager.StartCanary(req.ClusterID, req.Namespace, req.ServiceName, req.NewImage, req.Config)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(release, c)
}

// PromoteCanary 推进发布
func PromoteCanary(c *gin.Context) {
        id := c.Param("id")
        manager := canary.NewCanaryManager()
        release, err := manager.Promote(parseInt(id))
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(release, c)
}

// CompleteCanary 完成发布
func CompleteCanary(c *gin.Context) {
        id := c.Param("id")
        manager := canary.NewCanaryManager()
        release, err := manager.Complete(parseInt(id))
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(release, c)
}

// RollbackCanary 回滚发布
func RollbackCanary(c *gin.Context) {
        id := c.Param("id")
        var req struct {
                Reason string `json:"reason"`
        }
        c.ShouldBindJSON(&req)

        manager := canary.NewCanaryManager()
        release, err := manager.Rollback(parseInt(id), req.Reason)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(release, c)
}

// PauseCanary 暂停发布
func PauseCanary(c *gin.Context) {
        id := c.Param("id")
        manager := canary.NewCanaryManager()
        release, err := manager.Pause(parseInt(id))
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(release, c)
}

// AbortCanary 中止发布
func AbortCanary(c *gin.Context) {
        id := c.Param("id")
        manager := canary.NewCanaryManager()
        release, err := manager.Abort(parseInt(id))
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(release, c)
}

// GetConfigs 获取灰度配置
func GetConfigs(c *gin.Context) {
        clusterID := c.Query("clusterId")
        var configs []canary.CanaryConfig
        query := global.DB.Model(&canary.CanaryConfig{})
        if clusterID != "" {
                query = query.Where("cluster_id = ?", clusterID)
        }
        query.Find(&configs)
        response.OkWithData(configs, c)
}

// UpdateConfig 更新灰度配置
func UpdateConfig(c *gin.Context) {
        var config canary.CanaryConfig
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

func parseInt(s string) uint {
        var result uint
        for _, c := range s {
                if c >= '0' && c <= '9' {
                        result = result*10 + uint(c-'0')
                }
        }
        return result
}
