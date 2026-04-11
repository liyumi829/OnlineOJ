package breaker

import "time"

type CircuitBreakerConfig struct {
	FailureThreshold int           // 连续错误阈值
	OpenTimeout      time.Duration // 熔断器打开时间
}

func (c *CircuitBreakerConfig) SetDefault() {
	if c.FailureThreshold <= 0 {
		c.FailureThreshold = 3
	}
	if c.OpenTimeout <= 0 {
		c.OpenTimeout = 5 * time.Second
	}
}
