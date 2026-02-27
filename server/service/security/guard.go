package security

import (
        "encoding/json"
        "fmt"
        "net"
        "regexp"
        "strings"
        "time"

        "yunwei/global"
)

// SecurityEventType 安全事件类型
type SecurityEventType string

const (
        SecurityEventLogin        SecurityEventType = "login"
        SecurityEventLoginFailed  SecurityEventType = "login_failed"
        SecurityEventLoginAbnormal SecurityEventType = "login_abnormal"
        SecurityEventBruteForce   SecurityEventType = "brute_force"
        SecurityEventPortScan     SecurityEventType = "port_scan"
        SecurityEventDDoS         SecurityEventType = "ddos"
        SecurityEventMaliciousIP  SecurityEventType = "malicious_ip"
        SecurityEventSuspicious   SecurityEventType = "suspicious"
)

// GuardSecurityLevel 安全级别
type GuardSecurityLevel string

const (
        GuardSecurityLevelLow      GuardSecurityLevel = "low"
        GuardSecurityLevelMedium   GuardSecurityLevel = "medium"
        GuardSecurityLevelHigh     GuardSecurityLevel = "high"
        GuardSecurityLevelCritical GuardSecurityLevel = "critical"
)

// SecurityEvent 安全事件
type SecurityEvent struct {
        ID          uint             `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time        `json:"createdAt"`
        
        ServerID    uint             `json:"serverId" gorm:"index"`
        ServerName  string           `json:"serverName" gorm:"type:varchar(64)"`
        
        EventType   SecurityEventType `json:"eventType" gorm:"type:varchar(32)"`
        Level       GuardSecurityLevel     `json:"level" gorm:"type:varchar(16)"`
        
        // 来源信息
        SourceIP    string           `json:"sourceIp" gorm:"type:varchar(45);index"`
        SourcePort  int              `json:"sourcePort"`
        GeoLocation string           `json:"geoLocation" gorm:"type:varchar(128)"`
        
        // 目标信息
        TargetUser  string           `json:"targetUser" gorm:"type:varchar(64)"`
        TargetPort  int              `json:"targetPort"`
        TargetService string         `json:"targetService" gorm:"type:varchar(64)"`
        
        // 详情
        Description string           `json:"description" gorm:"type:text"`
        RawLog      string           `json:"rawLog" gorm:"type:text"`
        
        // 处理
        Status      string           `json:"status" gorm:"type:varchar(16);default:'new'"` // new, processing, resolved, ignored
        HandledBy   uint             `json:"handledBy"`
        HandledAt   *time.Time       `json:"handledAt"`
        Action      string           `json:"action" gorm:"type:varchar(64)"` // ban, whitelist, ignore
}

func (SecurityEvent) TableName() string {
        return "security_events"
}

// IPBlacklist IP黑名单
type IPBlacklist struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`
        
        IP          string     `json:"ip" gorm:"type:varchar(45);uniqueIndex"`
        CIDR        string     `json:"cidr" gorm:"type:varchar(64)"` // CIDR格式
        
        // 来源
        Reason      string     `json:"reason" gorm:"type:varchar(255)"`
        EventType   SecurityEventType `json:"eventType" gorm:"type:varchar(32)"`
        EventID     uint       `json:"eventId"`
        
        // 自动/手动
        AutoBanned  bool       `json:"autoBanned"`
        BannedBy    uint       `json:"bannedBy"`
        
        // 过期
        Permanent   bool       `json:"permanent"`
        ExpiresAt   *time.Time `json:"expiresAt"`
        
        // 状态
        Enabled     bool       `json:"enabled" gorm:"default:true"`
        
        // 统计
        BanCount    int        `json:"banCount"` // 被封禁次数
        AttackCount int        `json:"attackCount"` // 攻击次数
}

func (IPBlacklist) TableName() string {
        return "ip_blacklist"
}

