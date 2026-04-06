package config

// Config Judge 配置

type JudgeConfig struct {
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	LogPath       string `yaml:"log_path"`       // 日志目录
	TempPath      string `yaml:"temp_path"`      // 判题临时目录
	GlobalTimeout int64  `yaml:"global_timeout"` // 总体超时时间
}

// 用户合成配置
type rawJudgeConfig struct {
	App *rawAppConfig `yaml:"app"`
}

type rawAppConfig struct {
	LogPath       *string `yaml:"log_path"`       // 日志目录
	TempPath      *string `yaml:"temp_path"`      // 判题临时目录
	GlobalTimeout *int64  `yaml:"global_timeout"` // 总体超时时间
}
