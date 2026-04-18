package service

import (
	"fmt"
	"online-oj/api/proto/judge"
	"online-oj/judge/internal/cache"
	"online-oj/judge/internal/config"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// 实现grpc调用
type JudgeServer struct {
	judge.UnimplementedJudgeServiceServer // 服务端权柄
	cache                                 *cache.ReuseCache
	judgeName                             string        // 服务实例名
	storagePath                           string        // 存储路径
	globalTimeout                         time.Duration // 一个可执行程序总的超时时间
}

// NewServer 创建一个 judge 服务
func NewJudgeServer(cfg config.JudgeConfig) (*JudgeServer, error) {
	if _, err := os.Stat(cfg.App.TempPath); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.App.TempPath, 0755); err != nil {
			return nil, fmt.Errorf("create storage path failed: %w", err)
		}
	}
	// 创建 Redis 客户端
	opt := &redis.UniversalOptions{
		Addrs:        cfg.Redis.Addrs,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	}

	client := redis.NewUniversalClient(opt)

	return &JudgeServer{
		cache:         cache.NewReuseCache(client, -1),
		judgeName:     cfg.Rpc.InstanceName,
		storagePath:   cfg.App.TempPath,
		globalTimeout: time.Duration(cfg.App.GlobalTimeout) * time.Second,
	}, nil
}
