package control

import (
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type JudgeController struct {
	BaseController
	judgeService *service.JudgeService
}

func NewJudgeController(s *service.JudgeService) *JudgeController {
	return &JudgeController{judgeService: s}
}

// --------------------- 提交代码判题 ---------------------
func (con *JudgeController) SubmitCode(c *gin.Context) {
	zap.L().Debug("Submit Code begin: ")
	// 1. 绑定参数
	var req dto.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		con.Fail(c, "Paramter error："+err.Error())
		return
	}
	// 2. 调用Service（RPC判题）
	resp, err := con.judgeService.JudgeCode(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("Judge fail")
		con.Fail(c, "Judge fail")
		return
	}

	// 3. 返回结果给前端
	con.Success(c, resp)
}
