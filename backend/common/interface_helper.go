package common

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HTTPServer interface {
	GetPort() int
	GetRouter() *gin.Engine
	Run()
}

type Config interface {
}

type DBProvider interface {
	GetDB() *gorm.DB
}
