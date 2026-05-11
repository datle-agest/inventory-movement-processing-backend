package main

import (
	"flag"
	migration "inventory-movement-processing/cmd/migrations"
	v1 "inventory-movement-processing/cmd/server/routes/v1"
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/ginc"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	"inventory-movement-processing/pkg/components/gormc"
	"inventory-movement-processing/pkg/components/redisc"
	"inventory-movement-processing/pkg/components/workerc"
	sctx "inventory-movement-processing/pkg/service_context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DBProvider interface {
	GetDB() *gorm.DB
}

func newServiceContext() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("inventory-movement-processing"),
		sctx.WithComponent(configc.NewConfigComponent(common.KeyComponentConfig)),
		sctx.WithComponent(ginc.NewGin(common.KeyComponentGin)),
		sctx.WithComponent(workerc.NewPool(common.KeyCompWorkerPool, 1, 1)),
		sctx.WithComponent(gormc.NewGormDB(common.KeyComponentPostgres, "")),
		sctx.WithComponent(redisc.NewRedis(common.KeyComponentRedis)),
	)
}

func setupRouter(serviceCtx sctx.ServiceContext, router *gin.Engine) {
	router.Use(gin.Logger(), gin.Recovery(), middleware.Recovery(serviceCtx))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	v1.Register(serviceCtx, api)
	// registerV2Routes(serviceCtx, api)
}

func main() {
	printEnv := flag.Bool("print-env", false, "print resolved environment variables")

	runMigrate := flag.Bool("migrate", false, "run database migrations on startup")

	flag.Parse()

	serviceCtx := newServiceContext()

	if *printEnv {
		serviceCtx.OutEnv()
		return
	}

	if err := serviceCtx.Load(); err != nil {
		log.Fatalln(err)
	}
	defer serviceCtx.Stop()

	if *runMigrate {
		gormComp := serviceCtx.MustGet("gorm").(DBProvider)
		db := gormComp.GetDB()

		if err := migration.RunMigration(db); err != nil {
			log.Fatalf("AutoMigrate failed on startup: %v", err)
		}
	}

	ginComp := serviceCtx.MustGet(common.KeyComponentGin).(common.HTTPServer)
	setupRouter(serviceCtx, ginComp.GetRouter())
	ginComp.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down application...")
}
