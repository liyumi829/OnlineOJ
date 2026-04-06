package compile

import (
	"context"
	"fmt"
	pkglogger "online-oj/pkg/logger"
	"testing"
	"time"
)

var path = "../../../temp"

//基准测试结果：

// 初始化日志（必须 否则测试中 zap.L() 会 panic）
func init() {
	pkglogger.InitLogger("debug", "../../../logs", "test_compile", 1001)
}

// 辅助函数：打印编译结果
func printCompileResult(result *CompileResult) {
	fmt.Println("=====================================")
	fmt.Println("          编译结果详情")
	fmt.Println("=====================================")
	fmt.Printf("状态(Status): %s\n", result.Status)
	fmt.Printf("可执行路径(BinPath): %s\n", result.BinPath)
	fmt.Printf("错误信息(Stderr):\n%s\n", result.Stderr)
	fmt.Println("=====================================")
}

// ------------------------------
// 测试1：正常 Go 代码编译
// ------------------------------
func TestCompiler_GoCompile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 一段正确的 Go 代码
	goCode := `
package main
func main() {
	println("hello go compile")
}
`
	compiler := Compiler{
		CodeType: GoType,
		Code:     goCode,
	}

	// 执行编译
	result, err := compiler.Compile(ctx, path)
	if err != nil {
		t.Fatalf("Go 编译执行失败: %v\n", err)
	}

	// 打印完整结果
	printCompileResult(result)

	// 断言
	if result.Status != "OK" {
		t.Error("Go 编译预期状态 OK，实际：", result.Status)
	}
	if result.BinPath == "" {
		t.Error("Go 编译未生成可执行文件\n")
	}
}

// ------------------------------
// 测试2：错误 Go 代码编译（语法错误）
// ------------------------------
func TestCompiler_GoCompile_Error(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 错误的 Go 代码（少括号）
	badGoCode := `
package main
func main( 
	println("error")
}
`
	compiler := Compiler{
		CodeType: GoType,
		Code:     badGoCode,
	}

	result, err := compiler.Compile(ctx, path)
	if err == nil {
		t.Fatal("预期编译失败，但实际成功了\n")
	}

	// 打印完整结果
	printCompileResult(result)

	if result.Status != "CE" {
		t.Error("预期编译错误 CE，实际：", result.Status)
	}
	if result.Stderr == "" {
		t.Error("未捕获到编译错误信息\n")
	}
}

// ------------------------------
// 测试3：正常 C++ 代码编译
// ------------------------------
func TestCompiler_CppCompile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 正确 C++ 代码
	cppCode := `
#include <iostream>
using namespace std;
int main() {
	cout << "hello cpp compile" << endl;
	return 0;
}
`
	compiler := Compiler{
		CodeType: CppType,
		Code:     cppCode,
	}

	result, err := compiler.Compile(ctx, path)
	if err != nil {
		t.Fatalf("C++ 编译执行失败: %v\n", err)
	}

	// 打印完整结果
	printCompileResult(result)

	if result.Status != "OK" {
		t.Error("C++ 编译预期状态 OK，实际：", result.Status)
	}
	if result.BinPath == "" {
		t.Error("C++ 编译未生成可执行文件\n")
	}
}

// ------------------------------
// 测试4：错误 C++ 代码编译
// ------------------------------
func TestCompiler_CppCompile_Error(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 错误 C++ 代码
	badCppCode := `
#include <iostream>
int main() {
	cout << "error" << endl  // 少分号
	return 0
}
`
	compiler := Compiler{
		CodeType: CppType,
		Code:     badCppCode,
	}

	result, err := compiler.Compile(ctx, path)
	if err == nil {
		t.Fatal("预期C++编译失败，但实际成功了\n")
	}

	// 打印完整结果
	printCompileResult(result)

	if result.Status != "CE" {
		t.Error("预期编译错误 CE，实际：", result.Status)
	}
}
