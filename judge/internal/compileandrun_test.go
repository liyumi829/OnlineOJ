package internal

import (
	"context"
	pkg "online---oj/pkg/logger"
	"os"
	"testing"
	"time"
)

func init() {
	config := pkg.Config{
		Id:           1001,
		InstanceName: "run-1",
		Mode:         "prod",
		StoragePath:  "../../logs",
	}
	pkg.InitLogger(config)
}

// TestCompiler 测试编译器功能
func TestCompilerAndRun(t *testing.T) {
	// 创建存储路径
	path := "./temp"
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("创建存储目录失败: %v", err)
	}
	// sdefer os.RemoveAll(path)

	// 创建 Runner 用于运行编译后的程序
	runner := &Runner{
		CpuLimit:     10 * time.Second,
		MemoKiBLimit: 1024 * 64, // 64MB 内存限制
	}

	// ============================================================================
	// 测试1: Go 程序 - 正常编译并成功运行
	// ============================================================================
	t.Log("\n========== 测试1: Go 程序 - 正常编译并成功运行 ==========")
	goSuccessCode := `package main
import "fmt"
func main() {
	fmt.Println("Hello, World!")
	fmt.Println("This is a test program")
}
`
	compiler := &Compiler{codeType: GoType, code: goSuccessCode}
	compileResult, err := compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		// 运行编译后的程序
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		t.Logf("标准输出:\n%s", runResult.Stdout)
		if runResult.Stderr != "" {
			t.Logf("标准错误:\n%s", runResult.Stderr)
		}
	}

	// ============================================================================
	// 测试2: Go 程序 - 编译成功但运行时发生 panic
	// ============================================================================
	t.Log("\n========== 测试2: Go 程序 - 编译成功但运行时发生 panic ==========")
	goPanicCode := `package main
func main() {
	var a []int
	_ = a[100] // 索引越界，触发 panic
}
`
	compiler = &Compiler{codeType: GoType, code: goPanicCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		if runResult.Stdout != "" {
			t.Logf("标准输出:\n%s", runResult.Stdout)
		}
		t.Logf("标准错误:\n%s", runResult.Stderr)
	}

	// ============================================================================
	// 测试3: Go 程序 - 编译失败（语法错误）
	// ============================================================================
	t.Log("\n========== 测试3: Go 程序 - 编译失败（语法错误） ==========")
	goSyntaxErrorCode := `package main
import "fmt"
func main() {
	fmt.Println("Hello" // 缺少右括号，语法错误
}
`
	compiler = &Compiler{codeType: GoType, code: goSyntaxErrorCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s", compileResult.Status, compileResult.Stderr)
		t.Logf("编译错误信息:\n%s", compileResult.Stderr)
	} else {
		t.Logf("编译成功: BinPath=%s（预期应该编译失败）", compileResult.BinPath)
	}

	// ============================================================================
	// 测试4: Go 程序 - 编译成功但运行超时（死循环）
	// ============================================================================
	t.Log("\n========== 测试4: Go 程序 - 编译成功但运行超时（死循环） ==========")
	goTimeoutCode := `package main
func main() {
	for {
		// 死循环，会触发超时
	}
}
`
	compiler = &Compiler{codeType: GoType, code: goTimeoutCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
	}

	// ============================================================================
	// 测试5: C++ 程序 - 正常编译并成功运行
	// ============================================================================
	t.Log("\n========== 测试5: C++ 程序 - 正常编译并成功运行 ==========")
	cppSuccessCode := `#include <iostream>
using namespace std;
int main() {
    cout << "Hello from C++!" << endl;
    cout << "This is a test program" << endl;
    return 0;
}
`
	compiler = &Compiler{codeType: CppType, code: cppSuccessCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		t.Logf("标准输出:\n%s", runResult.Stdout)
		if runResult.Stderr != "" {
			t.Logf("标准错误:\n%s", runResult.Stderr)
		}
	}

	// ============================================================================
	// 测试6: C++ 程序 - 编译成功但运行时发生段错误
	// ============================================================================
	t.Log("\n========== 测试6: C++ 程序 - 编译成功但运行时发生段错误 ==========")
	cppSegfaultCode := `int main() {
    int* p = nullptr;
    *p = 42; // 空指针解引用，导致段错误
    return 0;
}
`
	compiler = &Compiler{codeType: CppType, code: cppSegfaultCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		if runResult.Stdout != "" {
			t.Logf("标准输出:\n%s", runResult.Stdout)
		}
		t.Logf("标准错误:\n%s", runResult.Stderr)
	}

	// ============================================================================
	// 测试7: C++ 程序 - 编译失败（语法错误）
	// ============================================================================
	t.Log("\n========== 测试7: C++ 程序 - 编译失败（语法错误） ==========")
	cppSyntaxErrorCode := `#include <iostream>
using namespace std;
int main() {
    cout << "Hello" // 缺少分号，语法错误
    return 0;
}
`
	compiler = &Compiler{codeType: CppType, code: cppSyntaxErrorCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s", compileResult.Status, compileResult.Stderr)
		t.Logf("编译错误信息:\n%s", compileResult.Stderr)
	} else {
		t.Logf("编译成功: BinPath=%s（预期应该编译失败）", compileResult.BinPath)
	}

	// ============================================================================
	// 测试8: C++ 程序 - 内存超限测试
	// ============================================================================
	t.Log("\n========== 测试8: C++ 程序 - 内存超限测试 ==========")
	cppMemoryLimitCode := `#include <vector>
#include <iostream>
int main() {
    std::vector<int*> ptrs;
    // 分配大量内存，超过限制
    for (int i = 0; i < 20000; i++) {
        ptrs.push_back(new int[1024 * 1024]); // 4MB
    }
    std::cout << "Memory allocated" << std::endl;
    return 0;
}
`
	compiler = &Compiler{codeType: CppType, code: cppMemoryLimitCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB (限制: %dKB)",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal, runner.MemoKiBLimit)
	}

	// ============================================================================
	// 测试9: Go 程序 - 输出超限测试
	// ============================================================================
	t.Log("\n========== 测试9: Go 程序 - 输出超限测试 ==========")
	goOutputLimitCode := `package main
import "fmt"
func main() {
    for i := 0; i < 1000000; i++ {
        fmt.Println("This is a very long line that will cause output limit exceeded")
    }
}
`
	compiler = &Compiler{codeType: GoType, code: goOutputLimitCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		t.Logf("标准输出长度: %d 字节", len(runResult.Stdout))
	}

	// ============================================================================
	// 测试10: C++ 程序 - 正常退出并返回自定义退出码
	// ============================================================================
	t.Log("\n========== 测试10: C++ 程序 - 正常退出并返回自定义退出码 ==========")
	cppExitCodeCode := `#include <iostream>
int main() {
    std::cout << "Exiting with code 42" << std::endl;
    return 42;
}
`
	compiler = &Compiler{codeType: CppType, code: cppExitCodeCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.runSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		t.Logf("标准输出:\n%s", runResult.Stdout)
	}
}

