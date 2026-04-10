package service

import (
	"context"
	"online-oj/api/proto/judge"
	"time"
)

// Ping 健康检查 RPC
func (s *JudgeServer) Ping(ctx context.Context, req *judge.PingRequest) (*judge.PingResponse, error) {
	return &judge.PingResponse{
		Ok:          true,
		NodeId:      s.judgeName,
		TimestampMs: time.Now().UnixMilli(),
		Message:     "pong",
	}, nil
}
