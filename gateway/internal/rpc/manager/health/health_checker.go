// 实现检测节点的健康状态
package health

import (
	"context"
	pb "online-oj/api/proto/judge"
	"online-oj/gateway/internal/rpc/node"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PingHealthChecker 是基于 Ping RPC 的心跳检测器
type PingHealthChecker struct {
	nodes    []*node.JudgeNode // 节点集合
	timeout  time.Duration     // 健康检测超时时间
	interval time.Duration     // 健康检测间隔时间
}

// NewPingHealthChecker 创建心跳检测器
func NewPingHealthChecker(nodes []*node.JudgeNode, timeout, interval time.Duration) *PingHealthChecker {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &PingHealthChecker{
		nodes:    nodes,
		timeout:  timeout,
		interval: interval,
	}
}

// Start 启动定时轮询
func (h *PingHealthChecker) Start(ctx context.Context) {
	// 启动后先立即做一次全量检查
	h.CheckAllNodes(ctx)

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("[rpc][heartbeat] checker stopped")
			return
		case <-ticker.C:
			h.CheckAllNodes(ctx)
		}
	}
}

// CheckAllNodes 检查全部节点
func (h *PingHealthChecker) CheckAllNodes(ctx context.Context) {
	var wg sync.WaitGroup // 等待检测任务完成
	wg.Add(len(h.nodes))

	for _, node := range h.nodes {
		currentNode := node // 赋值防止闭包引用
		go func() {
			defer wg.Done()
			h.checkOne(ctx, currentNode) // 检测当前节点
		}()
	}

	wg.Wait() // 等待所有检测完成
}

// checkOne 检查单个节点
func (h *PingHealthChecker) checkOne(ctx context.Context, node *node.JudgeNode) {
	if node == nil || node.Client == nil {
		return
	}

	probeCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	resp, err := node.Client.Ping(probeCtx, &pb.PingRequest{})
	if err != nil {
		node.MarkHeartbeatFailure(err)
		zap.L().Warn("[rpc][heartbeat] ping failed",
			zap.String("addr", node.Addr),
			zap.String("node info", node.String()),
			zap.String("error", err.Error()))
		return
	}

	if resp == nil || !resp.GetOk() {
		node.MarkHeartbeatFailure(errInvalidPingResponse) // 标记心跳失败
		zap.L().Warn("[rpc][heartbeat] ping invalid response",
			zap.String("addr", node.Addr),
			zap.String("instance_name", resp.GetNodeId()),
			zap.String("message", resp.GetMessage()))
		return
	}

	node.MarkHeartbeatSuccess() // 标记心跳成功
	// zap.L().Debug("[rpc][heartbeat] ping success",
	// 	zap.String("addr", node.Addr),
	// 	zap.String("instance_name", resp.GetNodeId()),
	// 	zap.String("time", time.UnixMilli(resp.GetTimestampMs()).Format("2006-01-02 15:04:05")),
	// 	zap.String("message", resp.Message))
}
