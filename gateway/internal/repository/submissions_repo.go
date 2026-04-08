package repository

import (
	"context"
	"errors"
	"online-oj/gateway/internal/model/entity"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 主要实现的业务接口（非业务没有列出）：
// 1. Create 在当前事务/连接中创建一条 submission 记录
// 2. 根据 submission_id 查询提交记录
// 3. 在事务中带行锁查询
// 4. 标记 submission 为运行中/系统错误/任意状态
// 5. 更新 submission 的最终状态与结果（运行状态就可以直接更新了）
// 6. 判断 submission 是否存在
// 7. 根据 submission_id 删除记录
// 8. Save 直接保存整个对象

// SubmissionRepository 提交记录仓储
type SubmissionRepository struct {
	repo *Repository
}

// SubmissionFinishPayload 用于更新最终结果
type SubmissionFinishPayload struct {
	Status        entity.SubmissionStatus // 状态
	Stdout        *string                 // 标准输出
	Stderr        *string                 // 标准错误
	CompileOutput *string                 // 编译错误输出
	ErrorMessage  *string                 // 系统错误信息
	RuntimeMS     int64                   // 运行时间 ms
	MemoryKB      int64                   // 内存消耗 kb
	ResultJSON    datatypes.JSON          // 运行结果 JSon
	FinishTime    time.Time               // 运行时间
}

// 创建 Submission 仓储
func NewSubmissionRepository(repo *Repository) *SubmissionRepository {
	return &SubmissionRepository{repo: repo}
}

// Create 在当前事务/连接中创建一条 submission 记录
func (sr *SubmissionRepository) Create(ctx context.Context, tx *gorm.DB, submission *entity.Submission) error {
	return sr.getDB(tx).WithContext(ctx).Create(submission).Error
}

// GetBySubmissionID 根据 submission_id 查询提交记录
func (sr *SubmissionRepository) GetBySubmissionID(ctx context.Context, submissionID string) (*entity.Submission, error) {
	var submission entity.Submission
	err := sr.repo.db.WithContext(ctx).
		Where("submission_id = ?", submissionID).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetBySubmissionIDForUpdate 在事务中带行锁查询
// 适用于需要先读后改、且希望更强一致性时使用
// 在事务中，执行该操作后，对应的行不能被修改，直到提交事务（行锁）
func (sr *SubmissionRepository) GetBySubmissionIDForUpdate(ctx context.Context, tx *gorm.DB, submissionID string) (*entity.Submission, error) {
	var submission entity.Submission
	err := sr.getDB(tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("submission_id = ?", submissionID).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// MarkRunning 标记 submission 为运行中状态
func (sr *SubmissionRepository) MarkRunning(ctx context.Context, tx *gorm.DB, submissionID string, startTime time.Time) error {
	return sr.getDB(tx).WithContext(ctx).Model(&entity.Submission{}).
		Where("submission_id = ?", submissionID).
		Updates(map[string]interface{}{
			"status":     entity.SubmissionStatusRunning,
			"start_time": startTime,
		}).Error
}

// UpdateSystemError 将 submission 标记为系统错误
// 用于 Worker/Judge RPC 调用失败时
func (sr *SubmissionRepository) UpdateSystemError(ctx context.Context, tx *gorm.DB, submissionID string, errMsg string, finishTime time.Time) error {
	return sr.getDB(tx).WithContext(ctx).Model(&entity.Submission{}).
		Where("submission_id = ?", submissionID).
		Updates(map[string]interface{}{
			"status":        entity.SubmissionStatusSystemError,
			"error_message": errMsg,
			"finish_time":   finishTime,
		}).Error
}

// UpdateStatusAndError 通用更新 submission 状态与错误信息
func (sr *SubmissionRepository) UpdateStatusAndErrorMessage(
	ctx context.Context,
	tx *gorm.DB,
	submissionID string,
	status entity.SubmissionStatus,
	errMsg string,
	finishTime *time.Time,
) error {
	updates := map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
	}
	if finishTime != nil {
		updates["finish_time"] = *finishTime
	}

	return sr.getDB(tx).WithContext(ctx).Model(&entity.Submission{}).
		Where("submission_id = ?", submissionID).
		Updates(updates).Error
}

// FinishSubmission 更新 submission 的最终状态与结果
func (sr *SubmissionRepository) FinishSubmission(ctx context.Context, tx *gorm.DB, submissionID string, payload SubmissionFinishPayload) error {
	updates := map[string]interface{}{ // 更新信息
		"status":         payload.Status,
		"stdout":         payload.Stdout,
		"stderr":         payload.Stderr,
		"compile_output": payload.CompileOutput,
		"error_message":  payload.ErrorMessage,
		"runtime_ms":     payload.RuntimeMS,
		"memory_kb":      payload.MemoryKB,
		"result_json":    payload.ResultJSON,
		"finish_time":    payload.FinishTime,
	}

	return sr.getDB(tx).WithContext(ctx).Model(&entity.Submission{}).
		Where("submission_id = ?", submissionID).
		Updates(updates).Error
}

// ListByStatus 按状态查询 submission 列表
// 可用于后台统计、管理接口
func (sr *SubmissionRepository) ListByStatus(ctx context.Context, status entity.SubmissionStatus, limit int) ([]entity.Submission, error) {
	var list []entity.Submission
	err := sr.repo.db.WithContext(ctx).
		Where("status = ?", status).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// ExistsBySubmissionID 判断 submission 是否存在
// 可用于等幂性
func (sr *SubmissionRepository) ExistsBySubmissionID(ctx context.Context, submissionID string) (bool, error) {
	var count int64
	err := sr.repo.db.WithContext(ctx).
		Model(&entity.Submission{}).
		Where("submission_id = ?", submissionID).
		Count(&count).Error
	return count > 0, err
}

// DeleteBySubmissionID 根据 submission_id 删除记录
// 注意：业务上通常不建议直接删除 submission，这里保留通用能力
func (sr *SubmissionRepository) DeleteBySubmissionID(ctx context.Context, tx *gorm.DB, submissionID string) error {
	db := sr.getDB(tx).WithContext(ctx)
	return db.Where("submission_id = ?", submissionID).Delete(&entity.Submission{}).Error
}

// Save 直接保存整个对象
// 如果对应的主键存在，则更新字段；否则，插入一条数据
func (sr *SubmissionRepository) Save(ctx context.Context, tx *gorm.DB, submission *entity.Submission) error {
	if submission == nil {
		return errors.New("submission is nil")
	}
	db := sr.getDB(tx).WithContext(ctx)
	return db.Save(submission).Error
}

// getDB 优先使用事务 tx；若 tx 为空则回退到默认 db
func (sr *SubmissionRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return sr.repo.db
}
