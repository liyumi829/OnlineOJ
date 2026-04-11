package breaker

import "time"

// CircuitBreakerSnapshot 是熔断器快照
type CircuitBreakerSnapshot struct {
	State            CircuitState
	FailureCount     int
	FailureThreshold int
	OpenedAt         time.Time
	LastFailureAt    time.Time
	LastSuccessAt    time.Time
	LastFailureErr   error
	HalfOpenInFlight bool
	OpenTimeout      time.Duration
}

// Snapshot 返回熔断器状态快照
func (b *CircuitBreaker) Snapshot() CircuitBreakerSnapshot {
	if b == nil {
		return CircuitBreakerSnapshot{
			State: CircuitClosed,
		}
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return CircuitBreakerSnapshot{
		State:            b.state,
		FailureCount:     b.failureCount,
		FailureThreshold: b.failureThreshold,
		OpenedAt:         b.openedAt,
		LastFailureAt:    b.lastFailureAt,
		LastSuccessAt:    b.lastSuccessAt,
		LastFailureErr:   b.lastFailureErr,
		HalfOpenInFlight: b.halfOpenInFlight,
		OpenTimeout:      b.openTimeout,
	}
}
