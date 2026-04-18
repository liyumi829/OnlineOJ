package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/redis/go-redis/v9"
)

// 职责:
//  1. 本地缓存 + Redis 二级缓存
//  2. key 必须包含完整版本信息，防止错误复用
//  3. TTL 随机抖动，防止雪崩
//  4. 支持 singleflight 防重复判题
//  5. 提供空值缓存，防止热点 key 反复回源

// ReuseCache Judge 结果复用缓存
type ReuseCache struct {
	mu       sync.RWMutex          // 互斥锁，保护缓存资源
	local    map[string]*localItem // 本地缓存
	redis    redis.UniversalClient // Redis缓存
	sf       singleflight.Group    // 单飞模式
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewReuseCache 创建结果复用缓存
func NewReuseCache(redisClient redis.UniversalClient, cleanInterval time.Duration) *ReuseCache {
	if cleanInterval <= 0 {
		cleanInterval = 1 * time.Minute // 默认每 1minute 进行一次清除
	}

	c := &ReuseCache{
		local:  make(map[string]*localItem),
		redis:  redisClient,
		stopCh: make(chan struct{}),
	}
	go c.cleanupLoop(cleanInterval) // 开启本地缓存删除
	return c
}

// BuildReuseKey 构造结果复用缓存 key
func BuildReuseKey(in ReuseKeyInput) string {
	raw := fmt.Sprintf(
		"problem_id=%d|language=%s|code=%s|testcase_version=%s|runtime_version=%s",
		in.ProblemID,
		strings.ToLower(strings.TrimSpace(in.Language)),
		in.Code,
		in.TestcaseVersion,
		in.RuntimeVersion,
	)

	sum := sha256.Sum256([]byte(raw))
	return "judge_reuse:" + hex.EncodeToString(sum[:])
}

// Get 先查本地，再查 Redis
func (c *ReuseCache) Get(ctx context.Context, key string) (*CachedJudgeResult, bool, error) {
	// 1. 本地缓存
	c.mu.RLock()
	item, ok := c.local[key]
	c.mu.RUnlock()
	if ok && item != nil && time.Now().Before(item.ExpireAt) {
		// 存在且没有没有过期
		cp := *item.Value
		return &cp, true, nil
	}

	// 2. Redis 缓存
	if c.redis == nil {
		return nil, false, nil
	}

	raw, err := c.redis.Get(ctx, key).Bytes() // 获取Redis中存储的数据
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var result CachedJudgeResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, false, err
	}

	// 回填本地缓存
	c.setLocal(key, &result, LocalReuseTTL) // 设置到本地缓存

	return &result, true, nil
}

// Set 写本地缓存 + Redis
func (c *ReuseCache) Set(ctx context.Context, key string, value *CachedJudgeResult) error {
	if key == "" || value == nil {
		return nil
	}

	// 1. 本地缓存
	c.setLocal(key, value, LocalReuseTTL)

	// 2. Redis 缓存
	if c.redis == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	ttl := withJitter(RedisReuseTTL)              // TTL 设置抖动，防止雪崩
	return c.redis.Set(ctx, key, data, ttl).Err() // 设置缓存
}

// SetEmpty 写空值缓存，防止热点 key 反复回源
func (c *ReuseCache) SetEmpty(ctx context.Context, key string) error {
	v := &CachedJudgeResult{
		Empty: true,
	}
	c.setLocal(key, v, ReuseEmptyTTL)

	if c.redis == nil {
		return nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, b, ReuseEmptyTTL).Err()
}

// Delete 删除缓存
func (c *ReuseCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	delete(c.local, key)
	c.mu.Unlock()

	if c.redis != nil {
		if err := c.redis.Del(ctx, key).Err(); err != nil && !errors.Is(err, redis.Nil) {
			return err
		}
	}
	return nil
}

// DoWithSingleflight 防止相同 key 并发重复判题
func (c *ReuseCache) DoWithSingleflight(
	key string,
	fn func() (*CachedJudgeResult, error),
) (*CachedJudgeResult, error, bool) {
	// shared：是不是共享别人的结果
	//= true → 你没有执行 fn ()，是等别人的结果
	//= false → 你是那个执行 fn () 的人
	value, err, shared := c.sf.Do(key, func() (any, error) {
		return fn() // 单飞模式进行判题
	})
	if err != nil {
		return nil, err, shared
	}

	result, ok := value.(*CachedJudgeResult)
	if !ok {
		return nil, errors.New("invalid cached judge result type"), shared
	}
	return result, nil, shared
}

// Close 关闭清理协程
func (c *ReuseCache) Close() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// setLocal 写本地缓存
func (c *ReuseCache) setLocal(key string, value *CachedJudgeResult, ttl time.Duration) {
	if key == "" || value == nil {
		return
	}

	cp := *value // 拷贝一份

	c.mu.Lock()
	c.local[key] = &localItem{
		Value:    &cp,                 // 将拷贝值写入本地缓存
		ExpireAt: time.Now().Add(ttl), // TTL时间
	}
	c.mu.Unlock()
}

// cleanupLoop 定时清理过期本地缓存
func (c *ReuseCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for key, item := range c.local {
				if item == nil || now.After(item.ExpireAt) {
					delete(c.local, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

// withJitter TTL 抖动，防止雪崩
func withJitter(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	n := rand.Int63n(int64(MaxTTLJitter) + 1) // 生成一个随机抖动的数据
	return base + time.Duration(n)
}