// IPWhitelist IP白名单
type IPWhitelist struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`
        
        IP          string     `json:"ip" gorm:"type:varchar(45);uniqueIndex"`
        CIDR        string     `json:"cidr" gorm:"type:varchar(64)"`
        Description string     `json:"description" gorm:"type:varchar(255)"`
        
        AddedBy     uint       `json:"addedBy"`
        Enabled     bool       `json:"enabled" gorm:"default:true"`
}

func (IPWhitelist) TableName() string {
        return "ip_whitelist"
}

// LoginRecord 登录记录
type LoginRecord struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        
        ServerID    uint       `json:"serverId" gorm:"index"`
        ServerName  string     `json:"serverName" gorm:"type:varchar(64)"`
        
        User        string     `json:"user" gorm:"type:varchar(64);index"`
        IP          string     `json:"ip" gorm:"type:varchar(45);index"`
        Port        int        `json:"port"`
        
        // 登录结果
        Success     bool       `json:"success"`
        Method      string     `json:"method" gorm:"type:varchar(32)"` // ssh, web, api
        
        // 地理位置
        GeoLocation string     `json:"geoLocation" gorm:"type:varchar(128)"`
        
        // 异常标记
        IsAbnormal  bool       `json:"isAbnormal"`
        AbnormalReason string  `json:"abnormalReason" gorm:"type:varchar(255)"`
}

func (LoginRecord) TableName() string {
        return "login_records"
}

// SecurityRule 安全规则
type SecurityRule struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`
        
        Name        string            `json:"name" gorm:"type:varchar(64)"`
        Type        SecurityEventType `json:"type" gorm:"type:varchar(32)"`
        Enabled     bool              `json:"enabled" gorm:"default:true"`
        
        // 触发条件
        Threshold   int               `json:"threshold"` // 阈值
        TimeWindow  int               `json:"timeWindow"` // 时间窗口(秒)
        
        // 动作
        Action      string            `json:"action" gorm:"type:varchar(32)"` // ban, alert, log
        BanDuration int               `json:"banDuration"` // 封禁时长(秒)，0为永久
        
        // 级别
        Level       GuardSecurityLevel     `json:"level" gorm:"type:varchar(16)"`
}

func (SecurityRule) TableName() string {
        return "security_rules"
}

// SecurityGuard 安全防护
type SecurityGuard struct {
        rules []SecurityRule
}

// NewSecurityGuard 创建安全防护
func NewSecurityGuard() *SecurityGuard {
        return &SecurityGuard{
                rules: GetDefaultSecurityRules(),
        }
}

// GetDefaultSecurityRules 获取默认安全规则
func GetDefaultSecurityRules() []SecurityRule {
        return []SecurityRule{
                // SSH暴力破解防护
                {
                        Name:        "SSH暴力破解防护",
                        Type:        SecurityEventBruteForce,
                        Enabled:     true,
                        Threshold:   5,      // 5次失败
                        TimeWindow:  300,    // 5分钟内
                        Action:      "ban",
                        BanDuration: 3600,   // 封禁1小时
                        Level:       GuardSecurityLevelHigh,
                },
                // 异常登录检测
                {
                        Name:        "异常IP登录检测",
                        Type:        SecurityEventLoginAbnormal,
                        Enabled:     true,
                        Threshold:   1,
                        TimeWindow:  0,
                        Action:      "alert",
                        Level:       GuardSecurityLevelMedium,
                },
                // 端口扫描检测
                {
                        Name:        "端口扫描检测",
                        Type:        SecurityEventPortScan,
                        Enabled:     true,
                        Threshold:   10,     // 扫描10个端口
                        TimeWindow:  60,     // 1分钟内
                        Action:      "ban",
                        BanDuration: 7200,   // 封禁2小时
                        Level:       GuardSecurityLevelHigh,
                },
                // DDoS检测
                {
                        Name:        "DDoS攻击检测",
                        Type:        SecurityEventDDoS,
                        Enabled:     true,
                        Threshold:   1000,   // 1000个连接
                        TimeWindow:  10,     // 10秒内
                        Action:      "ban",
                        BanDuration: 0,      // 永久封禁
                        Level:       GuardSecurityLevelCritical,
                },
                // 恶意IP检测
                {
                        Name:        "恶意IP检测",
                        Type:        SecurityEventMaliciousIP,
                        Enabled:     true,
                        Threshold:   1,
                        TimeWindow:  0,
                        Action:      "ban",
                        BanDuration: 0,
                        Level:       GuardSecurityLevelCritical,
                },
        }
}

