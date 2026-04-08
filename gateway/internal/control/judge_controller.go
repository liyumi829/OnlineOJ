package control

import (
	"net/http"
	"online-oj/gateway/internal/model/dto"
	"online-oj/gateway/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type JudgeController struct {
	BaseController
	submitService *service.SubmitService          // 提交服务
	queryService  *service.SubmissionQueryService // 查询服务
}

// NewJudgeController 创建 JudgeController
func NewJudgeController(
	submitService *service.SubmitService,
	queryService *service.SubmissionQueryService,
) *JudgeController {
	return &JudgeController{
		submitService: submitService,
		queryService:  queryService,
	}
}

// Submit 提交代码
func (jc *JudgeController) Submit(c *gin.Context) {
	var req dto.SubmitRequest

	// 1. 绑定 JSON 请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		jc.Fail(c, http.StatusBadRequest, err.Error(), "invalid request body")
		return
	}

	// 2. 调用提交服务
	resp, err := jc.submitService.Submit(c.Request.Context(), &req)
	if err != nil {
		jc.Fail(c, http.StatusInternalServerError, err.Error(), "submit fail")
		return
	}
	// 3. 调用成功返回结构
	jc.Success(c, resp)
}

// Query 请求代码的处理结果
func (jc *JudgeController) Query(c *gin.Context) {
	var req dto.SubmitQueryRequest

	// 支持 RESTful 风格路径参数
	submissionID := c.Param("submission_id")
	if submissionID != "" {
		req.SubmissionID = submissionID
	} else {
		// 否则读取 query 参数
		req.SubmissionID = c.Query("submission_id")
	}

	if req.SubmissionID == "" {
		jc.Fail(c, http.StatusBadRequest, "", "submission_id is required")
		return
	}
	zap.L().Info("", zap.String("submission_id", req.SubmissionID))

	// 调用查询服务
	resp, err := jc.queryService.SubmitQuery(c.Request.Context(), &req)
	if err != nil {
		jc.Fail(c, http.StatusInternalServerError, err.Error(), "query submission failed")
		return
	}

	jc.Success(c, resp)
}
