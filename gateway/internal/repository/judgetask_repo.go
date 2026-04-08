package repository

import (
	"context"
	"errors"
	"online-oj/gateway/internal/model/entity"
	"time"

	"gorm.io/gorm"
)

// 实现的接口：
// 1. 创建task记录
// 2. 通过taskID查询任务信息
// 3. 通过submissionID查询任务信息
// 4. 取一条待执行任务，只负责“查到候选任务”
// 5. 批量取待执行任务（批量拉去，有可能没有抢到）
// 6. 尝试抢占任务（抢去任务）配合4、5使用
// 7. 标记任务成功/失败
// 8. 根据任务ID，任务重入数据库中（执行失败）
// 9. 根据主键重入数据库
// 10. 增加重试次数并标记失败 -- (8和7结合)当任务超过最大重试次数时可直接调用
// 11. 查找长时间未完成的 RUNNING 任务
// 12. 删除较早完成的任务
// 13. 直接保存整个对象

// JudgeTaskRepository 任务仓储
type JudgeTaskRepository struct {
	repo *Repository
}

func NewJudgeTaskRepository(repo *Repository) *JudgeTaskRepository {
	return &JudgeTaskRepository{repo: repo}
}

// Create 在创建一条 judge_task 记录
func (jtr *JudgeTaskRepository) Create(ctx context.Context, tx *gorm.DB, task *entity.JudgeTask) error {
	return jtr.getDB(tx).WithContext(ctx).Create(task).Error
}

