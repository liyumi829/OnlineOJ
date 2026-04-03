package repository

import "online-oj/gateway/internal/model/entity"

// 主要实现接口：
// 1. 获取一个题目的样例测试
// 2. 获取一个题目的所有测试用例

// GetSampleCases 获取样例
// 参数:
//
//	problemID 问题的ID
//
// 返回值:用户可见的测试用例
func (r *Repository) GetSampleCases(problemID uint64) ([]ProblemSimpleVO, error) {
	var cases []ProblemSimpleVO
	err := r.db.
		Model(&entity.ProblemTestCase{}).
		Select("input", "output").
		Where("problem_id = ? AND is_sample = ?", problemID, true).
		Order("sort_order").Find(&cases).Error
	return cases, err
}

// GetAllCases 获取所有测试用例
// 参数:
//
//	problemID 问题的ID
//
// 返回值:一个测试有序的所有的测试用例
func (r *Repository) GetAllCases(problemID uint64) ([]ProblemSimpleVO, error) {
	var cases []ProblemSimpleVO
	err := r.db.
		Model(&entity.ProblemTestCase{}).
		Select("input", "output").
		Where("problem_id = ?", problemID).
		Order("sort_order").Find(&cases).Error
	return cases, err
}
