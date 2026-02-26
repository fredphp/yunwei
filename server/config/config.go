package config

import (
        "github.com/spf13/viper"
)

var CONFIG Server

type Server struct {
        System   System
        Mysql    Mysql
        Redis    Redis
        JWT      JWT
        AI       AI
        Security Security
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
        Enabled     bool   `mapstructure:"enabled"`
        APIKey      string `mapstructure:"api-key"`
        BaseURL     string `mapstructure:"base-url"`
        Model       string `mapstructure:"model"`
        MaxTokens   int    `mapstructure:"max-tokens"`
        Temperature float64 `mapstructure:"temperature"`
        AutoExecute bool   `mapstructure:"auto-execute"` // 是否自动执行低风险命令
}

type Security struct {
        EnableWhitelist   bool `mapstructure:"enable-whitelist"`
        EnableBlacklist   bool `mapstructure:"enable-blacklist"`
        RequireApproval   bool `mapstructure:"require-approval"`   // 高危命令需要审批
        AuditEnabled      bool `mapstructure:"audit-enabled"`      // 启用审计日志
        AuditRetentionDays int  `mapstructure:"audit-retention-days"` // 审计日志保留天数
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
                                SigningKey: "yunwei-secret-key",
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
                }
                return
        }
        
        v.Unmarshal(&CONFIG)
}
