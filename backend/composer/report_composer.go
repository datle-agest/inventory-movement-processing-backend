package composer

import (
	"inventory-movement-processing/common"
	movementRepository "inventory-movement-processing/internal/movement/repository/postgres"
	reportRepository "inventory-movement-processing/internal/report/repository/postgres"
	"inventory-movement-processing/internal/report/service"
	"inventory-movement-processing/internal/report/transport/http"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type reportHandler interface {
	GenerateDailySummary() gin.HandlerFunc
	ListTopActiveItems() gin.HandlerFunc
}

func ComposeReportService(serviceCtx sctx.ServiceContext) reportHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()

	redisComp := serviceCtx.MustGet(common.KeyComponentRedis).(common.CacheProvider)
	configComp := serviceCtx.MustGet(common.KeyComponentConfig).(common.Config)

	movementRepo := movementRepository.NewMovementRepository(db)
	reportRepo := reportRepository.NewReportRepository(db)

	reportSv := service.NewReportService(reportRepo, movementRepo, redisComp, configComp)

	reportHdl := http.NewReportHandler(reportSv)

	return reportHdl
}
