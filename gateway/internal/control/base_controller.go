package control

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaseController struct{}

// Render 渲染页面
//
// 参数:
//
//	c gin框架上下文
//	name 获取的html资源，文件名
//	data 需要渲染的数据
//
// 返回值: 无
func (b *BaseController) Render(c *gin.Context, name string, data any) {
	c.HTML(http.StatusOK, name, data)
}

// Fail JSON失败
func (b *BaseController) Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
}

// Success JSON成功
func (b *BaseController) Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}
