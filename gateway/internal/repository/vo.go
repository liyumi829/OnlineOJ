package repository

type ProblemSampleVO struct {
	Id     uint64 // 题目ID
	Number string // 题目编号
	Title  string // 标题
	Star   string // 难度
}

type ProblemSimpleVO struct {
	Input  string
	Output string
}

type TemplateCodeVO struct {
	TemplateCode string
	Enabled      bool
}
