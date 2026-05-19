package service

import (
	"context"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/logger"
	"time"
)

type reportRepository interface {
	UpsertDailyItemSummary(
		ctx context.Context,
		data []*reportEntity.DailyItemSummary,
	) error
	ListTopActiveItemsByDate(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error)
}

type movementService interface {
	AggregateDailyItemSummaryFromMovement(
		ctx context.Context,
		date time.Time,
	) ([]*reportEntity.DailyItemSummary, error)
}

type itemService interface {
	ListLowStockItems(ctx context.Context) ([]*itemEntity.Item, error)
}

type reportService struct {
	reportRepository reportRepository
	movementUseCase  movementService
	itemRepository   itemService
	cacheStore       common.CacheProvider
	config           common.Config
	logger           logger.Logger
}

func NewReportService(
	reportRepository reportRepository,
	movementUseCase movementService,
	itemRepository itemService,
	cacheStore common.CacheProvider,
	config common.Config,
	logger logger.Logger,
) *reportService {
	return &reportService{
		reportRepository: reportRepository,
		movementUseCase:  movementUseCase,
		itemRepository:   itemRepository,
		cacheStore:       cacheStore,
		config:           config,
		logger:           logger,
	}
}
