package entity

// 常量文件，标记提交的代码状态

type SubmissionStatus string // 提交状态(9)，带*都是结束的状态(4->3+1[6])
type JudgeTaskStatus string  // 任务状态(4)：p/r/s/f

const (
	SubmissionStatusPending             SubmissionStatus = "PENDING"               // 提交已创建，等待调度
	SubmissionStatusRunning             SubmissionStatus = "RUNNING"               // 判题执行中
	SubmissionStatusSystemError         SubmissionStatus = "SYSTEM_ERROR"          // 系统级错误
	SubmissionStatusAccepted            SubmissionStatus = "ACCEPTED"              // *判题成功通过
	SubmissionStatusWrongAnswer         SubmissionStatus = "WRONG_ANSWER"          // *答案错误
	SubmissionStatusCompileError        SubmissionStatus = "COMPILE_ERROR"         // *编译错误
	SubmissionStatusRuntimeError        SubmissionStatus = "RUNTIME_ERROR"         // *运行时错误
	SubmissionStatusTimeLimitExceeded   SubmissionStatus = "TIME_LIMIT_EXCEEDED"   // *超时
	SubmissionStatusMemoryLimitExceeded SubmissionStatus = "MEMORY_LIMIT_EXCEEDED" // *超内存
)

const (
	JudgeTaskStatusPending JudgeTaskStatus = "PENDING" // 等待执行
	JudgeTaskStatusRunning JudgeTaskStatus = "RUNNING" // 执行中
	JudgeTaskStatusSuccess JudgeTaskStatus = "SUCCESS" // 执行结束（不代表业务上一定 AC，只代表任务执行链路完成）
	JudgeTaskStatusFailed  JudgeTaskStatus = "FAILED"  // 执行失败（通常是系统级错误）
)
