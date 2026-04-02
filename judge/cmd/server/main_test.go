package main

import (
	"context"
	"fmt"
	"online-oj/api/proto/judge"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ===================== 配置 =====================
const (
	serverAddr     = "127.0.0.1:8080"
	codeTypeGo     = 1
	codeTypeCpp    = 2
	cpuLimitNormal = 3 * 1000_000_000 // 3s → ns
	cpuLimitShort  = 500_000_000      // 0.5s → ns（超时测试用）
	memLimitNormal = 64 * 1024        // 64MB → KB
	memLimitLow    = 10 * 1024        // 10MB → KB（内存超限用）
)

// 全局客户端
var client judge.JudgeServiceClient

func init() {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic("connect server fail: " + err.Error())
	}
	client = judge.NewJudgeServiceClient(conn)
}

// ====================== 工具 ======================
// 分隔符
func sep() {
	fmt.Println("\n==================================================")
}

// 永不超时 Context（你确认过的）
func ctx() context.Context {
	return context.Background()
}

// 打印完整结果（完整版：输出所有 CaseResult）
func printResp(resp *judge.JudgeResponse, err error) {
	if err != nil {
		fmt.Println("❌ RPC调用错误:", err)
		return
	}

	sep()
	fmt.Printf("✅ 总状态: %s\n", resp.Status)
	fmt.Printf("📤 总标准输出: %s\n", resp.Stdout)
	fmt.Printf("📥 总标准错误: %s\n", resp.Stderr)
	fmt.Printf("⏱ Max耗时: %.3f ms\n", float64(resp.Time)/1_000_000.0)
	fmt.Printf("📊 Max内存: %d KB\n", resp.Memory)
	sep()

	// ========== 这里是关键：完整输出所有测试用例结果 ==========
	fmt.Println("========== 测试用例详细结果 ==========")
	if len(resp.Results) == 0 {
		fmt.Println("⚠️  无测试用例结果 (可能编译失败/超时/内存溢出)")
	}

	for i, r := range resp.Results {
		fmt.Printf("🔹 用例 #%d\n", i+1)
		fmt.Printf("   ✅ 是否通过: %t\n", r.Passed)
		fmt.Printf("   🚗 结果状态: %s\n", r.Status)
		fmt.Printf("   📤 实际输出: %s\n", r.Output)
		fmt.Printf("   ⏱  用例耗时: %.3f ms\n", float64(r.Time)/1_000_000.0)
		fmt.Printf("   📊 用例内存: %d KB\n", r.Memory)
		fmt.Println("------------------------------------")
	}
	fmt.Println("====================================")
	fmt.Println()
}

// ====================== 1. 两数之和 - Go ======================
func Test_Judge_TwoSum_Go(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：两数之和 - Go")

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
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "4\n2 7 11 15\n9", Output: "0 1"},
			{Input: "3\n3 2 4\n6", Output: "1 2"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 2. 两数之和 - C++ ======================
func Test_Judge_TwoSum_Cpp(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：两数之和 - C++")

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
		CodeType:  codeTypeCpp,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "4\n2 7 11 15\n9", Output: "0 1"},
			{Input: "3\n3 2 4\n6", Output: "1 2"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 3. 回文数 - Go ======================
func Test_Judge_Palindrome_Go(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：回文数 - Go")

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
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "121", Output: "true"},
			{Input: "-121", Output: "false"},
			{Input: "10", Output: "false"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 4. 回文数 - C++ ======================
func Test_Judge_Palindrome_Cpp(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：回文数 - C++")

	code := `
#include <iostream>
using namespace std;
bool isPalindrome(int x) {
    if (x < 0 || (x % 10 == 0 && x != 0)) return false;
    int rev = 0, org = x;
    while (x > 0) {
        rev = rev*10 + x%10;
        x /= 10;
    }
    return org == rev;
}
int main() {
    int x;
    cin >> x;
    cout << boolalpha << isPalindrome(x) << endl;
    return 0;
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  codeTypeCpp,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "121", Output: "true"},
			{Input: "-121", Output: "false"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 5. 编译错误 - Go ======================
func Test_Judge_Compile_Error(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：编译错误 - Go")

	// 语法错误代码
	code := `
package main
import "fmt"
func main() {
    fmt.Println(123  // 少括号
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{{Input: "1", Output: "1"}},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 6. 单个测试用例答案错误 ======================
func Test_Judge_Single_Case_Wrong(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：单个用例答案错误 - Go")

	code := `
package main
import "fmt"
func main() {
    var a int
    fmt.Scan(&a)
    fmt.Println(100) // 故意输出错误答案
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "5", Output: "5"}, // 预期 5，实际输出 100
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 7. 3个用例 2个错误（部分错误） ======================
func Test_Judge_Multi_Case_Partial_Wrong(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：3用例 2个错误 - Go")

	code := `
package main
import "fmt"
func main() {
    var a int
    fmt.Scan(&a)
    if a == 1 {
        fmt.Println(1)
    } else {
        fmt.Println(999)
    }
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "1", Output: "1"}, // 正确
			{Input: "2", Output: "2"}, // 错误
			{Input: "3", Output: "3"}, // 错误
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 8. 代码超时（死循环） ======================
func Test_Judge_Timeout_TLE(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：代码超时 TLE - Go")

	code := `
package main
func main() {
    for {} // 死循环
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitShort, // 0.5s 必超时
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{{Input: "1", Output: "1"}},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 9. 内存超限 MLE（Go 大切片分配） ======================
func Test_Judge_Memory_Exceeded_MLE(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：内存超限 MLE - Go（大切片分配）")

	// 直接分配超大切片，必触发 10MB 限制
	code := `
package main
func main() {
    arr := make([]int, 1024*1024*20) // 约 160MB
	for i := range arr {
		arr[i] = 1
	}
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  codeTypeGo,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitLow, // 10MB 限制
		TestCases: []*judge.TestCase{{Input: "1", Output: "1"}},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 段错误（空指针访问） ======================
func Test_Judge_SegmentFault_CPP(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：段错误 Segmentation Fault - C++")

	// 直接访问空指针，必现段错误
	code := `
#include <iostream>
using namespace std;
int main() {
	cerr << "segment fault occurred" << endl;
    int *p = nullptr;
    *p = 123;
    cout << *p << endl;
    return 0;
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  2,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "1", Output: "1"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}
