package security

import (
	"encoding/json"
	"time"

	"yunwei/global"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditService 审计服务
type AuditService struct {
	db *gorm.DB
}

// NewAuditService 创建审计服务
func NewAuditService() *AuditService {
	return &AuditService{db: global.DB}
}

// AuditAction 审计动作类型
type AuditAction string

const (
	AuditActionLogin      AuditAction = "login"
	AuditActionLogout     AuditAction = "logout"
	AuditActionExecute    AuditAction = "execute"
	AuditActionApprove    AuditAction = "approve"
	AuditActionReject     AuditAction = "reject"
	AuditActionCancel     AuditAction = "cancel"
	AuditActionCreate     AuditAction = "create"
	AuditActionUpdate     AuditAction = "update"
	AuditActionDelete     AuditAction = "delete"
	AuditActionView       AuditAction = "view"
	AuditActionExport     AuditAction = "export"
	AuditActionImport     AuditAction = "import"
	AuditActionConfig     AuditAction = "config"
)

// LogParams 日志参数
type LogParams struct {
	UserID     uint
	Username   string
	ServerID   uint
	ServerName string
	Action     AuditAction
	Resource   string
	Command    string
	Result     string
	Details    map[string]interface{}
}

// Log 记录审计日志
func (s *AuditService) Log(params LogParams) error {
	detailsJSON, _ := json.Marshal(params.Details)

	log := AuditLog{
		UserID:     params.UserID,
		Username:   params.Username,
		ServerID:   params.ServerID,
		ServerName: params.ServerName,
		Action:     string(params.Action),
		Resource:   params.Resource,
		Command:    params.Command,
		Result:     params.Result,
		IP:         "",
		UserAgent:  "",
	}

	// 存储详细信息
	if len(detailsJSON) > 0 {
		log.Result = string(detailsJSON)
	}

	return global.DB.Create(&log).Error
}

// LogFromGin 从Gin上下文记录日志
func (s *AuditService) LogFromGin(c *gin.Context, params LogParams) error {
	// 获取IP
	params.Details["ip"] = c.ClientIP()
	params.Details["userAgent"] = c.GetHeader("User-Agent")
	params.Details["path"] = c.Request.URL.Path
	params.Details["method"] = c.Request.Method

	// 如果没有用户名，尝试从上下文获取
	if params.UserID == 0 {
		if userID, exists := c.Get("userID"); exists {
			params.UserID = userID.(uint)
		}
	}
	if params.Username == "" {
		if username, exists := c.Get("username"); exists {
			params.Username = username.(string)
		}
	}

	log := AuditLog{
		UserID:     params.UserID,
		Username:   params.Username,
		ServerID:   params.ServerID,
		ServerName: params.ServerName,
		Action:     string(params.Action),
		Resource:   params.Resource,
		Command:    params.Command,
		Result:     params.Result,
		IP:         c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	return global.DB.Create(&log).Error
}

// LogLogin 记录登录
func (s *AuditService) LogLogin(userID uint, username, ip string, success bool) error {
	result := "success"
	if !success {
		result = "failed"
	}

	log := AuditLog{
		UserID:   userID,
		Username: username,
		Action:   string(AuditActionLogin),
		Resource: "auth",
		Result:   result,
		IP:       ip,
	}

	return global.DB.Create(&log).Error
}

// LogCommand 记录命令执行
func (s *AuditService) LogCommand(userID uint, username string, serverID uint, serverName, command, result string) error {
	log := AuditLog{
		UserID:     userID,
		Username:   username,
		ServerID:   serverID,
		ServerName: serverName,
		Action:     string(AuditActionExecute),
		Resource:   "command",
		Command:    command,
		Result:     result,
	}

	return global.DB.Create(&log).Error
}

// LogApproval 记录审批
func (s *AuditService) LogApproval(userID uint, username string, executionID uint, approved bool, reason string) error {
	result := "approved"
	if !approved {
		result = "rejected"
	}

	log := AuditLog{
		UserID:   userID,
		Username: username,
		Action:   string(AuditActionApprove),
		Resource: "execution",
		Result:   result,
		Command:  reason,
	}

	return global.DB.Create(&log).Error
}

// GetLogs 获取日志列表
func (s *AuditService) GetLogs(params map[string]interface{}, page, pageSize int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := global.DB.Model(&AuditLog{})

	// 过滤条件
	if userID, ok := params["userId"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if serverID, ok := params["serverId"].(uint); ok && serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	if action, ok := params["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if result, ok := params["result"].(string); ok && result != "" {
		query = query.Where("result = ?", result)
	}
	if startTime, ok := params["startTime"].(time.Time); ok {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := params["endTime"].(time.Time); ok {
		query = query.Where("created_at <= ?", endTime)
	}
	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		query = query.Where("username LIKE ? OR command LIKE ? OR resource LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLogsByUser 获取用户日志
func (s *AuditService) GetLogsByUser(userID uint, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	err := global.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetLogsByServer 获取服务器日志
func (s *AuditService) GetLogsByServer(serverID uint, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	err := global.DB.Where("server_id = ?", serverID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetLoginHistory 获取登录历史
func (s *AuditService) GetLoginHistory(userID uint, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := global.DB.Where("action = ?", AuditActionLogin)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// GetCommandHistory 获取命令执行历史
func (s *AuditService) GetCommandHistory(serverID uint, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := global.DB.Where("action = ?", AuditActionExecute)
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// GetUserStatistics 获取用户操作统计
func (s *AuditService) GetUserStatistics(userID uint, days int) map[string]int64 {
	startTime := time.Now().AddDate(0, 0, -days)
	stats := make(map[string]int64)

	// 各类操作次数
	actions := []string{
		string(AuditActionExecute),
		string(AuditActionApprove),
		string(AuditActionReject),
		string(AuditActionLogin),
	}

	for _, action := range actions {
		var count int64
		global.DB.Model(&AuditLog{}).
			Where("user_id = ? AND action = ? AND created_at > ?", userID, action, startTime).
			Count(&count)
		stats[action] = count
	}

	return stats
}

// GetServerStatistics 获取服务器操作统计
func (s *AuditService) GetServerStatistics(serverID uint, days int) map[string]int64 {
	startTime := time.Now().AddDate(0, 0, -days)
	stats := make(map[string]int64)

	// 命令执行次数
	var execCount int64
	global.DB.Model(&AuditLog{}).
		Where("server_id = ? AND action = ? AND created_at > ?", serverID, AuditActionExecute, startTime).
		Count(&execCount)
	stats["executions"] = execCount

	// 成功次数
	var successCount int64
	global.DB.Model(&AuditLog{}).
		Where("server_id = ? AND action = ? AND result = ? AND created_at > ?", serverID, AuditActionExecute, "success", startTime).
		Count(&successCount)
	stats["success"] = successCount

	// 失败次数
	stats["failed"] = execCount - successCount

	return stats
}

// CleanOldLogs 清理旧日志
func (s *AuditService) CleanOldLogs(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return global.DB.Where("created_at < ?", cutoff).Delete(&AuditLog{}).Error
}

// ExportLogs 导出日志
func (s *AuditService) ExportLogs(params map[string]interface{}, startTime, endTime time.Time) ([]AuditLog, error) {
	var logs []AuditLog

	query := global.DB.Model(&AuditLog{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime)

	// 应用其他过滤条件
	if userID, ok := params["userId"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := params["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}

	err := query.Order("created_at ASC").Find(&logs).Error
	return logs, err
}

// GetDailyStats 获取每日统计
func (s *AuditService) GetDailyStats(days int) []map[string]interface{} {
	var results []map[string]interface{}

	startDate := time.Now().AddDate(0, 0, -days)

	global.DB.Model(&AuditLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count, action").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at), action").
		Order("date DESC").
		Scan(&results)

	return results
}

// AuditMiddleware 审计中间件
func AuditMiddleware(action AuditAction, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 执行请求
		c.Next()

		// 记录日志
		auditService := NewAuditService()

		var userID uint
		var username string

		if id, exists := c.Get("userID"); exists {
			userID = id.(uint)
		}
		if name, exists := c.Get("username"); exists {
			username = name.(string)
		}

		result := "success"
		if len(c.Errors) > 0 {
			result = "failed"
		}

		auditService.Log(LogParams{
			UserID:   userID,
			Username: username,
			Action:   action,
			Resource: resource,
			Result:   result,
		})
	}
}
