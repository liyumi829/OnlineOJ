// 实现连接 rpc 节点
package node

import (
	"errors"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/rpc/node/breaker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DialJudgeNode(cfg *JudgeNodeConfig) (*JudgeNode, error) {
	cfg.SetDefault() // 设置默认值
	if cfg.Addr == "" {
		// 如果节点地址为空
		return nil, errors.New("addr is empty...")
	}
	clientConn, err := grpc.NewClient(
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, errors.New("dial judge server failed")
	}
	// 构造JudgeNode所需其他字段
	// clientConn
	judgeClient := pb.NewJudgeServiceClient(clientConn)                    // 创建 judge 客户端
	circuitBreaker := breaker.NewCircuitBreaker(&cfg.CircuitBreakerConfig) // 创建熔断器
	return NewJudgeNode(cfg.Addr, clientConn, judgeClient, circuitBreaker), nil
}
