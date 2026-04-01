package run

import (
	"context"
	"online-oj/judge/internal/compile"
	"os"
	"testing"
	"time"
)

var path = "../temp"

func init() {
	os.Mkdir(path, 0755) // 创建了临时文件还没有删除 运行完成bin之后删除保存 tempDir
}

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
		Code := `package main
import "fmt"
func main() {
	fmt.Print(1)
}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &compile.Compiler{CodeType: compile.GoType, Code: Code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.RunSandboxed(context.Background())
		}
	})

	// 2. Go 中等代码（正常业务）
	b.Run("Go_Medium", func(b *testing.B) {
		Code := `package main
	import "fmt"
	func test(a,b int)int{return a+b}
	func main(){
		s:=0
		for i:=0;i<100;i++{s+=test(i,i*2)}
		fmt.Println("result:",s)
	}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &compile.Compiler{CodeType: compile.GoType, Code: Code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.RunSandboxed(context.Background())
		}
	})

	// 3. Go 长代码（复杂逻辑）
	b.Run("Go_Long", func(b *testing.B) {
		Code := `package main
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
			compiler := &compile.Compiler{CodeType: compile.GoType, Code: Code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.RunSandboxed(context.Background())
		}
	})

	// ==========================
	// C++ 三种规模代码
	// ==========================

	// 1. C++ 短代码
	b.Run("Cpp_Short", func(b *testing.B) {
		Code := `#include<iostream>
using namespace std;
int main() {
	cout<<1;
	return 0;
}`
		b.ResetTimer()
		for b.Loop() {
			compiler := &compile.Compiler{CodeType: compile.CppType, Code: Code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.RunSandboxed(context.Background())
		}
	})

	// 2. C++ 中等代码
	b.Run("Cpp_Medium", func(b *testing.B) {
		Code := `#include <iostream>
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
			compiler := &compile.Compiler{CodeType: compile.CppType, Code: Code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.RunSandboxed(context.Background())
		}
	})

	// 3. C++ 长代码
	b.Run("Cpp_Long", func(b *testing.B) {
		Code := `#include <iostream>
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
			compiler := &compile.Compiler{CodeType: compile.CppType, Code: Code}
			res, _ := compiler.Compile(context.Background(), path)
			runner.Bin = res.BinPath
			_, _ = runner.RunSandboxed(context.Background())
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
