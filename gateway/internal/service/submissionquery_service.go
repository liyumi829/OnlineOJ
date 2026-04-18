package service

import (
	"context"
	"encoding/json"
	"errors"
	gwcache "online-oj/gateway/internal/cache"
	"online-oj/gateway/internal/common"
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/model/entity"
	"online-oj/gateway/internal/repository"

	"go.uber.org/zap"
)

type SubmissionQueryService struct {
	sta   *SubmissionTaskAggregate // 操作句柄
	cache *gwcache.PollCache       // 缓存
}

func NewSubmissionQueryService(repo *repository.Repository, cache *gwcache.PollCache) *SubmissionQueryService {
	return &SubmissionQueryService{
		sta:   NewSubmissionTaskAggregate(repo),
		cache: cache,
	}
}

// 返回前端询问的请求
//  1. 先查本地缓存
//     - 查到了：(空值缓存……) 未完成则返回轻量级响应；完成则进行数据库查询
//     - 未查到：查询数据库；进行缓存更新；
func (sqs *SubmissionQueryService) SubmitQuery(ctx context.Context, req *dto.SubmitQueryRequest) (*dto.SubmitQueryResponse, error) {
	if req == nil {
		zap.L().Info("Request is empty")
		return nil, errors.New("Request is empty")
	}
	if req.SubmissionID == "" {
		return nil, errors.New("submissionID is empty")
	}
	// 1. 先进行缓存查询
	if pollState, ok := sqs.cache.Get(req.SubmissionID); ok && pollState != nil {
		// 空值缓存 防止缓存击穿 大量不存在的请求打到数据库
		if pollState.NotFound {
			return nil, errors.New("submission not found")
		}

		// 未完成直接返回轻量响应
		if !pollState.Done {
			return &dto.SubmitQueryResponse{
				SubmissionID:    pollState.SubmissionID,
				Phase:           string(pollState.Phase),
				Done:            pollState.Done,
				Polling:         pollState.Polling,
				NextPollAfterMS: pollState.NextPollAfterMS,
			}, nil
		}
		// 已完成则继续查 DB 返回完整结果
	}

	// 2. 查看数据库中是否存在对应的submission
	ok, err := sqs.sta.submissionRepo.ExistsBySubmissionID(ctx, req.SubmissionID)
	if err != nil {
		zap.L().Error("ExistsBySubmissionID fail", zap.String("error", err.Error()))
		return nil, err
	}
	// 查找到结果
	if !ok {
		// 不存在
		zap.L().Error("Unknown submission ID", zap.String("Submission ID", req.SubmissionID))
		// 进行空值缓存，防穿透
		sqs.cache.SetNotFound(req.SubmissionID)
		return &dto.SubmitQueryResponse{
			SubmissionID: req.SubmissionID,
			Done:         false,
			Polling:      false,
		}, errors.New("Unknown submission ID")
	}
	// 3. 查看数据库中对应的submission
	submission, err := sqs.sta.submissionRepo.GetBySubmissionID(ctx, req.SubmissionID)
	if err != nil {
		zap.L().Error("GetBySubmissionID fail", zap.String("error", err.Error()))
		return &dto.SubmitQueryResponse{
			SubmissionID: req.SubmissionID,
			Done:         false,
			Polling:      false,
		}, err
	}
	// 填写轻量级别缓存 --> 下次查询可用直接找缓存，而不是数据库
	respCache := &gwcache.SubmissionPollState{
		SubmissionID: submission.SubmissionID,
	}
	fillQueryCache(respCache, submission.Status)                          // 根据当前状态构造字段
	sqs.cache.Set(respCache, gwcache.PhaseTTL(string(submission.Status))) // 根据状态设置过期时间
	// 4. 查找成功构造字段
	resp := &dto.SubmitQueryResponse{
		SubmissionID:    submission.SubmissionID,
		Status:          string(submission.Status),
		RunTimeMS:       submission.RuntimeMS,
		MemoryKB:        submission.MemoryKB,
		Stdout:          safePtrString(submission.Stdout),
		Stderr:          safePtrString(submission.Stderr),
		ErrorMsg:        safePtrString(submission.ErrorMessage),
		Cases:           parseCaseResults(submission.ResultJSON),
		NextPollAfterMS: common.NextPollAfterMS1, // 下一次轮询时间
	}
	fillQueryProgress(resp, submission.Status) // 填写轮询字段
	return resp, nil
}

// fillQueryCache 根据 submission 状态填写缓存
func fillQueryCache(pollState *gwcache.SubmissionPollState, status entity.SubmissionStatus) {
	switch status {
	case entity.SubmissionStatusPending:
		pollState.Phase = "QUEUED"
		pollState.Done = false
		pollState.Polling = true
		pollState.NextPollAfterMS = common.NextPollAfterMS2

	case entity.SubmissionStatusRunning:
		pollState.Phase = "JUDGING"
		pollState.Done = false
		pollState.Polling = true
		pollState.NextPollAfterMS = common.NextPollAfterMS1

	case entity.SubmissionStatusAccepted,
		entity.SubmissionStatusWrongAnswer,
		entity.SubmissionStatusCompileError,
		entity.SubmissionStatusRuntimeError,
		entity.SubmissionStatusTimeLimitExceeded,
		entity.SubmissionStatusMemoryLimitExceeded:
		pollState.Phase = "FINISHED"
		pollState.Done = true
		pollState.Polling = false
		pollState.NextPollAfterMS = 0

	case entity.SubmissionStatusSystemError:
		pollState.Phase = "FAILED"
		pollState.Done = true
		pollState.Polling = false
		pollState.NextPollAfterMS = 0

	default:
		pollState.Phase = "UNKNOWN"
		pollState.Done = false
		pollState.Polling = true
		pollState.NextPollAfterMS = common.NextPollAfterMS3
	}
}

// fillQueryProgress 根据 submission 状态补充前端轮询控制字段
func fillQueryProgress(resp *dto.SubmitQueryResponse, status entity.SubmissionStatus) {
	switch status {
	case entity.SubmissionStatusPending:
		resp.Phase = "QUEUED"
		resp.Done = false
		resp.Polling = true
		resp.NextPollAfterMS = common.NextPollAfterMS2

	case entity.SubmissionStatusRunning:
		resp.Phase = "JUDGING"
		resp.Done = false
		resp.Polling = true
		resp.NextPollAfterMS = common.NextPollAfterMS1

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
		resp.NextPollAfterMS = common.NextPollAfterMS3
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
