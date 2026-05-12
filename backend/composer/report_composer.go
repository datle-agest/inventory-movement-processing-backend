package composer

import (
	"inventory-movement-processing/common"
	itemRepository "inventory-movement-processing/internal/item/repository/postgres"
	movementRepository "inventory-movement-processing/internal/movement/repository/postgres"
	reportRepository "inventory-movement-processing/internal/report/repository/postgres"
	"inventory-movement-processing/internal/report/service"
	"inventory-movement-processing/internal/report/transport/http"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type reportHandler interface {
	CreateReport() gin.HandlerFunc
	GetReport() gin.HandlerFunc
}

func ComposeReportService(serviceCtx sctx.ServiceContext) reportHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()

	itemRepo := itemRepository.NewItemRepository(db)
	movementRepo := movementRepository.NewMovementRepository(db)
	reportRepo := reportRepository.NewReportRepository(db)

	reportSv := service.NewReportService(reportRepo, movementRepo, itemRepo)

	reportHdl := http.NewReportHandler(reportSv)

	return reportHdl
}
