package entity

import "time"

// JudgeTask 对应 judge_tasks 表
type JudgeTask struct {
	ID           uint64          `gorm:"column:id;primaryKey;autoIncrement;comment:数据库自增主键"`                                                          // 数据库内部主键
	TaskID       string          `gorm:"column:task_id;type:varchar(64);not null;uniqueIndex:uk_task_id;comment:任务唯一ID"`                              // 任务唯一 ID
	SubmissionID string          `gorm:"column:submission_id;type:varchar(64);not null;index:idx_submission_id;comment:关联的提交ID"`                      // 关联的提交 ID
	Status       JudgeTaskStatus `gorm:"column:status;type:varchar(32);not null;default:PENDING;index:idx_status_created_at,priority:1;comment:任务状态"` // 当前调度状态
	RetryCount   uint32          `gorm:"column:retry_count;type:int unsigned;not null;default:0;comment:当前已重试次数"`                                     // 已重试次数
	MaxRetry     uint32          `gorm:"column:max_retry;type:int unsigned;not null;comment:最大重试次数"`                                                  // 最大重试次数
	WorkerID     *string         `gorm:"column:worker_id;type:varchar(64);index:idx_worker_id;comment:执行该任务的Worker标识"`                                // 被哪个 Worker 抢占执行
	JudgeNode    *string         `gorm:"column:judge_node;type:varchar(128);comment:执行该任务的Judge节点"`                                                   // 实际执行的 Judge 节点
	LastError    *string         `gorm:"column:last_error;type:text;comment:最近一次错误信息"`                                                                // 最近一次错误信息
	ScheduledAt  *time.Time      `gorm:"column:scheduled_at;comment:任务被调度时间"`                                                                         // 调度时间
	StartTime    *time.Time      `gorm:"column:start_time;comment:任务开始执行时间"`                                                                          // 开始执行时间
	FinishTime   *time.Time      `gorm:"column:finish_time;comment:任务完成时间"`                                                                           // 完成时间
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;index:idx_status_created_at,priority:2;comment:创建时间"`    // 创建时间
	UpdatedAt    time.Time       `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"`                            // 更新时间
}

// TableName 指定表名
func (JudgeTask) TableName() string {
	return "judge_tasks"
}
