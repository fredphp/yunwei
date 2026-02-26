package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func Ok(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  "操作成功",
	})
}

func OkWithMessage(message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: nil,
		Msg:  message,
	})
}

func OkWithData(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  "操作成功",
	})
}

func OkWithPage(list interface{}, total int64, page, pageSize int, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: PageResult{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Msg: "获取成功",
	})
}

func Fail(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 1,
		Data: nil,
		Msg:  "操作失败",
	})
}

func FailWithMessage(message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 1,
		Data: nil,
		Msg:  message,
	})
}
