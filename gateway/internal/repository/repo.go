package repository

import "gorm.io/gorm"

// Repository 核心：把 DB 包在这里
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建一个数据管理器（只初始化一次DB）
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}
