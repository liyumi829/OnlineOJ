package breaker

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu               sync.Mutex    // 互斥锁保护熔断信息
	state            CircuitState  // 熔断状态
	failureCount     int           // 失败次数
	failureThreshold int           // 状态进行转换的阈值
	openTimeout      time.Duration // 熔断器开启的时长
	openedAt         time.Time     // 熔断器开始的时间
	halfOpenInFlight bool          // 熔断器是否半开启，半开启有任务正在试探
	lastFailureAt    time.Time     // 上一次失败时间
	lastSuccessAt    time.Time     // 上一次成功实践
	lastFailureErr   error         // 上一次错误信息
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(cfg *CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: cfg.FailureThreshold,
		openTimeout:      cfg.OpenTimeout,
	}
}

// IsAllow 判断当前是否允许请求通过
//
// 说明：发生错误则不允许通过/未发生错误则允许通过
func (b *CircuitBreaker) IsAllow() error {
	if b == nil {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	zap.L().Debug("breaker:", zap.String("state", b.state.String()))
	switch b.state { // 判断当前熔断器状态

	case CircuitClosed:
		// 熔断器处于关闭状态
		return nil // 允许通过

	case CircuitOpen:
		// 熔断器处于开启状态
		if now.Sub(b.openedAt) >= b.openTimeout { // Open 超时后转为 Half-Open，并允许一个试探请求通过
			b.state = CircuitHalfOpen
			b.halfOpenInFlight = true // 半开状态
			return nil                // 允许请求通过
		}
		return ErrCircuitOpen

	case CircuitHalfOpen:
		// Half-Open 状态只允许一个试探请求
		if b.halfOpenInFlight { // 有任务正在执行
			return ErrCircuitHalfOpenBusy
		}
		b.halfOpenInFlight = true // 如果没有任务在执行，就放一个任务执行
		return nil

	default:
		return ErrCircuitOpen
	}
}

// OnSuccess 标记业务调用成功
//
// 业务成功：更新成功状态
func (b *CircuitBreaker) OnSuccess() {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.failureCount = 0
	b.lastFailureErr = nil
	b.lastSuccessAt = time.Now()
	b.halfOpenInFlight = false

	// 成功后统一回到 Closed
	b.state = CircuitClosed
}

// OnFailure 标记业务调用失败
//
// 业务失败：失败时间，信息记录。
//
// 先看状态，如果是半开状态则，打开熔断器；如果是关闭状态，判断是否达到阈值
func (b *CircuitBreaker) OnFailure(err error) {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastFailureAt = time.Now()
	b.lastFailureErr = err
	b.halfOpenInFlight = false

	switch b.state {
	case CircuitClosed:
		b.failureCount++
		if b.failureCount >= b.failureThreshold {
			b.open() // 如果到达阈值就打开熔断器
		}

	case CircuitHalfOpen:
		// Half-Open 试探失败，立即重新 Open
		b.open()

	case CircuitOpen:
		// Open 状态下理论上不会发起请求，这里只更新时间
		b.open()

	default:
		b.open()
	}
}

// State 返回当前熔断状态
func (b *CircuitBreaker) State() CircuitState {
	if b == nil {
		return CircuitClosed
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.state
}

// open 打开熔断器
func (b *CircuitBreaker) open() {
	b.state = CircuitOpen
	b.openedAt = time.Now()
	b.halfOpenInFlight = false
}