// AnalyzeLogin 分析登录
func (g *SecurityGuard) AnalyzeLogin(record *LoginRecord) (*SecurityEvent, error) {
        // 检查白名单
        if g.IsWhitelisted(record.IP) {
                return nil, nil
        }

        // 检查是否异常
        var abnormalReasons []string

        // 检查黑名单
        if g.IsBlacklisted(record.IP) {
                abnormalReasons = append(abnormalReasons, "IP在黑名单中")
        }

        // 检查失败登录次数
        failCount := g.GetFailedLoginCount(record.IP, 300)
        if failCount >= 5 {
                abnormalReasons = append(abnormalReasons, fmt.Sprintf("失败登录%d次", failCount))
        }

        // 检查异常地点（这里简化处理）
        // if !g.IsNormalLocation(record.IP, record.GeoLocation) {
        //     abnormalReasons = append(abnormalReasons, "异常登录地点")
        // }

        // 检查异常时间
        if g.IsAbnormalTime(record.CreatedAt) {
                abnormalReasons = append(abnormalReasons, "异常登录时间")
        }

        // 检查新IP
        if g.IsNewIP(record.IP, record.User) {
                abnormalReasons = append(abnormalReasons, "新IP登录")
        }

        if len(abnormalReasons) > 0 {
                record.IsAbnormal = true
                record.AbnormalReason = strings.Join(abnormalReasons, "; ")
                global.DB.Save(record)

                // 创建安全事件
                event := &SecurityEvent{
                        ServerID:    record.ServerID,
                        ServerName:  record.ServerName,
                        EventType:   SecurityEventLoginAbnormal,
                        Level:       GuardSecurityLevelMedium,
                        SourceIP:    record.IP,
                        SourcePort:  record.Port,
                        TargetUser:  record.User,
                        Description: fmt.Sprintf("检测到异常登录: %s", record.AbnormalReason),
                        Status:      "new",
                }
                global.DB.Create(event)

                return event, nil
        }

        return nil, nil
}

// DetectBruteForce 检测暴力破解
func (g *SecurityGuard) DetectBruteForce(ip string, timeWindow int) (*SecurityEvent, error) {
        // 检查白名单
        if g.IsWhitelisted(ip) {
                return nil, nil
        }

        // 统计失败次数
        failCount := g.GetFailedLoginCount(ip, timeWindow)

        // 查找暴力破解规则
        for _, rule := range g.rules {
                if rule.Type == SecurityEventBruteForce && rule.Enabled {
                        if failCount >= rule.Threshold {
                                // 创建安全事件
                                event := &SecurityEvent{
                                        EventType:   SecurityEventBruteForce,
                                        Level:       GuardSecurityLevel(rule.Level),
                                        SourceIP:    ip,
                                        Description: fmt.Sprintf("检测到SSH暴力破解，%d秒内失败%d次", timeWindow, failCount),
                                        Status:      "new",
                                }
                                global.DB.Create(event)

                                // 执行封禁
                                if rule.Action == "ban" {
                                        g.BanIP(ip, rule.BanDuration, "SSH暴力破解", SecurityEventBruteForce, event.ID)
                                }

                                return event, nil
                        }
                }
        }

        return nil, nil
}

