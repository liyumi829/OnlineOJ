package pick

import (
	"online-oj/gateway/internal/rpc/node"
	"sync/atomic"

	"go.uber.org/zap"
)

// RoundRobinPicker 基于轮询的选择器
type RoundRobinPicker struct {
	counter uint64
}

// NewRoundRobinPicker 创建轮询选择器
func NewRoundRobinPicker() *RoundRobinPicker {
	return &RoundRobinPicker{}
}

// Pick 从 heartbeat healthy 且 breaker allow 的节点中选择一个
func (p *RoundRobinPicker) Pick(nodes []*node.JudgeNode, excluded map[string]struct{}) (*node.JudgeNode, error) {
	if len(nodes) == 0 {
		return nil, ErrNoAvailableNode
	}

	// rr轮转，首先选取一个节点
	nodesLen := len(nodes)
	var node *node.JudgeNode
	for i := 0; i < nodesLen; i++ {
		idx := atomic.AddUint64(&p.counter, 1)      // 轮转
		pickNode := nodes[(idx-1)%uint64(nodesLen)] // 选择一个健康的节点
		if pickNode == nil {                        // 节点为空
			continue
		}
		if excluded != nil {
			if _, ok := excluded[pickNode.Addr]; ok {
				continue // 排除的节点
			}
		}
		if err := pickNode.AllowBusinessRequest(); err != nil {
			continue
		}
		// 选择成功
		// 健康、没有被排除、允许业务处理的节点
		node = pickNode
		break
	}

	zap.L().Debug("pick a node", zap.String("addr", node.Addr), zap.String("breaker_stat", node.CircuitBreaker.State().String()))
	return node, nil
}
