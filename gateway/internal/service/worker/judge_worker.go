package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/common"
	"online-oj/gateway/internal/model/entity"
	"online-oj/gateway/internal/repository"
	"online-oj/gateway/internal/service"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/datatypes"
)

const (
	Go  = "go"
	Cpp = "cpp"
)

// JudgeWorker 负责执行判题分发任务
type JudgeWorker struct {
	workerID        string                           // 当前 Worker 唯一标识
	judgeClient     JudgeInvoker                     // gRPC Judge 客户端
	problemProvider ProblemDataProvider              // 题目数据提供者
	cacheUpdater    CacheUpdater                     // 缓存更新器
	taskChan        <-chan *entity.JudgeTask         // 任务队列
	submissionRepo  *repository.SubmissionRepository // submission 仓储
	taskService     *service.SubmissionTaskAggregate // 组合事务服务
}

// newJudgeWorker 创建 JudgeWorker 由 Manager 创建
// 关键参数说明:
//   - judgeClient 接口。实现 Judge 的方法。主要中文件rpc中，实现了超时
//   - problemProvider 接口。主要提供题目信息和限制条件的接口
//   - CacheUpdater 接口。主要实现更新缓存的 Status 状态
func NewJudgeWorker(
	workerID string,
	judgeClient JudgeInvoker,
	taskChan <-chan *entity.JudgeTask,
	submissionRepo *repository.SubmissionRepository,
	problemProvider ProblemDataProvider,
	taskService *service.SubmissionTaskAggregate,
	cache CacheUpdater,
) *JudgeWorker {

	return &JudgeWorker{
		workerID:        workerID,
		judgeClient:     judgeClient,
		problemProvider: problemProvider,
		taskChan:        taskChan,
		submissionRepo:  submissionRepo,
		taskService:     taskService,
		cacheUpdater:    cache,
	}
}

// WorkerID 返回当前 Worker 标识
func (w *JudgeWorker) WorkerID() string {
	return w.workerID
}

func (w *JudgeWorker) startWork(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-w.taskChan:
			if !ok {
				return
			}
			w.ExecuteTask(ctx, task) // 忽略错误
		}
	}
}

