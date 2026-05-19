package service

import (
	"context"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"time"
)

type reportRepository interface {
	UpsertDailyItemSummary(
		ctx context.Context,
		data []*reportEntity.DailyItemSummary,
	) error
	ListTopActiveItemsByDate(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error)
}

type movementUseCase interface {
	AggregateDailyItemSummaryFromMovement(
		ctx context.Context,
		date time.Time,
	) ([]*reportEntity.DailyItemSummary, error)
}

type itemRepository interface {
	ListLowStockItems(ctx context.Context) ([]*itemEntity.Item, error)
}

type reportService struct {
	reportRepository reportRepository
	movementUseCase  movementUseCase
	itemRepository   itemRepository
	cacheStore       common.CacheProvider
	config           common.Config
}

func NewReportService(
	reportRepository reportRepository,
	movementUseCase movementUseCase,
	itemRepository itemRepository,
	cacheStore common.CacheProvider,
	config common.Config,
) *reportService {
	return &reportService{
		reportRepository: reportRepository,
		movementUseCase:  movementUseCase,
		itemRepository:   itemRepository,
		cacheStore:       cacheStore,
		config:           config,
	}
}
