package repository

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化连接数据库
//
// 参数:
//
//	dsn 连接路径及参数 格式: "%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
//	mode 模式 debug调试/prod运行
//
// 返回值: 权柄
func InitDB(dsn, mode string) *gorm.DB {
	// 1. 打开数据库连接
	var cfg *gorm.Config // 配置信息
	if mode == "debug" {
		cfg = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // 打印 SQL
		}
	} else {
		// 自定义 GORM 日志（输出到 stdout，生产环境用）
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出到标准输出
			logger.Config{
				SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
				LogLevel:                  logger.Warn,            // 只打印 Warn/Error
				IgnoreRecordNotFoundError: true,                   // 忽略记录不存在错误
				Colorful:                  false,                  // 生产环境关闭颜色
			},
		)
		cfg = &gorm.Config{
			Logger: newLogger, // 使用自定义日志器
		}
	}
	db, err := gorm.Open(mysql.Open(dsn), cfg)
	if err != nil {
		zap.L().Fatal("connect database failed", zap.String("error", err.Error()))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("get sql.DB instance failed", zap.String("error", err.Error()))
	}

	// ===================== 连接池核心配置 =====================
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数（根据MySQL配置调整，一般 100~200）
	sqlDB.SetMaxIdleConns(20)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(30 * time.Second) // 连接最大生命周期（避免长连接失效）
	sqlDB.SetConnMaxIdleTime(10 * time.Second) // 连接最大空闲时间
	zap.L().Info("database connected successfully")

	return db
}
