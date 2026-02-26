package tenant

import (
	"time"

	"gorm.io/gorm"
)

// Tenant 租户
type Tenant struct {
	ID          string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Slug        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"` // URL友好标识
	Domain      string         `gorm:"type:varchar(255);uniqueIndex" json:"domain"`       // 自定义域名
	Logo        string         `gorm:"type:varchar(500)" json:"logo"`
	Description string         `gorm:"type:text" json:"description"`

	// 租户状态
	Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active, suspended, deleted
	Plan        string         `gorm:"type:varchar(50);default:'free'" json:"plan"`     // free, starter, pro, enterprise
	BillingCycle string        `gorm:"type:varchar(20)" json:"billing_cycle"`           // monthly, yearly

	// 联系信息
	ContactName  string        `gorm:"type:varchar(100)" json:"contact_name"`
	ContactEmail string        `gorm:"type:varchar(255)" json:"contact_email"`
	ContactPhone string        `gorm:"type:varchar(50)" json:"contact_phone"`
	Address      string        `gorm:"type:text" json:"address"`

	// 配置
	Settings     JSON          `gorm:"type:json" json:"settings"`
	Features     JSON          `gorm:"type:json" json:"features"` // 启用的功能列表

	// 时间戳
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Quota        *TenantQuota   `gorm:"foreignKey:TenantID" json:"quota,omitempty"`
	Users        []TenantUser   `gorm:"foreignKey:TenantID" json:"users,omitempty"`
	Roles        []TenantRole   `gorm:"foreignKey:TenantID" json:"roles,omitempty"`
}

