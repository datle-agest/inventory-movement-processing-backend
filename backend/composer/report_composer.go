package composer

import (
	"inventory-movement-processing/common"
	itemRepository "inventory-movement-processing/internal/item/repository/postgres"
	itemService "inventory-movement-processing/internal/item/service"
	movementRepository "inventory-movement-processing/internal/movement/repository/postgres"
	movementService "inventory-movement-processing/internal/movement/service"
	reportRepository "inventory-movement-processing/internal/report/repository/postgres"
	"inventory-movement-processing/internal/report/service"
	"inventory-movement-processing/internal/report/transport/http"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type reportHandler interface {
	GetDailyReport() gin.HandlerFunc
}

func ComposeReportService(serviceCtx sctx.ServiceContext) reportHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()

	redisComp := serviceCtx.MustGet(common.KeyComponentRedis).(common.CacheProvider)
	logger := serviceCtx.Logger("report-service")

	movementRepo := movementRepository.NewMovementRepository(db)
	reportRepo := reportRepository.NewReportRepository(db)
	itemRepo := itemRepository.NewItemRepository(db)

	itemSv := itemService.NewItemService(itemRepo, logger)
	movementSv := movementService.NewMovementService(movementRepo, itemSv, nil, nil, logger)
	cacheConfig := service.DefaultReportCacheConfig()

	reportSv := service.NewReportService(reportRepo, movementSv, itemSv, redisComp, cacheConfig, logger)

	reportHdl := http.NewReportHandler(reportSv)

	return reportHdl
}
