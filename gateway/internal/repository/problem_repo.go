package repository

import (
	"context"
	"online-oj/gateway/internal/model/entity"
)

// 数据访问层：仅负责数据库原始查询，不处理业务、不组装VO
// 主要实现接口：
// 0. 获取数据库中的所有题目（可扩展，以后添加类型）
// 1. 通过ID获取单个题目完整信息
// 2. 分页获取题目简略列表
// 3. 获取指定题目所有语言模板
// 4. 获取指定题目+语言的测试代码
// 5. 获取指定题目的【样例测试用例】
// 6. 获取指定题目【所有测试用例】（评分用）
// 全是查询所有没有使用事务

type ProblemRepository struct {
	repo *Repository
}

func NewProblemRepositoty(repo *Repository) *ProblemRepository {
	return &ProblemRepository{
		repo: repo,
	}
}

// ===================== 1. 题目相关 =====================

// GetAllProblemCount 获取总共的题目数量
func (pr *ProblemRepository) GetAllProblemCount(ctx context.Context) int64 {
	var count int64
	pr.repo.db.
		WithContext(ctx).
		Model(&entity.Problem{}).
		Count(&count)
	return count
}

// GetProblemByID 通过problemID获取单个题目完整信息
func (pr *ProblemRepository) GetProblemByID(ctx context.Context, problemID uint64) (*entity.Problem, error) {
	var problem entity.Problem
	err := pr.repo.db.
		WithContext(ctx).
		First(&problem, problemID).Error
	return &problem, err
}

// GetProblemSimpleList 分页获取题目简略列表(已完成排序)
// 参数：
//
//	offset  偏移量（从第几条开始取）1 + 偏移量
//	limit   取多少条
//
// 返回值：题目简略列表、错误
func (pr *ProblemRepository) GetProblemSimpleList(ctx context.Context, offset int, limit int) ([]entity.Problem, error) {
	var list []entity.Problem

	err := pr.repo.db.
		WithContext(ctx).
		Model(&entity.Problem{}).
		Select("id", "number", "title", "star"). // 只查需要的字段
		Order("id ASC").                         // 按 ID 正序（保证顺序不乱）
		Offset(offset).                          // 从第几条开始
		Limit(limit).                            // 取多少条
		Find(&list).Error

	return list, err
}

// ===================== 2. 模板代码相关 =====================

// GetTemplateCodesByProblemID 获取指定题目所有语言模板
func (pr *ProblemRepository) GetTemplateCodesByProblemID(ctx context.Context, problemID uint64) ([]entity.ProblemLanguageTemplate, error) {
	var templates []entity.ProblemLanguageTemplate

	err := pr.repo.db.
		WithContext(ctx).
		Where("problem_id = ?", problemID).
		Find(&templates).Error

	return templates, err
}

// GetTestCodeByLang 获取指定题目的语言的预设和测试代码
func (pr *ProblemRepository) GetTestCodeByLang(ctx context.Context, problemID uint64, lang string) (*entity.ProblemLanguageTemplate, error) {
	var codes entity.ProblemLanguageTemplate
	err := pr.repo.db.
		WithContext(ctx).
		Model(&entity.ProblemLanguageTemplate{}).
		Select("prepend_code", "test_code").
		Where("problem_id = ? AND language = ?", problemID, lang).
		First(&codes).Error

	return &codes, err
}

// ===================== 3. 测试用例相关 =====================

// GetSampleCases 获取指定题目的【样例测试用例】（用户可见）
func (pr *ProblemRepository) GetSampleCases(ctx context.Context, problemID uint64) ([]entity.ProblemTestCase, error) {
	var cases []entity.ProblemTestCase

	err := pr.repo.db.
		WithContext(ctx).
		Where("problem_id = ? AND is_sample = ?", problemID, true).
		Order("sort_order ASC").
		Find(&cases).Error

	return cases, err
}

// GetAllTestCases 获取指定题目【所有测试用例】(排序之后的)
func (pr *ProblemRepository) GetAllTestCases(ctx context.Context, problemID uint64) ([]entity.ProblemTestCase, error) {
	var cases []entity.ProblemTestCase

	err := pr.repo.db.
		WithContext(ctx).
		Where("problem_id = ?", problemID).
		Order("sort_order ASC").
		Find(&cases).Error

	return cases, err
}
