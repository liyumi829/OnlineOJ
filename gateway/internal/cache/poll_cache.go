//   1. 只缓存轮询必要字段
//   2. 不缓存完整判题结果
//   3. 减少前端高频轮询对 DB 的压力
//   4. 提供空值缓存，防止 submission_id 穿透

package cache

import (
	"online-oj/gateway/internal/common"
	"sync"
	"time"
)

// PollCache Gateway 本地缓存
type PollCache struct {
	mu       sync.RWMutex                    // 保护缓存资源
	items    map[string]*SubmissionPollState // 缓存信息
	stopCh   chan struct{}                   // 停止信息
	stopOnce sync.Once
}

// NewPollCache 创建轮询缓存，并启动清理协程
func NewPollCache(cleanInterval time.Duration) *PollCache {
	if cleanInterval <= 0 {
		cleanInterval = 10 * time.Second
	}

	c := &PollCache{
		items:  make(map[string]*SubmissionPollState),
		stopCh: make(chan struct{}),
	}
	go c.cleanupLoop(cleanInterval) // 启动清理缓存线程 每10s进行清除
	return c
}

// Close 停止后台清理协程
func (c *PollCache) Close() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// BuildPollKey 构造缓存 key
func BuildPollKey(submissionID string) string {
	return "submission_poll:" + submissionID
}

// Get 获取缓存
func (c *PollCache) Get(submissionID string) (*SubmissionPollState, bool) {
	key := BuildPollKey(submissionID)

	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || item == nil {
		return nil, false
	}

	// 过期则删除并返回未命中
	if time.Now().After(item.ExpireAt) {
		c.Delete(submissionID)
		return nil, false
	}

	cp := *item // 拷贝一份，避免返回 map 中的指针元素
	return &cp, true
}

// Set 设置缓存
func (c *PollCache) Set(state *SubmissionPollState, ttl time.Duration) {
	if state == nil || state.SubmissionID == "" {
		return
	}
	if ttl <= 0 {
		ttl = PollTTLQueued
	}

	now := time.Now()
	cp := *state
	cp.UpdatedAt = now
	cp.ExpireAt = now.Add(ttl)

	key := BuildPollKey(state.SubmissionID)

	c.mu.Lock()
	c.items[key] = &cp
	c.mu.Unlock()
}

// Delete 删除缓存
func (c *PollCache) Delete(submissionID string) {
	key := BuildPollKey(submissionID)

	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// SetQueued 设置排队态
func (c *PollCache) SetQueued(submissionID string, nextPollAfterMS int64) {
	if nextPollAfterMS <= 0 {
		nextPollAfterMS = common.NextPollAfterMS2
	}
	c.Set(&SubmissionPollState{
		SubmissionID:    submissionID,
		Phase:           PhaseQueued,
		Done:            false,
		Polling:         true,
		NextPollAfterMS: nextPollAfterMS,
	}, PollTTLQueued)
}

// SetJudging 设置判题中
func (c *PollCache) SetRunning(submissionID string, nextPollAfterMS int64) {
	if nextPollAfterMS <= 0 {
		nextPollAfterMS = common.NextPollAfterMS1
	}
	c.Set(&SubmissionPollState{
		SubmissionID:    submissionID,
		Phase:           PhaseRunning,
		Done:            false,
		Polling:         true,
		NextPollAfterMS: nextPollAfterMS,
	}, PollTTLJudging)
}

// SetFinished 设置完成态
func (c *PollCache) SetAccpted(submissionID string) {
	c.Set(&SubmissionPollState{
		SubmissionID:    submissionID,
		Phase:           PhaseAccpted,
		Done:            true,
		Polling:         false,
		NextPollAfterMS: 0,
	}, PollTTLFinished)
}

// SetFailed 设置失败态
func (c *PollCache) SetFailed(submissionID string, nextPollAfterMS int64) {
	c.Set(&SubmissionPollState{
		SubmissionID:    submissionID,
		Phase:           PhaseFailed,
		Done:            true,
		Polling:         false,
		NextPollAfterMS: nextPollAfterMS,
	}, PollTTLFailed)
}

// SetNotFound 设置空值缓存，防止穿透
func (c *PollCache) SetNotFound(submissionID string) {
	c.Set(&SubmissionPollState{
		SubmissionID:    submissionID,
		Phase:           PhaseFailed,
		Done:            true,
		Polling:         false,
		NextPollAfterMS: 0,
		NotFound:        true,
	}, PollTTLEmpty)
}

// cleanupLoop 定时清理过期缓存
func (c *PollCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for key, item := range c.items {
				if item == nil || now.After(item.ExpireAt) {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}
