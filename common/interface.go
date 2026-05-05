package common

import "github.com/gin-gonic/gin"

type GinComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
	Run()
}

type HTTPServer interface {
	GetPort() int
	GetRouter() *gin.Engine
	Run()
}
