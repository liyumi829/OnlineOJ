package run

import (
	"context"
	"online-oj/judge/internal/compile"
	pkg "online-oj/pkg/logger"
	"os"
	"testing"
	"time"
)

// 如果需要运行特定的测试子集，可以使用以下方式：
// go test -v -run TestCompilerAndRun
func init() {
	config := pkg.Config{
		Id:           1001,
		InstanceName: "run-1",
		Mode:         "prod",
		StoragePath:  "../../../logs",
	}
	pkg.InitLogger(config)
}

// TestCompiler 测试编译器功能
func TestCompilerAndRun(t *testing.T) {
	// 创建存储路径
	path := "../temp"
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
	compiler := &compile.Compiler{CodeType: compile.GoType, Code: goSuccessCode}
	compileResult, err := compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		// 运行编译后的程序
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
	compiler = &compile.Compiler{CodeType: compile.GoType, Code: goPanicCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
	compiler = &compile.Compiler{CodeType: compile.GoType, Code: goSyntaxErrorCode}
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
	compiler = &compile.Compiler{CodeType: compile.GoType, Code: goTimeoutCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
	compiler = &compile.Compiler{CodeType: compile.CppType, Code: cppSuccessCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
	compiler = &compile.Compiler{CodeType: compile.CppType, Code: cppSegfaultCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
	compiler = &compile.Compiler{CodeType: compile.CppType, Code: cppSyntaxErrorCode}
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
	compiler = &compile.Compiler{CodeType: compile.CppType, Code: cppMemoryLimitCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
	compiler = &compile.Compiler{CodeType: compile.GoType, Code: goOutputLimitCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
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
    std::cout << "Exiting with Code 42" << std::endl;
    return 42;
}
`
	compiler = &compile.Compiler{CodeType: compile.CppType, Code: cppExitCodeCode}
	compileResult, err = compiler.Compile(context.Background(), path)
	if err != nil || compileResult.Status != "OK" {
		t.Logf("编译结果: Status=%s, Stderr=%s, Error=%v", compileResult.Status, compileResult.Stderr, err)
	} else {
		t.Logf("编译成功: BinPath=%s", compileResult.BinPath)
		runner.Bin = compileResult.BinPath
		runResult, runErr := runner.RunSandboxed(context.Background())
		if runErr != nil {
			t.Logf("运行错误: %v", runErr)
		}
		t.Logf("运行结果: ExitCode=%d, Status=%s, TimeReal=%v, MemoKiBReal=%dKB",
			runResult.ExitCode, runResult.Status, runResult.TimeReal, runResult.MemoKiBReal)
		t.Logf("标准输出:\n%s", runResult.Stdout)
	}
}
