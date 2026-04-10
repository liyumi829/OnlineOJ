package config

// Config Judge 配置

type JudgeConfig struct {
	App AppConfig `yaml:"app"`
	Rpc RpcConfig `yaml:"rpc"`
}

type AppConfig struct {
	LogPath       string `yaml:"log_path"`       // 日志目录
	TempPath      string `yaml:"temp_path"`      // 判题临时目录
	GlobalTimeout int64  `yaml:"global_timeout"` // 总体超时时间
}

type RpcConfig struct {
	InstanceName string `yaml:"instance_name"` // judge 服务的名称
	Addr         string `yaml:"addr"`          // 监听的地址
}

// 用户合成配置
type rawJudgeConfig struct {
	App *rawAppConfig `yaml:"app"`
	Rpc *rawRpcConfig `yaml:"rpc"`
}

type rawAppConfig struct {
	LogPath       *string `yaml:"log_path"`
	TempPath      *string `yaml:"temp_path"`
	GlobalTimeout *int64  `yaml:"global_timeout"`
}

type rawRpcConfig struct {
	InstanceName *string `yaml:"instance_name"`
	Addr         *string `yaml:"addr"`
}
