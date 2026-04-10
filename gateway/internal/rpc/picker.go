package rpc

import (
	"errors"
	"sync/atomic"

	"go.uber.org/zap"
)

// HeartbeatPicker 基于心跳状态选择节点
type HeartbeatPicker struct {
	counter uint64
}

// NewHeartbeatPicker 创建 Picker
func NewHeartbeatPicker() *HeartbeatPicker {
	return &HeartbeatPicker{}
}

// Pick 从心跳健康节点中轮询选择一个节点
func (p *HeartbeatPicker) Pick(nodes []*JudgeNode) (*JudgeNode, error) {
	if p == nil {
		return nil, errors.New("heartbeatPicker is nil")
	}
	if len(nodes) == 0 {
		return nil, errors.New("no judge nodes configured")
	}

	healthyNodes := make([]*JudgeNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if node.IsHeartbeatHealthy() {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil, errors.New("no heartbeat healthy judge node available")
	}

	idx := atomic.AddUint64(&p.counter, 1)
	healthyNode := healthyNodes[(idx-1)%uint64(len(healthyNodes))]
	zap.L().Debug("pick a healthy node", zap.String("addr", healthyNode.Addr))
	return healthyNode, nil
}
