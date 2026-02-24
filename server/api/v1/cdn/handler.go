package cdn

import (
	"net/http"

	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/service/cdn"

	"github.com/gin-gonic/gin"
)

// GetDomains 获取 CDN 域名列表
func GetDomains(c *gin.Context) {
	domains, err := cdn.GetDomains()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(domains, c)
}

// GetDomain 获取域名详情
func GetDomain(c *gin.Context) {
	id := c.Param("id")
	domain, err := cdn.GetDomain(parseInt(id))
	if err != nil {
		response.FailWithMessage("域名不存在", c)
		return
	}
	response.OkWithData(domain, c)
}

// AddDomain 添加域名
func AddDomain(c *gin.Context) {
	var domain cdn.CDNDomain
	if err := c.ShouldBindJSON(&domain); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := cdn.AddDomain(&domain); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(domain, c)
}

// UpdateDomain 更新域名
func UpdateDomain(c *gin.Context) {
	id := c.Param("id")
	var domain cdn.CDNDomain
	if err := global.DB.First(&domain, id).Error; err != nil {
		response.FailWithMessage("域名不存在", c)
		return
	}
	if err := c.ShouldBindJSON(&domain); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := cdn.UpdateDomain(&domain); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(domain, c)
}

// DeleteDomain 删除域名
func DeleteDomain(c *gin.Context) {
	id := c.Param("id")
	if err := cdn.DeleteDomain(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// OptimizeCDN 优化 CDN
func OptimizeCDN(c *gin.Context) {
	id := c.Param("id")
	domain, err := cdn.GetDomain(parseInt(id))
	if err != nil {
		response.FailWithMessage("域名不存在", c)
		return
	}
	manager := cdn.NewCDNManager()
	record, err := manager.AnalyzeAndOptimize(domain)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(record, c)
}

// PurgeCache 刷新缓存
func PurgeCache(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		URLs []string `json:"urls" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	manager := cdn.NewCDNManager()
	if err := manager.PurgeCache(parseInt(id), req.URLs); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// PreheatCache 预热缓存
func PreheatCache(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		URLs []string `json:"urls" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	manager := cdn.NewCDNManager()
	if err := manager.PreheatCache(parseInt(id), req.URLs); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetNodeStatus 获取节点状态
func GetNodeStatus(c *gin.Context) {
	id := c.Param("id")
	manager := cdn.NewCDNManager()
	nodes, err := manager.GetNodeStatus(parseInt(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(nodes, c)
}

// GetCacheRules 获取缓存规则
func GetCacheRules(c *gin.Context) {
	id := c.Param("id")
	rules, err := cdn.GetCacheRules(parseInt(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rules, c)
}

// AddCacheRule 添加缓存规则
func AddCacheRule(c *gin.Context) {
	var rule cdn.CDNCacheRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.DB.Create(&rule).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rule, c)
}

// UpdateCacheRule 更新缓存规则
func UpdateCacheRule(c *gin.Context) {
	var rule cdn.CDNCacheRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.DB.Save(&rule).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rule, c)
}

// DeleteCacheRule 删除缓存规则
func DeleteCacheRule(c *gin.Context) {
	id := c.Param("id")
	if err := global.DB.Delete(&cdn.CDNCacheRule{}, id).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetOptimizationHistory 获取优化历史
func GetOptimizationHistory(c *gin.Context) {
	domainID := c.Query("domainId")
	records, err := cdn.GetOptimizationHistory(parseInt(domainID), 50)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(records, c)
}

// CalculateCost 计算成本
func CalculateCost(c *gin.Context) {
	id := c.Param("id")
	domain, err := cdn.GetDomain(parseInt(id))
	if err != nil {
		response.FailWithMessage("域名不存在", c)
		return
	}
	manager := cdn.NewCDNManager()
	cost := manager.CalculateCost(domain)
	response.OkWithData(gin.H{"cost": cost}, c)
}

// OptimizeCost 成本优化
func OptimizeCost(c *gin.Context) {
	id := c.Param("id")
	manager := cdn.NewCDNManager()
	record, err := manager.OptimizeCost(parseInt(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(record, c)
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
