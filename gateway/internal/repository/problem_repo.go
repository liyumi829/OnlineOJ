package repository

import (
	"online-oj/gateway/internal/model/entity"
)

// 主要实现接口：
// 1. 获取单个题目详细信息
// 2. 获取多个题目编号、标题、难度

// GetProblemByID 通过题目编号获取一个题目的详细信息
// 参数:
//
//	id 题目的编号
//
// 返回值: 题目的详细信息
func (r *Repository) GetProblemByID(id uint64) (entity.Problem, error) {
	var problem entity.Problem
	err := r.db.First(&problem, id).Error
	return problem, err
}

// GetQuestionLists 获取题目列表
// 参数:
//
//	可扩展 // 根据分类……
//
// 返回值: 题目的编号，标题、难度
// GetProblemSimpleList 获取题目简略列表（多条）
func (r *Repository) GetProblemSimpleList() ([]ProblemSampleVO, error) {
	var list []ProblemSampleVO

	err := r.db.
		Model(&entity.Problem{}).
		Select("id", "number", "title", "star"). // 只查需要的3个字段
		Find(&list).Error

	return list, err
}
