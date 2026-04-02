package main

import (
	"flag"
	"online-oj/judge/internal/execute"
	pkg "online-oj/pkg/logger"

	"go.uber.org/zap"
)

var mode = flag.String("m", "debug", "mode: debug/prod")
var logPath = flag.String("lP", "./logs", "log storage path")
var tempPath = flag.String("tp", "./temp", "exe storage temp file path")
var host = flag.String("h", "127.0.0.1", "listen host")
var port = flag.String("p", "8080", "listen port")
var instanceName = flag.String("name", "judge", "Instance Name")
var id = flag.Uint64("id", 1, "Instance Id")

func InitLogger() {
	config := pkg.Config{
		Id:           *id,
		InstanceName: *instanceName,
		Mode:         *mode,
		StoragePath:  *logPath,
	}
	pkg.InitLogger(config)
}

func main() {
	flag.Parse()
	InitLogger()
	err := execute.StartGRPCServer(*host+":"+*port, *tempPath)
	if err != nil {
		zap.L().Fatal("Server failed to start.", zap.String("error", err.Error()))
	}
}
