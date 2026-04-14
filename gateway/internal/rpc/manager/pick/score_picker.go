package pick

import (
	"math"
	"sync/atomic"
	"time"

	"online-oj/gateway/internal/rpc/node"
)

// ScorePicker 是第二版加强型负载均衡器。
// 核心策略：
//  1. 先过滤不可用节点（excluded / heartbeat / breaker）
//  2. 对候选节点进行综合打分
//  3. 分数越低，优先级越高
//  4. 若分数相同或极接近，则轮询打散
//
// 当前评分项：
// - activeRequests 惩罚
// - 平均耗时惩罚
// - 失败率惩罚
// - warm-up 惩罚
type ScorePicker struct {
	counter uint64            // 轮询选择器
	cfg     ScorePickerConfig // 权重配置
}

// NewScorePicker 创建 ScorePicker
func NewScorePicker(cfg ScorePickerConfig) *ScorePicker {
	cfg.SetDefault()
	return &ScorePicker{
		cfg: cfg,
	}
}

// Name 返回选择器名称
func (p *ScorePicker) Name() string {
	return "score_picker"
}

// Pick 从节点中选择综合得分最低的节点
func (p *ScorePicker) Pick(nodes []*node.JudgeNode, excluded map[string]struct{}) (*node.JudgeNode, error) {
	if len(nodes) == 0 {
		return nil, ErrNoAvailableNode
	}

	// 节点选择
	type scoredNode struct {
		node  *node.JudgeNode
		score float64
	}

	// 候选节点
	candidates := make([]scoredNode, 0, len(nodes))

	for _, n := range nodes {
		if n == nil {
			continue
		}

		// 过滤本轮已尝试节点
		if excluded != nil {
			if _, ok := excluded[n.Addr]; ok {
				continue
			}
		}

		// 必须通过健康和熔断准入
		if err := n.AllowBusinessRequest(); err != nil {
			continue
		}

		score := p.calculateScore(n) // 计算当前节点评分
		candidates = append(candidates, scoredNode{
			node:  n,
			score: score,
		}) // 添加到候选节点
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableNode
	}

	// 下面进行节点选择
	bestScore := candidates[0].score
	bestNodes := make([]*node.JudgeNode, 0, len(candidates))

	for _, item := range candidates {
		switch {
		case item.score < bestScore:
			// 选择得分最低的节点
			bestScore = item.score
			bestNodes = bestNodes[:0]
			bestNodes = append(bestNodes, item.node)
		case nearlyEqual(item.score, bestScore):
			// 近似判断浮点数相等
			bestNodes = append(bestNodes, item.node)
		}
	}
	// 评分相同的节点随机选择
	idx := atomic.AddUint64(&p.counter, 1)
	return bestNodes[(idx-1)%uint64(len(bestNodes))], nil
}

// calculateScore 计算综合惩罚分，分数越低越优先
//
//   - 活跃请求越多，越差
//   - 平均耗时越高，越差
//   - 失败率越高，越差
//   - 处于 warm-up，越差
func (p *ScorePicker) calculateScore(n *node.JudgeNode) float64 {
	metrics := n.MetricsSnapshot()

	activePenalty := p.calculateActivePenalty(metrics.ActiveRequests)
	latencyPenalty := p.calculateLatencyPenalty(metrics.AvgLatency)
	failurePenalty := p.calculateFailureRatePenalty(metrics.TotalRequests, metrics.FailedRequests)
	warmupPenalty := p.calculateWarmupPenalty(n.InWarmup())
	return activePenalty + latencyPenalty + failurePenalty + warmupPenalty
}

// calculateActivePenalty 计算活跃请求惩罚
func (p *ScorePicker) calculateActivePenalty(active int64) float64 {
	if active <= 0 {
		return 0
	}
	return float64(active) * p.cfg.ActiveWeight
}

// calculateLatencyPenalty 计算平均耗时惩罚
func (p *ScorePicker) calculateLatencyPenalty(avgLatency time.Duration) float64 {
	if avgLatency <= 0 || p.cfg.LatencyBase <= 0 {
		return 0
	}

	// 比率：平均延迟/基准延迟
	ratio := float64(avgLatency) / float64(p.cfg.LatencyBase)

	// 线性惩罚，延迟越低惩罚分越低
	return ratio * p.cfg.LatencyWeight
}

// calculateFailureRatePenalty 计算失败率惩罚
func (p *ScorePicker) calculateFailureRatePenalty(totalRequests, failedRequests uint64) float64 {
	if totalRequests == 0 || failedRequests == 0 {
		return 0
	}

	// 失败比率
	failureRate := float64(failedRequests) / float64(totalRequests)
	// 样本不足时降低影响，避免新节点冷启动被误伤
	sampleFactor := 1.0
	// 冷启动 --> 没有多少总请求导致的失败比例大大增加
	// 减少失败比率
	if totalRequests < p.cfg.MinSampleCount {
		sampleFactor = float64(totalRequests) / float64(p.cfg.MinSampleCount)
	}

	return failureRate * p.cfg.FailureRateWeight * sampleFactor
}

// calculateWarmupPenalty 计算 warm-up 惩罚
func (p *ScorePicker) calculateWarmupPenalty(inWarmup bool) float64 {
	if !inWarmup {
		return 0
	}
	return p.cfg.WarmupPenalty
}

// nearlyEqual 判断两个浮点数是否近似相等
func nearlyEqual(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) <= epsilon
}
