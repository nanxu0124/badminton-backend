package web

import "github.com/gin-gonic/gin"

// Result  API 响应的统一格式
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type handler interface {
	RegisterRoutes(s *gin.Engine)
}
