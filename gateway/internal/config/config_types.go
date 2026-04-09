package config

// Config Gateway 最终配置
type GatewayConfig struct {
	App   AppConfig   `yaml:"app"`   // 应用配置
	MySQL MySQLConfig `yaml:"mysql"` // MySQL 配置
	RPC   RPCConfig   `yaml:"rpc"`   // RPC 配置
}

// AppConfig 应用配置
type AppConfig struct {
	LogPath  string `yaml:"log_path"`  // 日志目录
	ViewPath string `yaml:"view_path"` // view存储目录
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host      string `yaml:"host"`       // 数据库主机
	Port      int    `yaml:"port"`       // 数据库端口
	User      string `yaml:"user"`       // 数据库用户名
	Password  string `yaml:"password"`   // 数据库密码
	Database  string `yaml:"database"`   // 数据库名
	Charset   string `yaml:"charset"`    // 字符集
	ParseTime bool   `yaml:"parse_time"` // 是否解析时间
	Loc       string `yaml:"loc"`        // 时区
}

// RPCConfig RPC 配置
type RPCConfig struct {
	Addrs                 []string `yaml:"addrs"`                   // Judge节点地址
	RequestTimeoutSeconds int      `yaml:"request_timeout_seconds"` // RPC 请求超时秒数
}

// 用户合成配置
// rawGatewayConfig 用于安全合并配置，全部使用指针字段避免零值污染
type rawGatewayConfig struct {
	App   *rawAppConfig   `yaml:"app"`
	MySQL *rawMySQLConfig `yaml:"mysql"`
	RPC   *rawRPCConfig   `yaml:"rpc"`
}

type rawAppConfig struct {
	LogPath  *string   `yaml:"log_path"`
	ViewPath *string   `yaml:"view_path"`
	Addrs    *[]string `yaml:"addrs"`
}

type rawMySQLConfig struct {
	Host      *string `yaml:"host"`
	Port      *int    `yaml:"port"`
	User      *string `yaml:"user"`
	Password  *string `yaml:"password"`
	Database  *string `yaml:"database"`
	Charset   *string `yaml:"charset"`
	ParseTime *bool   `yaml:"parse_time"`
	Loc       *string `yaml:"loc"`
}

type rawRPCConfig struct {
	Addrs                 *[]string `yaml:"addrs"`
	RequestTimeoutSeconds *int      `yaml:"request_timeout_seconds"`
}
