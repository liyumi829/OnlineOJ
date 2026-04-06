package service

import (
	"context"
	"errors"

	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/common"
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/repository"

	"go.uber.org/zap"
)

const (
	Go  = "go"
	Cpp = "cpp"
)

// JudgeService 业务层服务。
type JudgeService struct {
	judgeClient JudgeClient            // 接口
	repo        *repository.Repository // 获取测试用例和测试代码
}

// NewJudgeService 创建业务服务。
func NewJudgeService(judgeClient JudgeClient, r *repository.Repository) *JudgeService {
	return &JudgeService{
		judgeClient: judgeClient,
		repo:        r,
	}
}

// JudgeCode 执行判题。
func (s *JudgeService) JudgeCode(ctx context.Context, req *dto.SubmitRequest) (*dto.SubmitResponse, error) {
	if req == nil {
		zap.L().Error("judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	zap.L().Info("Start JudgeCode task",
		zap.String("problem_id", req.ID),
		zap.Int("language", int(req.Language)),
		zap.Int("code_length", len(req.Code)),
	)

	id, err := common.StrToUint64(req.ID)
	if err != nil {
		zap.L().Error("parse string to uint64 error", zap.String("id", req.ID), zap.Error(err))
		return nil, err
	}

	// 获取题目信息
	problem, err := s.repo.GetProblemByID(id)
	if err != nil {
		zap.L().Error("GetProblemByID failed", zap.Uint64("id", id), zap.Error(err))
		return nil, err
	}

	// 获取测试用例
	testCases, err := s.repo.GetAllTestCases(id)
	if err != nil {
		zap.L().Error("GetAllTestCases failed", zap.Uint64("id", id), zap.Error(err))
		return nil, err
	}

	// 获取测试代码
	var language string
	switch req.Language {
	case 1:
		language = Go
	case 2:
		language = Cpp
	default:
		return nil, errors.New("unknown type")
	}
	Codes, err := s.repo.GetTestCodeByLang(id, language)
	if err != nil {
		zap.L().Error("GetTestCodeByLang failed", zap.Uint64("id", id), zap.Uint32("language", req.Language), zap.Error(err))
		return nil, err
	}

	zap.L().Info("Successfully retrieved problem data",
		zap.Uint64("id", id),
		zap.Int("test_cases_count", len(testCases)),
		zap.Int("prepend_code", len(Codes.PrependCode)),
		zap.Int("test_code", len(Codes.TestCode)),
		zap.Int64("cpu_limit", problem.CPULimit),
		zap.Int64("mem_limit", problem.MemLimit),
	)

	// 构建rpc可用测试用例
	cases := make([]*pb.TestCase, 0, len(testCases))
	for _, testCase := range testCases {
		cases = append(cases, &pb.TestCase{
			Input:  testCase.Input,
			Output: testCase.Output,
		})
	}

	rpcReq := &pb.JudgeRequest{
		Code:      Codes.PrependCode + "\n\n" + req.Code + "\n\n" + Codes.TestCode,
		CodeType:  int32(req.Language),
		CpuLimit:  problem.CPULimit,
		MemoLimit: problem.MemLimit,
		TestCases: cases,
	}

	zap.L().Info("Calling RPC Judge engine...")

	// 3. 下面进行rpc调用
	rpcRsp, err := s.judgeClient.Judge(ctx, rpcReq)
	if err != nil {
		zap.L().Error("RPC call failure judgment", zap.Error(err))
		return nil, err
	}

	// [INFO 日志：RPC 返回结果监控]
	zap.L().Info("RPC Judge finished",
		zap.String("overall_status", rpcRsp.Status),
		zap.Int("results_count", len(rpcRsp.Results)),
	)

	var totalTime, totalMemory int64
	resCases := make([]dto.TestCaseResult, 0, len(rpcRsp.Results))
	for index, result := range rpcRsp.Results {
		totalTime += result.Time
		totalMemory += result.Memory
		resCases = append(resCases, dto.TestCaseResult{
			Id:     uint64(index + 1),
			Status: result.Status,
		})
	}

	runTimeMs := float64(totalTime) / 1_000_000.0
	memoryMb := float64(totalMemory) / 1024.0
	zap.L().Info("JudgeCode completed successfully",
		zap.Float64("run_time_ms", runTimeMs),
		zap.Float64("memory_mb", memoryMb),
		zap.String("final_status", rpcRsp.Status),
	)

	return &dto.SubmitResponse{
		Status:  rpcRsp.Status,
		Stdout:  rpcRsp.Stdout,
		Stderr:  rpcRsp.Stderr,
		RunTime: runTimeMs,
		Memory:  memoryMb,
		Cases:   resCases,
	}, nil
}
