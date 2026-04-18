package config

// Config Judge 配置

type JudgeConfig struct {
	App   AppConfig   `yaml:"app"`
	Rpc   RpcConfig   `yaml:"rpc"`
	Redis RedisConfig `yaml:"redis"`
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

type RedisConfig struct {
	Addrs        []string `yaml:"addrs"`        // 集群地址
	Password     string   `yaml:"password"`     // 密码
	DB           int      `yaml:"db"`           // 数据库
	PoolSize     int      `yaml:"poolSize"`     // 连接池
	MinIdleConns int      `yaml:"minIdleConns"` // 最小连接数
}

// 用户合成配置
type rawJudgeConfig struct {
	App   *rawAppConfig   `yaml:"app"`
	Rpc   *rawRpcConfig   `yaml:"rpc"`
	Redis *rawRedisConfig `yaml:"redis"`
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

type rawRedisConfig struct {
	Addrs        *[]string `yaml:"addrs"`
	Password     *string   `yaml:"password"`
	DB           *int      `yaml:"db"`
	PoolSize     *int      `yaml:"poolSize"`
	MinIdleConns *int      `yaml:"minIdleConns"`
}
