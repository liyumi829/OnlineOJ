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
	Conn   *grpc.ClientConn      // 节点连接
	Client pb.JudgeServiceClient // 连接客户端
	mu     sync.Mutex            // 互斥锁，并发安全

	// 健康状态

	heartbeatHealthy       bool      // 健康状态
	lastHeartbeatSuccessAt time.Time // 最近一次心跳成功时间
	lastHeartbeatFailAt    time.Time // 最近一次心跳失败时间
	lastHeartbeatErr       error     // 最近一次心跳错误

	// 业务处理

	CircuitBreaker   BreakerInvoker // 熔断器
	lastBizSuccessAt time.Time      // 上次成功时间
	lastBizFailAt    time.Time      // 上次失败时间
	lastBizErr       error          // 上次错误信息
}

// NewJudgeNode 创建一个judge节点
//
// 默认该节点是不健康的状态需要进行ping/pong心跳检测
func NewJudgeNode(addr string, conn *grpc.ClientConn, client pb.JudgeServiceClient, circuitBreaker BreakerInvoker) *JudgeNode {
	now := time.Now()
	return &JudgeNode{
		Addr:                   addr,
		Conn:                   conn,
		Client:                 client,
		heartbeatHealthy:       false, // 初始未知，等待第一次心跳探测
		lastHeartbeatSuccessAt: time.Time{},
		lastHeartbeatFailAt:    time.Time{},
		lastHeartbeatErr:       nil,
		CircuitBreaker:         circuitBreaker,
		lastBizSuccessAt:       now,
		lastBizFailAt:          time.Time{},
		lastBizErr:             nil,
	}
}

// Judge 业务处理
func (n *JudgeNode) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if req == nil {
		zap.L().Error("[rpc][node]judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	resp, err := n.Client.Judge(ctx, req) // 真正调用业务逻辑
	if err != nil {
		return nil, fmt.Errorf("rpc judge call failed: %w", err)
	}
	return resp, nil
}

// Close 关闭底层连接。
func (n *JudgeNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.Conn == nil {
		return nil
	}
	return n.Conn.Close()
}

// MarkHeartbeatSuccess 标记心跳成功
func (n *JudgeNode) MarkHeartbeatSuccess() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.heartbeatHealthy = true
	n.lastHeartbeatSuccessAt = time.Now()
	n.lastHeartbeatErr = nil
}

// MarkHeartbeatFailure 标记心跳失败
func (n *JudgeNode) MarkHeartbeatFailure(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.heartbeatHealthy = false
	n.lastHeartbeatFailAt = time.Now()
	n.lastHeartbeatErr = err
}

// IsHeartbeatHealthy 返回节点心跳状态
func (n *JudgeNode) IsHeartbeatHealthy() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.heartbeatHealthy
}

// IsHealthy 返回节点的健康状态
func (n *JudgeNode) IsHealthy() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.heartbeatHealthy
}

// AllowBusinessRequest 判断该节点是否允许承接业务请求
func (n *JudgeNode) AllowBusinessRequest() error {
	if n == nil {
		return breaker.ErrCircuitOpen
	}
	if !n.IsHeartbeatHealthy() {
		return ErrHeartbeatUnhealthy
	}
	if n.CircuitBreaker == nil {
		return nil
	}
	return n.CircuitBreaker.IsAllow()
}

// MarkBizSuccess 标记业务成功
func (n *JudgeNode) MarkBizSuccess() {
	n.mu.Lock()
	n.lastBizSuccessAt = time.Now()
	n.lastBizErr = nil
	n.mu.Unlock()

	if n.CircuitBreaker != nil {
		n.CircuitBreaker.OnSuccess()
	}
}

// MarkBizFailure 标记业务处理失败
func (n *JudgeNode) MarkBizFailure(err error) {
	n.mu.Lock()

	n.lastBizErr = err
	n.lastBizFailAt = time.Now()
	n.mu.Unlock()

	if n.CircuitBreaker != nil {
		n.CircuitBreaker.OnFailure(err)
	}
}

// String() 实现 fmt.Stringer 接口，格式化输出
func (n *JudgeNode) String() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	// 健康状态显示
	healthyStr := "✅ HEALTHY"
	if !n.heartbeatHealthy {
		healthyStr = "❌ UNHEALTHY"
	}

	// 错误信息：有错误显示，没有显示 <none>
	var lastErrStr string
	if n.lastHeartbeatErr != nil {
		lastErrStr = n.lastHeartbeatErr.Error()
	} else {
		lastErrStr = "<none>"
	}

	// 时间格式化：零值显示 <nil>
	timeFmt := "2006-01-02 15:04:05"
	lastFailAtStr := n.lastHeartbeatFailAt.Format(timeFmt)
	if n.lastHeartbeatFailAt.IsZero() {
		lastFailAtStr = "<nil>"
	}
	lastSuccessAtStr := n.lastHeartbeatSuccessAt.Format(timeFmt)
	if n.lastHeartbeatSuccessAt.IsZero() {
		lastSuccessAtStr = "<nil>"
	}

	return fmt.Sprintf(
		"Status: %s | Last Error: %s | Last Fail: %s | Last Success: %s",
		healthyStr,
		lastErrStr,
		lastFailAtStr,
		lastSuccessAtStr,
	)
}
