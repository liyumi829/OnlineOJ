package compile

import (
	"context"
	"fmt"
	pkg "online---oj/pkg/logger"
	"testing"
	"time"
)

var path = "../temp"

//基准测试结果：

// 初始化日志（必须 否则测试中 zap.L() 会 panic）
func init() {
	config := pkg.Config{
		Id:           1001,
		InstanceName: "test_compile",
		Mode:         "prod",
		StoragePath:  "../../logs",
	}
	pkg.InitLogger(config)
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

///////////////////////////////////////////////////////////////////////////////////////////////////////

// 基准测试：Go 代码编译性能（重复执行，测试性能）
// go test . -bench=Benchmark -run=^$
// go test . -bench=Benchmark -run=^$ -benchmem -count=5
// -----------------------------
// 1. 极简 Go 代码（最快编译）
// ------------------------------
func BenchmarkCompiler_Go_Minimal(b *testing.B) {
	Code := `package main;func main(){}`
	b.ResetTimer()
	for b.Loop() {
		c := Compiler{CodeType: GoType, Code: Code}
		res, _ := c.Compile(context.Background(), path)
		if res.Status != "OK" {
			b.Fatal("fail")
		}
	}
}

// ------------------------------
// 2. 标准 Go 代码（最常用）
// ------------------------------
func BenchmarkCompiler_Go_Normal(b *testing.B) {
	Code := `
package main
import "fmt"
func main() {
	fmt.Println("hello")
}
`
	b.ResetTimer()
	for b.Loop() {
		c := Compiler{CodeType: GoType, Code: Code}
		res, _ := c.Compile(context.Background(), path)
		if res.Status != "OK" {
			b.Fatal("fail")
		}
	}
}

// ------------------------------
// 3. 大型 Go 代码（测试压力编译）
// ------------------------------
func BenchmarkCompiler_Go_Large(b *testing.B) {
	Code := `
package main
import "fmt"
func f1(){fmt.Print(1)}
func f2(){fmt.Print(2)}
func f3(){fmt.Print(3)}
func f4(){fmt.Print(4)}
func f5(){fmt.Print(5)}
func f6(){fmt.Print(6)}
func f7(){fmt.Print(7)}
func f8(){fmt.Print(8)}
func f9(){fmt.Print(9)}
func main() {
	f1();f2();f3();f4();f5();f6();f7();f8();f9()
}
`
	b.ResetTimer()
	for b.Loop() {
		c := Compiler{CodeType: GoType, Code: Code}
		res, _ := c.Compile(context.Background(), path)
		if res.Status != "OK" {
			b.Fatal("fail")
		}
	}
}

// ------------------------------
// 4. 极简 C++ 代码
// ------------------------------
func BenchmarkCompiler_Cpp_Minimal(b *testing.B) {
	Code := `int main(){return 0;}`
	b.ResetTimer()
	for b.Loop() {
		c := Compiler{CodeType: CppType, Code: Code}
		res, _ := c.Compile(context.Background(), path)
		if res.Status != "OK" {
			b.Fatal("fail")
		}
	}
}

// ------------------------------
// 5. 标准 C++ 代码
// ------------------------------
func BenchmarkCompiler_Cpp_Normal(b *testing.B) {
	Code := `
#include <iostream>
using namespace std;
int main() {
	cout << "hello" << endl;
	return 0;
}
`
	b.ResetTimer()
	for b.Loop() {
		c := Compiler{CodeType: CppType, Code: Code}
		res, _ := c.Compile(context.Background(), path)
		if res.Status != "OK" {
			b.Fatal("fail")
		}
	}
}

// ------------------------------
// 6. 大型 C++ 代码（多函数、压力）
// ------------------------------
func BenchmarkCompiler_Cpp_Large(b *testing.B) {
	Code := `
#include <iostream>
using namespace std;
void f1(){cout<<1;}
void f2(){cout<<2;}
void f3(){cout<<3;}
void f4(){cout<<4;}
void f5(){cout<<5;}
int main() {
	f1();f2();f3();f4();f5();
	return 0;
}
`
	b.ResetTimer()
	for b.Loop() {
		c := Compiler{CodeType: CppType, Code: Code}
		res, _ := c.Compile(context.Background(), path)
		if res.Status != "OK" {
			b.Fatal("fail")
		}
	}
}

// bench 测试结果：
// BenchmarkCompiler_Go_Minimal-4                16          69664750 ns/op
// BenchmarkCompiler_Go_Minimal-4                16          70503862 ns/op
// BenchmarkCompiler_Go_Minimal-4                16          69811202 ns/op
// BenchmarkCompiler_Go_Minimal-4                15          72386232 ns/op
// BenchmarkCompiler_Go_Minimal-4                15          71818389 ns/op

// BenchmarkCompiler_Go_Normal-4                 10         107640616 ns/op
// BenchmarkCompiler_Go_Normal-4                 10         107899800 ns/op
// BenchmarkCompiler_Go_Normal-4                 10         108075268 ns/op
// BenchmarkCompiler_Go_Normal-4                 10         112188822 ns/op
// BenchmarkCompiler_Go_Normal-4                 10         109024336 ns/op

// BenchmarkCompiler_Go_Large-4                  10         107633556 ns/op
// BenchmarkCompiler_Go_Large-4                  10         108023575 ns/op
// BenchmarkCompiler_Go_Large-4                   9         112009291 ns/op
// BenchmarkCompiler_Go_Large-4                  10         108365305 ns/op
// BenchmarkCompiler_Go_Large-4                  10         110176805 ns/op

// BenchmarkCompiler_Cpp_Minimal-4               27          39632035 ns/op
// BenchmarkCompiler_Cpp_Minimal-4               31          39604141 ns/op
// BenchmarkCompiler_Cpp_Minimal-4               30          39496645 ns/op
// BenchmarkCompiler_Cpp_Minimal-4               30          38785588 ns/op
// BenchmarkCompiler_Cpp_Minimal-4               30          40050444 ns/op

// BenchmarkCompiler_Cpp_Normal-4                 4         267410723 ns/op
// BenchmarkCompiler_Cpp_Normal-4                 4         277829814 ns/op
// BenchmarkCompiler_Cpp_Normal-4                 4         287057187 ns/op
// BenchmarkCompiler_Cpp_Normal-4                 4         266239152 ns/op
// BenchmarkCompiler_Cpp_Normal-4                 4         265208433 ns/op

// BenchmarkCompiler_Cpp_Large-4                  4         269136882 ns/op // 269ms
// BenchmarkCompiler_Cpp_Large-4                  4         284802034 ns/op
// BenchmarkCompiler_Cpp_Large-4                  4         289403877 ns/op
// BenchmarkCompiler_Cpp_Large-4                  4         276858384 ns/op
// BenchmarkCompiler_Cpp_Large-4                  4         273941744 ns/op
// PASS
// ok      online---oj/judge/internal      51.834s
