package rpc

import (
	"context"
	"errors"
	"fmt"
	pb "online-oj/api/proto/judge"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// 实现rpc调用的单个节点

// rpc 调用的单个节点
type JudgeNode struct {
	Addr                   string                // 节点地址
	Conn                   *grpc.ClientConn      // 节点连接
	Client                 pb.JudgeServiceClient // 连接客户端
	mu                     sync.Mutex            // 互斥锁，并发安全
	heartbeatHealthy       bool                  // 健康状态
	lastHeartbeatSuccessAt time.Time             // 最近一次心跳成功时间
	lastHeartbeatFailAt    time.Time             // 最近一次心跳失败时间
	lastHeartbeatErr       error                 // 最近一次心跳错误
	bizConsecutiveErr      int                   // 连续错误次数
	lastBizSuccessAt       time.Time             // 上次成功时间
	lastBizFailAt          time.Time             // 上次失败时间
	lastBizErr             error                 // 上次错误信息
}

// NewJudgeNode 创建一个judge节点
//
// 说明:
//
// 默认该节点是不健康的状态
func NewJudgeNode(addr string, conn *grpc.ClientConn, client pb.JudgeServiceClient) (*JudgeNode, error) {
	now := time.Now()
	return &JudgeNode{
		Addr:                   addr,
		Conn:                   conn,
		Client:                 client,
		heartbeatHealthy:       false, // 初始未知，等待第一次心跳探测
		lastHeartbeatSuccessAt: time.Time{},
		lastHeartbeatFailAt:    time.Time{},
		lastHeartbeatErr:       nil,
		bizConsecutiveErr:      0,
		lastBizSuccessAt:       now,
		lastBizFailAt:          time.Time{},
		lastBizErr:             nil,
	}, nil
}

// Judge 业务处理
func (n *JudgeNode) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	if req == nil {
		zap.L().Error("judge request is nil")
		return nil, errors.New("judge request is nil")
	}

	resp, err := n.Client.Judge(ctx, req)
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

// MarkBizSuccess 标记节点请求成功 健康状态
func (n *JudgeNode) MarkBizSuccess() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.heartbeatHealthy = true // 请求成功了，说明状态就是健康的
	n.bizConsecutiveErr = 0
	n.lastBizErr = nil
	n.lastBizSuccessAt = time.Now() // 获取当前成功实践
}

// MarkBizFailure 标记节点请求失败
func (n *JudgeNode) MarkBizFailure(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.heartbeatHealthy = false // 请求成功了，说明状态就是健康的
	n.bizConsecutiveErr++
	n.lastBizErr = err
	n.lastBizFailAt = time.Now()
}

// ForceHealthy 强制恢复节点的健康状态
func (n *JudgeNode) ForceHealthy() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.heartbeatHealthy = true
	n.bizConsecutiveErr = 0
	n.lastBizSuccessAt = time.Now()
	n.lastBizErr = nil
}

// HeartbeatSnapshot 返回心跳状态快照
func (n *JudgeNode) HeartbeatSnapshot() (healthy bool, lastSuccessAt time.Time, lastFailAt time.Time, lastErr error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.heartbeatHealthy, n.lastHeartbeatSuccessAt, n.lastHeartbeatFailAt, n.lastHeartbeatErr
}

// BizSnapshot 返回业务状态快照
func (n *JudgeNode) BizSnapshot() (consecutiveErr int, lastSuccessAt time.Time, lastFailAt time.Time, lastErr error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.bizConsecutiveErr, n.lastBizSuccessAt, n.lastBizFailAt, n.lastBizErr
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
		"Status: %s | Consecutive Errors: %d | Last Error: %s | Last Fail: %s | Last Success: %s",
		healthyStr,
		n.bizConsecutiveErr,
		lastErrStr,
		lastFailAtStr,
		lastSuccessAtStr,
	)
}
