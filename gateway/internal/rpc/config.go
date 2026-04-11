package rpc

import (
	"time"
)

// Config 是 RPC Client 配置
type RpcClientManagerConfig struct {
	Addrs                   []string        // Addrs 是多个 Judge 服务地址。
	RequestTimeout          time.Duration   // RequestTimeout 是单次 RPC 请求超时时间。
	HeartbeatTimeout        time.Duration   // 健康检查 RPC 超时
	HeartbeatInterval       time.Duration   // 健康检查扫描间隔
	CircuitBreakOpenTimeout time.Duration   // 熔断器开启时间
	FailureThreshold        int             // 连续失败阈值
	MaxRetries              int             // 例如 MaxRetries=2 表示：首次调用 + 最多 2 次重试
	RetryBackoffs           []time.Duration // 重试退避时间列表
}

// Default 为Config设置默认值
func (c *RpcClientManagerConfig) SetDefault() {
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 2 * time.Second
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = 2 * time.Second // 超时时间 2 s
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = 5 * time.Second // 每5s进行一次检测
	}
	if c.CircuitBreakOpenTimeout <= 0 {
		c.CircuitBreakOpenTimeout = 5 * time.Second // 熔断时长 5 s
	}
	// 最大重试次数 < 0 不进行重试 == 0 默认重试 2次
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 2
	}
	if len(c.RetryBackoffs) == 0 {
		c.RetryBackoffs = []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
		}
	}
}
