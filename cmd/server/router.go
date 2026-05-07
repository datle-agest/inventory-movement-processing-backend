package main

import (
	v1 "inventory-movement-processing/cmd/server/routes/v1"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	sctx "inventory-movement-processing/pkg/service_context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRouter(serviceCtx sctx.ServiceContext, router *gin.Engine) {
	router.Use(gin.Logger(), gin.Recovery(), middleware.Recovery(serviceCtx))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	v1.Register(serviceCtx, api)
	// registerV2Routes(serviceCtx, api)
}
