package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"yunwei/global"
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	ChannelTelegram NotificationChannel = "telegram"
	ChannelWechat   NotificationChannel = "wechat"
	ChannelEmail    NotificationChannel = "email"
	ChannelWebhook  NotificationChannel = "webhook"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotifyAlert      NotificationType = "alert"
	NotifyReport     NotificationType = "report"
	NotifyDecision   NotificationType = "decision"
	NotifySystem     NotificationType = "system"
	NotifySecurity   NotificationType = "security"
)

// NotificationRecord 通知记录
type NotificationRecord struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`

	Channel   NotificationChannel `json:"channel" gorm:"type:varchar(16)"`
	Type      NotificationType    `json:"type" gorm:"type:varchar(16)"`
	Title     string              `json:"title" gorm:"type:varchar(255)"`
	Content   string              `json:"content" gorm:"type:text"`
	
	Success   bool                `json:"success"`
	Error     string              `json:"error" gorm:"type:text"`
	Response  string              `json:"response" gorm:"type:text"`
}

func (NotificationRecord) TableName() string {
	return "notification_records"
}

// TelegramConfig Telegram配置
type TelegramConfig struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
	ParseMode string `json:"parseMode"` // HTML, Markdown
}

// WechatConfig 企业微信配置
type WechatConfig struct {
	WebhookURL string `json:"webhookUrl"`
	CorpID     string `json:"corpId"`
	AgentID    string `json:"agentId"`
	Secret     string `json:"secret"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	From         string `json:"from"`
	FromName     string `json:"fromName"`
}

// NotifierService 通知服务
type NotifierService struct {
	telegram TelegramConfig
	wechat   WechatConfig
	email    EmailConfig
	client   *http.Client
}

// NewNotifierService 创建通知服务
func NewNotifierService() *NotifierService {
	return &NotifierService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetTelegram 设置Telegram配置
func (n *NotifierService) SetTelegram(config TelegramConfig) {
	n.telegram = config
}

// SetWechat 设置企业微信配置
func (n *NotifierService) SetWechat(config WechatConfig) {
	n.wechat = config
}

// SetEmail 设置邮件配置
func (n *NotifierService) SetEmail(config EmailConfig) {
	n.email = config
}

// ==================== Telegram ====================

// SendTelegram 发送Telegram消息
func (n *NotifierService) SendTelegram(title, content string) error {
	if n.telegram.BotToken == "" || n.telegram.ChatID == "" {
		return fmt.Errorf("Telegram配置不完整")
	}

	text := fmt.Sprintf("<b>%s</b>\n\n%s", title, content)
	if n.telegram.ParseMode == "" {
		n.telegram.ParseMode = "HTML"
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.telegram.BotToken)

	payload := map[string]interface{}{
		"chat_id":    n.telegram.ChatID,
		"text":       text,
		"parse_mode": n.telegram.ParseMode,
	}

	return n.sendRequest(ChannelTelegram, url, payload, title, content)
}

// SendTelegramAlert 发送Telegram告警
func (n *NotifierService) SendTelegramAlert(serverName, alertType, message string, level string) error {
	emoji := "⚠️"
	if level == "critical" {
		emoji = "🔴"
	} else if level == "warning" {
		emoji = "🟡"
	}

	title := fmt.Sprintf("%s 告警通知", emoji)
	content := fmt.Sprintf(
		"服务器: %s\n类型: %s\n级别: %s\n\n%s\n\n时间: %s",
		serverName, alertType, level, message, time.Now().Format("2006-01-02 15:04:05"),
	)

	return n.SendTelegram(title, content)
}

// ==================== 企业微信 ====================

// SendWechat 发送企业微信消息
func (n *NotifierService) SendWechat(title, content string) error {
	if n.wechat.WebhookURL == "" {
		return fmt.Errorf("企业微信Webhook未配置")
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": fmt.Sprintf("# %s\n\n%s", title, content),
		},
	}

	return n.sendRequest(ChannelWechat, n.wechat.WebhookURL, payload, title, content)
}

// SendWechatAlert 发送企业微信告警
func (n *NotifierService) SendWechatAlert(serverName, alertType, message string, level string) error {
	color := "warning"
	if level == "critical" {
		color = "warning"
	} else if level == "warning" {
		color = "comment"
	}

	title := fmt.Sprintf("【告警通知】%s", serverName)
	content := fmt.Sprintf(
		"> 类型: <font color=\"%s\">%s</font>\n> 级别: %s\n> 详情: %s\n> 时间: %s",
		color, alertType, level, message, time.Now().Format("2006-01-02 15:04:05"),
	)

	return n.SendWechat(title, content)
}

// SendWechatCard 发送企业微信卡片消息
func (n *NotifierService) SendWechatCard(title, description, url, btntext string) error {
	if n.wechat.WebhookURL == "" {
		return fmt.Errorf("企业微信Webhook未配置")
	}

	payload := map[string]interface{}{
		"msgtype": "template_card",
		"template_card": map[string]interface{}{
			"card_type": "text_notice",
			"main_title": map[string]string{
				"title": title,
			},
			"sub_title_text": description,
			"card_action": map[string]interface{}{
				"type": 1,
				"url":  url,
			},
		},
	}

	return n.sendRequest(ChannelWechat, n.wechat.WebhookURL, payload, title, description)
}

