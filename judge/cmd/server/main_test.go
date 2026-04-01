package main

import (
	"context"
	"fmt"
	"online-oj/api/proto/execute"
	pkg "online-oj/pkg/logger"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ===================== 配置 =====================
const (
	serverAddr  = "127.0.0.1:8080" // 你的服务端地址
	rpcTimeout  = 10 * time.Second // RPC调用超时（足够大）
	cpuLimit    = 3000_000_000     // 代码执行3秒
	memLimitMLE = 64 * 1024        // 64MB（超限测试）
)

// 全局客户端
var client execute.CompileAndRunClient

func init() {
	// 初始化日志
	config := pkg.Config{
		Id:           1,
		InstanceName: "gRpcClient",
		Mode:         "prod",
		StoragePath:  "../../../logs",
	}
	pkg.InitLogger(config)

	// 连接服务端
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic("connect server fail: " + err.Error())
	}
	client = execute.NewCompileAndRunClient(conn)
}

// ===================== 工具函数 =====================
func ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), rpcTimeout)
	return ctx
}

// 分隔线
func sep() {
	fmt.Println("\n" + "==================================================================================")
}

// 打印结果
func printResult(name string, start time.Time, resp *execute.ExecuteResponse, err error) {
	cost := time.Since(start)
	fmt.Printf("🧪 测试用例：%s\n", name)
	fmt.Printf("⏱  RPC总耗时：%v\n", cost)
	if err != nil {
		fmt.Println("❌ 错误：", err)
		return
	}
	fmt.Printf("✅ 状态：%s\n", resp.Status)
	fmt.Printf("📤 标准输出：%s\n", resp.Stdout)
	fmt.Printf("📥 标准错误：%s\n", resp.Stderr)
	fmt.Printf("⏱  代码耗时：%.2f ms\n", float64(resp.Time)/1_000_000.0)
	fmt.Printf("📊 代码内存：%d KB\n", resp.Memory)
}

// ===================== 测试开始 =====================

// 1. 极短代码
func TestRPC_ShortCode(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `package main;func main(){println("short")}`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("短代码", start, resp, err)

	// 结果校验
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Error("status != OK")
	}
	if resp.Stdout != "" {
		t.Error("stdout not empty")
	}
	if resp.Stderr == "" {
		t.Error("stderr empty")
	}
}

// 2. 中等代码
func TestRPC_MidCode(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `
package main
import "fmt"
func main(){
	s:=0
	for i:=0;i<1000;i++{s+=i}
	fmt.Println(s)
}`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("中等代码", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Error("status fail")
	}
}

// 3. 长代码
func TestRPC_LongCode(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `
package main
import "fmt"
func f1()int{return 1}
func f2()int{return f1()+2}
func f3()int{return f2()+3}
func f4()int{return f3()+4}
func f5()int{return f4()+5}
func main(){fmt.Println(f5())}
`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("长代码", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Error("status fail")
	}
}

// 4. 递归代码
func TestRPC_RecurseCode(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `
package main
import "fmt"
func fib(n int)int{
	if n<=2{return 1}
	return fib(n-1)+fib(n-2)
}
func main(){fmt.Println(fib(20))}
`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("递归代码", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Error("status fail")
	}
}

// 5. 计算密集
func TestRPC_CPUHeavyCode(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `
package main
func main(){
	s:=0
	for i:=0;i<1000000;i++{s+=i}
}`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("计算密集代码", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Error("status fail")
	}
}

// 6. 编译错误
func TestRPC_CompileError(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `package main;func main(){syntax error}`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("编译错误代码", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "CE" {
		t.Error("should be CE")
	}
}

// 7. 超时代码 TLE
func TestRPC_Timeout(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `
#include <iostream>
int main() {
	while(true){};
}
`
	req := &execute.ExecuteRequest{CodeType: 2, Code: code, CpuLimit: 1000_000_000, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("超时代码(TLE)", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "TLE" {
		t.Error("should be TLE")
	}
}

// 8. 内存超限 MLE
func TestRPC_MemoryExceeded(t *testing.T) {
	sep()
	defer sep()
	start := time.Now()
	code := `package main
func main() {
	a := make([]int,1024*1024*80)
	for i := range a {
    	a[i] = 1 // 必须赋值！
	}
}`
	req := &execute.ExecuteRequest{CodeType: 1, Code: code, CpuLimit: cpuLimit, MemoLimit: memLimitMLE}
	resp, err := client.ExecuteCode(ctx(), req)
	printResult("内存超限(MLE)", start, resp, err)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "MLE" {
		t.Error("should be MLE")
	}
}
