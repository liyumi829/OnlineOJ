package rpc

import (
	"errors"
	"time"
)

// Config 是 RPC Client 配置
type Config struct {
	Addrs          []string      // Addrs 是多个 Judge 服务地址。
	RequestTimeout time.Duration // RequestTimeout 是单次 RPC 请求超时时间。
}

// setDefault 为Config设置默认值
func (c *Config) setDefault() {
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 5 * time.Second
	}
}

// NewJudgeNodesConfig
func NewConfig(Addrs []string, timeout time.Duration) (*Config, error) {
	if len(Addrs) == 0 {
		return nil, errors.New("Addrs is empty.")
	}
	config := &Config{}

	config.setDefault()
	config.Addrs = Addrs
	if timeout != 0 {
		config.RequestTimeout = timeout
	}
	return config, nil
}
