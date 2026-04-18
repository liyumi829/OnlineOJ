package cache

import "time"

const (
	// Gateway 轮询阶段常量 --> Submission 表的状态
	PhaseQueued  = "PENDING"
	PhaseRunning = "RUNNING"
	PhaseFailed  = "SYSTEM_ERROR"
	PhaseAccpted = "ACCEPTED" // 其他都是完成
)

const (
	// Gateway 本地缓存 TTL
	PollTTLQueued   = 5 * time.Second  // 排队阶段短 TTL
	PollTTLJudging  = 5 * time.Second  // 判题中短 TTL
	PollTTLFinished = 30 * time.Second // 完成态稍长 TTL
	PollTTLFailed   = 10 * time.Second // 失败态中短 TTL
	PollTTLEmpty    = 3 * time.Second  // 空值缓存，防穿透
)

// SubmissionPollState Gateway 轮询轻量缓存对象
type SubmissionPollState struct {
	SubmissionID    string    `json:"submission_id"`      // 提交唯一标识
	Phase           string    `json:"phase"`              // QUEUED / JUDGING / FINISHED / FAILED
	Done            bool      `json:"done"`               // 是否完成
	Polling         bool      `json:"polling"`            // 是否继续轮询
	NextPollAfterMS int64     `json:"next_poll_after_ms"` // 建议下次轮询时间
	NotFound        bool      `json:"not_found"`          // 是否为空值缓存
	UpdatedAt       time.Time `json:"updated_at"`         // 最近更新时间
	ExpireAt        time.Time `json:"expire_at"`          // 过期时间
}
