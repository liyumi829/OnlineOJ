package entity

import "time"

// ProblemTestCase 对应 problem_test_case 表
type ProblemTestCase struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProblemID  uint64    `gorm:"column:problem_id" json:"problem_id"`
	Input      string    `gorm:"column:input" json:"input"`           // 输入
	Output     string    `gorm:"column:output" json:"output"`         // 输出
	IsSample   bool      `gorm:"column:is_sample" json:"is_sample"`   // 是否样例
	SortOrder  int       `gorm:"column:sort_order" json:"sort_order"` // 顺序
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
}

func (ProblemTestCase) TableName() string {
	return "problem_test_case"
}
