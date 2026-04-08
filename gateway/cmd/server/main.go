package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"online-oj/gateway/cmd/router"
	"online-oj/gateway/internal/config"
	"online-oj/gateway/internal/control"
	"online-oj/gateway/internal/repository"
	"online-oj/gateway/internal/rpc"
	"online-oj/gateway/internal/service"
	"online-oj/gateway/internal/service/worker"
	pkglogger "online-oj/pkg/logger"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	// 配置文件路径（敏感配置和路径配置）
	configPath      = flag.String("config", "gateway/config/gateway.yaml", "gateway config file path")
	configLocalPath = flag.String("config-local", "gateway/config/gateway.local.yaml", "gateway local config file path")
	// 内部配置
	mode         = flag.String("m", "debug", "mode: debug/prod")
	host         = flag.String("h", "127.0.0.1", "gateway listen host")
	port         = flag.String("p", "9000", "gateway listen port")
	instanceName = flag.String("name", "gateway", "instance name(storage log dir name)")
	id           = flag.Uint64("id", 1, "instance id")
	// 外部配置 -- 支持命令行/环境变量
	judgeAddrs = flag.String("judge-addrs", "", "judge grpc addresses, comma separated") // 现在不支持多个，方便后面进行扩展
)

func main() {
	// 根上下文：
	// 整个进程的生命周期由该 ctx 控制。
	// 当收到退出信号时，调用 cancel()，后台 WorkerManager 就会感知并停止循环。
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 解析参数和配置
	flag.Parse()                                           // 命令行参数解析
	cfg, err := config.Load(*configPath, *configLocalPath) // 读取配置文件（MySQL / Redis / 日志路径）
	if err != nil {
		panic(err)
	}
	checkConfig(cfg) // 检查密码

	// 设置全局日志模式
	pkglogger.InitLogger(*mode, cfg.App.LogPath, *instanceName, *id) // 初始日志

	// Gorm服务 连接数据库及初始化
	repo := repository.NewRepository(repository.InitDB(cfg.MySQLDSN(), *mode)) // 初始化仓储层

	// 初始化rpc客户端
	addrs := splitJudgeAddr(*judgeAddrs) // 对地址进行解析
	judgeClient, err := rpc.NewClinet(ctx,
		rpc.Config{
			Addr:           addrs[0], // 先写死使用第一个
			RequestTimeout: time.Duration(cfg.RPC.RequestTimeoutSeconds) * time.Second,
		},
	) // 创建连接 conn Client
	defer func() {
		_ = judgeClient.Close()
	}()
	if err != nil {
		zap.L().Fatal("rpc client init failed", zap.String("error", err.Error()))
	}

	// 初始化 worker，通过 worke manager 进行管理
	workerCount := runtime.NumCPU()
	workers := make([]*worker.JudgeWorker, 0, workerCount)
	for i := 0; i < workerCount; i++ {
		w := worker.NewJudgeWorker(
			"worker-"+strconv.Itoa(i+1),
			judgeClient, // 所有 Worker 共用同一个 gRPC client
			nil,         // 让manager 传入
			repository.NewSubmissionRepository(repo),
			service.NewSubmissionTaskAggregate(repo),
			repository.NewProblemRepositoty(repo),
		)
		workers = append(workers, w)
	}
	// 初始化 worker manager 统一启动和管理 worker
	workerManager, err := worker.NewJudgeWorkerManager(
		repository.NewJudgeTaskRepository(repo),
		workers,
		20,
		1*time.Second,
	)
	if err != nil {
		zap.L().Fatal("[gateway] init worker manager failed.", zap.String("error", err.Error()))
		return
	}

	go func() { // 正式启动 manager
		zap.L().Info("[gateway] worker manager started")
		workerManager.Run(ctx)
		zap.L().Info("[gateway] worker manager stopped")
	}()

	// 初始化业务服务及其对应的控制器
	problemService := service.NewProblemService(repo)
	submitService := service.NewSubmitService(repo)
	submissionQueryService := service.NewSubmissionQueryService(repo)
	problemCtl := control.NewProblemController(problemService)
	judgeCtl := control.NewJudgeController(submitService, submissionQueryService)

	// Gin 框架设置
	if *mode == "prod" {
		gin.DisableConsoleColor()
		// gin.DefaultWriter = io.Discard
		gin.SetMode(gin.ReleaseMode) // 设置gin框架日志模式
	}
	r := gin.Default()               // 创建一个默认的Gin框架引擎实例
	r.LoadHTMLGlob(cfg.App.ViewPath) // 配置View
	// 注册路由
	r.GET("/", problemCtl.IndexPage) // 根目录
	api := r.Group("/api/v1")        // 判题
	router.RegisterJudgeRoutes(api, judgeCtl)
	problem := r.Group("/problem") // 题目显示
	router.RegisterProblemRoutes(problem, problemCtl)
	// 设置服务器
	server := &http.Server{
		Addr:    net.JoinHostPort(*host, *port),
		Handler: r, // Gin 实现了 http.Handler
	}
	go func() {
		zap.L().Info("gateway server starting",
			zap.String("addr", server.Addr),
			zap.String("judgeAddr", addrs[0]))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("gateway server start failed", zap.String("error", err.Error()))
		}
	}()
	// 退出信号
	<-quit
	zap.L().Info("[gateway] shutdown signal received")

	// 发出取消信号，通知后台 WorkerManager 停止轮询与调度
	cancel()

	// 优雅关闭 HTTP 服务（最多等待 10 秒）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("[gateway] http server shutdown failed",
			zap.String("error", err.Error()))
	}

	zap.L().Info("[gateway] exited")
}

// 从环境变量读取默认值
func getEnv(key, fallback string) string {
	// 查找系统环境变量：key=环境变量名
	// val=变量值，ok=是否找到该变量
	if val, ok := os.LookupEnv(key); ok && val != "" {
		// 找到变量 且 值不为空字符串 → 返回真实环境变量值
		return val
	}
	// 没找到变量 / 变量值为空 → 返回传入的默认值(fallback)
	return fallback
}

func splitJudgeAddr(a string) []string {
	return strings.Split(a, ",;")
}

// 检查配置是否完整。主要检查第三方服务的密码配置
func checkConfig(cfg *config.GatewayConfig) {
	if cfg.MySQL.Password == "" {
		// 支持从环境变量注入
		passwordStr := getEnv("MYSQL_PASSWORD", "")
		if passwordStr == "" {
			panic("Failed to read MYSQL password from both configuration file and environment variables!")
		}
	}
}
