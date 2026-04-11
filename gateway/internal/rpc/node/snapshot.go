package node

import "time"

// HeartbeatSnapshot 心跳快照 -- 用于调试
type HeartbeatSnapshot struct {
	Healthy       bool
	LastSuccessAt time.Time
	LastFailAt    time.Time
	LastErr       error
}

// HeartbeatSnapshot 返回心跳状态快照
func (n *JudgeNode) GetHeartbeatSnapshot() HeartbeatSnapshot {
	n.mu.Lock()
	defer n.mu.Unlock()

	return HeartbeatSnapshot{
		Healthy:       n.heartbeatHealthy,
		LastSuccessAt: n.lastHeartbeatSuccessAt,
		LastFailAt:    n.lastHeartbeatFailAt,
		LastErr:       n.lastHeartbeatErr,
	}
}

// BizSnapshot  业务快照 -- 用于调试
type BizSnapshot struct {
	LastSuccessAt time.Time
	LastFailAt    time.Time
	LastErr       error
}

// BizSnapshot 返回业务调用快照
func (n *JudgeNode) GetBizSnapshot() BizSnapshot {
	n.mu.Lock()
	defer n.mu.Unlock()

	return BizSnapshot{
		LastSuccessAt: n.lastBizSuccessAt,
		LastFailAt:    n.lastBizFailAt,
		LastErr:       n.lastBizErr,
	}
}
