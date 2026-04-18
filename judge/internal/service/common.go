package service

import (
	"online-oj/api/proto/judge"
	jdcache "online-oj/judge/internal/cache"
	"online-oj/judge/internal/compile"
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

func getCodeType(value int32) compile.Type {
	switch value {
	case 1:
		return compile.GoType
	case 2:
		return compile.CppType
	default:
	}
	return compile.UnKnownType
}

// 缓存结构 → proto 响应
func toJudgeResponse(c *jdcache.CachedJudgeResult) *judge.JudgeResponse {
	return &judge.JudgeResponse{
		SubmissionId: c.SubmissionID,
		Status:       c.Status,
		Stdout:       c.Stdout,
		Stderr:       c.Stderr,
		Time:         c.TimeNS,
		Memory:       c.MemoryKB,
		Results:      toProtoResults(c.CaseResults),
	}
}

// proto 响应 → 缓存结构
func toCachedResult(r *judge.JudgeResponse) *jdcache.CachedJudgeResult {
	return &jdcache.CachedJudgeResult{
		SubmissionID: r.SubmissionId,
		Status:       r.Status,
		Stdout:       r.Stdout,
		Stderr:       r.Stderr,
		TimeNS:       r.Time,
		MemoryKB:     r.Memory,
		CaseResults:  toCacheResults(r.Results),
	}
}

// 转换 CaseResults
func toCacheResults(rs []*judge.CaseResult) []jdcache.CachedCaseResult {
	var out []jdcache.CachedCaseResult
	for index, r := range rs {
		out = append(out, jdcache.CachedCaseResult{
			CaseID:    uint64(index + 1),
			Passed:    r.Passed,
			Status:    r.Status,
			RunTimeNS: r.Time,
			MemoryKB:  r.Memory,
			Output:    r.Output,
		})
	}
	return out
}

func toProtoResults(rs []jdcache.CachedCaseResult) []*judge.CaseResult {
	var out []*judge.CaseResult
	for _, r := range rs {
		out = append(out, &judge.CaseResult{
			Passed: r.Passed,
			Time:   r.RunTimeNS,
			Memory: r.MemoryKB,
			Status: r.Status,
			Output: r.Output,
		})
	}
	return out
}

// 语言数字转字符串
func getLanguageString(t int32) string {
	switch t {
	case 1:
		return "go"
	case 2:
		return "cpp"
	default:
		return "unknown"
	}
}
