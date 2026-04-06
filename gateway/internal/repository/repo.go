package repository

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Repository 核心：把 DB 包在这里
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建一个数据管理器（只初始化一次DB）
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

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
		cfg = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
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