// TenantQuota 租户资源配额
type TenantQuota struct {
	ID         string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID   string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"tenant_id"`

	// 用户配额
	MaxUsers         int `gorm:"default:5" json:"max_users"`
	MaxAdmins        int `gorm:"default:2" json:"max_admins"`

	// 资源配额
	MaxResources     int `gorm:"default:100" json:"max_resources"`
	MaxServers       int `gorm:"default:50" json:"max_servers"`
	MaxDatabases     int `gorm:"default:20" json:"max_databases"`

	// 监控配额
	MaxMonitors      int   `gorm:"default:100" json:"max_monitors"`
	MaxAlertRules    int   `gorm:"default:50" json:"max_alert_rules"`
	MetricsRetention int   `gorm:"default:30" json:"metrics_retention"` // 天

	// 成本配额
	MaxCloudAccounts int     `gorm:"default:5" json:"max_cloud_accounts"`
	BudgetLimit      float64 `gorm:"default:0" json:"budget_limit"` // 0表示无限制

	// 存储配额
	MaxStorageGB     int `gorm:"default:100" json:"max_storage_gb"`
	MaxBackupGB      int `gorm:"default:500" json:"max_backup_gb"`

	// API配额
	MaxAPICalls      int `gorm:"default:10000" json:"max_api_calls"`       // 每日
	MaxWebhooks      int `gorm:"default:10" json:"max_webhooks"`

	// 当前使用量
	CurrentUsers     int `gorm:"default:0" json:"current_users"`
	CurrentResources int `gorm:"default:0" json:"current_resources"`
	CurrentStorage   int `gorm:"default:0" json:"current_storage_gb"`
	CurrentAPICalls  int `gorm:"default:0" json:"current_api_calls"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantUser 租户用户关联
type TenantUser struct {
	ID         string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID   string    `gorm:"type:varchar(36);index;not null" json:"tenant_id"`
	UserID     string    `gorm:"type:varchar(36);index;not null" json:"user_id"`

	// 用户信息
	Email      string    `gorm:"type:varchar(255);index" json:"email"`
	Name       string    `gorm:"type:varchar(100)" json:"name"`
	Avatar     string    `gorm:"type:varchar(500)" json:"avatar"`

	// 角色与权限
	RoleID     string    `gorm:"type:varchar(36);index" json:"role_id"`
	RoleName   string    `gorm:"type:varchar(50)" json:"role_name"` // 冗余，方便查询
	IsOwner    bool      `gorm:"default:false" json:"is_owner"`
	IsAdmin    bool      `gorm:"default:false" json:"is_admin"`

	// 状态
	Status     string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, pending
	InvitedBy  string    `gorm:"type:varchar(36)" json:"invited_by"`
	JoinedAt   time.Time `json:"joined_at"`

	// 最后活跃
	LastActiveAt time.Time `json:"last_active_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Tenant     *Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Role       *TenantRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// TenantRole 租户角色
type TenantRole struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID    string    `gorm:"type:varchar(36);index;not null" json:"tenant_id"`

	// 角色信息
	Name        string    `gorm:"type:varchar(50);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(50);not null" json:"slug"` // owner, admin, operator, viewer
	Description string    `gorm:"type:text" json:"description"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"` // 系统预设角色不可删除

	// 权限列表
	Permissions JSON      `gorm:"type:json" json:"permissions"`

	// 范围控制
	Scope       string    `gorm:"type:varchar(20);default:'tenant'" json:"scope"` // tenant, department, project

	// 继承
	ParentID    string    `gorm:"type:varchar(36)" json:"parent_id"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Users       []TenantUser `gorm:"foreignKey:RoleID" json:"users,omitempty"`
}

// TenantInvitation 租户邀请
type TenantInvitation struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID   string     `gorm:"type:varchar(36);index;not null" json:"tenant_id"`

	// 邀请信息
	Email      string     `gorm:"type:varchar(255);index;not null" json:"email"`
	RoleID     string     `gorm:"type:varchar(36)" json:"role_id"`
	RoleName   string     `gorm:"type:varchar(50)" json:"role_name"`

	// 邀请状态
	Status     string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, accepted, declined, expired
	Token      string     `gorm:"type:varchar(64);uniqueIndex" json:"token"`

	// 邀请人
	InvitedBy  string     `gorm:"type:varchar(36)" json:"invited_by"`
	InviterName string    `gorm:"type:varchar(100)" json:"inviter_name"`

	// 时间
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at"`

	CreatedAt  time.Time  `json:"created_at"`

	// 关联
	Tenant     *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// TenantResourceUsage 租户资源使用记录（用于计费和统计）
type TenantResourceUsage struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID     string    `gorm:"type:varchar(36);index;not null" json:"tenant_id"`

	// 时间维度
	Date         time.Time `gorm:"type:date;index" json:"date"`
	Hour         int       `gorm:"default:-1" json:"hour"` // -1表示天级别聚合

	// 用户指标
	UserCount    int `gorm:"default:0" json:"user_count"`
	ActiveUsers  int `gorm:"default:0" json:"active_users"`

	// 资源指标
	ResourceCount int `gorm:"default:0" json:"resource_count"`
	ServerCount   int `gorm:"default:0" json:"server_count"`
	DatabaseCount int `gorm:"default:0" json:"database_count"`

	// 监控指标
	MonitorCount   int `gorm:"default:0" json:"monitor_count"`
	AlertCount     int `gorm:"default:0" json:"alert_count"`
	MetricsDataMB  int `gorm:"default:0" json:"metrics_data_mb"`

	// 成本指标
	TotalCost      float64 `gorm:"default:0" json:"total_cost"`
	CloudCost      float64 `gorm:"default:0" json:"cloud_cost"`

	// 存储指标
	StorageUsedMB  int `gorm:"default:0" json:"storage_used_mb"`
	BackupUsedMB   int `gorm:"default:0" json:"backup_used_mb"`

	// API指标
	APICalls       int `gorm:"default:0" json:"api_calls"`

	CreatedAt     time.Time `json:"created_at"`
}

// TenantBilling 租户账单
type TenantBilling struct {
	ID              string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID        string     `gorm:"type:varchar(36);index;not null" json:"tenant_id"`

	// 账单周期
	BillingPeriod   string     `gorm:"type:varchar(20);not null" json:"billing_period"` // YYYY-MM
	DueDate         time.Time  `json:"due_date"`

	// 金额明细
	BaseAmount      float64    `gorm:"default:0" json:"base_amount"`      // 基础费用
	UsageAmount     float64    `gorm:"default:0" json:"usage_amount"`     // 用量费用
	OverageAmount   float64    `gorm:"default:0" json:"overage_amount"`   // 超额费用
	DiscountAmount  float64    `gorm:"default:0" json:"discount_amount"`  // 折扣
	TaxAmount       float64    `gorm:"default:0" json:"tax_amount"`       // 税费
	TotalAmount     float64    `gorm:"default:0" json:"total_amount"`     // 总计

	// 使用量明细
	UsageDetails    JSON       `gorm:"type:json" json:"usage_details"`

	// 支付信息
	Status          string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, paid, overdue, cancelled
	PaymentMethod   string     `gorm:"type:varchar(50)" json:"payment_method"`
	PaymentID       string     `gorm:"type:varchar(100)" json:"payment_id"`
	PaidAt          *time.Time `json:"paid_at"`

	// 发票
	InvoiceNumber   string     `gorm:"type:varchar(50)" json:"invoice_number"`
	InvoiceURL      string     `gorm:"type:varchar(500)" json:"invoice_url"`

	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 关联
	Tenant          *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// TenantAuditLog 租户审计日志
type TenantAuditLog struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	TenantID    string    `gorm:"type:varchar(36);index;not null" json:"tenant_id"`

	// 操作信息
	UserID      string    `gorm:"type:varchar(36);index" json:"user_id"`
	UserName    string    `gorm:"type:varchar(100)" json:"user_name"`
	UserEmail   string    `gorm:"type:varchar(255)" json:"user_email"`

	// 操作详情
	Action      string    `gorm:"type:varchar(100);index;not null" json:"action"` // create, update, delete, login, etc.
	Resource    string    `gorm:"type:varchar(100)" json:"resource"`              // server, database, user, etc.
	ResourceID  string    `gorm:"type:varchar(36);index" json:"resource_id"`
	ResourceName string   `gorm:"type:varchar(255)" json:"resource_name"`

	// 变更内容
	OldValue    JSON      `gorm:"type:json" json:"old_value"`
	NewValue    JSON      `gorm:"type:json" json:"new_value"`
	Changes     JSON      `gorm:"type:json" json:"changes"`

	// 请求信息
	IPAddress   string    `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent   string    `gorm:"type:varchar(500)" json:"user_agent"`
	RequestID   string    `gorm:"type:varchar(36)" json:"request_id"`

	// 状态
	Status      string    `gorm:"type:varchar(20);default:'success'" json:"status"` // success, failed
	ErrorMsg    string    `gorm:"type:text" json:"error_msg"`

	CreatedAt   time.Time `gorm:"index" json:"created_at"`

	// 关联
	Tenant      *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// JSON 类型别名
type JSON map[string]interface{}

// TableName 方法
func (Tenant) TableName() string            { return "tenants" }
func (TenantQuota) TableName() string       { return "tenant_quotas" }
func (TenantUser) TableName() string        { return "tenant_users" }
func (TenantRole) TableName() string        { return "tenant_roles" }
func (TenantInvitation) TableName() string  { return "tenant_invitations" }
func (TenantResourceUsage) TableName() string { return "tenant_resource_usage" }
func (TenantBilling) TableName() string     { return "tenant_billings" }
func (TenantAuditLog) TableName() string    { return "tenant_audit_logs" }

// 预设角色权限
var DefaultRoles = map[string]map[string]interface{}{
	"owner": {
		"description": "租户所有者，拥有完全控制权",
		"permissions": []string{"*"},
	},
	"admin": {
		"description": "租户管理员，可管理租户内所有资源",
		"permissions": []string{
			"users:*", "roles:read", "resources:*", "monitors:*",
			"alerts:*", "backups:*", "costs:*", "settings:*",
		},
	},
	"operator": {
		"description": "运维人员，可操作资源但不可管理用户",
		"permissions": []string{
			"resources:*", "monitors:*", "alerts:*", "backups:read,execute",
			"costs:read", "settings:read",
		},
	},
	"viewer": {
		"description": "只读用户，只能查看数据",
		"permissions": []string{
			"resources:read", "monitors:read", "alerts:read", "backups:read",
			"costs:read", "settings:read",
		},
	},
}

// PlanQuotas 套餐配额预设
var PlanQuotas = map[string]TenantQuota{
	"free": {
		MaxUsers: 3, MaxAdmins: 1, MaxResources: 20, MaxServers: 10,
		MaxDatabases: 5, MaxMonitors: 20, MaxAlertRules: 10,
		MetricsRetention: 7, MaxCloudAccounts: 1, BudgetLimit: 1000,
		MaxStorageGB: 10, MaxBackupGB: 20, MaxAPICalls: 1000, MaxWebhooks: 2,
	},
	"starter": {
		MaxUsers: 10, MaxAdmins: 3, MaxResources: 100, MaxServers: 50,
		MaxDatabases: 20, MaxMonitors: 100, MaxAlertRules: 50,
		MetricsRetention: 30, MaxCloudAccounts: 3, BudgetLimit: 10000,
		MaxStorageGB: 50, MaxBackupGB: 200, MaxAPICalls: 10000, MaxWebhooks: 10,
	},
	"pro": {
		MaxUsers: 50, MaxAdmins: 10, MaxResources: 500, MaxServers: 200,
		MaxDatabases: 100, MaxMonitors: 500, MaxAlertRules: 200,
		MetricsRetention: 90, MaxCloudAccounts: 10, BudgetLimit: 100000,
		MaxStorageGB: 200, MaxBackupGB: 1000, MaxAPICalls: 100000, MaxWebhooks: 50,
	},
	"enterprise": {
		MaxUsers: -1, MaxAdmins: -1, MaxResources: -1, MaxServers: -1,
		MaxDatabases: -1, MaxMonitors: -1, MaxAlertRules: -1,
		MetricsRetention: 365, MaxCloudAccounts: -1, BudgetLimit: 0,
		MaxStorageGB: -1, MaxBackupGB: -1, MaxAPICalls: -1, MaxWebhooks: -1,
	},
}
