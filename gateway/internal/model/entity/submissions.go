package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Submission 对应 submissions 表
type Submission struct {
	ID            uint64           `gorm:"column:id;primaryKey;autoIncrement;comment:数据库自增主键"`                                               // 数据库内部主键
	SubmissionID  string           `gorm:"column:submission_id;type:varchar(64);not null;uniqueIndex:uk_submission_id;comment:服务端生成的业务提交ID"` // 对外暴露给前端的业务提交 ID
	ProblemID     uint64           `gorm:"column:problem_id;type:bigint unsigned;not null;index:idx_problem_id;comment:题目ID"`                // 题目 ID
	Language      string           `gorm:"column:language;type:varchar(32);not null;comment:提交语言类型"`                                         // 语言类型 "go" "cpp"
	SourceCode    string           `gorm:"column:source_code;type:mediumtext;not null;comment:用户提交的源代码"`                                     // 用户源代码
	Status        SubmissionStatus `gorm:"column:status;type:varchar(32);not null;default:PENDING;index:idx_status;comment:提交业务状态"`          // 当前业务状态
	Stdout        *string          `gorm:"column:stdout;type:mediumtext;comment:程序标准输出"`                                                     // 程序标准输出
	Stderr        *string          `gorm:"column:stderr;type:mediumtext;comment:程序标准错误输出"`                                                   // 程序标准错误输出
	CompileOutput *string          `gorm:"column:compile_output;type:mediumtext;comment:编译输出"`                                               // 编译输出
	ErrorMessage  *string          `gorm:"column:error_message;type:text;comment:系统错误信息"`                                                    // 系统错误信息
	RuntimeMS     int64            `gorm:"column:runtime_ms;type:bigint;not null;default:0;comment:总运行时间(毫秒)"`                               // 总运行时间，毫秒
	MemoryKB      int64            `gorm:"column:memory_kb;type:bigint;not null;default:0;comment:总内存使用(KB)"`                                // 总内存，KB
	ResultJSON    datatypes.JSON   `gorm:"column:result_json;type:json;comment:判题详细结果JSON"`                                                  // 详细判题结果，直接存 JSON
	SubmitTime    time.Time        `gorm:"column:submit_time;not null;default:CURRENT_TIMESTAMP;index:idx_submit_time;comment:提交时间"`         // 提交时间
	StartTime     *time.Time       `gorm:"column:start_time;comment:开始判题时间"`                                                                 // 开始执行时间
	FinishTime    *time.Time       `gorm:"column:finish_time;comment:判题完成时间"`                                                                // 完成时间
	CreatedAt     time.Time        `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:记录创建时间"`                              // 创建时间
	UpdatedAt     time.Time        `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:记录更新时间"`               // 更新时间
}

// TableName 指定表名
func (Submission) TableName() string {
	return "submissions"
}

// 判题结果：
// {
//   "cases": [
//     {
//       "case_id": 1,
//       "status": "ACCEPTED",
//       "runtime_ms": 2,
//       "memory_kb": 512
//     },
//     {
//       "case_id": 2,
//       "status": "WRONG_ANSWER",
//       "runtime_ms": 1,
//       "memory_kb": 520,
//       "stdout": "4",
//       "expected": "3"
//     }
//   ]
// }
