package entity

import "time"

// Problem 对应 problem 表
type Problem struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Number      string    `gorm:"column:number;uniqueIndex" json:"number"` // 题目编号 LC-9
	Title       string    `gorm:"column:title" json:"title"`               // 标题
	Star        string    `gorm:"column:star" json:"star"`                 // 难度 Easy/Medium/Hard
	CPULimit    int64     `gorm:"column:cpulimit" json:"cpu_limit"`        // 时间限制
	MemLimit    int64     `gorm:"column:memlimit" json:"mem_limit"`        // 内存限制
	Description string    `gorm:"column:description" json:"description"`   // 题目描述
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`

	// 关联
	Templates []ProblemLanguageTemplate `gorm:"foreignKey:ProblemID"`
	TestCases []ProblemTestCase         `gorm:"foreignKey:ProblemID"`
}

// 表名
func (Problem) TableName() string {
	return "problem"
}
