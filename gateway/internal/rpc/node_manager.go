// 连接管理。实现全局管理所有 judge 节点
package rpc

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"time"

	"go.uber.org/zap"
)

// Picker 选择节点接口
type Picker interface {
	Pick(nodes []*JudgeNode) (*JudgeNode, error) // 选择节点
}

// HealthChecker 定义心跳检测器接口
type HealthChecker interface {
	Start(ctx context.Context)         // Start 启动后台心跳检测
	CheckAllNodes(ctx context.Context) // CheckAllNodes 立即主动检查所有节点
}
type JudgeNodeManager struct {
	nodes            []*JudgeNode  // 节点集合
	requestTimeout   time.Duration // 调用rpc超时时间
	picker           Picker        // 节点选择器
	healthChecker    HealthChecker // 心跳检测器
	failureThreshold int           // rpc调用错误阈值
	counter          uint64        // 轮转计数器 负载均衡采用 rr 轮转方式 -- 配合原子操作防止多个请求打到同一个节点
}

// NewJudgeNodeManager 创建 judge 调用管理者。内部自动创建节点
func NewJudgeNodeManager(ctx context.Context, cfg *Config) (*JudgeNodeManager, error) {
	if len(cfg.Addrs) == 0 {
		return nil, errors.New("The address information in the configuration is empty")
	}
	// 创建 JudgeNode
	nodes := make([]*JudgeNode, 0, len(cfg.Addrs)) // 避免扩容消耗性能
	for _, addr := range cfg.Addrs {               // 循环创建rpc客户端
		newNode, err := dialJudgeNode(addr)
		if err != nil {
			zap.L().Warn("[rpc] The current rpc node is unavailable.", zap.String("host:port", addr))
			continue
		}
		zap.L().Info("[rpc] The current rpc node is available.", zap.String("host:port", addr))
		nodes = append(nodes, newNode) // 新增连接
	}
	if len(nodes) == 0 {
		zap.L().Error("[rpc] The current RPC client cannot connect to any node.")
		return nil, errors.New("The current RPC client cannot connect to any node")
	}
	// 创建心跳检测器
	healthChecker := NewPingHealthChecker(
		nodes,
		cfg.HeartbeatTimeout,
		cfg.HeartbeatInterval)
	// 创建节点选择器
	nodesPicker := NewHeartbeatPicker()
	m := &JudgeNodeManager{
		nodes:            nodes,
		requestTimeout:   cfg.RequestTimeout,
		healthChecker:    healthChecker,
		picker:           nodesPicker,
		failureThreshold: cfg.FailureThreshold,
		counter:          0,
	}

	// 创建manager完成
	go m.healthChecker.Start(ctx) // 启动心跳检测

	return m, nil
}

// Judge 实现接口。实现调用方和实现方解耦合
func (m *JudgeNodeManager) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if m == nil {
		return nil, errors.New("multi judge client is nil")
	}
	if req == nil {
		zap.L().Error("judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	// 选择节点
	node, err := m.picker.Pick(m.nodes) // 选择一个健康的节点
	if err != nil {
		return nil, err
	}
	// 选择到一个节点，然后节点进行调用
	rpcCtx, cancel := context.WithTimeout(ctx, m.requestTimeout)
	defer cancel()

	resp, err := node.Client.Judge(rpcCtx, req)
	if err != nil {
		node.MarkBizFailure(err) // 业务调用失败
		return nil, fmt.Errorf("judge rpc failed, addr=%s, err=%w", node.Addr, err)
	}

	node.MarkBizSuccess() // 记录业务成功
	return resp, nil
}

// Close 关闭所有底层 gRPC 连接
func (m *JudgeNodeManager) Close() error {
	if m == nil {
		return nil
	}

	var firstErr error
	for _, node := range m.nodes {
		if node == nil || node.Conn == nil {
			continue
		}
		if err := node.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
