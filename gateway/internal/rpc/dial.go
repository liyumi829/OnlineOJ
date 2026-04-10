package rpc

import (
	"errors"
	pb "online-oj/api/proto/judge"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 实现连接 rpc 节点

func dialJudgeNode(addr string) (*JudgeNode, error) {
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
	// 创建 judge 客户端
	judgeClient := pb.NewJudgeServiceClient(clientConn)
	// 创建连接成功 创建一个节点
	return NewJudgeNode(addr, clientConn, judgeClient)
}
