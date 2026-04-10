package rpc

import (
	"time"
)

// Config 是 RPC Client 配置
type Config struct {
	Addrs             []string      // Addrs 是多个 Judge 服务地址。
	RequestTimeout    time.Duration // RequestTimeout 是单次 RPC 请求超时时间。
	HeartbeatTimeout  time.Duration // 健康检查 RPC 超时
	HeartbeatInterval time.Duration // 健康检查扫描间隔
	FailureThreshold  int           // 连续失败阈值
}

// Default 为Config设置默认值
func Default() *Config {
	c := &Config{}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 2 * time.Second
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = 2 * time.Second // 超时时间 2 s
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = 5 * time.Second // 每5s进行一次检测
	}
	if c.FailureThreshold <= 0 {
		c.FailureThreshold = 3 // 默认连续失败阈值为3
	}
	return c
}
