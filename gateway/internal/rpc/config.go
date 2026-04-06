package rpc

import "time"

// Config 是一个 RPC Client 配置。
type Config struct {
	Addr           string        // Addr 是单个 Judge 服务地址。
	RequestTimeout time.Duration // RequestTimeout 是单次 RPC 请求超时时间。
}

// setDefault 为Config设置默认值。连接超时3s、请求超时5s
func (c *Config) setDefault() {
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 5 * time.Second
	}
}
