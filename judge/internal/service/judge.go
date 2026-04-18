package service

import (
	"context"
	"fmt"
	"log"
	"online-oj/api/proto/judge"
	"online-oj/judge/internal/cache"
	"online-oj/judge/internal/compile"
	"online-oj/judge/internal/run"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// 全局计数器：真正执行判题的次数（没有命中任何缓存）
var realJudgeCount atomic.Int64

// Judge 带有缓存的判题
func (s *JudgeServer) Judge(ctx context.Context, req *judge.JudgeRequest) (*judge.JudgeResponse, error) {
	// 1. 构造缓存 KEY
	key := cache.BuildReuseKey(cache.ReuseKeyInput{
		ProblemID:       0,                                    // 题目ID
		Language:        getLanguageString(req.GetCodeType()), // 语言
		Code:            req.GetCode(),                        // 代码
		TestcaseVersion: "v1",                                 // 测试用例
		RuntimeVersion:  "go 1.25.5",                          // 编译器版本号
	})

	// 2. 先查缓存
	result, found, err := s.cache.Get(ctx, key)
	if err != nil {
		// 缓存出错，不影响主流程，直接跳过
		zap.L().Warn("cache get failed", zap.Error(err))
	}
	if found && result != nil {
		// 缓存命中！直接返回
		zap.L().Info("judge result cache HIT", zap.String("key", key))
		return toJudgeResponse(result), nil
	}

	// 3. 缓存未命中 → singleflight 防止并发判题
	res, err, _ := s.cache.DoWithSingleflight(key, func() (*cache.CachedJudgeResult, error) {
		// 真正执行判题（调用你原来的函数）
		resp, err := s.realJudge(ctx, req)
		if err != nil {
			return nil, err
		}

		// 转成缓存结构
		cached := toCachedResult(resp)

		// 写入缓存
		if err := s.cache.Set(ctx, key, cached); err != nil {
			zap.L().Error("cache set failed", zap.Error(err))
		}

		return cached, nil
	})

	if err != nil {
		return nil, err
	}

	// 4. 返回结果
	return toJudgeResponse(res), nil
}

// realJudge 执行真正的判题逻辑
func (s *JudgeServer) realJudge(ctx context.Context, req *judge.JudgeRequest) (*judge.JudgeResponse, error) {
	// 日志：判题开始
	realJudgeCount.Add(1) // 统计判题次数
	log.Println("真实判题次数：", realJudgeCount.Load())
	zap.L().Info("judge server start processing judge request",
		zap.Int32("code_type", req.GetCodeType()),
		zap.Int("test_case_count", len(req.GetTestCases())),
	)

	codeType := getCodeType(req.GetCodeType()) // 获取类型
	if codeType == compile.UnKnownType {
		zap.L().Warn("user submit a unknown code type",
			zap.Int32("type_code", req.GetCodeType()),
		)
		return nil, fmt.Errorf("user submit a unknown code type, type code: %d", req.GetCodeType())
	}
	// 日志：代码类型校验通过
	zap.L().Debug("code type check passed")

	code := req.GetCode() // 获取代码
	compiler := &compile.Compiler{
		CodeType: codeType,
		Code:     code,
	}

	// 日志：开始编译
	zap.L().Info("start compiling code")

	cRes, err := compiler.Compile(ctx, s.storagePath) // 编译调用
	if err != nil {
		// 编译内部发生问题，不应该被用户知晓
		zap.L().Error("compilation internal error",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	// 编译结果日志
	if cRes.Status == "CE" {
		zap.L().Warn("code compilation failed (CE)",
			zap.String("stderr", cRes.Stderr),
		)
	} else {
		zap.L().Info("code compilation success",
			zap.String("bin_path", cRes.BinPath),
		)
	}

	rsp := &judge.JudgeResponse{}
	// 发生编译错误
	if cRes.Status == "CE" {
		rsp.Status = "CE"
		rsp.Stderr = cRes.Stderr

		// 日志：返回编译错误结果
		zap.L().Info("judge request finished with compilation error",
			zap.String("status", rsp.Status),
		)
		return rsp, nil
	}

	zap.L().Info("start executing code in sandbox",
		zap.Duration("cpu_limit", time.Duration(req.CpuLimit)),
		zap.Int64("mem_kb_limit", req.MemoLimit),
	)

	runner := &run.Runner{
		Bin:          cRes.BinPath,
		CpuLimit:     time.Duration(req.CpuLimit),
		MemoKiBLimit: req.MemoLimit,
		TestCases:    req.GetTestCases(),
	}

	rRes, err := runner.RunSandboxed(ctx, s.globalTimeout)
	if err != nil {
		zap.L().Error("code execution internal error",
			zap.String("bin_path", cRes.BinPath),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	zap.L().Info("code execution finished",
		zap.String("status", rRes.Status),
		zap.Duration("real_time", rRes.TimeReal),
		zap.Int64("real_mem_kb", rRes.MemoKiBReal),
	)

	// 组装响应
	rsp.Status = rRes.Status
	rsp.Stderr = rRes.Stderr
	rsp.Stdout = summarizeCaseStatus(rRes.CaseRusults)
	rsp.Time = rRes.TimeReal.Nanoseconds()
	rsp.Memory = rRes.MemoKiBReal
	rsp.Results = rRes.CaseRusults

	// 日志：整个判题流程完成
	zap.L().Info("judge request process success",
		zap.String("final_status", rsp.Status),
		zap.Int64("time_ns", rsp.Time),
		zap.Int64("memory_kb", rsp.Memory),
	)

	return rsp, nil
}
