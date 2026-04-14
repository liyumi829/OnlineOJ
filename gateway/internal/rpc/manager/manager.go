// 节点管理者
package manager

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/rpc/manager/health"
	"online-oj/gateway/internal/rpc/manager/pick"
	"online-oj/gateway/internal/rpc/manager/retry"
	"online-oj/gateway/internal/rpc/node"
	"time"

	"go.uber.org/zap"
)

// JudgeNodeManager 管理判题节点
//
//   - 选择健康节点
//   - 进行心跳检测
//   - 进行重试机制
type JudgeNodeManager struct {
	// 管理的节点集合
	nodes []*node.JudgeNode
	// 负载均衡节点选择器(RR算法)
	picker Picker
	// 心跳检测器
	healthChecker HealthChecker
	// 调用失败重试策略
	retryPolicy RetryPolicy
	// 预热时长
	warmupDuration time.Duration
}

// NewJudgeNodeManager 构造节点管理者
func NewJudgeNodeManager(
	nodes []*node.JudgeNode,
	picker Picker,
	healthChecker HealthChecker,
	retryPolicy RetryPolicy,
	warmupDuration time.Duration,
) (*JudgeNodeManager, error) {
	if len(nodes) == 0 {
		return nil, errors.New("node manager requires at least one node")
	}
	// 如果没有创建默认的
	if picker == nil {
		picker = pick.NewRoundRobinPicker()
	}
	if healthChecker == nil {
		healthChecker = health.NewPingHealthChecker(nodes, 0, 0)
	}
	if retryPolicy == nil {
		retryPolicy = retry.NewExponentialRetryPolicy(2, nil)
	}
	return &JudgeNodeManager{
		nodes:          nodes,
		picker:         picker,
		healthChecker:  healthChecker,
		retryPolicy:    retryPolicy,
		warmupDuration: warmupDuration,
	}, nil
}

// Start 启动。开始管理健康管理节点
func (m *JudgeNodeManager) Start(ctx context.Context) {
	if m == nil {
		return
	}
	if m.healthChecker == nil {
		return
	}

	go m.healthChecker.Start(ctx)
}

// InvokeJudgeWithRetry 执行 Judge RPC，并带跨节点重试
//
// 逻辑：维护一个set，尽量避免每一次重试请求打到同一个节点
//
// 注意：如果重试次数大于健康节点数，那么就需要清空set，避免无法继续重试
func (m *JudgeNodeManager) InvokeJudgeWithRetry(
	ctx context.Context,
	req *pb.JudgeRequest,
	call func(context.Context, *node.JudgeNode, *pb.JudgeRequest) (*pb.JudgeResponse, error),
) (*pb.JudgeResponse, error) {
	if m == nil {
		zap.L().Error("[rpc][manager]JudgeNodeManager is nil")
		return nil, errors.New("node manager is nil")
	}
	if req == nil {
		zap.L().Error("[rpc][manager]Judge request is nil")
		return nil, errors.New("judge request is nil")
	}
	if call == nil {
		zap.L().Error("[rpc][manager]Judge RPC call function is nil")
		return nil, errors.New("judge rpc call func is nil")
	}

	maxRetries := m.retryPolicy.MaxRetries()
	totalAttempts := maxRetries + 1
	zap.L().Info("[rpc][manager]Start judge RPC with retry",
		zap.Int("max_retries", maxRetries),
		zap.Int("total_attempts", totalAttempts))

	// excluded 用于尽量避免重试打到同一个失败节点
	excluded := make(map[string]struct{}, totalAttempts) // 记录选择过的节点
	var lastErr error                                    // 记录

	// 根据退避策略进行重试选择节点进行执行
	for attempt := 0; attempt < totalAttempts; attempt++ {
		attemptLog := zap.L().With(
			zap.Int("attempt", attempt+1),
			zap.Int("total_attempts", totalAttempts)) // 携带第几次重试的日志器

		if attempt > 0 {
			attemptLog.Debug("[rpc][manager]Waiting before next judge RPC attempt")
			if err := m.retryPolicy.Wait(ctx, attempt); err != nil { // 重试等待
				attemptLog.Error("[rpc][manager]Retry wait failed", zap.Error(err))
				return nil, err
			}
		}

		node, err := m.picker.Pick(m.nodes, excluded) // 选择节点
		if err != nil {
			// 下面进行双重判断：避免节点数少于重试次数时无法继续
			// 如果排除后找不到节点，则允许清空 excluded 再尝试一次
			if err == pick.ErrNoAvailableNode && len(excluded) > 0 {
				attemptLog.Warn("[rpc][manager]No available nodes with exclusion, clearing excluded nodes",
					zap.Int("excluded_count", len(excluded)))

				excluded = make(map[string]struct{}, totalAttempts) // 情况标记的节点
				node, err = m.picker.Pick(m.nodes, excluded)        // 重新选择节点
			}

			// 这里再次判断上面的重新选择节点是否选择成功
			if err != nil {
				if lastErr != nil {
					attemptLog.Error("[rpc][manager]No available judge node after attempts",
						zap.Error(err),
						zap.NamedError("last_error", lastErr))

					return nil, fmt.Errorf("no available node after attempts, last_err=%w", lastErr)
				}
				attemptLog.Error("Failed to pick judge node", zap.Error(err))
				return nil, err
			}
		}

		excluded[node.Addr] = struct{}{}                                 // 标记该节点已经选择过了
		attemptLog = attemptLog.With(zap.String("node_addr", node.Addr)) // 已经选择节点的地址日志器

		// 记录请求开始
		node.StartRequest()
		zap.L().Debug("[rpc][manager]Invoking judge RPC on node",
			zap.String("addr", node.Addr),
			zap.Int64("active", node.ActiveRequests()))
		startAt := time.Now()
		resp, err := call(ctx, node, req) // 执行rpc调用
		latency := time.Since(startAt)    // 调用时间
		timeout := isTimeoutError(err)    // 是否超时
		success := err == nil             // 是否成功
		node.FinishRequest(success, timeout, latency, err, m.warmupDuration)

		if err == nil {
			zap.L().Warn("[rpc][manager] invoke judge success",
				zap.String("addr", node.Addr),
				zap.Int64("active", node.ActiveRequests()))
			return resp, nil
		}
		lastErr = err

		zap.L().Warn("[rpc][manager] invoke judge failed",
			zap.String("addr", node.Addr),
			zap.Int("attempt", attempt+1),
			zap.Int("total_attempts", totalAttempts),
			zap.Int64("active", node.ActiveRequests()),
			zap.Error(err))
	}

	// 如果重试次数内没有完成，本次调用失败
	zap.L().Error("[rpc][manager]Judge RPC failed after all attempts",
		zap.Int("total_attempts", totalAttempts),
		zap.NamedError("last_error", lastErr))

	return nil, fmt.Errorf("judge rpc failed after %d attempts, last_err=%w", totalAttempts, lastErr)
}

// Close 关闭全部节点连接
func (m *JudgeNodeManager) Close() error {
	if m == nil {
		return nil
	}

	var firstErr error
	for _, n := range m.nodes {
		if n == nil {
			continue
		}
		if err := n.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
