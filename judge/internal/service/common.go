package service

import (
	"online-oj/api/proto/judge"
	"strconv"
	"strings"
)

var order = []string{"TLE", "MLE", "RE", "WA", "OLE", "AC"}

// SummarizeCaseStatus 统计测试用例状态分布
// 输入：CaseResult 切片
// 输出：格式如 "AC: [1,2,5]; WA: [3,6]; TLE: [4]"
func summarizeCaseStatus(results []*judge.CaseResult) string {
	// 使用 map 收集各状态的用例编号（自动去重且高效）
	statusMap := make(map[string][]int)
	// 遍历所有用例，按 Status 分类（用例编号从 1 开始）
	for idx, result := range results {
		caseNum := idx + 1
		statusMap[result.Status] = append(statusMap[result.Status], caseNum)
	}
	if len(results) == len(statusMap["AC"]) {
		// 如果所有用例都通过，返回简洁信息
		return "All cases passed"
	}
	// 按优先级顺序输出（TLE > MLE > RE > WA > OLE > AC）
	var sb strings.Builder
	sb.Grow(256) // 预分配容量，减少内存重分配

	// 定义输出顺序（从严重到轻微）
	first := true
	for _, status := range order {
		// fmt.Println("--------------", status, "------", len(statusMap[status]))
		if cases, exists := statusMap[status]; exists && len(cases) > 0 {
			if !first {
				sb.WriteString("; ")
			}
			first = false

			// 写入状态名
			sb.WriteString(status)
			sb.WriteString(": [")

			// 写入用例编号列表
			for i, num := range cases {
				if i > 0 {
					sb.WriteByte(',')
				}
				sb.WriteString(strconv.Itoa(num))
			}
			sb.WriteByte(']')
		}
	}

	return sb.String()
}
