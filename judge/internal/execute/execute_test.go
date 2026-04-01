// execute/server_test.go
package execute

import (
	"context"
	"fmt"
	"online-oj/api/proto/execute"
	"online-oj/judge/internal/compile"
	pkg "online-oj/pkg/logger"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// 初始化日志
func init() {
	config := pkg.Config{
		Id:           1,
		InstanceName: "gRpc-Server",
		Mode:         "debug",
		StoragePath:  "../../../logs",
	}
	pkg.InitLogger(config)
}

// 测试常量（代码执行CPU限制 3s）
const (
	testStoragePath = "./test_storage"
	testAddr        = "127.0.0.1:50052"
	CPU_LIMIT_S     = 3000_000_000 // 代码执行 3s 限制
	MEM_LIMIT_KB    = 128          // 128KB
)

// 无超时 context（永远不会出现 context deadline exceeded）
func backgroundCtx() context.Context {
	return context.Background()
}

// 清理测试目录
func cleanTestDir() {
	_ = os.RemoveAll(testStoragePath)
}

// ====================== 测试耗时工具 ======================
func testStart(t *testing.T) time.Time {
	t.Helper()
	fmt.Println("\n" + "----------------------------------------------------------------------")
	fmt.Printf("🧪 开始测试：%s\n", t.Name())
	return time.Now()
}

func testEnd(t *testing.T, start time.Time) {
	t.Helper()
	cost := time.Since(start)
	fmt.Printf("✅ 测试完成：%s，耗时：%v\n", t.Name(), cost)
	fmt.Println("----------------------------------------------------------------------")
	fmt.Println()
}

// ============================================================
// Test_NewServer 测试服务创建
// ============================================================
func Test_NewServer(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, err := NewServer(testStoragePath)
	assert.NoError(t, err)
	assert.NotNil(t, srv)
}

// ============================================================
// Test_getCodeType 测试语言类型映射
// ============================================================
func Test_getCodeType(t *testing.T) {
	start := testStart(t)
	defer testEnd(t, start)

	assert.Equal(t, compile.GoType, getCodeType(1))
	assert.Equal(t, compile.CppType, getCodeType(2))
	assert.Equal(t, compile.UnKnownType, getCodeType(0))
	assert.Equal(t, compile.UnKnownType, getCodeType(999))
}

// ============================================================
// Test_ExecuteCode_UnknownType 未知语言类型
// ============================================================
func Test_ExecuteCode_UnknownType(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, _ := NewServer(testStoragePath)
	req := &execute.ExecuteRequest{CodeType: 999, Code: "test"}
	rsp, err := srv.ExecuteCode(backgroundCtx(), req)

	assert.Error(t, err)
	assert.Nil(t, rsp)
}

// ============================================================
// Test_ExecuteCode_CompileError 编译错误 CE
// ============================================================
func Test_ExecuteCode_CompileError(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, _ := NewServer(testStoragePath)
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      "package main\nfunc main() { println( }",
		CpuLimit:  CPU_LIMIT_S,
		MemoLimit: MEM_LIMIT_KB << 20,
	}
	rsp, err := srv.ExecuteCode(backgroundCtx(), req)

	assert.NoError(t, err)
	assert.Equal(t, "CE", rsp.Status)
	assert.NotEmpty(t, rsp.Stderr)
	assert.Empty(t, rsp.Stdout)
	assert.Equal(t, int64(0), rsp.Time)
	assert.Equal(t, int64(0), rsp.Memory)
}

// ============================================================
// Test_ExecuteCode_GoSuccess Go 代码成功
// ============================================================
func Test_ExecuteCode_GoSuccess(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, _ := NewServer(testStoragePath)
	goCode := `
package main
import "fmt"
func main() { fmt.Println("hello test") }
`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      goCode,
		CpuLimit:  CPU_LIMIT_S,
		MemoLimit: MEM_LIMIT_KB << 20,
	}
	rsp, err := srv.ExecuteCode(backgroundCtx(), req)

	assert.NoError(t, err)
	assert.Equal(t, "OK", rsp.Status)
	assert.Contains(t, rsp.Stdout, "hello test")
	assert.Empty(t, rsp.Stderr)
	assert.Greater(t, rsp.Time, int64(0))
	assert.Greater(t, rsp.Memory, int64(0))
}

