package global

import (
        "fmt"
        "yunwei/config"

        "go.uber.org/zap"
        "gorm.io/driver/mysql"
        "gorm.io/gorm"
)

var (
        DB     *gorm.DB
        Logger *zap.Logger
)

func InitDB() {
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
                config.CONFIG.Mysql.Username,
                config.CONFIG.Mysql.Password,
                config.CONFIG.Mysql.Host,
                config.CONFIG.Mysql.Port,
                config.CONFIG.Mysql.Database,
        )

        var err error
        DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
        if err != nil {
                panic("数据库连接失败: " + err.Error())
        }
}

func InitLogger() {
        var err error
        Logger, err = zap.NewProduction()
        if err != nil {
                panic("日志初始化失败: " + err.Error())
        }
}
