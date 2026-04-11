package node

import "online-oj/gateway/internal/rpc/node/breaker"

type JudgeNodeConfig struct {
	Addr string
	breaker.CircuitBreakerConfig
}

func (c *JudgeNodeConfig) SetDefault() {
	(&c.CircuitBreakerConfig).SetDefault() // 设置默认值
}
