package manager

import (
	"context"
	"online-oj/gateway/internal/rpc/node"
	"time"
)

// Picker 选择节点接口
type Picker interface {
	Pick(nodes []*node.JudgeNode, excluded map[string]struct{}) (*node.JudgeNode, error) // 选择节点
}

// HealthChecker 定义心跳检测器接口
type HealthChecker interface {
	Start(ctx context.Context)         // Start 启动后台心跳检测
	CheckAllNodes(ctx context.Context) // CheckAllNodes 立即主动检查所有节点
}

// RetryPolicy 是重试策略接口
type RetryPolicy interface {
	// MaxRetries 返回最大重试次数
	MaxRetries() int

	// Backoff 返回第 retryIndex 次重试前的等待时间
	Backoff(retryIndex int) time.Duration

	// Wait 等待重试退避时间
	Wait(ctx context.Context, retryIndex int) error
}
