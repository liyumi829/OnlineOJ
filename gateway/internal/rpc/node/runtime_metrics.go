package node

import (
	"sync"
	"time"
)

// RuntimeMetrics 节点运行时指标（全锁保护，强一致）
type RuntimeMetrics struct {
	mu              sync.Mutex // 互斥锁保证读取/写入的强一致性
	activeRequests  int64      // 当前活跃请求
	totalRequests   uint64     // 总请求数
	successRequests uint64     // 成功数
	failedRequests  uint64     // 失败数
	timeoutRequests uint64     // 超时数
	totalLatencyNs  uint64     // 总耗时 ns
	lastLatencyNs   uint64     // 最后一次耗时 ns
}

// NewRuntimeMetrics 创建运行时指标对象
func NewRuntimeMetrics() *RuntimeMetrics {
	return &RuntimeMetrics{}
}

// StartRequest 在请求开始时调用
func (m *RuntimeMetrics) StartRequest() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeRequests++
	m.totalRequests++
}

// FinishRequest 在请求结束时调用
func (m *RuntimeMetrics) FinishRequest(success bool, timeout bool, latency time.Duration) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeRequests--

	if success {
		m.successRequests++
	} else {
		m.failedRequests++
	}

	if timeout {
		m.timeoutRequests++
	}

	ns := uint64(latency.Nanoseconds())
	m.totalLatencyNs += ns
	m.lastLatencyNs = ns
}

// ActiveRequests 返回当前活跃请求数
func (m *RuntimeMetrics) ActiveRequests() int64 {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeRequests
}

// AvgLatency 返回平均耗时
func (m *RuntimeMetrics) AvgLatency() time.Duration {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.totalRequests == 0 {
		return 0
	}
	return time.Duration(m.totalLatencyNs / m.totalRequests)
}

// RuntimeMetricsSnapshot 指标快照
type RuntimeMetricsSnapshot struct {
	ActiveRequests  int64
	TotalRequests   uint64
	SuccessRequests uint64
	FailedRequests  uint64
	TimeoutRequests uint64
	TotalLatency    time.Duration
	LastLatency     time.Duration
	AvgLatency      time.Duration
}

// Snapshot 强一致快照
func (m *RuntimeMetrics) Snapshot() RuntimeMetricsSnapshot {
	if m == nil {
		return RuntimeMetricsSnapshot{}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var avg uint64
	if m.totalRequests > 0 {
		avg = m.totalLatencyNs / m.totalRequests
	}

	return RuntimeMetricsSnapshot{
		ActiveRequests:  m.activeRequests,
		TotalRequests:   m.totalRequests,
		SuccessRequests: m.successRequests,
		FailedRequests:  m.failedRequests,
		TimeoutRequests: m.timeoutRequests,
		TotalLatency:    time.Duration(m.totalLatencyNs),
		LastLatency:     time.Duration(m.lastLatencyNs),
		AvgLatency:      time.Duration(avg),
	}
}
