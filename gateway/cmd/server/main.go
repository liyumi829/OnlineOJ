package main

import (
	"context"
	"flag"
	"net"
	"online-oj/gateway/internal/config"
	"online-oj/gateway/internal/control"
	"online-oj/gateway/internal/repository"
	"online-oj/gateway/internal/rpc"
	"online-oj/gateway/internal/service"
	pkglogger "online-oj/pkg/logger"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func initDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印 SQL
	})
	if err != nil {
		zap.L().Fatal("connect database failed", zap.String("error", err.Error()))
	}
	return db
}

func initLogger(mode string, logPath string, instanceName string, instanceID uint64) {
	pkglogger.InitLogger(pkglogger.Config{
		Id:           instanceID,
		InstanceName: instanceName,
		Mode:         mode,
		StoragePath:  logPath,
	})
}

func main() {
	// 解析参数和配置
	flag.Parse()                                           // 命令行参数解析
	cfg, err := config.Load(*configPath, *configLocalPath) // 读取配置文件（MySQL / Redis / 日志路径）
	if err != nil {
		panic(err)
	}
	checkConfig(cfg) // 检查密码

	// 设置全局日志模式
	initLogger(*mode, cfg.App.LogPath, *instanceName, *id) // 初始日志

	// 创建服务
	// 1. Gin
	if *mode == "prod" {
		gin.SetMode(gin.ReleaseMode) // 设置gin框架日志模式
	}
	r := gin.Default()               // 创建一个默认的Gin框架引擎实例
	r.LoadHTMLGlob(cfg.App.ViewPath) // 配置View
	// 2. rpc客户端
	addrs := splitJudgeAddr(*judgeAddrs) // 对地址进行解析
	conn, err := rpc.NewClinet(context.Background(),
		rpc.Config{
			Addr:           addrs[0], // 先写死使用第一个
			RequestTimeout: time.Duration(cfg.RPC.RequestTimeoutSeconds) * time.Second,
		},
	) // 创建连接 conn Client
	if err != nil {
		zap.L().Fatal("rpc client init failed", zap.String("error", err.Error()))
	}
	// 3. Gorm服务
	repo := repository.NewRepository(initDB(cfg.MySQLDSN())) // 初始化仓储层
	// 4. 初始化业务服务
	problemService := service.NewProblemService(repo)
	judgeService := service.NewJudgeService(conn, repo)
	// 5. 初始化控制器
	problemCtl := control.NewProblemController(problemService)
	judgeCtl := control.NewJudgeController(judgeService)

	// 注册路由
	r.GET("/", problemCtl.IndexPage)
	r.GET("/problems", problemCtl.ListPage)
	r.GET("/problem/:id", problemCtl.ProblemPage)
	r.POST("/judge/submit", judgeCtl.SubmitCode)

	addr := net.JoinHostPort(*host, *port)
	zap.L().Info("gateway server starting",
		zap.String("addr", addr),
		zap.String("judgeAddr", addrs[0]),
	)

	if err := r.Run(addr); err != nil {
		zap.L().Fatal("gateway server start failed", zap.String("error", err.Error()))
	}
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
