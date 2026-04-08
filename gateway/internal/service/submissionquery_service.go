package service

import (
	"context"
	"encoding/json"
	"errors"
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/model/entity"
	"online-oj/gateway/internal/repository"

	"go.uber.org/zap"
)

const (
	nextPollAfterMS1 = 200
	nextPollAfterMS2 = 400
	nextPollAfterMS3 = 800
)

type SubmissionQueryService struct {
	sta *SubmissionTaskAggregate // 操作句柄
}

func NewSubmissionQueryService(repo *repository.Repository) *SubmissionQueryService {
	return &SubmissionQueryService{
		sta: NewSubmissionTaskAggregate(repo),
	}
}

// 返回前端询问的请求
func (sqs *SubmissionQueryService) SubmitQuery(ctx context.Context, req *dto.SubmitQueryRequest) (*dto.SubmitQueryResponse, error) {
	if req == nil {
		zap.L().Info("Request is empty")
		return nil, errors.New("Request is empty")
	}
	if req.SubmissionID == "" {
		return nil, errors.New("submissionID is empty")
	}
	// 1. 查看数据库中是否存在对应的submission
	ok, err := sqs.sta.submissionRepo.ExistsBySubmissionID(ctx, req.SubmissionID)
	if err != nil {
		zap.L().Error("ExistsBySubmissionID fail", zap.String("error", err.Error()))
		return nil, err
	}
	// 查找到结果
	if !ok {
		// 不存在
		zap.L().Error("Unknown submission ID", zap.String("Submission ID", req.SubmissionID))
		return &dto.SubmitQueryResponse{
			SubmissionID: req.SubmissionID,
			Done:         false,
			Polling:      false,
		}, errors.New("Unknown submission ID")
	}
	// 2. 查看数据库中对应的submission
	submission, err := sqs.sta.submissionRepo.GetBySubmissionID(ctx, req.SubmissionID)
	if err != nil {
		zap.L().Error("GetBySubmissionID fail", zap.String("error", err.Error()))
		return &dto.SubmitQueryResponse{
			SubmissionID: req.SubmissionID,
			Done:         false,
			Polling:      false,
		}, err
	}
	// 3. 查找成功构造字段
	resp := &dto.SubmitQueryResponse{
		SubmissionID:    submission.SubmissionID,
		Status:          string(submission.Status),
		RunTimeMS:       submission.RuntimeMS,
		MemoryKB:        submission.MemoryKB,
		Stdout:          safePtrString(submission.Stdout),
		Stderr:          safePtrString(submission.Stderr),
		ErrorMsg:        safePtrString(submission.ErrorMessage),
		Cases:           parseCaseResults(submission.ResultJSON),
		NextPollAfterMS: nextPollAfterMS1, // 下一次轮询时间
	}
	fillQueryProgress(resp, submission.Status) // 填写轮询字段

	return resp, nil
}

// fillQueryProgress 根据 submission 状态补充前端轮询控制字段
func fillQueryProgress(resp *dto.SubmitQueryResponse, status entity.SubmissionStatus) {
	switch status {
	case entity.SubmissionStatusPending:
		resp.Phase = "QUEUED"
		resp.Done = false
		resp.Polling = true
		resp.NextPollAfterMS = nextPollAfterMS1

	case entity.SubmissionStatusRunning:
		resp.Phase = "JUDGING"
		resp.Done = false
		resp.Polling = true
		resp.NextPollAfterMS = nextPollAfterMS1

	case entity.SubmissionStatusAccepted,
		entity.SubmissionStatusWrongAnswer,
		entity.SubmissionStatusCompileError,
		entity.SubmissionStatusRuntimeError,
		entity.SubmissionStatusTimeLimitExceeded,
		entity.SubmissionStatusMemoryLimitExceeded:
		resp.Phase = "FINISHED"
		resp.Done = true
		resp.Polling = false
		resp.NextPollAfterMS = 0

	case entity.SubmissionStatusSystemError:
		resp.Phase = "FAILED"
		resp.Done = true
		resp.Polling = false
		resp.NextPollAfterMS = 0

	default:
		resp.Phase = "UNKNOWN"
		resp.Done = false
		resp.Polling = true
		resp.NextPollAfterMS = nextPollAfterMS3
	}
}

// safePtrString 安全转换字符串指针
func safePtrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// parseCaseResults 解析数据库中的 result_json
func parseCaseResults(raw []byte) []dto.TestCaseResult {
	if len(raw) == 0 {
		return nil
	}

	var wrapper struct {
		Cases []dto.TestCaseResult `json:"cases"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && len(wrapper.Cases) > 0 {
		return wrapper.Cases
	}

	// 兼容两种 JSON 格式
	var cases []dto.TestCaseResult
	if err := json.Unmarshal(raw, &cases); err == nil {
		return cases
	}

	return nil
}
