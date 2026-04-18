package worker

import (
	"context"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/model/entity"
)

// JudgeInvoker  是业务层使用的判题客户端接口。
// 接口定义在使用方（service）这里，符合 Go 的接口设计习惯。
type JudgeInvoker interface {
	// Judge 发起一次判题请求。
	Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error)
	// Close 关闭连接
	Close() error
}

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

// CacheUpdater 提供 cache 缓存的修改方法
type CacheUpdater interface {
	SetRunning(submissionID string, nextPollAfterMS int64)
	SetAccpted(submissionID string)
	SetFailed(submissionID string, nextPollAfterMS int64)
}
