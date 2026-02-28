package notify

import (
        "bytes"
        "encoding/json"
        "fmt"
        "net/http"
        "time"

        "yunwei/global"
        patrolModel "yunwei/model/patrol"
        "yunwei/service/detector"
)

// Notifier 通知器接口
type Notifier interface {
        SendPatrolReport(record *patrolModel.PatrolRecord) error
        SendAlert(alert *detector.Alert) error
        SendMessage(title, content string) error
}

// NotifyConfig 通知配置
type NotifyConfig struct {
        // Telegram
        TelegramEnabled bool   `json:"telegramEnabled"`
        TelegramToken   string `json:"telegramToken"`
        TelegramChatID  string `json:"telegramChatId"`

        // 企业微信
        WeChatEnabled   bool   `json:"weChatEnabled"`
        WeChatWebhook   string `json:"weChatWebhook"`

        // 钉钉
        DingTalkEnabled bool   `json:"dingTalkEnabled"`
        DingTalkWebhook string `json:"dingTalkWebhook"`

        // 邮件
        EmailEnabled  bool     `json:"emailEnabled"`
        SMTPHost      string   `json:"smtpHost"`
        SMTPPort      int      `json:"smtpPort"`
        SMTPUser      string   `json:"smtpUser"`
        SMTPPassword  string   `json:"smtpPassword"`
        EmailTo       []string `json:"emailTo"`

        // 飞书
        FeishuEnabled bool   `json:"feishuEnabled"`
        FeishuWebhook string `json:"feishuWebhook"`
}

