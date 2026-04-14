// 实现rpc调用的单个节点
package node

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/rpc/node/breaker"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// rpc 调用的单个节点
type JudgeNode struct {
	Addr   string                // 节点地址
	conn   *grpc.ClientConn      // 节点连接
	client pb.JudgeServiceClient // 连接客户端

	// 健康状态
	health *HealthyStatus

	// 业务状态
	business *BusinessStatus

	// 运行时指标
	metrics *RuntimeMetrics

	// 预热状态

	warmupMu    sync.Mutex // warm-up 状态锁
	warmupUntil time.Time  // warm-up 截止时间，零值表示不在 warm-up
}

// NewJudgeNode 创建一个judge节点
//
// 默认该节点是不健康的状态需要进行ping/pong心跳检测
func NewJudgeNode(addr string, conn *grpc.ClientConn, client pb.JudgeServiceClient, breaker BreakerInvoker) *JudgeNode {
	return &JudgeNode{
		Addr:     addr,
		conn:     conn,
		client:   client,
		health:   NewHealthyStatus(),
		business: NewBusinessStatus(breaker),
		metrics:  NewRuntimeMetrics(),
	}
}

// Judge 业务处理
func (n *JudgeNode) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if req == nil {
		zap.L().Error("[rpc][node]judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	resp, err := n.client.Judge(ctx, req) // 真正调用业务逻辑
	if err != nil {
		return nil, fmt.Errorf("rpc judge call failed: %w", err)
	}
	return resp, nil
}

// Ping 进行心跳检测
func (n *JudgeNode) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	if req == nil {
		zap.L().Error("[rpc][node]ping request is nil")
		return nil, errors.New("ping request is nil")
	}

	resp, err := n.client.Ping(ctx, req) // 调用方法
	if err != nil {
		return nil, fmt.Errorf("rpc ping call failed: %w", err)
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

// AllowBusinessRequest 判断节点是否允许承接业务请求
func (n *JudgeNode) AllowBusinessRequest() error {
	if n == nil {
		return errors.New("judge node is nil")
	}
	if n.health == nil || !n.health.IsHealthy() {
		return ErrHeartbeatUnhealthy
	}
	if n.business == nil {
		return nil
	}
	return n.business.AllowRequest()
}

// MarkHeartbeatSuccess 标记心跳成功
func (n *JudgeNode) MarkHeartbeatSuccess() {
	if n == nil || n.health == nil {
		return
	}
	n.health.MarkHeartbeatSuccess()
}

// MarkHeartbeatFailure 标记心跳失败
func (n *JudgeNode) MarkHeartbeatFailure(err error) {
	if n == nil || n.health == nil {
		return
	}
	n.health.MarkHeartbeatFailure(err)
}

// StartRequest 标记请求开始
func (n *JudgeNode) StartRequest() {
	if n == nil || n.metrics == nil {
		return
	}
	n.metrics.StartRequest()
}

// FinishRequest 标记请求结束
func (n *JudgeNode) FinishRequest(success bool, timeout bool, latency time.Duration, err error, warmup time.Duration) {
	if n == nil {
		return
	}

	if n.metrics != nil {
		n.metrics.FinishRequest(success, timeout, latency)
	}

	if n.business != nil {
		before := n.business.BreakerState() // 记录当前的熔断器状态
		if success {
			n.business.MarkBizSuccess()
		} else {
			n.business.MarkBizFailure(err)
		}
		after := n.business.BreakerState() // 记录标记后的熔断器状态
		// 请求结束，如果一个请求是在 half-open状态结束的，就需要进入预热状态，不能打太多的请求到该节点上
		// HalfOpen -> Closed 成功恢复后进入 warm-up
		if before == breaker.CircuitHalfOpen && after == breaker.CircuitClosed && success && warmup > 0 {
			n.EnterWarmup(warmup) // 成功并且状态进行转变，进入预热状态
		}
	}
}

// EnterWarmup 进入 warm-up（预热状态）
func (n *JudgeNode) EnterWarmup(duration time.Duration) {
	if n == nil || duration <= 0 {
		return
	}

	n.warmupMu.Lock()
	defer n.warmupMu.Unlock()

	n.warmupUntil = time.Now().Add(duration)
}

// InWarmup 判断当前节点是否处于预热状态
func (n *JudgeNode) InWarmup() bool {
	if n == nil {
		return false
	}

	n.warmupMu.Lock()
	defer n.warmupMu.Unlock()

	return !n.warmupUntil.IsZero() && time.Now().Before(n.warmupUntil)
}

// ActiveRequests 返回当前活跃请求数
func (n *JudgeNode) ActiveRequests() int64 {
	if n == nil || n.metrics == nil {
		return 0
	}
	return n.metrics.ActiveRequests()
}

// HealthSnapshot 返回健康状态快照
func (n *JudgeNode) HealthSnapshot() HealthyStatusSnapshot {
	if n == nil || n.health == nil {
		return HealthyStatusSnapshot{}
	}
	return n.health.Snapshot()
}

// BusinessSnapshot 返回业务状态快照
func (n *JudgeNode) BusinessSnapshot() BusinessStatusSnapshot {
	if n == nil || n.business == nil {
		return BusinessStatusSnapshot{}
	}
	return n.business.Snapshot()
}

// MetricsSnapshot 返回指标快照
func (n *JudgeNode) MetricsSnapshot() RuntimeMetricsSnapshot {
	if n == nil || n.metrics == nil {
		return RuntimeMetricsSnapshot{}
	}
	return n.metrics.Snapshot()
}
