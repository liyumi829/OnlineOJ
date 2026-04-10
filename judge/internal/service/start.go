package service

import (
	"net"
	"online-oj/api/proto/judge"
	"online-oj/judge/internal/config"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer 启动gRPC服务端
func StartGRPCServer(cfg config.JudgeConfig) error {
	lis, err := net.Listen("tcp", cfg.Rpc.Addr)
	if err != nil {
		return err
	}

	srv, err := NewJudgeServer(cfg) // 创建服务实例
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	judge.RegisterJudgeServiceServer(s, srv)

	// 优雅关闭（监听系统信号）
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		zap.L().Info("shutting down gRPC server...")
		s.GracefulStop()
	}()

	zap.L().Info("gRPC server started...",
		zap.String("addr", cfg.Rpc.Addr),
		zap.String("instance_name", cfg.Rpc.InstanceName))
	return s.Serve(lis)
}
