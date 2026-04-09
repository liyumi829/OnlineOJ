package repository

import (
	"context"

	"gorm.io/gorm"
)

// Repository 核心：把 DB 包在这里
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建一个数据管理器（只初始化一次DB）
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB 返回底层 DB，必要时可供高级查询使用
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// Transaction 执行事务(主要用于submissions和judge_task的更新)
// 用于保证 submission 和 judge_task 的一致性创建/更新
func (r *Repository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
