package tenant

import (
        "net/http"
        "strconv"
        "time"

        tenantModel "yunwei/model/tenant"
        "github.com/gin-gonic/gin"
)

// ==================== 账单管理接口 ====================

// ListBillings 列出账单
func (h *Handler) ListBillings(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
        status := c.Query("status")
        period := c.Query("period")
        tenantID := c.Query("tenant_id")

        var billings []tenantModel.TenantBilling
        var total int64

        query := h.db.Model(&tenantModel.TenantBilling{}).Preload("Tenant")
        if status != "" {
                query = query.Where("status = ?", status)
        }
        if period != "" {
                query = query.Where("billing_period = ?", period)
        }
        if tenantID != "" {
                query = query.Where("tenant_id = ?", tenantID)
        }

        query.Count(&total)
        err := query.Order("created_at DESC").
                Offset((page - 1) * pageSize).
                Limit(pageSize).
                Find(&billings).Error

        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "data":      billings,
                "total":     total,
                "page":      page,
                "page_size": pageSize,
        })
}

// GetBilling 获取账单详情
func (h *Handler) GetBilling(c *gin.Context) {
        id := c.Param("id")
        var billing tenantModel.TenantBilling
        err := h.db.Preload("Tenant").First(&billing, "id = ?", id).Error
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "账单不存在"})
                return
        }
        c.JSON(http.StatusOK, gin.H{"data": billing})
}

// CreateBilling 创建账单
func (h *Handler) CreateBilling(c *gin.Context) {
        var req struct {
                TenantID      string  `json:"tenant_id" binding:"required"`
                BillingPeriod string  `json:"billing_period" binding:"required"`
                BaseAmount    float64 `json:"base_amount"`
                UsageAmount   float64 `json:"usage_amount"`
                OverageAmount float64 `json:"overage_amount"`
                DiscountAmount float64 `json:"discount_amount"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        // 计算总额
        totalAmount := req.BaseAmount + req.UsageAmount + req.OverageAmount - req.DiscountAmount

        billing := &tenantModel.TenantBilling{
                ID:             generateUUID(),
                TenantID:       req.TenantID,
                BillingPeriod:  req.BillingPeriod,
                DueDate:        time.Now().AddDate(0, 0, 15), // 15天后到期
                BaseAmount:     req.BaseAmount,
                UsageAmount:    req.UsageAmount,
                OverageAmount:  req.OverageAmount,
                DiscountAmount: req.DiscountAmount,
                TotalAmount:    totalAmount,
                Status:         "pending",
        }

        if err := h.db.Create(billing).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, gin.H{"data": billing})
}

// MarkBillingPaid 标记账单已支付
func (h *Handler) MarkBillingPaid(c *gin.Context) {
        id := c.Param("id")
        var req struct {
                PaymentMethod string `json:"payment_method"`
                PaymentID     string `json:"payment_id"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        now := time.Now()
        updates := map[string]interface{}{
                "status":         "paid",
                "paid_at":        &now,
                "payment_method": req.PaymentMethod,
                "payment_id":     req.PaymentID,
                "invoice_number": "INV-" + time.Now().Format("20060102") + "-" + id[:8],
        }

        if err := h.db.Model(&tenantModel.TenantBilling{}).Where("id = ?", id).Updates(updates).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{"message": "账单已标记为已支付"})
}

