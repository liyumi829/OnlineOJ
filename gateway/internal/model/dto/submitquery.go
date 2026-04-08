package dto

// SubmitQueryRequest 表示查询提交结果的请求
type SubmitQueryRequest struct {
	// 提交唯一标识
	SubmissionID string `json:"submission_id" form:"submission_id"`
}

// TestCaseResult 表示单个测试点结果
type TestCaseResult struct {
	CaseID    uint64 `json:"case_id"`    // 测试点编号
	Status    string `json:"status"`     // 单个测试点状态 AC/WA/OLE/RE/MLE/TLE
	RunTimeMS int64  `json:"runtime_ms"` // 运行时间，单位毫秒
	MemoryKB  int64  `json:"memory_kb"`  // 内存使用，单位 KB
	Stdout    string `json:"stdout"`     // 程序标准输出
	Stderr    string `json:"stderr"`     // 程序标准错误输出
	ErrorMsg  string `json:"error_msg"`  // 错误信息
}

// SubmitQueryResponse 表示查询判题结果的响应
type SubmitQueryResponse struct {
	// 提交唯一标识
	SubmissionID string `json:"submission_id"`

	// 当前阶段：QUEUED / JUDGING / FINISHED / FAILED
	//
	// 说明：
	//   - QUEUED：任务已提交，等待 Worker 执行
	//   - JUDGING：任务正在执行
	//   - FINISHED：任务已完成，业务结果见 Status
	//   - FAILED：系统级失败，例如 Judge RPC 多次失败
	Phase string `json:"phase"`

	// 是否已经完成
	//
	// 前端轮询时优先看这个字段：
	//   - false：继续轮询
	//   - true：停止轮询，展示最终结果
	Done bool `json:"done"`

	// 是否应该继续轮询
	//
	// 前端可使用：
	//   - polling=true：继续轮询
	//   - polling=false：停止轮询
	Polling bool `json:"polling"`

	NextPollAfterMS int64            `json:"next_poll_after_ms"` // 建议前端下次轮询间隔，单位毫秒
	Status          string           `json:"status"`             // 总状态：AC/WA/OLE/RE/MLE/TLE
	RunTimeMS       int64            `json:"runtime_ms"`         // 总运行时间，单位毫秒
	MemoryKB        int64            `json:"memory_kb"`          // 总内存使用，单位 KB
	Stdout          string           `json:"stdout"`             // 汇总标准输出
	Stderr          string           `json:"stderr"`             // 汇总标准错误输出
	ErrorMsg        string           `json:"error_msg"`          // 系统错误
	Cases           []TestCaseResult `json:"cases"`              // 每个测试点的详细结果
}
