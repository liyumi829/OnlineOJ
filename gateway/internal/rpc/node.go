package rpc

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 实现rpc调用的单个节点

// rpc 调用的单个节点
type JudgeNode struct {
	addr   string                // 节点地址
	conn   *grpc.ClientConn      // 节点连接
	client pb.JudgeServiceClient // 连接客户端
}

func NewJudgeNode(addr string) (*JudgeNode, error) {
	if addr == "" {
		// 如果节点地址为空
		return nil, errors.New("addr is empty...")
	}
	clientConn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.New("dial judge server failed")
	}
	// 创建连接成功
	return &JudgeNode{
		addr:   addr,
		conn:   clientConn,
		client: pb.NewJudgeServiceClient(clientConn),
	}, nil
}

func (n *JudgeNode) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if req == nil {
		zap.L().Error("judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	resp, err := n.client.Judge(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("rpc judge call failed: %w", err)
	}
	return resp, nil
}

// Close 关闭底层连接。
func (n *JudgeNode) Close() error {
	if n.conn == nil {
		return nil
	}
	return n.conn.Close()
}