// BanIP 封禁IP
func (g *SecurityGuard) BanIP(ip string, duration int, reason string, eventType SecurityEventType, eventID uint) error {
        // 检查白名单
        if g.IsWhitelisted(ip) {
                return fmt.Errorf("IP在白名单中，无法封禁")
        }

        // 检查是否已封禁
        var existing IPBlacklist
        result := global.DB.Where("ip = ? AND enabled = ?", ip, true).First(&existing)
        if result.Error == nil {
                // 已存在，增加封禁次数
                existing.BanCount++
                global.DB.Save(&existing)
                return nil
        }

        // 创建黑名单记录
        blacklist := &IPBlacklist{
                IP:         ip,
                Reason:     reason,
                EventType:  eventType,
                EventID:    eventID,
                AutoBanned: true,
                Enabled:    true,
        }

        if duration > 0 {
                expiresAt := time.Now().Add(time.Duration(duration) * time.Second)
                blacklist.ExpiresAt = &expiresAt
                blacklist.Permanent = false
        } else {
                blacklist.Permanent = true
        }

        global.DB.Create(blacklist)

        // 执行iptables封禁（这里只记录，实际执行需要SSH）
        // iptables -A INPUT -s ip -j DROP

        return nil
}

// UnbanIP 解封IP
func (g *SecurityGuard) UnbanIP(ip string, operatorID uint) error {
        result := global.DB.Model(&IPBlacklist{}).
                Where("ip = ?", ip).
                Update("enabled", false)

        if result.RowsAffected == 0 {
                return fmt.Errorf("IP不在黑名单中")
        }

        // 执行iptables解封
        // iptables -D INPUT -s ip -j DROP

        return nil
}

// ManualBan 手动封禁
func (g *SecurityGuard) ManualBan(ip string, duration int, reason string, operatorID uint) error {
        // 检查白名单
        if g.IsWhitelisted(ip) {
                return fmt.Errorf("IP在白名单中，无法封禁")
        }

        blacklist := &IPBlacklist{
                IP:         ip,
                Reason:     reason,
                AutoBanned: false,
                BannedBy:   operatorID,
                Enabled:    true,
        }

        if duration > 0 {
                expiresAt := time.Now().Add(time.Duration(duration) * time.Second)
                blacklist.ExpiresAt = &expiresAt
        } else {
                blacklist.Permanent = true
        }

        return global.DB.Create(blacklist).Error
}

// AddToWhitelist 添加到白名单
func (g *SecurityGuard) AddToWhitelist(ip, description string, operatorID uint) error {
        whitelist := &IPWhitelist{
                IP:          ip,
                Description: description,
                AddedBy:     operatorID,
                Enabled:     true,
        }
        return global.DB.Create(whitelist).Error
}

// RemoveFromWhitelist 从白名单移除
func (g *SecurityGuard) RemoveFromWhitelist(ip string) error {
        return global.DB.Where("ip = ?", ip).Delete(&IPWhitelist{}).Error
}

// IsBlacklisted 检查是否在黑名单
func (g *SecurityGuard) IsBlacklisted(ip string) bool {
        var count int64
        global.DB.Model(&IPBlacklist{}).
                Where("ip = ? AND enabled = ?", ip, true).
                Where("permanent = ? OR expires_at > ?", true, time.Now()).
                Count(&count)
        return count > 0
}

// IsWhitelisted 检查是否在白名单
func (g *SecurityGuard) IsWhitelisted(ip string) bool {
        var whitelist IPWhitelist
        result := global.DB.Where("ip = ? AND enabled = ?", ip, true).First(&whitelist)
        if result.Error == nil {
                return true
        }

        // 检查CIDR
        var whitelists []IPWhitelist
        global.DB.Where("enabled = ?", true).Find(&whitelists)
        for _, w := range whitelists {
                if w.CIDR != "" {
                        _, network, _ := net.ParseCIDR(w.CIDR)
                        if network != nil && network.Contains(net.ParseIP(ip)) {
                                return true
                        }
                }
        }

        return false
}

