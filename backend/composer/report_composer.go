package composer

import (
	"inventory-movement-processing/common"
	itemRepository "inventory-movement-processing/internal/item/repository/postgres"
	movementRepository "inventory-movement-processing/internal/movement/repository/postgres"
	movementService "inventory-movement-processing/internal/movement/service"
	reportRepository "inventory-movement-processing/internal/report/repository/postgres"
	"inventory-movement-processing/internal/report/service"
	"inventory-movement-processing/internal/report/transport/http"
	"inventory-movement-processing/pkg/components/workerc"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type reportHandler interface {
	GetDailyReport() gin.HandlerFunc
}

func ComposeReportService(serviceCtx sctx.ServiceContext) reportHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()
	workerPool := serviceCtx.MustGet(common.KeyCompWorkerPool).(workerc.WorkerPool)

	redisComp := serviceCtx.MustGet(common.KeyComponentRedis).(common.CacheProvider)
	configComp := serviceCtx.MustGet(common.KeyComponentConfig).(common.Config)

	movementRepo := movementRepository.NewMovementRepository(db)
	reportRepo := reportRepository.NewReportRepository(db)
	itemRepo := itemRepository.NewItemRepository(db)

	movementSv := movementService.NewMovementService(movementRepo, workerPool)

	reportSv := service.NewReportService(reportRepo, movementSv, itemRepo, redisComp, configComp)

	reportHdl := http.NewReportHandler(reportSv)

	return reportHdl
}
