package node

import "online-oj/gateway/internal/rpc/node/breaker"

type BreakerInvoker interface {
	IsAllow() error              // Allow 在发起业务 RPC 前调用，判断是否允许请求通过
	OnSuccess()                  // OnSuccess 在业务 RPC 成功后调用
	OnFailure(err error)         // OnFailure 在业务 RPC 失败后调用
	State() breaker.CircuitState // State 返回熔断器当前状态
}
