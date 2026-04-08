package router

import (
	"github.com/gin-gonic/gin"

	"online-oj/gateway/internal/control"
)

// RegisterJudgeRoutes 注册判题相关路由
//
// 说明:
//
//   - 判题 /judge
//
//     1. /judge/submit 提交代码
//
//     2. /judge/result 查看判题结果 使用query参数
//
//     3. /submissions/:submission_id 使用 RESTful 风格
func RegisterJudgeRoutes(rg *gin.RouterGroup, judgeController *control.JudgeController) {
	judgeGroup := rg.Group("/judge")
	{
		// 提交代码
		judgeGroup.POST("/submit", judgeController.Submit)

		// 查询判题结果（普通 query 参数方式）
		judgeGroup.GET("/result", judgeController.Query)
	}

	// RESTful 风格查询
	rg.GET("/submissions/:submission_id", judgeController.Query)
}
