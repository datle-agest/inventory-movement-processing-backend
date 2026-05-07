package main

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/composer"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/ginc"
	"inventory-movement-processing/pkg/components/workerc"
	sctx "inventory-movement-processing/pkg/service_context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func newServiceContext() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("inventory-movement-processing"),
		sctx.WithComponent(configc.NewConfigComponent(common.KeyComponentConfig)),
		sctx.WithComponent(ginc.NewGin(common.KeyComponentGin)),
		sctx.WithComponent(workerc.NewPool(common.KeyCompWorkerPool)),
	)
}

func setupRoute(serviceCtx sctx.ServiceContext, route *gin.RouterGroup) {
	route.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	itemHTTPHandler := composer.ComposeItemService(serviceCtx)

	v1 := route.Group("/v1")

	v1.GET("/item", itemHTTPHandler.GetItem())
}

func main() {
	serviceCtx := newServiceContext()

	if err := serviceCtx.Load(); err != nil {
		log.Fatalln(err)
	}
	defer serviceCtx.Stop()

	ginComp := serviceCtx.MustGet(common.KeyComponentGin).(common.HTTPServer)
	router := ginComp.GetRouter()

	apiGroup := router.Group("/api")

	setupRoute(serviceCtx, apiGroup)

	ginComp.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down application...")
}
