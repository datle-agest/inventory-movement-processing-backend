package main

import (
	"context"
	"flag"
	v1 "inventory-movement-processing/cmd/server/routes/v1"
	"inventory-movement-processing/common"
	"inventory-movement-processing/composer"
	_ "inventory-movement-processing/docs"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/cronc"
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
	"time"

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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGci..."

type DBProvider interface {
	GetDB() *gorm.DB
}

// CronConfig là interface để lấy cron schedule từ config component.
type CronConfig interface {
	GetDailyCronSchedule() string
}

func newServiceContext() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("inventory-movement-processing"),
		sctx.WithComponent(configc.NewConfigComponent(common.KeyComponentConfig)),
		sctx.WithComponent(ginc.NewGin(common.KeyComponentGin)),
		sctx.WithComponent(workerc.NewPool(common.KeyCompWorkerPool, 0, 0)),
		sctx.WithComponent(gormc.NewGormDB(common.KeyComponentPostgres, "")),
		sctx.WithComponent(redisc.NewRedis(common.KeyComponentRedis)),
		sctx.WithComponent(cronc.NewCron(common.KeyComponentCron)),
	)
}

func setupRouter(serviceCtx sctx.ServiceContext, router *gin.Engine) {
	sysLogger := serviceCtx.Logger("api-audit")

	router.Use(gin.Recovery(), middleware.Recovery(serviceCtx))
	router.Use(middleware.AuditLog(sysLogger))
	router.Use(middleware.Metrics())
	router.Use(middleware.ResponseTime())
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Expose-Headers", "X-Response-Time")
		c.Next()
	})

	router.GET("/metrics", middleware.PrometheusHandler())

	cfg := serviceCtx.MustGet(common.KeyComponentConfig).(middleware.Config)
	router.Use(middleware.AuthByRole(cfg))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	v1.Register(serviceCtx, api)
}

func setupJob(serviceCtx sctx.ServiceContext) {
	cronComp := serviceCtx.MustGet(common.KeyComponentCron).(cronc.CronComponent)
	cfg := serviceCtx.MustGet(common.KeyComponentConfig).(CronConfig)
	logger := serviceCtx.Logger("cron-setup")

	schedule := cfg.GetDailyCronSchedule()
	logger.Infof("registering daily-inventory-sync schedule=%q", schedule)

	reportSv := composer.ComposeCronJob(serviceCtx)

	if err := cronComp.AddJob(cronc.JobDefinition{
		Name:     "daily-inventory-sync",
		Schedule: schedule,
		Handler: func() {
			yesterday := time.Now().AddDate(0, 0, -1)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := reportSv.GenerateDailySummary(ctx, yesterday); err != nil {
				logger.Errorf("daily-inventory-sync failed date=%s err=%v", yesterday.Format("2006-01-02"), err)
				return
			}
			logger.Infof("daily-inventory-sync completed date=%s", yesterday.Format("2006-01-02"))
		},
	}); err != nil {
		log.Fatalf("register cron job: %v", err)
	}
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

	setupJob(serviceCtx)

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
