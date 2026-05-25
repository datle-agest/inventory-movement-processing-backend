package service

import (
	"context"
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

type cacheProvider interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
}

type reportService struct {
	reportRepository reportRepository
	movementService  movementService
	itemService      itemService
	cacheStore       cacheProvider
	cacheConfig      ReportCacheConfig
	logger           logger.Logger
}

func NewReportService(
	reportRepository reportRepository,
	movementService movementService,
	itemService itemService,
	cacheStore cacheProvider,
	cacheConfig ReportCacheConfig,
	logger logger.Logger,
) *reportService {
	return &reportService{
		reportRepository: reportRepository,
		movementService:  movementService,
		itemService:      itemService,
		cacheStore:       cacheStore,
		cacheConfig:      cacheConfig,
		logger:           logger,
	}
}
