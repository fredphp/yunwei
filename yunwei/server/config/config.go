package config

import (
	"github.com/spf13/viper"
)

var CONFIG Server

type Server struct {
	System    System
	Mysql     Mysql
	Redis     Redis
	JWT       JWT
	AI        AI
	Security  Security
	Notify    Notify
	Patrol    Patrol
	SelfHeal  SelfHeal
	Prediction Prediction
	Scheduler Scheduler
}

type System struct {
	Port     string `mapstructure:"port"`
	GrpcPort string `mapstructure:"grpc-port"`
	Env      string `mapstructure:"env"`
	Name     string `mapstructure:"name"`
}

type Mysql struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxIdleConns int    `mapstructure:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns"`
}

type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWT struct {
	SigningKey  string `mapstructure:"signing-key"`
	ExpiresTime string `mapstructure:"expires-time"`
	Issuer      string `mapstructure:"issuer"`
}

type AI struct {
	Enabled     bool    `mapstructure:"enabled"`
	APIKey      string  `mapstructure:"api-key"`
	BaseURL     string  `mapstructure:"base-url"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max-tokens"`
	Temperature float64 `mapstructure:"temperature"`
	AutoExecute bool    `mapstructure:"auto-execute"`
}

type Security struct {
	EnableWhitelist    bool `mapstructure:"enable-whitelist"`
	EnableBlacklist    bool `mapstructure:"enable-blacklist"`
	RequireApproval    bool `mapstructure:"require-approval"`
	AuditEnabled       bool `mapstructure:"audit-enabled"`
	AuditRetentionDays int  `mapstructure:"audit-retention-days"`
}

type Notify struct {
	Telegram  TelegramNotify
	WeChat    WeChatNotify
	DingTalk  DingTalkNotify
}

type TelegramNotify struct {
	Enabled bool   `mapstructure:"enabled"`
	Token   string `mapstructure:"token"`
	ChatID  string `mapstructure:"chat-id"`
}

type WeChatNotify struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

type DingTalkNotify struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

type Patrol struct {
	DailyCron  string `mapstructure:"daily-cron"`
	HourlyCron string `mapstructure:"hourly-cron"`
	AutoReport bool   `mapstructure:"auto-report"`
}

type SelfHeal struct {
	Enabled       bool `mapstructure:"enabled"`
	CheckInterval int  `mapstructure:"check-interval"`
	MaxRetry      int  `mapstructure:"max-retry"`
	Cooldown      int  `mapstructure:"cooldown"`
}

type Prediction struct {
	Enabled  bool `mapstructure:"enabled"`
	Interval int  `mapstructure:"interval"`
	HistoryDays int `mapstructure:"history-days"`
}

type Scheduler struct {
	Enabled       bool `mapstructure:"enabled"`
	MaxConcurrent int  `mapstructure:"max-concurrent"`
}

func Init() {
	v := viper.New()
	v.SetConfigFile("config/config.yaml")

	if err := v.ReadInConfig(); err != nil {
		// 使用默认配置
		CONFIG = Server{
			System: System{
				Port:     "8080",
				GrpcPort: "50051",
				Env:      "develop",
				Name:     "yunwei",
			},
			Mysql: Mysql{
				Host:     "127.0.0.1",
				Port:     3306,
				Username: "root",
				Password: "123456",
				Database: "yunwei",
			},
			JWT: JWT{
				SigningKey: "yunwei-secret-key-2024",
			},
			AI: AI{
				Enabled:     true,
				BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
				Model:       "glm-4",
				MaxTokens:   4096,
				Temperature: 0.7,
				AutoExecute: false,
			},
			Security: Security{
				EnableWhitelist:    true,
				EnableBlacklist:    true,
				RequireApproval:    true,
				AuditEnabled:       true,
				AuditRetentionDays: 90,
			},
			Notify: Notify{
				Telegram:  TelegramNotify{Enabled: false},
				WeChat:    WeChatNotify{Enabled: false},
				DingTalk:  DingTalkNotify{Enabled: false},
			},
			Patrol: Patrol{
				DailyCron:  "0 0 8 * * *",
				HourlyCron: "0 0 * * * *",
				AutoReport: true,
			},
			SelfHeal: SelfHeal{
				Enabled:       true,
				CheckInterval: 30,
				MaxRetry:      3,
				Cooldown:      60,
			},
			Prediction: Prediction{
				Enabled:     true,
				Interval:    6,
				HistoryDays: 30,
			},
			Scheduler: Scheduler{
				Enabled:       true,
				MaxConcurrent: 10,
			},
		}
		return
	}

	v.Unmarshal(&CONFIG)
}
