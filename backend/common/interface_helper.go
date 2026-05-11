package common

import "github.com/gin-gonic/gin"

type HTTPServer interface {
	GetPort() int
	GetRouter() *gin.Engine
	Run()
}

type Config interface {
}
