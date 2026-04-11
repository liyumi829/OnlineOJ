package node

import "errors"

var (
	// ErrHeartbeatUnhealthy 表示节点心跳不健康
	ErrHeartbeatUnhealthy = errors.New("node heartbeat is unhealthy")
)