// NotifyRecord 通知记录
type NotifyRecord struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`

        Type      string `json:"type" gorm:"type:varchar(32)"`  // patrol, alert, message
        Channel   string `json:"channel" gorm:"type:varchar(32)"` // telegram, wechat, dingtalk, email
        Title     string `json:"title" gorm:"type:varchar(255)"`
        Content   string `json:"content" gorm:"type:text"`
        Status    string `json:"status" gorm:"type:varchar(16)"` // success, failed
        Error     string `json:"error" gorm:"type:text"`
}

func (NotifyRecord) TableName() string {
        return "notify_records"
}

// TelegramNotifier Telegram通知器
type TelegramNotifier struct {
        Token  string
        ChatID string
}

// NewTelegramNotifier 创建Telegram通知器
func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
        return &TelegramNotifier{
                Token:  token,
                ChatID: chatID,
        }
}

// SendMessage 发送消息
func (t *TelegramNotifier) SendMessage(text string) error {
        url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)

        payload := map[string]interface{}{
                "chat_id":    t.ChatID,
                "text":       text,
                "parse_mode": "Markdown",
        }

        jsonData, _ := json.Marshal(payload)

        resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
        if err != nil {
                return fmt.Errorf("发送Telegram消息失败: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("Telegram API返回错误: %d", resp.StatusCode)
        }

        return nil
}

// SendPatrolReport 发送巡检报告
func (t *TelegramNotifier) SendPatrolReport(record *patrolModel.PatrolRecord) error {
        text := formatPatrolReport(record)
        return t.SendMessage(text)
}

// SendAlert 发送告警
func (t *TelegramNotifier) SendAlert(alert *detector.Alert) error {
        text := formatAlert(alert)
        return t.SendMessage(text)
}

// WeChatNotifier 企业微信通知器
type WeChatNotifier struct {
        Webhook string
}

// NewWeChatNotifier 创建企业微信通知器
func NewWeChatNotifier(webhook string) *WeChatNotifier {
        return &WeChatNotifier{
                Webhook: webhook,
        }
}

// SendMessage 发送消息
func (w *WeChatNotifier) SendMessage(content string) error {
        payload := map[string]interface{}{
                "msgtype": "markdown",
                "markdown": map[string]string{
                        "content": content,
                },
        }

        jsonData, _ := json.Marshal(payload)

        resp, err := http.Post(w.Webhook, "application/json", bytes.NewReader(jsonData))
        if err != nil {
                return fmt.Errorf("发送企业微信消息失败: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("企业微信API返回错误: %d", resp.StatusCode)
        }

        return nil
}

// SendPatrolReport 发送巡检报告
func (w *WeChatNotifier) SendPatrolReport(record *patrolModel.PatrolRecord) error {
        content := formatPatrolReportMarkdown(record)
        return w.SendMessage(content)
}

// SendAlert 发送告警
func (w *WeChatNotifier) SendAlert(alert *detector.Alert) error {
        content := formatAlertMarkdown(alert)
        return w.SendMessage(content)
}

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
        Webhook string
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier(webhook string) *DingTalkNotifier {
        return &DingTalkNotifier{
                Webhook: webhook,
        }
}

// SendMessage 发送消息
func (d *DingTalkNotifier) SendMessage(content string) error {
        payload := map[string]interface{}{
                "msgtype": "markdown",
                "markdown": map[string]string{
                        "title": "运维通知",
                        "text":  content,
                },
        }

        jsonData, _ := json.Marshal(payload)

        resp, err := http.Post(d.Webhook, "application/json", bytes.NewReader(jsonData))
        if err != nil {
                return fmt.Errorf("发送钉钉消息失败: %w", err)
        }
        defer resp.Body.Close()

        return nil
}

// SendPatrolReport 发送巡检报告
func (d *DingTalkNotifier) SendPatrolReport(record *patrolModel.PatrolRecord) error {
        content := formatPatrolReportMarkdown(record)
        return d.SendMessage(content)
}

// SendAlert 发送告警
func (d *DingTalkNotifier) SendAlert(alert *detector.Alert) error {
        content := formatAlertMarkdown(alert)
        return d.SendMessage(content)
}

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
        Webhook string
}

// NewFeishuNotifier 创建飞书通知器
func NewFeishuNotifier(webhook string) *FeishuNotifier {
        return &FeishuNotifier{
                Webhook: webhook,
        }
}

// SendMessage 发送消息
func (f *FeishuNotifier) SendMessage(content string) error {
        payload := map[string]interface{}{
                "msg_type": "interactive",
                "card": map[string]interface{}{
                        "elements": []map[string]interface{}{
                                {
                                        "tag": "markdown",
                                        "content": content,
                                },
                        },
                },
        }

        jsonData, _ := json.Marshal(payload)

        resp, err := http.Post(f.Webhook, "application/json", bytes.NewReader(jsonData))
        if err != nil {
                return fmt.Errorf("发送飞书消息失败: %w", err)
        }
        defer resp.Body.Close()

        return nil
}

// SendPatrolReport 发送巡检报告
func (f *FeishuNotifier) SendPatrolReport(record *patrolModel.PatrolRecord) error {
        content := formatPatrolReportMarkdown(record)
        return f.SendMessage(content)
}

// MultiNotifier 多通道通知器
type MultiNotifier struct {
        telegram *TelegramNotifier
        wechat   *WeChatNotifier
        dingtalk *DingTalkNotifier
        feishu   *FeishuNotifier
}

// NewMultiNotifier 创建多通道通知器
func NewMultiNotifier(config NotifyConfig) *MultiNotifier {
        n := &MultiNotifier{}

        if config.TelegramEnabled && config.TelegramToken != "" {
                n.telegram = NewTelegramNotifier(config.TelegramToken, config.TelegramChatID)
        }
        if config.WeChatEnabled && config.WeChatWebhook != "" {
                n.wechat = NewWeChatNotifier(config.WeChatWebhook)
        }
        if config.DingTalkEnabled && config.DingTalkWebhook != "" {
                n.dingtalk = NewDingTalkNotifier(config.DingTalkWebhook)
        }
        if config.FeishuEnabled && config.FeishuWebhook != "" {
                n.feishu = NewFeishuNotifier(config.FeishuWebhook)
        }

        return n
}

// SendPatrolReport 发送巡检报告到所有配置的通道
func (n *MultiNotifier) SendPatrolReport(record *patrolModel.PatrolRecord) error {
        if n.telegram != nil {
                if err := n.telegram.SendPatrolReport(record); err != nil {
                        n.logNotify("patrol", "telegram", "巡检报告", record.Summary, "failed", err.Error())
                } else {
                        n.logNotify("patrol", "telegram", "巡检报告", record.Summary, "success", "")
                }
        }

        if n.wechat != nil {
                if err := n.wechat.SendPatrolReport(record); err != nil {
                        n.logNotify("patrol", "wechat", "巡检报告", record.Summary, "failed", err.Error())
                } else {
                        n.logNotify("patrol", "wechat", "巡检报告", record.Summary, "success", "")
                }
        }

        if n.dingtalk != nil {
                if err := n.dingtalk.SendPatrolReport(record); err != nil {
                        n.logNotify("patrol", "dingtalk", "巡检报告", record.Summary, "failed", err.Error())
                } else {
                        n.logNotify("patrol", "dingtalk", "巡检报告", record.Summary, "success", "")
                }
        }

        if n.feishu != nil {
                if err := n.feishu.SendPatrolReport(record); err != nil {
                        n.logNotify("patrol", "feishu", "巡检报告", record.Summary, "failed", err.Error())
                } else {
                        n.logNotify("patrol", "feishu", "巡检报告", record.Summary, "success", "")
                }
        }

        return nil
}

// SendAlert 发送告警
func (n *MultiNotifier) SendAlert(alert *detector.Alert) error {
        content := fmt.Sprintf("%s: %s", alert.Title, alert.Message)

        if n.telegram != nil {
                n.telegram.SendAlert(alert)
                n.logNotify("alert", "telegram", alert.Title, content, "success", "")
        }

        if n.wechat != nil {
                n.wechat.SendAlert(alert)
                n.logNotify("alert", "wechat", alert.Title, content, "success", "")
        }

        if n.dingtalk != nil {
                n.dingtalk.SendAlert(alert)
                n.logNotify("alert", "dingtalk", alert.Title, content, "success", "")
        }

        return nil
}

// SendMessage 发送普通消息
func (n *MultiNotifier) SendMessage(title, content string) error {
        if n.telegram != nil {
                n.telegram.SendMessage(fmt.Sprintf("*%s*\n\n%s", title, content))
        }
        if n.wechat != nil {
                n.wechat.SendMessage(fmt.Sprintf("## %s\n\n%s", title, content))
        }
        if n.dingtalk != nil {
                n.dingtalk.SendMessage(fmt.Sprintf("### %s\n\n%s", title, content))
        }

        return nil
}

// logNotify 记录通知日志
func (n *MultiNotifier) logNotify(notifyType, channel, title, content, status, errMsg string) {
        record := NotifyRecord{
                Type:    notifyType,
                Channel: channel,
                Title:   title,
                Content: content,
                Status:  status,
                Error:   errMsg,
        }
        global.DB.Create(&record)
}

// Helper functions
func formatPatrolReport(record *patrolModel.PatrolRecord) string {
        return fmt.Sprintf(`🤖 *服务器巡检报告*

