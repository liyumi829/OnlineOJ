package service

import (
	"context"
	"errors"
	"online-oj/gateway/internal/common"
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/repository"

	"go.uber.org/zap"
)

const (
	maxRetry = 0 // 最大重试次数
)

// 提交创建结果状态

const (
	SubmitCreateFailed = "FAIL"    // 失败
	SubmitCreateOK     = "SUCCESS" // 成功
)

// 实现一个提交代码的服务
type SubmitService struct {
	sta *SubmissionTaskAggregate // 聚合事务操作 submission 表和 judge_task 表
}

func NewSubmitService(repo *repository.Repository) *SubmitService {
	return &SubmitService{
		sta: NewSubmissionTaskAggregate(repo),
	}
}

// Submit 生成SubmissionID 返回给调用者
func (sts *SubmitService) Submit(ctx context.Context, req *dto.SubmitRequest) (*dto.SubmitResponse, error) {
	if sts == nil || sts.sta == nil {
		zap.L().Error("submit service not initialized")
		return nil, errors.New("submit service not initialized")
	}
	if req == nil {
		zap.L().Warn("submit request is nil")
		return nil, errors.New("submit request is nil")
	}
	if req.Code == "" {
		zap.L().Warn("submit request code is empty")
		return nil, errors.New("code is required")
	}

	// 1. 构造提交内容
	submissionID := common.Uuid() // 生成一个随机的ID
	taskID := common.Uuid()
	zap.L().Info(
		"create submission and judge task",
		zap.String("submission_id", submissionID),
		zap.String("task_id", taskID),
		zap.Uint64("problem_id", req.ProblemID),
	)
	data := &CreateSubmissionWithTaskParams{
		SubmissionID: submissionID,
		ProblemID:    req.ProblemID,
		Language:     req.Language,
		SourceCode:   req.Code,
		TaskID:       taskID,
		MaxRetry:     maxRetry,
	}
	// zap.L().Debug("", zap.Uint32("MaxRetry", data.MaxRetry))
	// 2. 提交数据
	if err := sts.sta.CreateSubmissionWithTask(ctx, data); err != nil {
		zap.L().Error(
			"failed to create submission and judge task",
			zap.String("submission_id", submissionID),
			zap.String("task_id", taskID),
			zap.Error(err),
		)

		return &dto.SubmitResponse{
			SubmissionID: submissionID,
			Status:       SubmitCreateFailed,
			Message:      "failed to create submission and judge task records",
		}, err
	}
	// 3. 创建记录成功，返回
	zap.L().Info(
		"submission and judge task created successfully",
		zap.String("submission_id", submissionID),
		zap.String("task_id", taskID),
	)
	return &dto.SubmitResponse{
		SubmissionID: submissionID,
		Status:       SubmitCreateOK,
		Message:      "ok",
	}, nil
}
