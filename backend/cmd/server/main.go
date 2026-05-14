package main

import (
	"flag"
	v1 "inventory-movement-processing/cmd/server/routes/v1"
	"inventory-movement-processing/common"
	_ "inventory-movement-processing/docs"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/ginc"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	"inventory-movement-processing/pkg/components/gormc"
	"inventory-movement-processing/pkg/components/redisc"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/migrations"
	sctx "inventory-movement-processing/pkg/service_context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @title Inventory Movement Processing API
// @version 1.0
// @description API for inventory movement processing system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api
// @schemes http https

type DBProvider interface {
	GetDB() *gorm.DB
}

func newServiceContext() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("inventory-movement-processing"),
		sctx.WithComponent(configc.NewConfigComponent(common.KeyComponentConfig)),
		sctx.WithComponent(ginc.NewGin(common.KeyComponentGin)),
		sctx.WithComponent(workerc.NewPool(common.KeyCompWorkerPool, 0, 0)),
		sctx.WithComponent(gormc.NewGormDB(common.KeyComponentPostgres, "")),
		sctx.WithComponent(redisc.NewRedis(common.KeyComponentRedis)),
	)
}

func setupRouter(serviceCtx sctx.ServiceContext, router *gin.Engine) {
	router.Use(gin.Logger(), gin.Recovery(), middleware.Recovery(serviceCtx))
	//router.Use(middleware.AuthByRole("manager", "staff"))
	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		gormComp := serviceCtx.MustGet(common.KeyComponentPostgres).(DBProvider)
		db := gormComp.GetDB()

		if err := migrations.RunMigration(db); err != nil {
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
