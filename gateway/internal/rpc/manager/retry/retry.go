package retry

import (
	"context"
	"time"
)

// ExponentialRetryPolicy 是简单指数退避策略
type ExponentialRetryPolicy struct {
	maxRetries int
	backoffs   []time.Duration
}

// NewExponentialRetryPolicy 创建重试策略
func NewExponentialRetryPolicy(maxRetries int, backoffs []time.Duration) *ExponentialRetryPolicy {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries == 0 {
		maxRetries = 2
	}
	if len(backoffs) == 0 {
		backoffs = []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
		}
	}

	return &ExponentialRetryPolicy{
		maxRetries: maxRetries,
		backoffs:   backoffs,
	}
}

// MaxRetries 返回最大重试次数
func (p *ExponentialRetryPolicy) MaxRetries() int {
	if p == nil {
		return 0
	}
	return p.maxRetries
}

// Backoff 返回退避时间
func (p *ExponentialRetryPolicy) Backoff(retryIndex int) time.Duration {
	if p == nil || retryIndex <= 0 {
		return 0
	}

	index := retryIndex - 1
	if index < len(p.backoffs) {
		return p.backoffs[index]
	}

	return p.backoffs[len(p.backoffs)-1]
}

// Wait 等待退避时间，受 ctx 控制
func (p *ExponentialRetryPolicy) Wait(ctx context.Context, retryIndex int) error {
	backoff := p.Backoff(retryIndex)
	if backoff <= 0 {
		return nil
	}

	timer := time.NewTimer(backoff)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
