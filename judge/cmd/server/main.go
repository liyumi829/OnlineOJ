package main

import (
	"flag"
	"net"
	"online-oj/judge/internal/config"
	"online-oj/judge/internal/execute"
	pkglogger "online-oj/pkg/logger"

	"go.uber.org/zap"
)

var (
	// 配置文件路径（敏感配置和路径配置）
	configPath      = flag.String("config", "judge/config/judge.yaml", "judge config file path")
	configLocalPath = flag.String("config-local", "judge/config/judge.local.yaml", "judge local config file path")
	// 内部配置
	mode         = flag.String("m", "debug", "mode: debug/prod")
	host         = flag.String("h", "127.0.0.1", "judge listen host")
	port         = flag.String("p", "10000", "judge listen port")
	instanceName = flag.String("name", "judge", "instance name(storage log dir name)")
	id           = flag.Uint64("id", 1, "instance id")
)

func main() {
	// 解析参数、配置
	flag.Parse()                                           // 命令行参数
	cfg, err := config.Load(*configPath, *configLocalPath) // 读取配置文件（日志路径 / 临时目录）
	if err != nil {
		zap.L().Fatal("load config failed", zap.String("error", err.Error()))
	}
	zap.L().Info("Configuration file loaded successfully.",
		zap.String("logPath", cfg.App.LogPath),
		zap.String("tempPath", cfg.App.TempPath))
	// 初始化日志
	pkglogger.InitLogger(*mode, cfg.App.LogPath, *instanceName, *id)

	addr := net.JoinHostPort(*host, *port)

	zap.L().Info("judge grpc server starting",
		zap.String("addr", addr),
		zap.String("logPath", cfg.App.LogPath),
		zap.String("tempPath", cfg.App.TempPath),
		zap.Uint64("instanceId", *id),
	)

	if err := execute.StartGRPCServer(addr, &cfg.App); err != nil {
		zap.L().Fatal("server failed to start", zap.String("error", err.Error()))
	}
}
