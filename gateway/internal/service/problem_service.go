package service

import (
	"context"
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/model/entity"
	"online-oj/gateway/internal/repository"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// 实现数据交互的服务
// 单行结构（对应数据库一行）
type TemplateCodeItem struct {
	TemplateCode string
	Enabled      bool
}

// 多行数据 → 切片
type TemplateCodeVO struct {
	Items []TemplateCodeItem
}

type ProblemService struct {
	repo       *repository.ProblemRepository // 操作数据库的权柄
	listCache  sync.Map                      // 分页数据缓存 [key: page, value: 题目列表Slice]
	totalCount int64                         // 题目总数缓存
	countOnce  sync.Once                     // 保证总数只查一次（简单处理方案，如有新增需写个方法重置它）
}

func NewProblemService(repo *repository.Repository) *ProblemService {
	return &ProblemService{
		repo: repository.NewProblemRepositoty(repo),
	}
}

// 1. 获取题目总数
func (ps *ProblemService) GetAllProblemCount(ctx context.Context) int64 {
	ps.countOnce.Do(func() {
		// 调用仓储层/数据库去真实查询一次总数
		ps.totalCount = ps.repo.GetAllProblemCount(ctx)
	})
	return ps.totalCount
}

// 分页获取题目列表
func (ps *ProblemService) GetProblemListByPage(ctx context.Context, page int, pageSize int) (interface{}, error) {
	// 1. 先看缓存有没有这页的数据
	if cachedList, ok := ps.listCache.Load(page); ok {
		return cachedList, nil
	}

	// 2. 缓存没命中，计算真实的分页偏移量（按需加载）
	offset := (page - 1) * pageSize

	// 去数据库里查
	list, err := ps.repo.GetProblemSimpleList(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}
	res := make([]dto.ProblemVO, 0, len(list))
	for _, problem := range list {
		res = append(res, *newProblemVO(&problem))
	}

	// 3. 把新查出来的这页的数据塞进缓存
	ps.listCache.Store(page, res)

	return res, nil
}

// GetProblemDetailVO 获取一个题目的详细描述 用于前端渲染
func (ps *ProblemService) GetProblemDetailVO(ctx context.Context, id uint64) (*dto.ProblemDetailVO, error) {
	problem, err := ps.repo.GetProblemByID(ctx, id) // 获取详细信息
	if err != nil {
		return nil, err
	}
	templateCodes, err := ps.repo.GetTemplateCodesByProblemID(ctx, id)
	if err != nil {
		return nil, err
	}
	var goCode, cppCode string
	const errorCode = "can not use this language!"
	for _, code := range templateCodes {
		switch code.Language {
		case "go":
			if code.Enabled {
				goCode = code.TemplateCode
			} else {
				goCode = errorCode
			}
		case "cpp":
			if code.Enabled {
				cppCode = code.TemplateCode
			} else {
				cppCode = errorCode
			}
		}
	}
	testCases, err := ps.repo.GetSampleCases(ctx, id)
	sampleCases := make([]dto.TestCaseVO, 0, len(testCases))
	for _, testCase := range testCases {
		// 处理一下输入输出
		input := strings.ReplaceAll(testCase.Input, "\n", " ")   // 将 \n 统一为空格
		output := strings.ReplaceAll(testCase.Output, "\n", " ") // 将 \n 统一为空格
		zap.L().Debug("testCase", zap.String("testCase Input", input), zap.String("testCase Output", output))
		sampleCases = append(sampleCases, dto.TestCaseVO{
			Input:  input,
			Output: output,
		})
	}
	// 组装
	return &dto.ProblemDetailVO{
		ProblemVO:       *newProblemVO(problem),
		Desc:            problem.Description,
		TemplateCodeGo:  goCode,
		TemplateCodeCpp: cppCode,
		TestCases:       sampleCases,
	}, nil
}

func newProblemVO(problem *entity.Problem) *dto.ProblemVO {
	var res dto.ProblemVO
	res.ID = strconv.FormatUint(problem.ID, 10)
	var sbuild strings.Builder
	sbuild.WriteString(problem.Title)
	sbuild.WriteRune('⌈')
	sbuild.WriteString(problem.Number)
	sbuild.WriteRune('⌋')
	res.TitleNumber = sbuild.String()
	zap.L().Debug(res.TitleNumber, zap.Uint64("id", problem.ID))
	res.Star = problem.Star
	return &res
}
