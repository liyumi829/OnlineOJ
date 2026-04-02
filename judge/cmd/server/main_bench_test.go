package main

import (
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

// ====================== Benchmark 压测 ======================
// 压测正常代码：两数之和 Go - 多测试用例
func BenchmarkJudge_TwoSum_Go(b *testing.B) {
	code := `
package main
import "fmt"
func twoSum(nums []int, target int) []int {
    m := make(map[int]int)
    for i, num := range nums {
        if j, ok := m[target-num]; ok {
            return []int{j, i}
        }
        m[num] = i
    }
    return nil
}
func main() {
    var n, target int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }
    fmt.Scan(&target)
    res := twoSum(nums, target)
    fmt.Println(res[0], res[1])
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  1,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
		TestCases: []*judge.TestCase{
			{Input: "4\n2 7 11 15\n9", Output: "0 1"},
			{Input: "3\n3 2 4\n6", Output: "1 2"},
		},
	}

	// 重置计时器，排除初始化耗时
	b.ResetTimer()

	// 压测循环
	for i := 0; i < b.N; i++ {
		// 使用永不超时 ctx
		resp, err := client.Judge(ctx(), req)
		if err != nil {
			b.Fatalf("bench failed: %v", err)
		}
		if resp.Status != "AC" {
			b.Fatalf("expect AC, got %s", resp.Status)
		}
	}
}

// 压测 C++ 正常代码
func BenchmarkJudge_TwoSum_Cpp(b *testing.B) {
	code := `
#include <iostream>
#include <vector>
#include <unordered_map>
using namespace std;
vector<int> twoSum(vector<int>& nums, int target) {
    unordered_map<int, int> m;
    for (int i=0; i<nums.size(); i++) {
        if (m.count(target-nums[i])) {
            return {m[target-nums[i]], i};
        }
        m[nums[i]] = i;
    }
    return {};
}
int main() {
    int n, target;
    cin >> n;
    vector<int> nums(n);
    for (int i=0; i<n; i++) cin >> nums[i];
    cin >> target;
    auto res = twoSum(nums, target);
    cout << res[0] << " " << res[1] << endl;
    return 0;
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  2,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
		TestCases: []*judge.TestCase{
			{Input: "4\n2 7 11 15\n9", Output: "0 1"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Judge(ctx(), req)
		if err != nil || resp.Status != "AC" {
			b.Fatalf("bench err: %v, status: %s", err, resp.Status)
		}
	}
}

// 压测 回文数 Go
func BenchmarkJudge_Palindrome_Go(b *testing.B) {
	code := `
package main
import "fmt"
func isPalindrome(x int) bool {
	if x < 0 || (x%10 == 0 && x != 0) {
		return false
	}
	rev := 0
	org := x
	for x > 0 {
		rev = rev*10 + x%10
		x /= 10
	}
	return org == rev
}
func main() {
	var x int
	fmt.Scan(&x)
	fmt.Println(isPalindrome(x))
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  1,
		CpuLimit:  3000_000_000,
		MemoLimit: 64 * 1024,
		TestCases: []*judge.TestCase{
			{Input: "121", Output: "true"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Judge(ctx(), req)
		if err != nil || resp.Status != "AC" {
			b.FailNow()
		}
	}
}

// BenchmarkJudge_TwoSum_Go-4                    10         116562620 ns/op
// BenchmarkJudge_TwoSum_Go-4                     9         111984341 ns/op
// BenchmarkJudge_TwoSum_Go-4                     9         116066673 ns/op
// BenchmarkJudge_TwoSum_Go-4                     9         117058255 ns/op
// BenchmarkJudge_TwoSum_Go-4                     9         119532649 ns/op

// BenchmarkJudge_TwoSum_Cpp-4                    2         510560644 ns/op
// BenchmarkJudge_TwoSum_Cpp-4                    2         514149572 ns/op
// BenchmarkJudge_TwoSum_Cpp-4                    3         494158342 ns/op
// BenchmarkJudge_TwoSum_Cpp-4                    3         493187547 ns/op
// BenchmarkJudge_TwoSum_Cpp-4                    3         488908519 ns/op

// BenchmarkJudge_Palindrome_Go-4                10         114373289 ns/op
// BenchmarkJudge_Palindrome_Go-4                 9         112523754 ns/op
// BenchmarkJudge_Palindrome_Go-4                10         110150980 ns/op
// BenchmarkJudge_Palindrome_Go-4                10         112201023 ns/op
// BenchmarkJudge_Palindrome_Go-4                 9         125675459 ns/op
// PASS
// ok      online-oj/judge/cmd/server      26.914s
