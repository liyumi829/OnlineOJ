package control

import (
	"net/http"
	"online-oj/gateway/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProblemController struct {
	BaseController
	problemService *service.ProblemService
}

func NewProblemController(s *service.ProblemService) *ProblemController {
	return &ProblemController{problemService: s}
}

// 首页：访问 / 自动渲染 index.html
func (con *ProblemController) IndexPage(c *gin.Context) {
	con.Render(c, "index.html", nil)
}

// --------------------- 页面路由 ---------------------
func (con *ProblemController) ListPage(c *gin.Context) {
	// 获取并处理参数
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	pageSize := 10 // 每页10条

	// 获取总数，并计算页数防越界
	total := int(con.problemService.GetAllProblemCount()) // 获取题目总数
	totalPage := (total + pageSize - 1) / pageSize        // 向上取整算总页数
	// 防止恶意页码
	if page < 1 {
		page = 1
	}
	// 防止用户强行输入超大页码
	if totalPage > 0 && page > totalPage {
		page = totalPage
	}

	// 3. 一行代码获取数据（Service 会自动处理命中缓存还是走 DB）
	list, err := con.problemService.GetProblemListByPage(page, pageSize)
	if err != nil {
		zap.L().Error("Get problem list fail", zap.String("error", err.Error()))
		con.Fail(c, "Failed to get question list.")
		return
	}

	// 4. 计算分页按钮逻辑
	prevPage := page - 1
	nextPage := page + 1
	if prevPage < 1 {
		prevPage = 1
	}
	if nextPage > totalPage {
		nextPage = totalPage
	}

	// 5. 传给前端模板
	c.HTML(http.StatusOK, "all_problems.html", gin.H{
		"QuestionList": list,
		"CurrentPage":  page,
		"PrevPage":     prevPage,
		"NextPage":     nextPage,
		"Total":        total,
		"TotalPage":    totalPage,
	})
}

func (con *ProblemController) ProblemPage(c *gin.Context) {
	// 题目详情页
	idStr := c.Param("id")
	uintId, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		con.Fail(c, "Illegal question ID parameter")
		return
	}
	vo, err := con.problemService.GetProblemDetailVO(uintId)

	if err != nil {
		zap.L().Error("Get a problem Detail fail", zap.String("error", err.Error()))
		con.Fail(c, "Get a problem Detail fail")
		return
	}
	con.Render(c, "one_problem.html", vo)
}
