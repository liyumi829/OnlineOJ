package dto

// 前端交给网关的请求
// 主要设计接口：
// 提交编写的代码之后查看提交结果

type SubmitRequest struct {
	ID       string `json:"id"`       // 题目ID
	Code     string `json:"code"`     // 用户提交的源代码
	Language uint32 `json:"language"` // 代码类型 1=Go 2=C++（用于区分语言）
	// Input         string `json:"-"`          // 用户自己的测试用例输出（未实现）后台没有标准答案的代码
}
