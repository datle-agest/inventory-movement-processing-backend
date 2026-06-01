package composer

import (
	"context"
	"inventory-movement-processing/common"
	itemRepository "inventory-movement-processing/internal/item/repository/postgres"
	itemService "inventory-movement-processing/internal/item/service"
	movementRepository "inventory-movement-processing/internal/movement/repository/postgres"
	movementService "inventory-movement-processing/internal/movement/service"
	reportRepository "inventory-movement-processing/internal/report/repository/postgres"
	"inventory-movement-processing/internal/report/service"
	sctx "inventory-movement-processing/pkg/service_context"
	"time"
)

type reportService interface {
	GenerateDailySummary(
		ctx context.Context,
		date time.Time,
	) error
}

func ComposeCronJob(serviceCtx sctx.ServiceContext) reportService {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()

	logger := serviceCtx.Logger("report-service")

	movementRepo := movementRepository.NewMovementRepository(db)
	reportRepo := reportRepository.NewReportRepository(db)
	itemRepo := itemRepository.NewItemRepository(db)

	itemSv := itemService.NewItemService(itemRepo, logger)
	movementSv := movementService.NewMovementService(movementRepo, itemSv, nil, nil, logger)

	return service.NewReportService(reportRepo, movementSv, nil, nil, nil, logger)
}
