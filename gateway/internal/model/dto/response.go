package dto

type TestCaseResult struct {
	Id     uint64 // 测试用例的编号
	Status string // 单个测试用例结果状态
}

type SubmitResponse struct {
	Status  string           `json:"status"`  // 总状态：AC/WA/TLE/CE/RE等
	Stdout  string           `json:"stdout"`  // 标准输出
	Stderr  string           `json:"stderr"`  // 标准错误
	RunTime float64          `json:"runTime"` // 运行时间（毫秒）
	Memory  float64          `json:"memory"`  // 内存使用情况（MB）
	Cases   []TestCaseResult `json:"cases"`   // 每个测试用例的详细结果
}