// 如果需要运行特定的测试子集，可以使用以下方式：
// go test -v -run TestCompilerAndRun

///////////////////////////////////////////////////////////////////////////////////////

// 基准测试公共初始化：只执行一次
// go test -bench ^BenchmarkMain$ -benchmem -run ^$ -count=5 只跑这一个测试
func BenchmarkMain(b *testing.B) {
	runner := &Runner{
		CpuLimit:     2 * time.Second,
		MemoKiBLimit: 64 * 1024,
	}

	// ==========================
	// Go 语言三种规模代码
	// ==========================

	// 1. Go 短代码（极简）
	b.Run("Go_Short", func(b *testing.B) {
		code := `package main;import "fmt";func main(){fmt.Print(1)}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &Compiler{codeType: GoType, code: code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.runSandboxed(context.Background())
		}
	})

	// 2. Go 中等代码（正常业务）
	b.Run("Go_Medium", func(b *testing.B) {
		code := `package main
	import "fmt"
	func test(a,b int)int{return a+b}
	func main(){
		s:=0
		for i:=0;i<100;i++{s+=test(i,i*2)}
		fmt.Println("result:",s)
	}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &Compiler{codeType: GoType, code: code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.runSandboxed(context.Background())
		}
	})

	// 3. Go 长代码（复杂逻辑）
	b.Run("Go_Long", func(b *testing.B) {
		code := `package main
	import "fmt"
	type Data struct{ID int;Value string}
	func NewData(id int,v string)*Data{return &Data{id,v}}
	func (d *Data)Show(){fmt.Println(d.ID,d.Value)}
	func test1()int{return 123}
	func test2()string{return "hello"}
	func main(){
		list:=make([]*Data,0,100)
		for i:=0;i<100;i++{
			list=append(list,NewData(i,test2()))
		}
		sum:=0
		for _,v:=range list{
			sum+=v.ID;v.Show()
		}
		fmt.Println("sum:",sum,test1())
	}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &Compiler{codeType: GoType, code: code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.runSandboxed(context.Background())
		}
	})

	// ==========================
	// C++ 三种规模代码
	// ==========================

	// 1. C++ 短代码
	b.Run("Cpp_Short", func(b *testing.B) {
		code := `#include<iostream>
using namespace std;
int main() {
	cout<<1;
	return 0;
}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &Compiler{codeType: CppType, code: code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.runSandboxed(context.Background())
		}
	})

	// 2. C++ 中等代码
	b.Run("Cpp_Medium", func(b *testing.B) {
		code := `#include <iostream>
using namespace std;
int add(int a,int b){return a+b;}
int main(){
	int s=0;
	for(int i=0;i<100;i++)s+=add(i,i*2);
	cout<<s<<endl;
	return 0;
}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &Compiler{codeType: CppType, code: code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.runSandboxed(context.Background())
		}
	})

	// 3. C++ 长代码
	b.Run("Cpp_Long", func(b *testing.B) {
		code := `#include <iostream>
#include <vector>
#include <string>
using namespace std;
struct Data{int id;string val;};
Data newData(int id,string v){return {id,v};}
void show(Data d){cout<<d.id<<" "<<d.val<<endl;}
int main(){
	vector<Data> vec;
	for(int i=0;i<100;i++) vec.push_back(newData(i,"test"));
	int sum=0;
	for(auto& d:vec){sum+=d.id;show(d);}
	cout<<"sum="<<sum<<endl;
	return 0;
}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &Compiler{codeType: CppType, code: code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.runSandboxed(context.Background())
		}
	})
}

