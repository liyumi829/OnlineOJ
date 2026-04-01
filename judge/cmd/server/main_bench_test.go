package main

import (
	"context"
	"online-oj/api/proto/judge"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const benchServerAddr = "127.0.0.1:8080"

var benchClient judge.JudgeServiceClient

// 测试前初始化客户端连接
func init() {
	// ✅ 修复：使用官方推荐的 insecure.NewCredentials() 替代 WithInsecure()
	conn, err := grpc.NewClient(benchServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("客户端连接服务端失败：" + err.Error())
	}

	// 创建客户端
	benchClient = judge.NewJudgeServiceClient(conn)
}

// ==============================================
// 场景 1：极短代码（最简单打印）
// ==============================================
func BenchmarkExecute_ShortCode_Go(b *testing.B) {
	code := `package main
import "fmt"
func main() {
	fmt.Print(1)
}`
	req := &judge.JudgeRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000, // 3s CPU限制
		MemoLimit: 64 * 1024,    // 64MB = 65536 KB
	}

	b.ResetTimer() // 开始计时

	// 压测循环
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 2：中等长度代码（正常逻辑）
// ==============================================
func BenchmarkExecute_MidCode_Go(b *testing.B) {
	code := `package main
import "fmt"
func test(a,b int)int{return a+b}
func main(){
	s:=0
	for i:=0;i<100;i++{s+=test(i,i*2)}
	fmt.Println("result:",s)
}`
	req := &judge.JudgeRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 3：长代码（大量逻辑、多行）
// ==============================================
func BenchmarkExecute_LongCode_Go(b *testing.B) {
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
	req := &judge.JudgeRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 4：递归代码（深度递归，测试性能）
// ==============================================
func BenchmarkExecute_RecurseCode_Go(b *testing.B) {
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
	req := &judge.JudgeRequest{
		CodeType:  1,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ==============================================
// 场景 5：计算密集型（压测CPU）
// ==============================================
func BenchmarkExecute_ShortCode_Cpp(b *testing.B) {
	code := `#include<iostream>
using namespace std;
int main() {
	cout<<1;
	return 0;
}`
	req := &judge.JudgeRequest{
		CodeType:  2,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecute_MediumCode_Cpp(b *testing.B) {
	code := `#include <iostream>
using namespace std;
int add(int a,int b){return a+b;}
int main(){
	int s=0;
	for(int i=0;i<100;i++)s+=add(i,i*2);
	cout<<s<<endl;
	return 0;
}`
	req := &judge.JudgeRequest{
		CodeType:  2,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecute_LongCode_Cpp(b *testing.B) {
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
	req := &judge.JudgeRequest{
		CodeType:  2,
		Code:      code,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := benchClient.Judge(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

///////////////////////////////////////////////////////
// 															100~300ms之间
// BenchmarkExecute_ShortCode_Go-4               10         109075201 ns/op
// BenchmarkExecute_ShortCode_Go-4               10         109672968 ns/op
// BenchmarkExecute_ShortCode_Go-4               10         109786091 ns/op
// BenchmarkExecute_ShortCode_Go-4               10         110348252 ns/op
// BenchmarkExecute_ShortCode_Go-4               10         110640906 ns/op

// BenchmarkExecute_MidCode_Go-4                 10         115969695 ns/op
// BenchmarkExecute_MidCode_Go-4                  9         115524553 ns/op
// BenchmarkExecute_MidCode_Go-4                  9         114395183 ns/op
// BenchmarkExecute_MidCode_Go-4                  9         116039551 ns/op
// BenchmarkExecute_MidCode_Go-4                 10         108689196 ns/op

// BenchmarkExecute_LongCode_Go-4                10         107512182 ns/op
// BenchmarkExecute_LongCode_Go-4                10         109466593 ns/op
// BenchmarkExecute_LongCode_Go-4                10         111880932 ns/op
// BenchmarkExecute_LongCode_Go-4                10         110081228 ns/op
// BenchmarkExecute_LongCode_Go-4                10         117567914 ns/op

// BenchmarkExecute_RecurseCode_Go-4             10         111404591 ns/op
// BenchmarkExecute_RecurseCode_Go-4              9         113388912 ns/op
// BenchmarkExecute_RecurseCode_Go-4             10         109063515 ns/op
// BenchmarkExecute_RecurseCode_Go-4             10         107941414 ns/op
// BenchmarkExecute_RecurseCode_Go-4              9         111985014 ns/op

// BenchmarkExecute_ShortCode_Cpp-4               4         279471396 ns/op
// BenchmarkExecute_ShortCode_Cpp-4               4         267665606 ns/op
// BenchmarkExecute_ShortCode_Cpp-4               4         291253578 ns/op
// BenchmarkExecute_ShortCode_Cpp-4               4         266253802 ns/op
// BenchmarkExecute_ShortCode_Cpp-4               4         262743927 ns/op

// BenchmarkExecute_MediumCode_Cpp-4              4         269005291 ns/op
// BenchmarkExecute_MediumCode_Cpp-4              4         265793952 ns/op
// BenchmarkExecute_MediumCode_Cpp-4              4         281753316 ns/op
// BenchmarkExecute_MediumCode_Cpp-4              4         270195343 ns/op
// BenchmarkExecute_MediumCode_Cpp-4              4         272684110 ns/op

// BenchmarkExecute_LongCode_Cpp-4                3         334136179 ns/op
// BenchmarkExecute_LongCode_Cpp-4                4         332150854 ns/op
// BenchmarkExecute_LongCode_Cpp-4                3         337274232 ns/op
// BenchmarkExecute_LongCode_Cpp-4                3         352949292 ns/op
// BenchmarkExecute_LongCode_Cpp-4                3         355002692 ns/op
