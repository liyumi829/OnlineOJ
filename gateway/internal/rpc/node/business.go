package node

import (
	"online-oj/gateway/internal/rpc/node/breaker"
	"sync"
	"time"
)

// 业务处理字段

type BusinessStatus struct {
	mu               sync.Mutex     // 互斥锁，并发安全
	CircuitBreaker   BreakerInvoker // 熔断器
	lastBizSuccessAt time.Time      // 上次成功时间
	lastBizFailAt    time.Time      // 上次失败时间
	lastBizErr       error          // 上次错误信息
}

// NewBusinessStatus 创建 BusinessStatus
func NewBusinessStatus(b BreakerInvoker) *BusinessStatus {
	if b == nil {
		b = breaker.NewCircuitBreaker(&breaker.CircuitBreakerConfig{})
	}

	return &BusinessStatus{
		CircuitBreaker:   b,
		lastBizSuccessAt: time.Time{},
		lastBizFailAt:    time.Time{},
		lastBizErr:       nil,
	}
}

// AllowRequest 判断是否允许请求通过
func (b *BusinessStatus) AllowRequest() error {
	if b == nil || b.CircuitBreaker == nil {
		return nil
	}
	return b.CircuitBreaker.IsAllow()
}

// MarkBizSuccess 标记业务成功
func (b *BusinessStatus) MarkBizSuccess() {
	if b == nil {
		return
	}

	b.mu.Lock()
	b.lastBizSuccessAt = time.Now()
	b.lastBizErr = nil
	b.mu.Unlock()

	if b.CircuitBreaker != nil {
		b.CircuitBreaker.OnSuccess()
	}
}

// MarkBizFailure 标记业务失败
func (b *BusinessStatus) MarkBizFailure(err error) {
	if b == nil {
		return
	}

	b.mu.Lock()
	b.lastBizFailAt = time.Now()
	b.lastBizErr = err
	b.mu.Unlock()

	if b.CircuitBreaker != nil {
		b.CircuitBreaker.OnFailure(err)
	}
}

// BreakerState 返回熔断器状态
func (b *BusinessStatus) BreakerState() breaker.CircuitState {
	if b == nil || b.CircuitBreaker == nil {
		return breaker.CircuitClosed
	}
	return b.CircuitBreaker.State()
}

// BusinessStatusSnapshot 是业务状态快照
type BusinessStatusSnapshot struct {
	LastBizSuccessAt time.Time
	LastBizFailAt    time.Time
	LastBizErr       error
	BreakerState     breaker.CircuitState
}

// Snapshot 返回业务状态快照
func (b *BusinessStatus) Snapshot() BusinessStatusSnapshot {
	if b == nil {
		return BusinessStatusSnapshot{}
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	state := breaker.CircuitClosed
	if b.CircuitBreaker != nil {
		state = b.CircuitBreaker.State()
	}

	return BusinessStatusSnapshot{
		LastBizSuccessAt: b.lastBizSuccessAt,
		LastBizFailAt:    b.lastBizFailAt,
		LastBizErr:       b.lastBizErr,
		BreakerState:     state,
	}
}
