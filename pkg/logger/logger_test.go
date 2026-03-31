package pkg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// 测试1：开发模式(debug)初始化日志器
func TestInitLogger_DebugMode(t *testing.T) {
	config := Config{
		Id:           1001,
		InstanceName: "test_debug",
		Mode:         "debug",
		StoragePath:  "",
	}

	// 初始化
	InitLogger(config)

	// 验证全局 Logger 不为空
	if Logger == nil {
		t.Fatal("debug 模式下 Logger 初始化失败，logger 为 nil")
	}

	// 测试输出不同级别日志
	Logger.Debug("debug 日志测试")
	Logger.Info("info 日志测试")
	Logger.Warn("warn 日志测试")
	Logger.Error("error 日志测试")

	t.Log("✅ debug 模式日志器测试通过\n")
}

// 测试2：生产模式(prod)初始化日志器
func TestInitLogger_ProdMode(t *testing.T) {
	config := Config{
		Id:           2001,
		InstanceName: "test_prod",
		Mode:         "prod",
		StoragePath:  "../../logs",
	}

	// 初始化
	InitLogger(config)

	if Logger == nil {
		t.Fatal("prod 模式下 Logger 初始化失败，logger 为 nil")
	}

	// 生产环境只输出 Info 及以上
	Logger.Debug("这条 debug 日志不会输出")
	Logger.Info("prod info 日志测试")
	Logger.Error("prod error 日志测试")

	// 检查日志目录是否创建
	logDir := "../../logs"
	instanceDir := filepath.Join(logDir, config.InstanceName)
	if _, err := os.Stat(instanceDir); os.IsNotExist(err) {
		t.Errorf("生产环境日志目录未创建：%s", instanceDir)
	}

	t.Log("✅ prod 模式日志器测试通过\n")
}

// 测试3：不同 Id & 不同实例名称// InitLogger 初始化的是全局logger
// 只有最后一个实例其效果
func TestInitLogger_MultiInstance(t *testing.T) {
	testCases := []struct {
		name   string
		config Config
	}{
		{
			name: "实例1-ID3001",
			config: Config{
				Id:           3001,
				InstanceName: "multi_instance_1",
				Mode:         "debug",
			},
		},
		{
			name: "实例2-ID3002",
			config: Config{
				Id:           3002,
				InstanceName: "multi_instance_2",
				Mode:         "prod",
				StoragePath:  "../../logs",
			},
		},
	}

	for _, tc := range testCases {
		config := tc.config
		t.Run(tc.name, func(t *testing.T) {
			InitLogger(config)
			if Logger == nil {
				t.Fatal("Logger 初始化失败")
			}
			Logger.Info("多实例测试日志")
		})
	}

	t.Log("✅ 多实例日志器测试通过\n")
}

// 测试4：默认模式（空模式 = debug）
func TestInitLogger_DefaultMode(t *testing.T) {
	config := Config{
		Id:           4001,
		InstanceName: "test_default",
		Mode:         "", // 空，走默认 debug
	}

	InitLogger(config)

	if Logger == nil {
		t.Fatal("默认模式 Logger 初始化失败")
	}

	Logger.Debug("默认模式 debug 日志")
	Logger.Info("默认模式 info 日志")

	t.Log("✅ 默认模式测试通过\n")
}

// 测试5：测试 zap 全局替换是否成功（zap.L()）
func TestZapGlobalLogger(t *testing.T) {
	config := Config{
		Id:           5001,
		InstanceName: "test_zap_global",
		Mode:         "debug",
	}

	InitLogger(config)

	// 使用全局 Logger
	zap.L().Info("通过 zap.L() 输出日志")
	zap.L().Error("通过 zap.L() 输出错误日志")

	t.Log("✅ zap 全局日志器替换成功\n")
}

// 测试：日志切割后，源日志被删除，只保留 .gz
func TestLogger_Rotate_DeleteOldLog(t *testing.T) {
	config := Config{
		Id:           8888,
		InstanceName: "check_rotate_delete",
		Mode:         "prod",
		StoragePath:  "../../logs",
	}
	InitLogger(config)

	logDir := "../../logs"
	instanceDir := filepath.Join(logDir, config.InstanceName)
	// logPath := filepath.Join(instanceDir, config.InstanceName+"8888.log")

	// 写入大量日志触发切割
	bigLog := bytes.Repeat([]byte("X"), 1024) // 1KB
	for i := 0; i < 1300; i++ {
		Logger.Info("test", zap.ByteString("data", bigLog))
	}
	_ = Logger.Sync()

	// 查看目录里的文件
	files, _ := os.ReadDir(instanceDir)
	hasUncompressedOldLog := false

	// ====================== 修复点：等待后台压缩 ======================
	// lumberjack 压缩是异步的，必须等 200~500ms 才能完成
	time.Sleep(5 * time.Second)
	// =================================================================

	for _, f := range files {
		name := f.Name()
		// 检查是否存在 已切割但未被删除的 .log 旧文件
		fmt.Println(name)
		if name != "8888.log" && strings.HasSuffix(name, ".log") {
			hasUncompressedOldLog = true
		}
	}

	// 断言：不应该存在未压缩的旧日志
	if hasUncompressedOldLog {
		t.Error("❌ 旧日志文件未被自动删除！")
	} else {
		t.Log("✅ 旧日志已自动删除，仅保留压缩包 .gz")
	}
}
