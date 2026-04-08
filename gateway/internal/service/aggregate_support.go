package service

import (
	"context"
	"errors"
	"online-oj/gateway/internal/model/entity"
	"online-oj/gateway/internal/repository"
	"time"

	"gorm.io/gorm"
)

type SubmissionTaskAggregate struct {
	repo           *repository.Repository // 持有仓储对象
	submissionRepo *repository.SubmissionRepository
	taskRepo       *repository.JudgeTaskRepository
}

// 构造函数
func NewSubmissionTaskAggregate(repo *repository.Repository) *SubmissionTaskAggregate {
	return &SubmissionTaskAggregate{
		repo:           repo,
		submissionRepo: repository.NewSubmissionRepository(repo),
		taskRepo:       repository.NewJudgeTaskRepository(repo),
	}
}

// CreateSubmissionWithTaskParams 创建提交与任务的参数
type CreateSubmissionWithTaskParams struct {
	SubmissionID string
	ProblemID    uint64
	Language     string
	SourceCode   string

	TaskID   string
	MaxRetry uint32
}

// CreateSubmissionWithTask 原子创建 submission 和 judge_task
//
// 说明：
//
// submission 和 task 必须一起成功或一起失败；
// 否则会出现：
// 1. submission 已创建，但没有 task，任务永远不会被执行；
//
// 2. task 已创建，但没有 submission，Worker 无法查询到代码。
//
// 所以这里必须使用事务。
func (sta *SubmissionTaskAggregate) CreateSubmissionWithTask(ctx context.Context, params *CreateSubmissionWithTaskParams) error {
	if params == nil {
		return errors.New("param is empty")
	}
	if params.SubmissionID == "" {
		return errors.New("submissionID is empty")
	}
	if params.TaskID == "" {
		return errors.New("taskID is empty")
	}
	if params.ProblemID == 0 {
		return errors.New("problemID is error")
	}
	if params.SourceCode == "" {
		return errors.New("sourceCode is empty")
	}
	now := time.Now()

	return sta.repo.Transaction(ctx,
		func(tx *gorm.DB) error { // 事务函数
			submission := &entity.Submission{
				SubmissionID: params.SubmissionID,
				ProblemID:    params.ProblemID,
				Language:     params.Language,
				SourceCode:   params.SourceCode,
				Status:       entity.SubmissionStatusPending,
				SubmitTime:   now,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := sta.submissionRepo.Create(ctx, tx, submission); err != nil {
				return err
			}

			task := &entity.JudgeTask{
				TaskID:       params.TaskID,
				SubmissionID: params.SubmissionID,
				Status:       entity.JudgeTaskStatusPending,
				RetryCount:   0,
				MaxRetry:     params.MaxRetry,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := sta.taskRepo.Create(ctx, tx, task); err != nil {
				return err
			}

			return nil
		})
}

// TryClaimTaskAndMarkRunning 尝试抢占任务并将 submission 标记为 RUNNING
//
// 重要说明：
//  1. 该方法必须放在同一个事务里，保证 task 与 submission 状态一致
//  2. task 的抢占依赖条件更新：只有 status=PENDING 时才能改为 RUNNING
//  3. 若 RowsAffected=0，表示任务已被其他 Worker 抢走
func (sta *SubmissionTaskAggregate) TryClaimTaskAndMarkRunning(
	ctx context.Context,
	taskID string,
	submissionID string,
	workerID string,
) (claimed bool, err error) {
	if taskID == "" {
		return false, errors.New("taskID is empty")
	}
	if submissionID == "" {
		return false, errors.New("submissionID is empty")
	}
	if workerID == "" {
		return false, errors.New("workerID is empty")
	}

	now := time.Now()

	err = sta.repo.Transaction(ctx, func(tx *gorm.DB) error {
		// 第一步：在事务内原子抢占任务
		claimed, err = sta.taskRepo.TryClaimTaskWithTx(ctx, tx, taskID, workerID, now, now)
		if err != nil {
			return err
		}
		if !claimed {
			// 没抢到任务，不返回错误，直接结束事务
			return nil
		}

		// 第二步：抢占成功后，同事务更新 submission 为 RUNNING
		if err = sta.submissionRepo.MarkRunning(ctx, tx, submissionID, now); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return claimed, nil
}

// FinalizeJudgeSuccess 提交判题执行成功后的事务更新
// 说明：这里的“Success”指任务链路执行成功，业务结果可能是 AC/WA/CE/RE 等
func (sta *SubmissionTaskAggregate) FinalizeJudgeSuccess(
	ctx context.Context,
	submissionID string,
	taskID string,
	submissionPayload repository.SubmissionFinishPayload,
	judgeNode *string,
) error {
	if submissionID == "" {
		return errors.New("submissionID is empty")
	}
	if taskID == "" {
		return errors.New("taskID is empty")
	}

	return sta.repo.Transaction(ctx,
		func(tx *gorm.DB) error {
			// 先写 submission 最终业务状态与结果
			if err := sta.submissionRepo.FinishSubmission(ctx, tx, submissionID, submissionPayload); err != nil {
				return err
			}
			// 再写 task 完成状态
			if err := sta.taskRepo.MarkSuccess(ctx, tx, taskID, judgeNode, submissionPayload.FinishTime); err != nil {
				return err
			}
			return nil
		})
}

// FinalizeJudgeFailure 判题链路失败后的更新
//
// requeue=true：
//  1. submission 回到 PENDING
//  2. judge_task 重新入队
//
// requeue=false：
//  1. submission 标记为 SYSTEM_ERROR
//  2. judge_task 标记为 FAILED
func (sta *SubmissionTaskAggregate) FinalizeJudgeFailure(
	ctx context.Context,
	submissionID string,
	taskID string,
	errMsg string,
	requeue bool,
) error {
	if submissionID == "" {
		return errors.New("submissionID is empty")
	}
	if taskID == "" {
		return errors.New("taskID is empty")
	}
	if errMsg == "" {
		errMsg = "unknown judge failure"
	}

	now := time.Now()

	return sta.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if requeue {
			// 可重试场景：submission 不应进入终态 SYSTEM_ERROR，而应回到 PENDING
			if err := sta.submissionRepo.UpdateStatusAndErrorMessage(
				ctx,
				tx,
				submissionID,
				entity.SubmissionStatusPending,
				errMsg,
				nil,
			); err != nil {
				return err
			}

			// task 重新入队
			if err := sta.taskRepo.Requeue(ctx, tx, taskID, errMsg); err != nil {
				return err
			}

			return nil
		}

		// 不再重试：submission 进入终态 SYSTEM_ERROR
		if err := sta.submissionRepo.UpdateSystemError(ctx, tx, submissionID, errMsg, now); err != nil {
			return err
		}

		// task 标记为 FAILED
		if err := sta.taskRepo.MarkFailed(ctx, tx, taskID, errMsg, now); err != nil {
			return err
		}

		return nil
	})
}
