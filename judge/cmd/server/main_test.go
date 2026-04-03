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

// ====================== 1. 两数之和 - Go（LeetCode标准格式）======================
func Test_Judge_TwoSum_Go(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：两数之和 - Go | 输入输出数组格式 [1,2,3] | 乱序自动AC")

	code := `
package main
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)
func twoSum(nums []int, target int) []int {
	hash := make(map[int]int)
	for i, num := range nums {
		if j, ok := hash[target-num]; ok {
			return []int{j, i}
		}
		hash[num] = i
	}
	return []int{-1, -1}
}
func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	line := sc.Text()
	line = line[1 : len(line)-1]
	parts := strings.Split(line, ",")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		num, _ := strconv.Atoi(strings.TrimSpace(p))
		nums = append(nums, num)
	}
	sc.Scan()
	target, _ := strconv.Atoi(sc.Text())
	res := twoSum(nums, target)
	fmt.Printf("[%d,%d]\n", res[0], res[1])
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  1,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "[2,7,11,15]\n9", Output: "[0,1]"},
			{Input: "[3,2,4]\n6", Output: "[1,2]"},
			{Input: "[3,3]\n6", Output: "[0,1]"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 2. 两数之和 - C++（LeetCode标准格式）======================
func Test_Judge_TwoSum_Cpp(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：两数之和 - C++ | 输入输出数组格式 [1,2,3] | 乱序自动AC")

	code := `
#include <iostream>
#include <vector>
#include <sstream>
#include <string>
#include <unordered_map>
using namespace std;

class Solution {
public:
    vector<int> twoSum(vector<int>& nums, int target) {
        unordered_map<int, int> map;
        for (int i = 0; i < nums.size(); ++i) {
            if (map.count(target - nums[i])) {
                return {map[target - nums[i]], i};
            }
            map[nums[i]] = i;
        }
        return {-1, -1};
    }
};

int main() {
    string line;
    getline(cin, line);
    line = line.substr(1, line.size() - 2);
    stringstream ss(line);
    vector<int> nums;
    string numStr;
    while (getline(ss, numStr, ',')) {
        nums.push_back(stoi(numStr));
    }
    int target;
    cin >> target;
    Solution sol;
    vector<int> res = sol.twoSum(nums, target);
    cout << "[" << res[0] << "," << res[1] << "]" << endl;
    return 0;
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  2,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "[2,7,11,15]\n9", Output: "[0,1]"},
			{Input: "[3,2,4]\n6", Output: "[1,2]"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 3. 回文数 - Go（LeetCode标准格式）======================
func Test_Judge_Palindrome_Go(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：回文数 - Go | LeetCode标准输入输出")

	code := `
package main
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)
func isPalindrome(x int) bool {
	if x < 0 {
		return false
	}
	original := x
	reverted := 0
	for x > 0 {
		reverted = reverted*10 + x%10
		x /= 10
	}
	return original == reverted
}
func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	x, _ := strconv.Atoi(sc.Text())
	res := isPalindrome(x)
	fmt.Println(res)
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  1,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "121", Output: "true"},
			{Input: "-121", Output: "false"},
			{Input: "10", Output: "false"},
			{Input: "0", Output: "true"},
		},
	}

	resp, err := client.Judge(ctx(), req)
	printResp(resp, err)
}

// ====================== 4. 回文数 - C++（LeetCode标准格式）======================
func Test_Judge_Palindrome_Cpp(t *testing.T) {
	sep()
	defer sep()
	t.Log("🧪 测试：回文数 - C++ | LeetCode标准输入输出")

	code := `
#include <iostream>
using namespace std;

class Solution {
public:
    bool isPalindrome(int x) {
        if (x < 0) return false;
        long long ori = x;
        long long rev = 0;
        while (x > 0) {
            rev = rev * 10 + x % 10;
            x /= 10;
        }
        return ori == rev;
    }
};

int main() {
    int x;
    cin >> x;
    Solution sol;
    cout << boolalpha << sol.isPalindrome(x) << endl;
    return 0;
}`

	req := &judge.JudgeRequest{
		Code:      code,
		CodeType:  2,
		CpuLimit:  cpuLimitNormal,
		MemoLimit: memLimitNormal,
		TestCases: []*judge.TestCase{
			{Input: "121", Output: "true"},
			{Input: "-121", Output: "false"},
			{Input: "12321", Output: "true"},
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