// GetFailedLoginCount 获取失败登录次数
func (g *SecurityGuard) GetFailedLoginCount(ip string, timeWindow int) int {
        var count int64
        since := time.Now().Add(-time.Duration(timeWindow) * time.Second)
        global.DB.Model(&LoginRecord{}).
                Where("ip = ? AND success = ? AND created_at > ?", ip, false, since).
                Count(&count)
        return int(count)
}

// IsAbnormalTime 检查是否异常时间
func (g *SecurityGuard) IsAbnormalTime(t time.Time) bool {
        hour := t.Hour()
        // 简单判断：凌晨2-6点为异常时间
        return hour >= 2 && hour < 6
}

// IsNewIP 检查是否新IP
func (g *SecurityGuard) IsNewIP(ip, user string) bool {
        var count int64
        global.DB.Model(&LoginRecord{}).
                Where("ip = ? AND user = ? AND success = ?", ip, user, true).
                Count(&count)
        return count == 0
}

// DetectPortScan 检测端口扫描
func (g *SecurityGuard) DetectPortScan(ip string, ports []int, timeWindow int) (*SecurityEvent, error) {
        if g.IsWhitelisted(ip) {
                return nil, nil
        }

        for _, rule := range g.rules {
                if rule.Type == SecurityEventPortScan && rule.Enabled {
                        if len(ports) >= rule.Threshold {
                                event := &SecurityEvent{
                                        EventType:   SecurityEventPortScan,
                                        Level:       GuardSecurityLevel(rule.Level),
                                        SourceIP:    ip,
                                        Description: fmt.Sprintf("检测到端口扫描，扫描了%d个端口", len(ports)),
                                        Status:      "new",
                                }
                                global.DB.Create(event)

                                if rule.Action == "ban" {
                                        g.BanIP(ip, rule.BanDuration, "端口扫描", SecurityEventPortScan, event.ID)
                                }

                                return event, nil
                        }
                }
        }

        return nil, nil
}

// ParseSSHDLog 解析SSH日志
func (g *SecurityGuard) ParseSSHDLog(log string) (*LoginRecord, error) {
        record := &LoginRecord{
                Method: "ssh",
        }

        // 解析失败登录
        // Failed password for user from ip port port
        failedPattern := regexp.MustCompile(`Failed password for (\w+) from ([\d.]+) port (\d+)`)
        if matches := failedPattern.FindStringSubmatch(log); len(matches) == 4 {
                record.User = matches[1]
                record.IP = matches[2]
                fmt.Sscanf(matches[3], "%d", &record.Port)
                record.Success = false
                return record, nil
        }

        // 解析成功登录
        // Accepted password for user from ip port port
        successPattern := regexp.MustCompile(`Accepted (?:password|publickey) for (\w+) from ([\d.]+) port (\d+)`)
        if matches := successPattern.FindStringSubmatch(log); len(matches) == 4 {
                record.User = matches[1]
                record.IP = matches[2]
                fmt.Sscanf(matches[3], "%d", &record.Port)
                record.Success = true
                return record, nil
        }

        return nil, fmt.Errorf("无法解析SSH日志")
}

// GetSecurityEvents 获取安全事件
func (g *SecurityGuard) GetSecurityEvents(params map[string]interface{}, page, pageSize int) ([]SecurityEvent, int64, error) {
        var events []SecurityEvent
        var total int64

        query := global.DB.Model(&SecurityEvent{})

        if ip, ok := params["ip"].(string); ok && ip != "" {
                query = query.Where("source_ip = ?", ip)
        }
        if eventType, ok := params["type"].(string); ok && eventType != "" {
                query = query.Where("event_type = ?", eventType)
        }
        if level, ok := params["level"].(string); ok && level != "" {
                query = query.Where("level = ?", level)
        }
        if status, ok := params["status"].(string); ok && status != "" {
                query = query.Where("status = ?", status)
        }

        query.Count(&total)

        offset := (page - 1) * pageSize
        err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error

        return events, total, err
}

