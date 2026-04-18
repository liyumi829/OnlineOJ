package cache

import "time"

// PhaseTTL 根据 phase 选择 TTL
func PhaseTTL(phase string) time.Duration {
	switch phase {
	case PhaseQueued:
		return PollTTLQueued
	case PhaseRunning:
		return PollTTLJudging
	case PhaseFailed:
		return PollTTLFailed
	case PhaseAccpted:
		fallthrough
	default:
		return PollTTLFinished
	}
}

// DefaultNextPollAfterMS 根据阶段设置默认轮询间隔
func DefaultNextPollAfterMS(phase string) int64 {
	switch phase {
	case PhaseQueued:
		return 200
	case PhaseRunning:
		return 300
	case PhaseAccpted, PhaseFailed:
		return 0
	default:
		return 400
	}
}