// ExecuteTask 执行单个判题任务
//
// 执行流程：
//  1. 抢占任务并标记 submission 为 RUNNING
//  2. 查询 submission
//  3. 查询题目限制和测试点
//  4. 构造 protobuf JudgeRequest
//  5. 调用 Judge RPC
//  6. 解析响应并回写 submission / judge_task
func (w *JudgeWorker) ExecuteTask(ctx context.Context, task *entity.JudgeTask) error {
	// 确保参数都是正确的
	if task == nil {
		return errors.New("judge task is nil")
	}
	if w == nil {
		return errors.New("judge worker is nil")
	}
	if w.judgeClient == nil {
		return errors.New("judge grpc client is nil")
	}
	if w.submissionRepo == nil {
		return errors.New("submission repository is nil")
	}
	if w.taskService == nil {
		return errors.New("submission task service is nil")
	}
	if w.problemProvider == nil {
		return errors.New("problem provider is nil")
	}

	// 1. 先抢占任务，防止被多个 Worker 同时执行
	claimed, err := w.taskService.TryClaimTaskAndMarkRunning(ctx, task.TaskID, task.SubmissionID, w.workerID)
	if err != nil {
		return err
	}
	if !claimed {
		// 说明任务已被其他 Worker 抢走，直接返回即可
		return nil
	}
	w.cacheUpdater.SetRunning(task.SubmissionID, common.NextPollAfterMS1)
	// 2. 抢占成功查询submission，读取提交数据获取：题目ID、语言类型、源码
	submission, err := w.submissionRepo.GetBySubmissionID(ctx, task.SubmissionID)
	if err != nil {
		_ = w.taskService.FinalizeJudgeFailure(ctx, task.SubmissionID, task.TaskID, fmt.Sprintf("load submission failed: %v", err), false)
		w.cacheUpdater.SetFailed(submission.SubmissionID, common.NextPollAfterMS0)
		return err
	}
	// 3. 读取题目限制、测试用例和测试代码
	// 题目限制
	problem, err := w.problemProvider.GetProblemByID(ctx, submission.ProblemID)
	if err != nil {
		_ = w.taskService.FinalizeJudgeFailure(ctx, task.SubmissionID, task.TaskID, fmt.Sprintf("build judge request data failed: %v", err), false)
		w.cacheUpdater.SetFailed(submission.SubmissionID, common.NextPollAfterMS0)
		return err
	}
	// 测试用例
	testCases, err := w.problemProvider.GetAllTestCases(ctx, submission.ProblemID)
	if err != nil {
		_ = w.taskService.FinalizeJudgeFailure(ctx, task.SubmissionID, task.TaskID, fmt.Sprintf("build judge request data failed: %v", err), false)
		w.cacheUpdater.SetFailed(submission.SubmissionID, common.NextPollAfterMS0)
		return err
	}
	cases := make([]*pb.TestCase, 0, len(testCases))
	for _, testCase := range testCases {
		cases = append(cases, &pb.TestCase{
			Input:  testCase.Input,
			Output: testCase.Output,
		})
	}
	// 获取测试代码
	Codes, err := w.problemProvider.GetTestCodeByLang(ctx, submission.ProblemID, submission.Language)
	if err != nil {
		zap.L().Error("GetTestCodeByLang failed", zap.Uint64("id", submission.ProblemID), zap.String("language", submission.Language), zap.Error(err))
		return err
	}
	// 日志输出
	zap.L().Debug("Successfully retrieved problem data",
		zap.Uint64("problemID", submission.ProblemID),
		zap.Int("test_cases_count", len(testCases)),
		zap.Int("prepend_code", len(Codes.PrependCode)),
		zap.Int("test_code", len(Codes.TestCode)),
		zap.Int64("cpu_limit", problem.CPULimit),
		zap.Int64("mem_limit", problem.MemLimit))

	// 测试代码
	// 4. 构建rpc调用request数据
	// 构造 protobuf JudgeRequest
	rpcReq := &pb.JudgeRequest{
		SubmissionId: submission.SubmissionID,
		Code:         Codes.PrependCode + "\n\n" + submission.SourceCode + "\n\n" + Codes.TestCode,
		CodeType:     getLanguageCode(submission.Language),
		CpuLimit:     problem.CPULimit,
		MemoLimit:    problem.MemLimit,
		TestCases:    cases,
	}
	// 5. 构造完成调用 grpc
	rpcResp, err := w.judgeClient.Judge(ctx, rpcReq)
	if err != nil {
		// 如果超时就尝试重试
		if status.Code(err) == codes.DeadlineExceeded {
			zap.L().Info("gRPC request failed, attempting retry.", zap.Uint32("retry times", task.RetryCount), zap.Uint32("max retry times", task.MaxRetry))
			if task.RetryCount < task.MaxRetry {
				// 进行重试
				err := w.taskService.FinalizeJudgeFailure(ctx, task.SubmissionID, task.TaskID, fmt.Sprintf("judge rpc failed: %v", err), true)
				if err == nil {
					return nil
				}
			}
		}
		_ = w.taskService.FinalizeJudgeFailure(ctx, task.SubmissionID, task.TaskID, fmt.Sprintf("judge rpc failed: %v", err), false)
		return err
	}
	zap.L().Debug("RPC Judge finished",
		zap.String("overall_status", rpcResp.Status),
		zap.Int("results_count", len(rpcResp.Results)),
	)
	// 6. 整理结果
	return w.finalizeSuccess(ctx, task.SubmissionID, task.TaskID, rpcResp)
}

