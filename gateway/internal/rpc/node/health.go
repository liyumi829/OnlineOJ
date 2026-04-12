// 节点的健康状态
package node

import (
	"sync"
	"time"
)

type HealthyStatus struct {
	mu                     sync.Mutex // 互斥锁，并发安全
	heartbeatHealthy       bool       // 健康状态
	lastHeartbeatSuccessAt time.Time  // 最近一次心跳成功时间
	lastHeartbeatFailAt    time.Time  // 最近一次心跳失败时间
	lastHeartbeatErr       error      // 最近一次心跳错误
}

// NewHealthyStatus 创建 HealthyStatus
func NewHealthyStatus() *HealthyStatus {
	return &HealthyStatus{
		heartbeatHealthy:       false,
		lastHeartbeatSuccessAt: time.Time{},
		lastHeartbeatFailAt:    time.Time{},
		lastHeartbeatErr:       nil,
	}
}

// MarkHeartbeatSuccess 标记心跳成功
func (h *HealthyStatus) MarkHeartbeatSuccess() {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	h.heartbeatHealthy = true
	h.lastHeartbeatSuccessAt = time.Now()
	h.lastHeartbeatErr = nil
}

// MarkHeartbeatFailure 标记心跳失败
func (h *HealthyStatus) MarkHeartbeatFailure(err error) {
	if h == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.heartbeatHealthy = false
	h.lastHeartbeatFailAt = time.Now()
	h.lastHeartbeatErr = err
}

// IsHealthy 返回节点的健康状态
func (h *HealthyStatus) IsHealthy() bool {
	if h == nil {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.heartbeatHealthy
}

// HealthyStatusSnapshot 心跳快照
type HealthyStatusSnapshot struct {
	Healthy       bool
	LastSuccessAt time.Time
	LastFailAt    time.Time
	LastErr       error
}

// HeartbeatSnapshot 返回心跳状态快照
func (h *HealthyStatus) Snapshot() HealthyStatusSnapshot {
	if h == nil {
		return HealthyStatusSnapshot{}
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	return HealthyStatusSnapshot{
		Healthy:       h.heartbeatHealthy,
		LastSuccessAt: h.lastHeartbeatSuccessAt,
		LastFailAt:    h.lastHeartbeatFailAt,
		LastErr:       h.lastHeartbeatErr,
	}
}
