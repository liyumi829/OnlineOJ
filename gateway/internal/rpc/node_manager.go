package rpc

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// 连接管理。实现全局管理所有 judge 节点

type JudgeNodeManager struct {
	nodes   []*JudgeNode  // 节点集合
	timeout time.Duration // 调用rpc超时时间
	counter uint64        // 轮转计数器 负载均衡采用 rr 轮转方式 -- 配合原子操作防止多个请求打到同一个节点
}

// NewJudgeNodeManager 创建 judge 调用管理者。内部自动创建节点
func NewJudgeNodeManager(cfg Config) (*JudgeNodeManager, error) {
	if len(cfg.Addrs) == 0 {
		return nil, errors.New("The address information in the configuration is empty")
	}
	nodes := make([]*JudgeNode, 0, len(cfg.Addrs)) // 避免扩容消耗性能
	fmt.Println(cfg.Addrs)
	for _, addr := range cfg.Addrs { // 循环创建rpc客户端
		newConn, err := NewJudgeNode(addr)
		if err != nil {
			zap.L().Warn("The current rpc node is unavailable.", zap.String("host:port", addr))
			continue
		}
		zap.L().Info("The current rpc node is available.", zap.String("host:port", addr))
		nodes = append(nodes, newConn) // 新增连接
	}
	if len(nodes) == 0 {
		zap.L().Error("The current RPC client cannot connect to any node.")
		return nil, errors.New("The current RPC client cannot connect to any node")
	}

	return &JudgeNodeManager{
		nodes:   nodes,
		timeout: cfg.RequestTimeout,
		counter: 0,
	}, nil
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
	rpcCtx, cancel := context.WithTimeout(ctx, m.timeout) // 调用的超时时间
	defer cancel()

	var resp *pb.JudgeResponse
	// 选择节点
	for {
		judgeNode, err := m.pickNode()
		if err != nil {
			return nil, err
		}
		// 选择到一个节点，然后节点进行调用
		resp, err = judgeNode.Judge(rpcCtx, req)
		if err != nil {
			// 说明节点不健康
			zap.L().Error("rpc judge call failed", zap.String("addr", judgeNode.addr), zap.String("error", err.Error()))
			continue // 选择下一个
		} else {
			break
		}
	}

	return resp, nil
}

// Close 关闭所有底层 gRPC 连接
func (m *JudgeNodeManager) Close() error {
	if m == nil {
		return nil
	}

	var firstErr error
	for _, node := range m.nodes {
		if node == nil || node.conn == nil {
			continue
		}
		if err := node.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// pickNode 选择合适的节点
func (m *JudgeNodeManager) pickNode() (*JudgeNode, error) {
	if m == nil || len(m.nodes) == 0 {
		return nil, errors.New("no available judge node")
	}

	idx := atomic.AddUint64(&m.counter, 1)        // 原子递增，保证并发安全
	node := m.nodes[(idx-1)%uint64(len(m.nodes))] // 负载均衡策略
	if node == nil || node.client == nil {
		return nil, errors.New("picked judge node is nil")
	}
	zap.L().Debug("pick a node", zap.String("addr", node.addr))
	return node, nil
}
