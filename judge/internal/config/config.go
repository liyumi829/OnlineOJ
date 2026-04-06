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
}
