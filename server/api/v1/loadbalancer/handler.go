package loadbalancer

import (
	"net/http"

	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/service/loadbalancer"

	"github.com/gin-gonic/gin"
)

// GetLBs 获取负载均衡器列表
func GetLBs(c *gin.Context) {
	lbs, err := loadbalancer.GetLBs()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(lbs, c)
}

// GetLB 获取负载均衡器详情
func GetLB(c *gin.Context) {
	id := c.Param("id")
	lb, err := loadbalancer.GetLB(parseInt(id))
	if err != nil {
		response.FailWithMessage("负载均衡器不存在", c)
		return
	}
	response.OkWithData(lb, c)
}

// AddLB 添加负载均衡器
func AddLB(c *gin.Context) {
	var lb loadbalancer.LoadBalancer
	if err := c.ShouldBindJSON(&lb); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.DB.Create(&lb).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(lb, c)
}

// UpdateLB 更新负载均衡器
func UpdateLB(c *gin.Context) {
	id := c.Param("id")
	var lb loadbalancer.LoadBalancer
	if err := global.DB.First(&lb, id).Error; err != nil {
		response.FailWithMessage("负载均衡器不存在", c)
		return
	}
	if err := c.ShouldBindJSON(&lb); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.DB.Save(&lb).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(lb, c)
}

// DeleteLB 删除负载均衡器
func DeleteLB(c *gin.Context) {
	id := c.Param("id")
	if err := global.DB.Delete(&loadbalancer.LoadBalancer{}, id).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetBackends 获取后端服务器
func GetBackends(c *gin.Context) {
	lbID := c.Param("id")
	backends, err := loadbalancer.GetBackends(parseInt(lbID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(backends, c)
}

// AddBackend 添加后端服务器
func AddBackend(c *gin.Context) {
	var backend loadbalancer.BackendServer
	if err := c.ShouldBindJSON(&backend); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.DB.Create(&backend).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(backend, c)
}

// UpdateBackend 更新后端服务器
func UpdateBackend(c *gin.Context) {
	id := c.Param("id")
	var backend loadbalancer.BackendServer
	if err := global.DB.First(&backend, id).Error; err != nil {
		response.FailWithMessage("后端服务器不存在", c)
		return
	}
	if err := c.ShouldBindJSON(&backend); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.DB.Save(&backend).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(backend, c)
}

// DeleteBackend 删除后端服务器
func DeleteBackend(c *gin.Context) {
	id := c.Param("id")
	if err := global.DB.Delete(&loadbalancer.BackendServer{}, id).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// OptimizeLB 优化负载均衡器
func OptimizeLB(c *gin.Context) {
	id := c.Param("id")
	lb, err := loadbalancer.GetLB(parseInt(id))
	if err != nil {
		response.FailWithMessage("负载均衡器不存在", c)
		return
	}

	optimizer := loadbalancer.NewLBOptimizer()
	record, err := optimizer.AnalyzeAndOptimize(lb)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(record, c)
}

// AutoBalance 自动负载均衡
func AutoBalance(c *gin.Context) {
	id := c.Param("id")
	optimizer := loadbalancer.NewLBOptimizer()
	if err := optimizer.AutoBalance(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	id := c.Param("id")
	optimizer := loadbalancer.NewLBOptimizer()
	if err := optimizer.HealthCheck(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetOptimizationHistory 获取优化历史
func GetOptimizationHistory(c *gin.Context) {
	lbID := c.Query("lbId")
	records, err := loadbalancer.GetOptimizationHistory(parseInt(lbID), 50)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(records, c)
}

// GetAlgorithmConfigs 获取算法配置
func GetAlgorithmConfigs(c *gin.Context) {
	lbID := c.Query("lbId")
	var configs []loadbalancer.AlgorithmConfig
	query := global.DB.Model(&loadbalancer.AlgorithmConfig{})
	if lbID != "" {
		query = query.Where("lb_id = ?", lbID)
	}
	query.Find(&configs)
	response.OkWithData(configs, c)
}

// UpdateAlgorithmConfig 更新算法配置
func UpdateAlgorithmConfig(c *gin.Context) {
	var config loadbalancer.AlgorithmConfig
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