// ==================== Webhook ====================

// SendWebhook 发送Webhook通知
func (n *NotifierService) SendWebhook(webhookURL string, data map[string]interface{}) error {
	return n.sendRequest(ChannelWebhook, webhookURL, data, "", "")
}

// ==================== 批量通知 ====================

// Broadcast 广播通知到所有渠道
func (n *NotifierService) Broadcast(title, content string) map[NotificationChannel]error {
	errors := make(map[NotificationChannel]error)

	// Telegram
	if n.telegram.BotToken != "" {
		if err := n.SendTelegram(title, content); err != nil {
			errors[ChannelTelegram] = err
		}
	}

	// 企业微信
	if n.wechat.WebhookURL != "" {
		if err := n.SendWechat(title, content); err != nil {
			errors[ChannelWechat] = err
		}
	}

	return errors
}

// BroadcastAlert 广播告警
func (n *NotifierService) BroadcastAlert(serverName, alertType, message, level string) map[NotificationChannel]error {
	errors := make(map[NotificationChannel]error)

	if n.telegram.BotToken != "" {
		if err := n.SendTelegramAlert(serverName, alertType, message, level); err != nil {
			errors[ChannelTelegram] = err
		}
	}

	if n.wechat.WebhookURL != "" {
		if err := n.SendWechatAlert(serverName, alertType, message, level); err != nil {
			errors[ChannelWechat] = err
		}
	}

	return errors
}

// ==================== HTTP请求 ====================

// sendRequest 发送HTTP请求
func (n *NotifierService) sendRequest(channel NotificationChannel, url string, payload interface{}, title, content string) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		n.recordNotification(channel, NotifySystem, title, content, false, err.Error(), "")
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		n.recordNotification(channel, NotifySystem, title, content, false, err.Error(), "")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		n.recordNotification(channel, NotifySystem, title, content, false, err.Error(), "")
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		n.recordNotification(channel, NotifySystem, title, content, false, errMsg, string(body))
		return fmt.Errorf(errMsg)
	}

	n.recordNotification(channel, NotifySystem, title, content, true, "", string(body))
	return nil
}

// recordNotification 记录通知
func (n *NotifierService) recordNotification(channel NotificationChannel, notifyType NotificationType, title, content string, success bool, errMsg, response string) {
	record := NotificationRecord{
		Channel:  channel,
		Type:     notifyType,
		Title:    title,
		Content:  content,
		Success:  success,
		Error:    errMsg,
		Response: response,
	}

	if global.DB != nil {
		global.DB.Create(&record)
	}
}

// ==================== 历史记录 ====================

// GetHistory 获取通知历史
func (n *NotifierService) GetHistory(channel NotificationChannel, limit int) ([]NotificationRecord, error) {
	var records []NotificationRecord
	query := global.DB.Order("created_at DESC")
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}

// GetFailedNotifications 获取失败的通知
func (n *NotifierService) GetFailedNotifications(limit int) ([]NotificationRecord, error) {
	var records []NotificationRecord
	err := global.DB.Where("success = ?", false).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// RetryFailed 重试失败的通知
func (n *NotifierService) RetryFailed(recordID uint) error {
	var record NotificationRecord
	if err := global.DB.First(&record, recordID).Error; err != nil {
		return err
	}

	switch record.Channel {
	case ChannelTelegram:
		return n.SendTelegram(record.Title, record.Content)
	case ChannelWechat:
		return n.SendWechat(record.Title, record.Content)
	}

	return fmt.Errorf("不支持的重试渠道: %s", record.Channel)
}

// ==================== 模板消息 ====================

// AlertTemplate 告警模板
type AlertTemplate struct {
	ServerName  string
	AlertType   string
	Level       string
	Message     string
	Value       float64
	Threshold   float64
	Timestamp   time.Time
}

// FormatAlert 格式化告警消息
func (n *NotifierService) FormatAlert(t AlertTemplate) string {
	return fmt.Sprintf(
		"🖥️ 服务器: %s\n📋 类型: %s\n⚡ 级别: %s\n📊 当前值: %.2f (阈值: %.2f)\n📝 详情: %s\n⏰ 时间: %s",
		t.ServerName, t.AlertType, t.Level, t.Value, t.Threshold, t.Message, t.Timestamp.Format("2006-01-02 15:04:05"),
	)
}

// ReportTemplate 报告模板
type ReportTemplate struct {
	Title       string
	Summary     string
	Details     []string
	Recommendations []string
	Timestamp   time.Time
}

// FormatReport 格式化报告消息
func (n *NotifierService) FormatReport(t ReportTemplate) string {
	content := fmt.Sprintf("📅 %s\n\n", t.Timestamp.Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("📊 %s\n\n", t.Summary)

	if len(t.Details) > 0 {
		content += "📋 详情:\n"
		for _, d := range t.Details {
			content += fmt.Sprintf("  • %s\n", d)
		}
		content += "\n"
	}

	if len(t.Recommendations) > 0 {
		content += "💡 建议:\n"
		for _, r := range t.Recommendations {
			content += fmt.Sprintf("  • %s\n", r)
		}
	}

	return content
}
