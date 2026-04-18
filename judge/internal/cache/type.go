package cache

import "time"

const (
	LocalReuseTTL = 10 * time.Minute // Judge 本地缓存基础 TTL
	RedisReuseTTL = 12 * time.Hour   // Judge Redis 缓存基础 TTL
	ReuseEmptyTTL = 30 * time.Second // 空值缓存 TTL
	MaxTTLJitter  = 30 * time.Minute // TTL 抖动上限
)

// CachedJudgeResult Judge 结果复用缓存对象
type CachedJudgeResult struct {
	SubmissionID    string             `json:"submission_id"`
	Status          string             `json:"status"`           // AC/WA/TLE/MLE/RE...
	Stdout          string             `json:"stdout"`           // 标准输出
	Stderr          string             `json:"stderr"`           // 标准错误
	TimeNS          int64              `json:"time_ns"`          // 总运行时间(ns)
	MemoryKB        int64              `json:"memory_kb"`        // 总内存(kb)
	ErrorMsg        string             `json:"error_msg"`        // 系统错误
	CaseResults     []CachedCaseResult `json:"case_results"`     // 测试点结果
	TestcaseVersion string             `json:"testcase_version"` // 测试数据版本
	RuntimeVersion  string             `json:"runtime_version"`  // 编译器/运行时版本
	Empty           bool               `json:"empty"`            // 空值标记
}

// CachedCaseResult 测试点缓存结果
type CachedCaseResult struct {
	CaseID    uint64 `json:"case_id"`
	Passed    bool   `json:"pass"`
	Status    string `json:"status"`
	RunTimeNS int64  `json:"runtime_ns"`
	MemoryKB  int64  `json:"memory_kb"`
	Output    string `json:"output"`
}

// localItem 本地缓存项
type localItem struct {
	Value    *CachedJudgeResult // 缓存内容
	ExpireAt time.Time          // 过期时间
}

// ReuseKeyInput 构造缓存 key 的必要字段
type ReuseKeyInput struct {
	ProblemID       uint64 // 题目编号 -- 固定
	Language        string // 语言
	Code            string // 用户代码
	TestcaseVersion string // 测试数据版本 -- 固定
	RuntimeVersion  string // 编译器/运行时版本 -- 固定
}
