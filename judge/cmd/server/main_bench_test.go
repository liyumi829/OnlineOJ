package main

import (
	"context"
	"online-oj/api/proto/execute"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const benchServerAddr = "127.0.0.1:8080"

var benchClient execute.CompileAndRunClient

// 测试前初始化客户端连接
func init() {
	// ✅ 修复：使用官方推荐的 insecure.NewCredentials() 替代 WithInsecure()
	conn, err := grpc.NewClient(benchServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("客户端连接服务端失败：" + err.Error())
	}

	// 创建客户端
	benchClient = execute.NewCompileAndRunClient(conn)
}

// ==============================================
// 场景 1：极短代码（最简单打印）
// ==============================================
func BenchmarkExecute_ShortCode(b *testing.B) {
	code := `package main;func main(){println(1)}`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000, // 3s CPU限制
		MemoLimit: 64 * 1024,    // 64MB = 65536 KB
	}

	b.ResetTimer() // 开始计时

	// 压测循环
	for b.Loop() {
		_, err := benchClient.ExecuteCode(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 2：中等长度代码（正常逻辑）
// ==============================================
func BenchmarkExecute_MidCode(b *testing.B) {
	code := `
package main
import "fmt"
func main() {
	a := 0
	for i := 0; i < 1000; i++ {
		a += i
	}
	fmt.Println(a)
}
`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.ExecuteCode(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 3：长代码（大量逻辑、多行）
// ==============================================
func BenchmarkExecute_LongCode(b *testing.B) {
	code := `
package main
import "fmt"
func test1() int { return 1 }
func test2() int { return test1() + 2 }
func test3() int { return test2() + 3 }
func test4() int { return test3() + 4 }
func test5() int { return test4() + 5 }
func test6() int { return test5() + 6 }
func test7() int { return test6() + 7 }
func test8() int { return test7() + 8 }
func test9() int { return test8() + 9 }
func main() {
	fmt.Println(test9())
}
`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.ExecuteCode(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 4：递归代码（深度递归，测试性能）
// ==============================================
func BenchmarkExecute_RecurseCode(b *testing.B) {
	code := `
package main
import "fmt"
func fib(n int) int {
	if n <= 2 { return 1 }
	return fib(n-1) + fib(n-2)
}
func main() {
	fmt.Println(fib(20))
}
`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.ExecuteCode(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 5：计算密集型（压测CPU）
// ==============================================
func BenchmarkExecute_CPUHeavyCode(b *testing.B) {
	code := `
package main
func main() {
	a := 0
	for i := 0; i < 1000000; i++ {
		a += i
	}
}
`
	req := &execute.ExecuteRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.ExecuteCode(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

///////////////////////////////////////////////////////
// 																70~120ms之间
// BenchmarkExecute_ShortCode-4                  16          73971987 ns/op           15689 B/op        137 allocs/op
// BenchmarkExecute_ShortCode-4                  16          76133474 ns/op            5166 B/op         87 allocs/op
// BenchmarkExecute_ShortCode-4                  16          73560451 ns/op            5138 B/op         87 allocs/op
// BenchmarkExecute_MidCode-4                     9         115705742 ns/op            5273 B/op         88 allocs/op
// BenchmarkExecute_MidCode-4                    10         108931336 ns/op            5266 B/op         88 allocs/op
// BenchmarkExecute_MidCode-4                    10         108366896 ns/op            5288 B/op         88 allocs/op
// BenchmarkExecute_LongCode-4                   10         109829918 ns/op            5548 B/op         87 allocs/op
// BenchmarkExecute_LongCode-4                    9         112430149 ns/op            5586 B/op         88 allocs/op
// BenchmarkExecute_LongCode-4                    9         113785745 ns/op            5566 B/op         87 allocs/op
// BenchmarkExecute_RecurseCode-4                 9         122512453 ns/op            5280 B/op         87 allocs/op
// BenchmarkExecute_RecurseCode-4                 9         115786747 ns/op            5275 B/op         87 allocs/op
// BenchmarkExecute_RecurseCode-4                 9         111151998 ns/op            5291 B/op         88 allocs/op
// BenchmarkExecute_CPUHeavyCode-4               15          76960555 ns/op            5168 B/op         86 allocs/op
// BenchmarkExecute_CPUHeavyCode-4               14          73921039 ns/op            5185 B/op         86 allocs/op
// BenchmarkExecute_CPUHeavyCode-4               15          72635903 ns/op            5166 B/op         86 allocs/op
