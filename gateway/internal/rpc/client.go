package rpc

import (
	"context"
	"fmt"
	pb "online-oj/api/proto/judge"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 是第一版单节点 Judge RPC Client。
type Client struct {
	conn    *grpc.ClientConn      // 连接器
	client  pb.JudgeServiceClient // 客户端
	timeout time.Duration         // 调用超时时间
}

// New 创建单节点 RPC Client。
// 把最基础链路跑通。
func NewClinet(ctx context.Context, cfg Config) (*Client, error) {
	cfg.setDefault()

	if cfg.Addr == "" {
		zap.L().Error("judge server address is empty")
		return nil, fmt.Errorf("judge server address is empty")
	}

	// 新版 gRPC 推荐使用 grpc.NewClient。
	conn, err := grpc.NewClient(
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, fmt.Errorf("dial judge server failed: %w", err)
	}

	return &Client{
		conn:    conn,
		client:  pb.NewJudgeServiceClient(conn),
		timeout: cfg.RequestTimeout,
	}, nil
}

// Judge 发起一次判题调用。
func (c *Client) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if req == nil {
		zap.L().Error("judge request is nil")
		return nil, fmt.Errorf("judge request is nil")
	}

	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.Judge(rpcCtx, req)
	if err != nil {
		return nil, fmt.Errorf("rpc judge call failed: %w", err)
	}

	return resp, nil
}

// Close 关闭底层连接。
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
