package config

import (
	pkgconfiger "online-oj/pkg/configer"
)

// Load 加载配置文件，后面的文件覆盖前面的文件
func Load(paths ...string) (*JudgeConfig, error) {
	newRaw := func() *rawJudgeConfig {
		return &rawJudgeConfig{}
	}
	return pkgconfiger.Load(defaultConfig(), newRaw, mergeJudgeConfig, paths...)
}

// defaultConfig 默认配置
func defaultConfig() *JudgeConfig {
	cfg := &JudgeConfig{}
	cfg.App.LogPath = "./logs"
	cfg.App.TempPath = "./temp"
	cfg.App.GlobalTimeout = 2

	cfg.Rpc.InstanceName = "judge[127.0.0.1:10000]"
	cfg.Rpc.Addr = "127.0.0.1:10000"

	cfg.Redis.Addrs = []string{
		"127.0.0.1:6379",
	}
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0
	cfg.Redis.PoolSize = 100
	cfg.Redis.MinIdleConns = 10
	return cfg
}

func mergeJudgeConfig(dst *JudgeConfig, src *rawJudgeConfig) {
	if src == nil {
		return
	}

	if src.App != nil {
		if src.App.LogPath != nil {
			dst.App.LogPath = *src.App.LogPath
		}
		if src.App.TempPath != nil {
			dst.App.TempPath = *src.App.TempPath
		}
		if src.App.GlobalTimeout != nil {
			dst.App.GlobalTimeout = *src.App.GlobalTimeout
		}
	}

	if src.Rpc != nil {
		if src.Rpc.InstanceName != nil {
			dst.Rpc.InstanceName = *src.Rpc.InstanceName
		}
		if src.Rpc.Addr != nil {
			dst.Rpc.Addr = *src.Rpc.Addr
		}
	}

	if src.Redis != nil {
		if src.Redis.Addrs != nil {
			dst.Redis.Addrs = *src.Redis.Addrs
		}
		if src.Redis.Password != nil {
			dst.Redis.Password = *src.Redis.Password
		}
		if src.Redis.DB != nil {
			dst.Redis.DB = *src.Redis.DB
		}
		if src.Redis.PoolSize != nil {
			dst.Redis.PoolSize = *src.Redis.PoolSize
		}
		if src.Redis.MinIdleConns != nil {
			dst.Redis.MinIdleConns = *src.Redis.MinIdleConns
		}
	}
}