// GenerateBillings 批量生成账单
func (h *Handler) GenerateBillings(c *gin.Context) {
        var req struct {
                Period string `json:"period" binding:"required"` // YYYY-MM 格式
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        // 获取所有活跃租户
        var tenants []tenantModel.Tenant
        h.db.Where("status = ?", "active").Find(&tenants)

        var created, skipped int
        for _, t := range tenants {
                // 检查是否已存在该周期账单
                var count int64
                h.db.Model(&tenantModel.TenantBilling{}).
                        Where("tenant_id = ? AND billing_period = ?", t.ID, req.Period).
                        Count(&count)
                if count > 0 {
                        skipped++
                        continue
                }

                // 获取套餐定价
                baseAmount := getPlanPrice(t.Plan)

                // 创建账单
                billing := &tenantModel.TenantBilling{
                        ID:            generateUUID(),
                        TenantID:      t.ID,
                        BillingPeriod: req.Period,
                        DueDate:       time.Now().AddDate(0, 0, 15),
                        BaseAmount:    baseAmount,
                        TotalAmount:   baseAmount,
                        Status:        "pending",
                }

                if err := h.db.Create(billing).Error; err == nil {
                        created++
                }
        }

        c.JSON(http.StatusOK, gin.H{
                "message":  "账单生成完成",
                "created":  created,
                "skipped":  skipped,
        })
}

// ==================== 审计日志接口 ====================

// ListAuditLogsAdmin 管理员查看所有审计日志
func (h *Handler) ListAuditLogsAdmin(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
        tenantID := c.Query("tenant_id")
        action := c.Query("action")
        resource := c.Query("resource")
        status := c.Query("status")

        var logs []tenantModel.TenantAuditLog
        var total int64

        query := h.db.Model(&tenantModel.TenantAuditLog{}).Preload("Tenant")
        if tenantID != "" {
                query = query.Where("tenant_id = ?", tenantID)
        }
        if action != "" {
                query = query.Where("action = ?", action)
        }
        if resource != "" {
                query = query.Where("resource = ?", resource)
        }
        if status != "" {
                query = query.Where("status = ?", status)
        }

        query.Count(&total)
        err := query.Order("created_at DESC").
                Offset((page - 1) * pageSize).
                Limit(pageSize).
                Find(&logs).Error

        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "data":      logs,
                "total":     total,
                "page":      page,
                "page_size": pageSize,
        })
}

// GetAuditLog 获取审计日志详情
func (h *Handler) GetAuditLog(c *gin.Context) {
        id := c.Param("id")
        var log tenantModel.TenantAuditLog
        err := h.db.Preload("Tenant").First(&log, "id = ?", id).Error
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "日志不存在"})
                return
        }
        c.JSON(http.StatusOK, gin.H{"data": log})
}

