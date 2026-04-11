package pick

import "errors"

var (
	// ErrNoAvailableNode 表示没有可用节点
	ErrNoAvailableNode = errors.New("no available judge node")
)