// finalizeSuccess 解析 Judge RPC 响应并写回数据库和缓存
func (w *JudgeWorker) finalizeSuccess(ctx context.Context, submissionID, taskID string, resp *pb.JudgeResponse) error {
	if resp == nil {
		err := w.taskService.FinalizeJudgeFailure(ctx, submissionID, taskID, "judge rpc response is nil", false)
		if err == nil {
			w.cacheUpdater.SetFailed(submissionID, common.NextPollAfterMS0)
		}
		return err
	}

	stdout := resp.GetStdout() // 总的标准输出
	stderr := resp.GetStderr() // 总的标准错误
	var compileOutput, errorMsg string
	if resp.Status == "CE" { // 编译错误输出
		compileOutput = stderr // 程序的编译输出错误在stderr中
	}

	status := mapRPCStatus(resp.GetStatus())                     // 转换为存储的状态
	resultJSON, marshalErr := buildResultJSON(resp.GetResults()) // Json化结果
	if marshalErr != nil {
		errorMsg = fmt.Sprintf("marshal judge results failed: %v", marshalErr)
	}

	// 若最终业务状态是非 AC 类错误，可补充错误信息
	if errorMsg == "" &&
		(status == entity.SubmissionStatusCompileError ||
			status == entity.SubmissionStatusRuntimeError ||
			status == entity.SubmissionStatusSystemError) {
		errorMsg = stderr
	}

	var totalTime, totalMemory int64
	for _, testCase := range resp.GetResults() {
		totalTime += testCase.Time
		totalMemory += testCase.Memory
	}

	finishTime := time.Now() // 结束时间

	// 写到数据库中
	err := w.taskService.FinalizeJudgeSuccess(
		ctx,
		submissionID,
		taskID,
		repository.SubmissionFinishPayload{
			Status:        status,
			Stdout:        stringPtrIfNotEmpty(stdout),
			Stderr:        stringPtrIfNotEmpty(stderr),
			CompileOutput: stringPtrIfNotEmpty(compileOutput),
			ErrorMessage:  stringPtrIfNotEmpty(errorMsg),
			RuntimeMS:     normalizeRuntimeToMS(totalTime),
			MemoryKB:      normalizeMemoryToKB(totalMemory),
			ResultJSON:    resultJSON,
			FinishTime:    finishTime,
		},
		nil,
	)
	// 写到缓存中 --> 任务已经完成了
	if err == nil {
		w.cacheUpdater.SetAccpted(submissionID)
	}
	return err
}

// getLanguageCode 将字符串转为数字
func getLanguageCode(lang string) int32 {
	switch lang {
	case "go":
		return 1
	case "cpp":
		return 2
	default:
		return 0
	}
}

// buildResultJSON 将 RPC 测试点结果转为 JSON 串
func buildResultJSON(results []*pb.CaseResult) (datatypes.JSON, error) {
	type caseResultJSON struct {
		CaseID    uint64 `json:"case_id"`
		Passed    bool   `json:"passed"`
		RunTimeMS int64  `json:"runtime_ms"`
		MemoryKB  int64  `json:"memory_kb"`
		Status    string `json:"status"`
		Stdout    string `json:"stdout"`
	}

	type wrapper struct {
		Cases []caseResultJSON `json:"cases"`
	}

	resp := wrapper{
		Cases: make([]caseResultJSON, 0, len(results)),
	}

	// 构建 wrapper
	for idx, item := range results {
		if item == nil {
			continue
		}
		resp.Cases = append(resp.Cases, caseResultJSON{
			CaseID:    uint64(idx + 1),
			Passed:    item.GetPassed(),
			RunTimeMS: normalizeRuntimeToMS(item.GetTime()),
			MemoryKB:  normalizeMemoryToKB(item.GetMemory()),
			Status:    item.GetStatus(),
			Stdout:    item.GetOutput(),
		})
	}

	raw, err := json.Marshal(resp) // json解析
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(raw), nil // 强转类型返回结果json
}

// mapRPCStatus 将 Judge RPC 的状态映射为 submission 状态
func mapRPCStatus(status string) entity.SubmissionStatus {
	switch status {
	case "AC":
		return entity.SubmissionStatusAccepted
	case "WA":
		return entity.SubmissionStatusWrongAnswer
	case "CE":
		return entity.SubmissionStatusCompileError
	case "RE":
		return entity.SubmissionStatusRuntimeError
	case "TLE":
		return entity.SubmissionStatusTimeLimitExceeded
	case "MLE":
		return entity.SubmissionStatusMemoryLimitExceeded
	default:
		return entity.SubmissionStatusSystemError
	}
}

// normalizeRuntimeToMS 统一运行时间单位到 ms
//
// 说明：
//  1. proto 中的 JudgeResponse.time 的字段单位写的是 ns
//  2. 当前数据库字段 runtime_ms 是 ms
//  3. 这里按 ns -> ms 进行换算
func normalizeRuntimeToMS(v int64) int64 {
	if v <= 0 {
		return 0
	}
	return v / int64(time.Millisecond) // ns -> ms
}

// normalizeMemoryToKB 统一内存单位到 KB
//
// 说明：
//  1. 若 Judge 服务返回的已经是 KB，则这里可以直接返回
func normalizeMemoryToKB(v int64) int64 {
	if v <= 0 {
		return 0
	}
	return v
}

// stringPtrIfNotEmpty 非空字符串转指针
func stringPtrIfNotEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