📅 时间: %s
📊 类型: %s

*服务器状态*
• 总数: %d
• 🟢 在线: %d
• 🔴 离线: %d
• ⚠️ 警告: %d
• 🔥 严重: %d

*告警统计*
• 总计: %d

⏱ 耗时: %dms`,
                record.CreatedAt.Format("2006-01-02 15:04"),
                record.Type,
                record.TotalServers,
                record.OnlineServers,
                record.OfflineServers,
                record.WarningCount,
                record.CriticalCount,
                record.AlertCount,
                record.Duration,
        )
}

func formatPatrolReportMarkdown(record *patrolModel.PatrolRecord) string {
        return fmt.Sprintf(`# 🤖 服务器巡检报告

> 时间: %s | 类型: %s

## 服务器状态

| 指标 | 数量 |
| --- | --- |
| 总数 | %d |
| 🟢 在线 | %d |
| 🔴 离线 | %d |
| ⚠️ 警告 | %d |
| 🔥 严重 | %d |

## 告警统计

总计: **%d** 条

---
⏱ 耗时: %dms`,
                record.CreatedAt.Format("2006-01-02 15:04"),
                record.Type,
                record.TotalServers,
                record.OnlineServers,
                record.OfflineServers,
                record.WarningCount,
                record.CriticalCount,
                record.AlertCount,
                record.Duration,
        )
}

func formatAlert(alert *detector.Alert) string {
        levelEmoji := map[detector.AlertLevel]string{
                detector.AlertLevelInfo:     "ℹ️",
                detector.AlertLevelWarning:  "⚠️",
                detector.AlertLevelCritical: "🔥",
                detector.AlertLevelEmergency: "🚨",
        }

        return fmt.Sprintf(`%s *告警通知*

*标题*: %s
*级别*: %s
*时间*: %s

*详情*:
%s`,
                levelEmoji[alert.Level],
                alert.Title,
                alert.Level,
                alert.CreatedAt.Format("2006-01-02 15:04:05"),
                alert.Message,
        )
}

func formatAlertMarkdown(alert *detector.Alert) string {
        levelColor := map[detector.AlertLevel]string{
                detector.AlertLevelInfo:     "蓝色",
                detector.AlertLevelWarning:  "橙色",
                detector.AlertLevelCritical: "红色",
                detector.AlertLevelEmergency: "紫色",
        }

        return fmt.Sprintf(`# 🚨 告警通知

> 级别: <font color="%s">%s</font>

**标题**: %s

**时间**: %s

**详情**:
%s`,
                levelColor[alert.Level],
                alert.Level,
                alert.Title,
                alert.CreatedAt.Format("2006-01-02 15:04:05"),
                alert.Message,
        )
}
