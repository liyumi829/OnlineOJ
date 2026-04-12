package pick

import (
	"sync/atomic"

	"online-oj/gateway/internal/rpc/node"
)

// LeastActivePicker 是负载均衡器：
//  1. 只从健康且允许接流量的节点中选择
//  2. 在候选节点中选择 activeRequests 最少的节点
//  3. 若最小 active 相同，则使用轮询打散，避免固定顺序偏置
type LeastActivePicker struct {
	counter uint64
}

// NewLeastActivePicker 创建最少活跃请求选择器
func NewLeastActivePicker() *LeastActivePicker {
	return &LeastActivePicker{}
}

// Name 返回选择器名称
func (p *LeastActivePicker) Name() string {
	return "least_active"
}

// Pick 选择最合适的节点
func (p *LeastActivePicker) Pick(nodes []*node.JudgeNode, excluded map[string]struct{}) (*node.JudgeNode, error) {
	if len(nodes) == 0 {
		return nil, ErrNoAvailableNode
	}

	normalCandidates := make([]*node.JudgeNode, 0, len(nodes))

	for _, n := range nodes {
		if n == nil {
			continue
		}

		// 跳过本轮已尝试过的节点
		if excluded != nil {
			if _, ok := excluded[n.Addr]; ok {
				continue
			}
		}

		// 必须通过健康检查和熔断器准入
		if err := n.AllowBusinessRequest(); err != nil {
			continue
		}
		// 通过选择
		normalCandidates = append(normalCandidates, n)
	}

	if len(normalCandidates) > 0 {
		return p.pickLeastActive(normalCandidates), nil
	}

	return nil, ErrNoAvailableNode
}

// pickLeastActive 从给定节点集合中挑选 activeRequests 最少的节点
func (p *LeastActivePicker) pickLeastActive(nodes []*node.JudgeNode) *node.JudgeNode {
	// 调用保证非空
	minActive := nodes[0].ActiveRequests()
	best := make([]*node.JudgeNode, 0, len(nodes))

	for _, n := range nodes {
		active := n.ActiveRequests()

		switch {
		case active < minActive:
			minActive = active
			best = best[:0]
			best = append(best, n)
		case active == minActive:
			best = append(best, n)
		}
	}

	// 同 active 时轮询打散
	idx := atomic.AddUint64(&p.counter, 1)
	return best[(idx-1)%uint64(len(best))]
}