// ============================================================
// Test_ExecuteCode_CppSuccess C++ 代码成功
// ============================================================
func Test_ExecuteCode_CppSuccess(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, _ := NewServer(testStoragePath)
	cppCode := `
#include <iostream>
using namespace std;
int main() { cout << "cpp test" << endl; return 0; }
`
	req := &execute.ExecuteRequest{
		CodeType:  2,
		Code:      cppCode,
		CpuLimit:  CPU_LIMIT_S,
		MemoLimit: MEM_LIMIT_KB << 20,
	}
	rsp, err := srv.ExecuteCode(backgroundCtx(), req)

	assert.NoError(t, err)
	assert.Equal(t, "OK", rsp.Status)
	assert.Contains(t, rsp.Stdout, "cpp test")
	assert.Empty(t, rsp.Stderr)
	assert.Greater(t, rsp.Time, int64(0))
	assert.Greater(t, rsp.Memory, int64(0))
}

// ============================================================
// Test_ExecuteCode_Timeout 代码执行超时 TLE（3s）
// ============================================================
func Test_ExecuteCode_Timeout(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, _ := NewServer(testStoragePath)
	loopCode := `package main; func main(){ for{} }` // 死循环

	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      loopCode,
		CpuLimit:  CPU_LIMIT_S, // 代码执行 3s 超时
		MemoLimit: MEM_LIMIT_KB << 20,
	}
	rsp, err := srv.ExecuteCode(backgroundCtx(), req)

	assert.NoError(t, err)
	assert.Equal(t, "TLE", rsp.Status)
	assert.NotEmpty(t, rsp.Stderr)
	assert.Empty(t, rsp.Stdout)
	assert.Greater(t, rsp.Time, int64(0))
	assert.GreaterOrEqual(t, rsp.Memory, int64(0))
}

// ============================================================
// Test_ExecuteCode_MemoryLimit 内存超限 MLE
// ============================================================
func Test_ExecuteCode_MemoryLimit(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	srv, _ := NewServer(testStoragePath)
	memCode := `package main
func main() {
	a := make([]int,1024*1024)
	for i := range a {
    	a[i] = 1 // 必须赋值！
	}
}`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      memCode,
		CpuLimit:  CPU_LIMIT_S,
		MemoLimit: 128,
	}
	rsp, err := srv.ExecuteCode(backgroundCtx(), req)

	assert.NoError(t, err)
	assert.Equal(t, "MLE", rsp.Status)
	assert.NotEmpty(t, rsp.Stderr)
	assert.Empty(t, rsp.Stdout)
	assert.Greater(t, rsp.Time, int64(0))
	assert.Greater(t, rsp.Memory, int64(0))
}

// ============================================================
// Test_StartGRPCServer gRPC 服务测试
// ============================================================
// Test_StartGRPCServer gRPC 服务测试（彻底修复grpc.Dial弃用+clientconn未定义问题）
func Test_StartGRPCServer(t *testing.T) {
	defer cleanTestDir()
	start := testStart(t)
	defer testEnd(t, start)

	// 异步启动服务
	go func() {
		_ = StartGRPCServer(testAddr, testStoragePath)
	}()
	time.Sleep(300 * time.Millisecond)

	// ✅ 最终正确写法：grpc.NewClient，兼容所有gRPC版本，无警告、无报错
	conn, err := grpc.NewClient(testAddr, grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := execute.NewCompileAndRunClient(conn)

	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      "package main\nfunc main() { println( }",
		CpuLimit:  CPU_LIMIT_S,
		MemoLimit: MEM_LIMIT_KB,
	}
	rsp, err := client.ExecuteCode(backgroundCtx(), req)

	// RPC全字段校验
	assert.NoError(t, err)
	assert.Equal(t, "CE", rsp.Status)
	assert.NotEmpty(t, rsp.Stderr)
	assert.Empty(t, rsp.Stdout)
	assert.Equal(t, int64(0), rsp.Time)
	assert.Equal(t, int64(0), rsp.Memory)
}
