package service

import (
	"context"
	"inventory-movement-processing/common"
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

type movementRepository interface {
	AggregateDailyItemSummaryFromMovement(
		ctx context.Context,
		date time.Time,
	) ([]*reportEntity.DailyItemSummary, error)
}

type reportService struct {
	reportRepository   reportRepository
	movementRepository movementRepository
	cacheStore         common.CacheProvider
}

func NewReportService(
	reportRepository reportRepository,
	movementRepository movementRepository,
	cacheStore common.CacheProvider,
) *reportService {
	return &reportService{
		reportRepository:   reportRepository,
		movementRepository: movementRepository,
		cacheStore:         cacheStore,
	}
}
