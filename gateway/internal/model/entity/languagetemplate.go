package entity

import "time"

// ProblemLanguageTemplate 对应 problem_language_template 表
type ProblemLanguageTemplate struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:自增主键" json:"id"`
	ProblemID    uint64    `gorm:"column:problem_id;not null;index:idx_problem_language,unique;comment:题目ID" json:"problem_id"`
	Language     string    `gorm:"column:language;type:varchar(32);not null;index:idx_problem_language,unique;comment:编程语言，如 go/cpp" json:"language"`
	PrependCode  string    `gorm:"column:prepend_code;type:mediumtext;not null;comment:隐藏的头部代码，如 Go package/import 或 C++ include" json:"prepend_code"`
	TemplateCode string    `gorm:"column:template_code;type:mediumtext;not null;comment:用户可见的预设代码" json:"template_code"`
	TestCode     string    `gorm:"column:test_code;type:mediumtext;not null;comment:评测入口代码" json:"test_code"`
	Enabled      bool      `gorm:"column:enabled;type:tinyint(1);not null;default:1;comment:是否启用" json:"enabled"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (ProblemLanguageTemplate) TableName() string {
	return "problem_language_template"
}
