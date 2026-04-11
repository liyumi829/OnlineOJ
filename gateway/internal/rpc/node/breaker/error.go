package breaker

import "errors"

var (
	ErrCircuitOpen         = errors.New("circuit breaker is open")                    // ErrCircuitOpen 表示当前熔断器打开，请求不允许通过
	ErrCircuitHalfOpenBusy = errors.New("circuit breaker half-open trial is running") // ErrCircuitHalfOpenBusy 表示半开状态已有试探请求正在执行
)
