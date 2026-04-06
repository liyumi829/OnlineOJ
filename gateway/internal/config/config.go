package config

import (
	"fmt"

	pkgconfiger "online-oj/pkg/configer"
)

// Load 加载 GateWay 配置文件，后面的文件覆盖前面的文件
func Load(paths ...string) (*GatewayConfig, error) {
	newRaw := func() *rawGatewayConfig {
		return &rawGatewayConfig{}
	}
	return pkgconfiger.Load(defaultConfig(), newRaw, mergeGatewayConfig, paths...)
}

// defaultConfig 默认配置
func defaultConfig() *GatewayConfig {
	cfg := &GatewayConfig{}
	cfg.App.LogPath = "./logs"
	cfg.App.ViewPath = "gateway/internal/web/views/*.html"

	cfg.MySQL.Host = "127.0.0.1"
	cfg.MySQL.Port = 3306
	cfg.MySQL.User = "root"
	cfg.MySQL.Password = ""
	cfg.MySQL.Database = "OnlineOj"
	cfg.MySQL.Charset = "utf8mb4"
	cfg.MySQL.ParseTime = true
	cfg.MySQL.Loc = "Local"

	cfg.RPC.RequestTimeoutSeconds = 5
	return cfg
}

// mergeGatewayConfig 按字段合并 Gateway 配置
func mergeGatewayConfig(dst *GatewayConfig, src *rawGatewayConfig) {
	if src == nil {
		return
	}

	if src.App != nil {
		if src.App.LogPath != nil {
			dst.App.LogPath = *src.App.LogPath
		}
	}

	if src.MySQL != nil {
		if src.MySQL.Host != nil {
			dst.MySQL.Host = *src.MySQL.Host
		}
		if src.MySQL.Port != nil {
			dst.MySQL.Port = *src.MySQL.Port
		}
		if src.MySQL.User != nil {
			dst.MySQL.User = *src.MySQL.User
		}
		if src.MySQL.Password != nil {
			dst.MySQL.Password = *src.MySQL.Password
		}
		if src.MySQL.Database != nil {
			dst.MySQL.Database = *src.MySQL.Database
		}
		if src.MySQL.Charset != nil {
			dst.MySQL.Charset = *src.MySQL.Charset
		}
		if src.MySQL.ParseTime != nil {
			dst.MySQL.ParseTime = *src.MySQL.ParseTime
		}
		if src.MySQL.Loc != nil {
			dst.MySQL.Loc = *src.MySQL.Loc
		}
	}

	if src.RPC != nil {
		if src.RPC.RequestTimeoutSeconds != nil {
			dst.RPC.RequestTimeoutSeconds = *src.RPC.RequestTimeoutSeconds
		}
	}
}

// MySQLDSN 生成 MySQL DSN
func (c *GatewayConfig) MySQLDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.MySQL.User,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.Database,
		c.MySQL.Charset,
		c.MySQL.ParseTime,
		c.MySQL.Loc,
	)
}
