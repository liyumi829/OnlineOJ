package run

import (
	"encoding/json"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// CompareMode 表示答案比较模式
type CompareMode string

const (
	CompareAuto         CompareMode = "auto"          // 自动识别类型比较
	CompareRawString    CompareMode = "raw_string"    // 原始字符串严格比较
	CompareTrimString   CompareMode = "trim_string"   // 去掉首尾空白后比较
	CompareNormalizeStr CompareMode = "normalize_str" // 标准化空白后比较
	CompareBool         CompareMode = "bool"          // 布尔比较
	CompareInt          CompareMode = "int"           // 整数比较
	CompareFloat        CompareMode = "float"         // 浮点比较
	CompareIntArray     CompareMode = "int_array"     // 整数数组比较
	CompareStringArray  CompareMode = "string_array"  // 字符串数组比较
	CompareJSON         CompareMode = "json"          // JSON 结构比较
)

// CompareAnswer 通用答案比较入口
// 建议在测试用例中增加 CompareMode 字段，由外部传入具体比较类型
func CompareAnswer(userOut, stdOut string, mode CompareMode) bool {
	switch mode {
	case CompareRawString:
		return userOut == stdOut

	case CompareTrimString:
		return strings.TrimSpace(userOut) == strings.TrimSpace(stdOut)

	case CompareNormalizeStr:
		return normalizeString(userOut) == normalizeString(stdOut)

	case CompareBool:
		return compareBool(userOut, stdOut)

	case CompareInt:
		return compareInt(userOut, stdOut)

	case CompareFloat:
		return compareFloat(userOut, stdOut, 1e-9)

	case CompareIntArray:
		return compareIntArray(userOut, stdOut, true)

	case CompareStringArray:
		return compareStringArray(userOut, stdOut, true)

	case CompareJSON:
		return compareJSON(userOut, stdOut)

	case CompareAuto:
		fallthrough
	default:
		return compareAuto(userOut, stdOut)
	}
}

// compareAuto 自动比较：按常见类型依次尝试
func compareAuto(userOut, stdOut string) bool {
	userTrim := strings.TrimSpace(userOut)
	stdTrim := strings.TrimSpace(stdOut)

	// 1. 先尝试布尔值
	if isBool(userTrim) && isBool(stdTrim) {
		return compareBool(userTrim, stdTrim)
	}

	// 2. 再尝试整数
	if isInt(userTrim) && isInt(stdTrim) {
		return compareInt(userTrim, stdTrim)
	}

	// 3. 再尝试浮点数
	if isFloat(userTrim) && isFloat(stdTrim) {
		return compareFloat(userTrim, stdTrim, 1e-9)
	}

	// 4. 再尝试 JSON
	if looksLikeJSON(userTrim) && looksLikeJSON(stdTrim) {
		if compareJSON(userTrim, stdTrim) {
			return true
		}
	}

	// 5. 再尝试整数数组
	if maybeArray(userTrim) || maybeArray(stdTrim) {
		if compareIntArray(userTrim, stdTrim, true) {
			return true
		}
		if compareStringArray(userTrim, stdTrim, true) {
			return true
		}
	}

	// 6. 最后做标准化字符串比较
	return normalizeString(userTrim) == normalizeString(stdTrim)
}

// compareBool 比较布尔值
func compareBool(a, b string) bool {
	av, errA := strconv.ParseBool(strings.ToLower(strings.TrimSpace(a)))
	bv, errB := strconv.ParseBool(strings.ToLower(strings.TrimSpace(b)))
	return errA == nil && errB == nil && av == bv
}

// compareInt 比较整数
func compareInt(a, b string) bool {
	av, errA := strconv.ParseInt(strings.TrimSpace(a), 10, 64)
	bv, errB := strconv.ParseInt(strings.TrimSpace(b), 10, 64)
	return errA == nil && errB == nil && av == bv
}

// compareFloat 比较浮点数
func compareFloat(a, b string, eps float64) bool {
	av, errA := strconv.ParseFloat(strings.TrimSpace(a), 64)
	bv, errB := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if errA != nil || errB != nil {
		return false
	}
	return math.Abs(av-bv) <= eps
}

// compareIntArray 比较整数数组
// ignoreOrder = true 表示忽略顺序
func compareIntArray(a, b string, ignoreOrder bool) bool {
	arrA, okA := parseToIntArray(a)
	arrB, okB := parseToIntArray(b)
	if !okA || !okB {
		return false
	}

	if ignoreOrder {
		sort.Ints(arrA)
		sort.Ints(arrB)
	}

	return reflect.DeepEqual(arrA, arrB)
}

// compareStringArray 比较字符串数组
// ignoreOrder = true 表示忽略顺序
func compareStringArray(a, b string, ignoreOrder bool) bool {
	arrA, okA := parseToStringArray(a)
	arrB, okB := parseToStringArray(b)
	if !okA || !okB {
		return false
	}

	if ignoreOrder {
		sort.Strings(arrA)
		sort.Strings(arrB)
	}

	return reflect.DeepEqual(arrA, arrB)
}

// compareJSON 比较 JSON 结构
func compareJSON(a, b string) bool {
	var objA any
	var objB any

	if err := json.Unmarshal([]byte(strings.TrimSpace(a)), &objA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(b)), &objB); err != nil {
		return false
	}

	return reflect.DeepEqual(objA, objB)
}

// parseToIntArray 将文本解析为整数数组
// 支持：
// [1,2,3]
// 1 2 3
// 1,2,3
// [ 1 , 2 , 3 ]
func parseToIntArray(s string) ([]int, bool) {
	s = cleanArrayText(s)
	if s == "" {
		return []int{}, true
	}

	parts := splitArrayParts(s)
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		v, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		result = append(result, v)
	}

	return result, true
}

// parseToStringArray 将文本解析为字符串数组
func parseToStringArray(s string) ([]string, bool) {
	s = cleanArrayText(s)
	if s == "" {
		return []string{}, true
	}

	parts := splitArrayParts(s)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.Trim(part, `"'`)
		result = append(result, part)
	}

	return result, true
}

// cleanArrayText 清理数组包裹符号
func cleanArrayText(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"(", "",
		")", "",
	)
	return strings.TrimSpace(replacer.Replace(s))
}

// splitArrayParts 统一分割数组元素
func splitArrayParts(s string) []string {
	// 将逗号统一为空格
	s = strings.ReplaceAll(s, ",", " ")
	// 按空白切分，自动兼容多个空格/换行/tab
	return strings.Fields(s)
}

// normalizeString 标准化字符串
// 适合大多数普通输出比较
func normalizeString(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// isBool 判断是否是布尔值
func isBool(s string) bool {
	_, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(s)))
	return err == nil
}

// isInt 判断是否是整数
func isInt(s string) bool {
	_, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return err == nil
}

// isFloat 判断是否是浮点数
func isFloat(s string) bool {
	_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil
}

// looksLikeJSON 判断是否像 JSON
func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

// maybeArray 判断是否可能是数组表达形式
func maybeArray(s string) bool {
	s = strings.TrimSpace(s)
	return strings.Contains(s, ",") ||
		strings.Contains(s, " ") ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) ||
		(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}"))
}
