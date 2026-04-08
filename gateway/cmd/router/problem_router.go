package router

import (
	"online-oj/gateway/internal/control"

	"github.com/gin-gonic/gin"
)

func RegisterProblemRoutes(rg *gin.RouterGroup, problemController *control.ProblemController) {
	{
		// 获取题目列表
		rg.GET("/list", problemController.ListPage)

		// 获取题目详细信息
		rg.GET("/:id", problemController.ProblemPage)
	}
}
