package main

import (
	"online-oj/judge/internal/execute"
	pkg "online-oj/pkg/logger"

	"go.uber.org/zap"
)

func init() {
	config := pkg.Config{
		Id:           1,
		InstanceName: "gRpcServer",
		Mode:         "prod",
		StoragePath:  "../../../logs",
	}
	pkg.InitLogger(config)
}

func main() {
	err := execute.StartGRPCServer("127.0.0.1:8080", "./temp")
	if err != nil {
		zap.L().Fatal("Server failed to start.", zap.String("error", err.Error()))
	}
}
