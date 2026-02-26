package cert

import (
        "yunwei/global"
        "yunwei/model/common/response"
        "yunwei/service/cert"

        "github.com/gin-gonic/gin"
)

// GetCertificates 获取证书列表
func GetCertificates(c *gin.Context) {
        certs, err := cert.GetCertificates()
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(certs, c)
}

// GetCertificate 获取证书详情
func GetCertificate(c *gin.Context) {
        id := c.Param("id")
        certObj, err := cert.GetCertificate(parseInt(id))
        if err != nil {
                response.FailWithMessage("证书不存在", c)
                return
        }
        response.OkWithData(certObj, c)
}

// AddCertificate 添加证书
func AddCertificate(c *gin.Context) {
        var certObj cert.Certificate
        if err := c.ShouldBindJSON(&certObj); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        if err := cert.AddCertificate(&certObj); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(certObj, c)
}

// UpdateCertificate 更新证书
func UpdateCertificate(c *gin.Context) {
        id := c.Param("id")
        var certObj cert.Certificate
        if err := global.DB.First(&certObj, id).Error; err != nil {
                response.FailWithMessage("证书不存在", c)
                return
        }
        if err := c.ShouldBindJSON(&certObj); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        if err := cert.UpdateCertificate(&certObj); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(certObj, c)
}

// DeleteCertificate 删除证书
func DeleteCertificate(c *gin.Context) {
        id := c.Param("id")
        if err := cert.DeleteCertificate(parseInt(id)); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.Ok(nil, c)
}

// RenewCertificate 续期证书
func RenewCertificate(c *gin.Context) {
        id := c.Param("id")
        manager := cert.NewCertRenewalManager()
        record, err := manager.RenewCertificate(parseInt(id))
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(record, c)
}

// CheckCertificate 检查证书状态
func CheckCertificate(c *gin.Context) {
        id := c.Param("id")
        certObj, err := cert.GetCertificate(parseInt(id))
        if err != nil {
                response.FailWithMessage("证书不存在", c)
                return
        }
        manager := cert.NewCertRenewalManager()
        if err := manager.CheckCertificate(certObj); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(certObj, c)
}

// CheckAllCertificates 检查所有证书
func CheckAllCertificates(c *gin.Context) {
        manager := cert.NewCertRenewalManager()
        if err := manager.CheckAllCertificates(); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.Ok(nil, c)
}

// GetRenewalHistory 获取续期历史
func GetRenewalHistory(c *gin.Context) {
        certID := c.Query("certId")
        records, err := cert.GetRenewalHistory(parseInt(certID), 50)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(records, c)
}

// RequestNewCert 申请新证书
func RequestNewCert(c *gin.Context) {
        var req struct {
                Domain      string `json:"domain" binding:"required"`
                Email       string `json:"email" binding:"required"`
                DNSProvider string `json:"dnsProvider"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }

        manager := cert.NewCertRenewalManager()
        certObj, err := manager.RequestNewCert(req.Domain, req.Email, req.DNSProvider)
        if err != nil {
                response.FailWithMessage(err.Error(), c)
                return
        }
        response.OkWithData(certObj, c)
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
