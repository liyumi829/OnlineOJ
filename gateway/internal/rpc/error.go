package rpc

import "errors"

var (
	// errInvalidPingResponse 表示 ping 返回非法响应
	errInvalidPingResponse = errors.New("invalid ping response")
)
