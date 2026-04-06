package dto

// 一个简易的题目处理
type ProblemVO struct {
	ID          string // 题目ID
	TitleNumber string // 标题+编号 格式 回文数[LC-9]
	Star        string // 难度：Easy/Medium/Difficult
}

// 测试样例的输入和输出
type TestCaseVO struct {
	Input  string // 测试用例的输入
	Output string // 测试用例的输出
}

// ProblemDetail 详细结构用于渲染详情页面
type ProblemDetailVO struct {
	ProblemVO
	Desc            string       // 题目描述
	TemplateCodeGo  string       // Go语言的样例代码 -- 用于前端缓存/快速切换
	TemplateCodeCpp string       // Cpp语言的样例代码
	TestCases       []TestCaseVO // 测试用例
}