// BenchmarkMain/Go_Short-4                      10         107539295 ns/op          125810 B/op        189 allocs/op
// BenchmarkMain/Go_Short-4                      10         110443472 ns/op          121242 B/op        182 allocs/op
// BenchmarkMain/Go_Short-4                      10         108330079 ns/op          122548 B/op        181 allocs/op
// BenchmarkMain/Go_Short-4                      10         106751597 ns/op          121048 B/op        180 allocs/op
// BenchmarkMain/Go_Short-4                      10         108692391 ns/op          121457 B/op        180 allocs/op

// BenchmarkMain/Go_Medium-4                     10         106126721 ns/op          123502 B/op        183 allocs/op
// BenchmarkMain/Go_Medium-4                      9         112476040 ns/op          122369 B/op        181 allocs/op
// BenchmarkMain/Go_Medium-4                     10         105654706 ns/op          122293 B/op        181 allocs/op
// BenchmarkMain/Go_Medium-4                     10         107051752 ns/op          122826 B/op        183 allocs/op
// BenchmarkMain/Go_Medium-4                     10         111091028 ns/op          121285 B/op        181 allocs/op

// BenchmarkMain/Go_Long-4                        9         114237539 ns/op          124126 B/op        183 allocs/op
// BenchmarkMain/Go_Long-4                       10         105130815 ns/op          125016 B/op        186 allocs/op
// BenchmarkMain/Go_Long-4                       10         113415445 ns/op          126680 B/op        185 allocs/op
// BenchmarkMain/Go_Long-4                       10         111358306 ns/op          124348 B/op        184 allocs/op
// BenchmarkMain/Go_Long-4                       10         111495152 ns/op          124816 B/op        183 allocs/op

// BenchmarkMain/Cpp_Short-4                      4         284595997 ns/op          125240 B/op        210 allocs/op
// BenchmarkMain/Cpp_Short-4                      4         262504515 ns/op          127790 B/op        212 allocs/op
// BenchmarkMain/Cpp_Short-4                      4         283179379 ns/op          125240 B/op        210 allocs/op
// BenchmarkMain/Cpp_Short-4                      4         283772507 ns/op          129884 B/op        213 allocs/op
// BenchmarkMain/Cpp_Short-4                      4         260835162 ns/op          126348 B/op        215 allocs/op

// BenchmarkMain/Cpp_Medium-4                     4         270793330 ns/op          129008 B/op        211 allocs/op
// BenchmarkMain/Cpp_Medium-4                     4         263519584 ns/op          122790 B/op        209 allocs/op
// BenchmarkMain/Cpp_Medium-4                     4         261482092 ns/op          129992 B/op        214 allocs/op
// BenchmarkMain/Cpp_Medium-4                     4         279691022 ns/op          123786 B/op        213 allocs/op
// BenchmarkMain/Cpp_Medium-4                     4         270661742 ns/op          127930 B/op        214 allocs/op

// BenchmarkMain/Cpp_Long-4                       4         324241492 ns/op          134116 B/op        220 allocs/op
// BenchmarkMain/Cpp_Long-4                       4         324396858 ns/op          125654 B/op        211 allocs/op
// BenchmarkMain/Cpp_Long-4                       4         331093788 ns/op          133356 B/op        218 allocs/op
// BenchmarkMain/Cpp_Long-4                       3         333917029 ns/op          127765 B/op        219 allocs/op
// BenchmarkMain/Cpp_Long-4                       4         322184256 ns/op          130126 B/op        214 allocs/op
