package compile

import (
	"context"
	"testing"
)

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
// ok      online-oj/judge/internal      51.834s
