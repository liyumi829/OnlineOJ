package worker

import (
	"context"
	"online-oj/gateway/internal/model/entity"
)

// ProblemDataProvider 提供题目判题所需的数据
//
// 说明：
//  1. Worker 只依赖接口，避免直接耦合具体题目仓储实现
//  2. 这里返回 JudgeRequest 所需的限制和测试点
type ProblemDataProvider interface {
	// GetAllTestCases 根据 problemID 获取测试用例
	GetAllTestCases(ctx context.Context, problemID uint64) ([]entity.ProblemTestCase, error)
	// GetProblemByID 通过 problemID 获取题目详细信息(用于cpu和limit限制)
	GetProblemByID(ctx context.Context, problemID uint64) (*entity.Problem, error)
	// GetTestCodeByLang 通过 problemID 和 语言获取题目的测试代码，前导代码
	GetTestCodeByLang(ctx context.Context, problemID uint64, lang string) (*entity.ProblemLanguageTemplate, error)
}
