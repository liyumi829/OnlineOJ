package entity

import "time"

// ProblemLanguageTemplate 对应 problem_language_template 表
type ProblemLanguageTemplate struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProblemID    uint64    `gorm:"column:problem_id;index:idx_problem_language,unique" json:"problem_id"`
	Language     string    `gorm:"column:language;index:idx_problem_language,unique" json:"language"` // go/cpp
	TemplateCode string    `gorm:"column:template_code" json:"template_code"`                         // 预设代码
	TestCode     string    `gorm:"column:test_code" json:"test_code"`                                 // 测试代码
	Enabled      bool      `gorm:"column:enabled" json:"enabled"`                                     // 是否启用
	CreateTime   time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time" json:"update_time"`
}

func (ProblemLanguageTemplate) TableName() string {
	return "problem_language_template"
}
