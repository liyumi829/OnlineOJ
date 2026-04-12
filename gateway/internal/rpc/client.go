// rpc客户端
package rpc

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/rpc/manager"
	"online-oj/gateway/internal/rpc/manager/health"
	pick "online-oj/gateway/internal/rpc/manager/pick"
	"online-oj/gateway/internal/rpc/manager/retry"
	"online-oj/gateway/internal/rpc/node"
	"online-oj/gateway/internal/rpc/node/breaker"
	"time"

	"go.uber.org/zap"
)

type RpcClientManager struct {
	// rpc调用超时时间
	requestTimeout time.Duration
	// 节点管理器
	nodeManager *manager.JudgeNodeManager
}

// NewRpcClientManager 创建 rpc 调用管理者。内部自动创建节点
func NewRpcClientManager(ctx context.Context, cfg *RpcClientManagerConfig) (*RpcClientManager, error) {
	cfg.SetDefault() // 合并用户配置与默认配置
	if len(cfg.Addrs) == 0 {
		return nil, errors.New("judge rpc addrs is empty")
	}

	nodes := make([]*node.JudgeNode, 0, len(cfg.Addrs))
	for _, addr := range cfg.Addrs {
		n, err := node.DialJudgeNode(&node.JudgeNodeConfig{
			Addr: addr,
			CircuitBreakerConfig: breaker.CircuitBreakerConfig{
				FailureThreshold: cfg.FailureThreshold,
				OpenTimeout:      cfg.CircuitBreakOpenTimeout,
			},
		})
		addrLogger := zap.L().With(zap.String("addr", addr))
		if err != nil {
			addrLogger.Warn("[rpc] dial judge node failed", zap.Error(err))
			continue
		}

		addrLogger.Info("[rpc] judge node connected")
		nodes = append(nodes, n)
	}

	if len(nodes) == 0 {
		zap.L().Error("[rpc]no available judge node")
		return nil, errors.New("no available judge node")
	}

	healthChecker := health.NewPingHealthChecker(
		nodes,
		cfg.HeartbeatTimeout,
		cfg.HeartbeatInterval,
	)

	picker := pick.NewLeastActivePicker()

	retryPolicy := retry.NewExponentialRetryPolicy(
		cfg.MaxRetries,
		cfg.RetryBackoffs,
	)

	nodeManager, err := manager.NewJudgeNodeManager(
		nodes,
		picker,
		healthChecker,
		retryPolicy,
	)
	if err != nil {
		return nil, err
	}

	client := &RpcClientManager{
		requestTimeout: cfg.RequestTimeout,
		nodeManager:    nodeManager,
	}

	client.nodeManager.Start(ctx)

	return client, nil
}

// Close 关闭底层连接
func (c *RpcClientManager) Close() error {
	if c == nil || c.nodeManager == nil {
		return nil
	}
	return c.nodeManager.Close()
}

// Judge 实现接口。实现调用方和实现方解耦合
func (m *RpcClientManager) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if m == nil {
		return nil, errors.New("multi judge client is nil")
	}
	if req == nil {
		zap.L().Error("judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	resp, err := m.nodeManager.InvokeJudgeWithRetry(
		ctx,
		req,
		func(parent context.Context, n *node.JudgeNode, r *pb.JudgeRequest) (*pb.JudgeResponse, error) {
			rpcCtx, cancel := context.WithTimeout(parent, m.requestTimeout)
			defer cancel()

			resp, err := n.Judge(rpcCtx, r)
			if err != nil {
				zap.L().Error("judge rpc failed",
					zap.String("addr", n.Addr),
					zap.Int64("active", n.ActiveRequests()),
					zap.String("breaker_state", n.BusinessSnapshot().BreakerState.String()),
					zap.Error(err))
				return nil, fmt.Errorf("judge rpc failed, addr=%s, breaker_state=%s: %w", n.Addr, n.BusinessSnapshot().BreakerState.String(), err)
			}

			return resp, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
