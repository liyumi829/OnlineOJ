package dto

// SubmitRequest 表示前端提交代码的请求体
type SubmitRequest struct {
	ProblemID uint64 `json:"problem_id"` // 题目 ID
	Code      string `json:"code"`       // 用户提交的源代码
	Language  string `json:"language"`   // 代码语言："go" "cpp" ...
}

// SubmitResponse 表示提交代码后的响应
// 异步架构下，提交成功后只返回 submission_id 和当前状态
type SubmitResponse struct {
	SubmissionID string `json:"submission_id"` // 服务端生成的提交唯一标识
	Status       string `json:"status"`        // 当前状态：Success/Fail
	Message      string `json:"message"`       // 可选提示信息
}
