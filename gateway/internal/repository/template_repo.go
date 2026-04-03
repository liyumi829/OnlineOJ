package repository

import "online-oj/gateway/internal/model/entity"

// 主要实现接口
// 1. 获取某个题目对应语言的模板代码
// 2. 获取某个题目的完整代码

// GetTemplateCode 获取样例代码
// 参数:
//
//	problemID 题目ID
//	codeType 语言类型 "go"/"cpp"
//
// 返回值：获取到的样例代码以及是否启用 错误
func (r *Repository) GetTemplateCode(problemID uint64, codeType string) (TemplateCodeVO, error) {
	var res TemplateCodeVO
	err := r.db.
		Model(&entity.ProblemLanguageTemplate{}).
		Select("template_code", "enabled").
		Where("problem_id = ? AND language = ?", problemID, codeType).
		First(&res).Error
	return res, err
}

// GetTestCode 获取一个题目的测试代码
// 参数:
//
//	problemID 题目ID
//	codeType 语言类型 "go"/"cpp"
//
// 返回值：获取到的测试代码 错误
func (r *Repository) GetTestCode(problemID uint64, codeType string) (string, error) {
	var code string
	err := r.db.
		Model(&entity.ProblemLanguageTemplate{}).
		Select("test_code").
		Where("problem_id = ? AND language = ?", problemID, codeType).
		Scan(&code).Error
	return code, err
}