// GetByTaskID 根据 task_id 查询任务
func (jtr *JudgeTaskRepository) GetByTaskID(ctx context.Context, taskID string) (*entity.JudgeTask, error) {
	var task entity.JudgeTask
	err := jtr.repo.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetBySubmissionID 根据 submission_id 查询任务
// 一条 submission 记录对应一条 task 记录
func (jtr *JudgeTaskRepository) GetBySubmissionID(ctx context.Context, submissionID string) (*entity.JudgeTask, error) {
	var task entity.JudgeTask
	err := jtr.repo.db.WithContext(ctx).
		Where("submission_id = ?", submissionID).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// PickOnePendingTask 取一条待执行任务
// 这里只负责“查到候选任务”，不保证一定抢占成功
func (jtr *JudgeTaskRepository) PickOnePendingTask(ctx context.Context) (*entity.JudgeTask, error) {
	var task entity.JudgeTask
	err := jtr.repo.db.WithContext(ctx).
		Where("status = ?", entity.JudgeTaskStatusPending).
		Order("id ASC").
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// PickPendingTasks 批量取待执行任务
// 可用于批量拉取，但仍然必须逐个 TryClaimTask 才算真正抢到
func (jtr *JudgeTaskRepository) PickPendingTasks(ctx context.Context, limit int) ([]entity.JudgeTask, error) {
	var tasks []entity.JudgeTask
	err := jtr.repo.db.WithContext(ctx).
		Where("status = ?", entity.JudgeTaskStatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// TryClaimTask 抢占任务
//
// 说明：
// 抢占任务：1. 查到pending状态 2.能将表中的状态从pending转换到running
// 完成上述两步，RowsAffected == 1 表示当前 Worker 抢占成功
// 当条sql能够保证原子性，不会被多个sql进行更新数据
// 这里添加一个事务参数，是为了保证逻辑的完整性，即：抢占到任务之后，更新submission表也能成功
func (jtr *JudgeTaskRepository) TryClaimTaskWithTx(
	ctx context.Context,
	tx *gorm.DB,
	taskID string,
	workerID string,
	scheduledAt time.Time,
	startTime time.Time,
) (bool, error) {
	result := jtr.getDB(tx).WithContext(ctx).
		Model(&entity.JudgeTask{}).
		Where("task_id = ? AND status = ?", taskID, entity.JudgeTaskStatusPending).
		Updates(map[string]interface{}{
			"status":       entity.JudgeTaskStatusRunning,
			"worker_id":    workerID,
			"scheduled_at": scheduledAt,
			"start_time":   startTime,
		})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

// MarkSuccess 标记任务执行成功
// 注意：这里的成功是“任务执行链路成功结束”，不是业务上一定 AC
func (jtr *JudgeTaskRepository) MarkSuccess(ctx context.Context, tx *gorm.DB, taskID string, judgeNode *string, finishTime time.Time) error {
	db := jtr.getDB(tx).WithContext(ctx)

	updates := map[string]interface{}{
		"status":      entity.JudgeTaskStatusSuccess,
		"finish_time": finishTime,
	}
	if judgeNode != nil {
		updates["judge_node"] = *judgeNode
	}

	return db.Model(&entity.JudgeTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

// MarkFailed 标记任务执行失败
func (jtr *JudgeTaskRepository) MarkFailed(ctx context.Context, tx *gorm.DB, taskID string, errMsg string, finishTime time.Time) error {
	db := jtr.getDB(tx).WithContext(ctx)
	return db.Model(&entity.JudgeTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"status":      entity.JudgeTaskStatusFailed,
			"last_error":  errMsg,
			"finish_time": finishTime,
		}).Error
}

// Requeue 任务重新入队
// 用于系统级错误重试：将 RUNNING/FAILED 任务重新置为 PENDING
func (jtr *JudgeTaskRepository) Requeue(ctx context.Context, tx *gorm.DB, taskID string, lastError string) error {
	db := jtr.getDB(tx).WithContext(ctx)
	return db.Model(&entity.JudgeTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       entity.JudgeTaskStatusPending,
			"retry_count":  gorm.Expr("retry_count + 1"),
			"last_error":   lastError,
			"worker_id":    nil,
			"scheduled_at": nil,
			"start_time":   nil,
			"finish_time":  nil,
		}).Error
}

// RequeueByID 根据主键重新入队
func (jtr *JudgeTaskRepository) RequeueByID(ctx context.Context, tx *gorm.DB, id uint64, lastError string) error {
	db := jtr.getDB(tx).WithContext(ctx)
	return db.Model(&entity.JudgeTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       entity.JudgeTaskStatusPending,
			"retry_count":  gorm.Expr("retry_count + 1"),
			"last_error":   lastError,
			"worker_id":    nil,
			"scheduled_at": nil,
			"start_time":   nil,
			"finish_time":  nil,
		}).Error
}

// IncrementRetryAndFail 增加重试次数并标记失败
// 当任务超过最大重试次数时可直接调用
func (jtr *JudgeTaskRepository) IncrementRetryAndFail(ctx context.Context, tx *gorm.DB, taskID string, errMsg string, finishTime time.Time) error {
	db := jtr.getDB(tx).WithContext(ctx)
	return db.Model(&entity.JudgeTask{}). // Model(&Submission{}) = 告诉 GORM 要更新哪张表！
						Where("task_id = ?", taskID).
						Updates(map[string]interface{}{
			"retry_count": gorm.Expr("retry_count + 1"),
			"status":      entity.JudgeTaskStatusFailed,
			"last_error":  errMsg,
			"finish_time": finishTime,
		}).Error
}

// FindStaleRunningTasks 查找长时间未完成的 RUNNING 任务
// 用于 Worker 崩溃恢复、系统重启后的脏任务修复
func (jtr *JudgeTaskRepository) FindStaleRunningTasks(ctx context.Context, before time.Time, limit int) ([]entity.JudgeTask, error) {
	var tasks []entity.JudgeTask
	err := jtr.repo.db.WithContext(ctx).
		Where("status = ? AND start_time IS NOT NULL AND start_time < ?", entity.JudgeTaskStatusRunning, before).
		Order("id ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// DeleteFinishedBefore 删除较早完成的任务
// 建议只用于定时清理 judge_tasks，不建议清理 submissions
func (jtr *JudgeTaskRepository) DeleteFinishedBefore(ctx context.Context, tx *gorm.DB, before time.Time) error {
	db := jtr.getDB(tx).WithContext(ctx)
	return db.Where("status IN ? AND finish_time IS NOT NULL AND finish_time < ?",
		[]entity.JudgeTaskStatus{entity.JudgeTaskStatusSuccess, entity.JudgeTaskStatusFailed}, before).
		Delete(&entity.JudgeTask{}).Error
	// &struct{} 只是告诉 GORM：操作哪张表
}

// Save 直接保存整个对象
// 不建议用于高并发调度主链路；这里保留备用
func (jtr *JudgeTaskRepository) Save(ctx context.Context, tx *gorm.DB, task *entity.JudgeTask) error {
	if task == nil {
		return errors.New("judge task is nil")
	}
	db := jtr.getDB(tx).WithContext(ctx)
	return db.Save(task).Error
}

// getDB 优先使用事务 tx；若 tx 为空则回退到默认 db
func (jtr *JudgeTaskRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return jtr.repo.db
}