// GetBlacklist 获取黑名单
func (g *SecurityGuard) GetBlacklist(page, pageSize int) ([]IPBlacklist, int64, error) {
        var list []IPBlacklist
        var total int64

        query := global.DB.Model(&IPBlacklist{}).Where("enabled = ?", true)
        query.Count(&total)

        offset := (page - 1) * pageSize
        err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error

        return list, total, err
}

// GetWhitelist 获取白名单
func (g *SecurityGuard) GetWhitelist() ([]IPWhitelist, error) {
        var list []IPWhitelist
        err := global.DB.Where("enabled = ?", true).Find(&list).Error
        return list, err
}

// CleanExpiredBans 清理过期封禁
func (g *SecurityGuard) CleanExpiredBans() error {
        return global.DB.Model(&IPBlacklist{}).
                Where("permanent = ? AND expires_at < ?", false, time.Now()).
                Update("enabled", false).Error
}

// GetSecurityStats 获取安全统计
func (g *SecurityGuard) GetSecurityStats(days int) map[string]int64 {
        stats := make(map[string]int64)
        since := time.Now().AddDate(0, 0, -days)

        // 事件统计
        var totalEvents, criticalEvents, highEvents int64
        global.DB.Model(&SecurityEvent{}).Where("created_at > ?", since).Count(&totalEvents)
        global.DB.Model(&SecurityEvent{}).Where("created_at > ? AND level = ?", since, GuardSecurityLevelCritical).Count(&criticalEvents)
        global.DB.Model(&SecurityEvent{}).Where("created_at > ? AND level = ?", since, GuardSecurityLevelHigh).Count(&highEvents)
        stats["totalEvents"] = totalEvents
        stats["criticalEvents"] = criticalEvents
        stats["highEvents"] = highEvents

        // 登录统计
        var totalLogins, successLogins, failedLogins, abnormalLogins int64
        global.DB.Model(&LoginRecord{}).Where("created_at > ?", since).Count(&totalLogins)
        global.DB.Model(&LoginRecord{}).Where("created_at > ? AND success = ?", since, true).Count(&successLogins)
        global.DB.Model(&LoginRecord{}).Where("created_at > ? AND success = ?", since, false).Count(&failedLogins)
        global.DB.Model(&LoginRecord{}).Where("created_at > ? AND is_abnormal = ?", since, true).Count(&abnormalLogins)
        stats["totalLogins"] = totalLogins
        stats["successLogins"] = successLogins
        stats["failedLogins"] = failedLogins
        stats["abnormalLogins"] = abnormalLogins

        // 黑名单统计
        var blacklistedIPs, autoBannedIPs int64
        global.DB.Model(&IPBlacklist{}).Where("enabled = ?", true).Count(&blacklistedIPs)
        global.DB.Model(&IPBlacklist{}).Where("enabled = ? AND auto_banned = ?", true, true).Count(&autoBannedIPs)
        stats["blacklistedIPs"] = blacklistedIPs
        stats["autoBannedIPs"] = autoBannedIPs

        return stats
}

// ExportBlacklist 导出黑名单
func (g *SecurityGuard) ExportBlacklist() ([]byte, error) {
        var list []IPBlacklist
        global.DB.Where("enabled = ?", true).Find(&list)
        return json.Marshal(list)
}

// ImportBlacklist 导入黑名单
func (g *SecurityGuard) ImportBlacklist(data []byte) error {
        var list []IPBlacklist
        if err := json.Unmarshal(data, &list); err != nil {
                return err
        }

        for _, item := range list {
                global.DB.FirstOrCreate(&item, IPBlacklist{IP: item.IP})
        }

        return nil
}