// CreateAuditLog 创建审计日志
func (h *Handler) CreateAuditLog(c *gin.Context) {
        var req struct {
                TenantID     string                 `json:"tenant_id" binding:"required"`
                UserID       string                 `json:"user_id"`
                UserName     string                 `json:"user_name"`
                UserEmail    string                 `json:"user_email"`
                Action       string                 `json:"action" binding:"required"`
                Resource     string                 `json:"resource"`
                ResourceID   string                 `json:"resource_id"`
                ResourceName string                 `json:"resource_name"`
                OldValue     map[string]interface{} `json:"old_value"`
                NewValue     map[string]interface{} `json:"new_value"`
                Changes      map[string]interface{} `json:"changes"`
                IPAddress    string                 `json:"ip_address"`
                UserAgent    string                 `json:"user_agent"`
                RequestID    string                 `json:"request_id"`
                Status       string                 `json:"status"`
                ErrorMsg     string                 `json:"error_msg"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        if req.Status == "" {
                req.Status = "success"
        }

        log := &tenantModel.TenantAuditLog{
                ID:           generateUUID(),
                TenantID:     req.TenantID,
                UserID:       req.UserID,
                UserName:     req.UserName,
                UserEmail:    req.UserEmail,
                Action:       req.Action,
                Resource:     req.Resource,
                ResourceID:   req.ResourceID,
                ResourceName: req.ResourceName,
                OldValue:     req.OldValue,
                NewValue:     req.NewValue,
                Changes:      req.Changes,
                IPAddress:    req.IPAddress,
                UserAgent:    req.UserAgent,
                RequestID:    req.RequestID,
                Status:       req.Status,
                ErrorMsg:     req.ErrorMsg,
        }

        if err := h.db.Create(log).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, gin.H{"data": log})
}

// ==================== 套餐管理接口 ====================

// ListPlans 列出套餐配置
func (h *Handler) ListPlans(c *gin.Context) {
        plans := []map[string]interface{}{}
        for key, quota := range tenantModel.PlanQuotas {
                plan := map[string]interface{}{
                        "key":         key,
                        "name":        getPlanName(key),
                        "price":       getPlanPrice(key),
                        "description": getPlanDescription(key),
                        "quota":       quota,
                }
                plans = append(plans, plan)
        }
        c.JSON(http.StatusOK, gin.H{"data": plans})
}

// UpdatePlan 更新套餐配置
func (h *Handler) UpdatePlan(c *gin.Context) {
        planKey := c.Param("key")
        var req struct {
                Name        string  `json:"name"`
                Price       float64 `json:"price"`
                Description string  `json:"description"`
                Quota       tenantModel.TenantQuota `json:"quota"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        // 更新预设配额
        tenantModel.PlanQuotas[planKey] = req.Quota

        c.JSON(http.StatusOK, gin.H{"message": "套餐配置已更新"})
}

// ==================== 统计接口 ====================

// GetBillingStats 获取账单统计
func (h *Handler) GetBillingStats(c *gin.Context) {
        var totalRevenue, paidAmount, pendingAmount float64
        var overdueCount int64

        // 计算本月收入
        h.db.Model(&tenantModel.TenantBilling{}).
                Where("billing_period = ?", time.Now().Format("2006-01")).
                Select("COALESCE(SUM(total_amount), 0)").
                Scan(&totalRevenue)

        // 已收款
        h.db.Model(&tenantModel.TenantBilling{}).
                Where("status = ?", "paid").
                Select("COALESCE(SUM(total_amount), 0)").
                Scan(&paidAmount)

        // 待收款
        h.db.Model(&tenantModel.TenantBilling{}).
                Where("status = ?", "pending").
                Select("COALESCE(SUM(total_amount), 0)").
                Scan(&pendingAmount)

        // 逾期账单
        h.db.Model(&tenantModel.TenantBilling{}).
                Where("status = ? AND due_date < ?", "pending", time.Now()).
                Count(&overdueCount)

        c.JSON(http.StatusOK, gin.H{
                "data": gin.H{
                        "total_revenue": totalRevenue,
                        "paid_amount":   paidAmount,
                        "pending_amount": pendingAmount,
                        "overdue_count": overdueCount,
                },
        })
}

// GetAuditStats 获取审计日志统计
func (h *Handler) GetAuditStats(c *gin.Context) {
        var successCount, failedCount int64

        h.db.Model(&tenantModel.TenantAuditLog{}).
                Where("status = ?", "success").Count(&successCount)
        h.db.Model(&tenantModel.TenantAuditLog{}).
                Where("status = ?", "failed").Count(&failedCount)

        c.JSON(http.StatusOK, gin.H{
                "data": gin.H{
                        "success": successCount,
                        "failed":  failedCount,
                },
        })
}

// ==================== 辅助函数 ====================

func generateUUID() string {
        return strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatInt(time.Now().UnixMicro()%1000000, 36)
}

func getPlanPrice(plan string) float64 {
        prices := map[string]float64{
                "free":       0,
                "starter":    99,
                "pro":        299,
                "enterprise": 0,
        }
        return prices[plan]
}

func getPlanName(plan string) string {
        names := map[string]string{
                "free":       "Free",
                "starter":    "Starter",
                "pro":        "Pro",
                "enterprise": "Enterprise",
        }
        return names[plan]
}

func getPlanDescription(plan string) string {
        descs := map[string]string{
                "free":       "免费版，适合个人学习使用",
                "starter":    "入门版，适合小型团队",
                "pro":        "专业版，适合中型企业",
                "enterprise": "企业版，联系销售定制",
        }
        return descs[plan]
}
