package worker

import (
	"context"
	"errors"
	"time"

	"online-oj/gateway/internal/model/entity"
	"online-oj/gateway/internal/repository"
)

const (
	DefaultQueueSize = 20 // 经过测试的最佳
)

// JudgeWorkerManager 负责后台调度任务
type JudgeWorkerManager struct {
	taskRepo     *repository.JudgeTaskRepository // 判题任务仓储
	workers      []*JudgeWorker                  // 任务
	queueSize    int                             // 任务队列的大小
	taskChan     chan<- *entity.JudgeTask        // 任务队列
	pollInterval time.Duration                   // 轮询间隔时间
}

// NewJudgeWorkerManager 创建 Worker Manager
//
// 说明:
//   - workers 工作 goroutine 集合。
//   - queueSize 队列大小。默认是15大小的缓冲区
//   - pollInterval 间隔拉取的时间。默认是 1s
func NewJudgeWorkerManager(
	repo *repository.JudgeTaskRepository,
	workers []*JudgeWorker,
	queueSize int,
	pollInterval time.Duration,
) (*JudgeWorkerManager, error) {
	if repo == nil {
		return nil, errors.New("judge task repository is nil")
	}
	if len(workers) == 0 {
		return nil, errors.New("judge request tranfrom worker is empty")
	}
	if queueSize <= 0 {
		queueSize = DefaultQueueSize
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	taskChan := make(chan *entity.JudgeTask, queueSize)
	for _, worker := range workers { // manager 给 worker 通道
		worker.taskChan = taskChan
	}
	return &JudgeWorkerManager{
		taskRepo:     repo,
		workers:      workers,
		queueSize:    queueSize,
		taskChan:     taskChan,
		pollInterval: pollInterval,
	}, nil
}

// Run 启动后台调度循环
//
// 逻辑：
//  1. 启动 worker 线程进行工作，统一发布任务
//  2. 根据缓冲区数量定时拉取 PENDING 任务
func (m *JudgeWorkerManager) Run(ctx context.Context) {

	for _, worker := range m.workers { // 启动worker
		// 启动固定数量的goroutine
		go worker.startWork(ctx)
	}

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()
	// 启动时立即执行一次，降低首轮等待
	m.dispatch(ctx) // 分发任务

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.dispatch(ctx)
		}
	}
}

// dispatch 拉取任务并调度执行
func (m *JudgeWorkerManager) dispatch(ctx context.Context) {
	// 1. 计算队列剩余空间
	free := cap(m.taskChan) - len(m.taskChan)

	// 规则：剩余空闲 < 1/2 队列容量 或 小于最小拉取数 → 不拉取
	minFetchSize := cap(m.taskChan) / 2 // 空闲至少占队列一半，才去拉

	if free < minFetchSize {
		// 队列还很满，不需要拉，减少DB查询压力
		return
	}

	// 2. 批量拉取任务（一次拉满）
	tasks, err := m.taskRepo.PickPendingTasks(ctx, free)
	if err != nil || len(tasks) == 0 {
		return
	}

	// 3. 扔进队列（无阻塞，因为前面已经算好空闲）
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		case m.taskChan <- &task:
		}
	}
}
