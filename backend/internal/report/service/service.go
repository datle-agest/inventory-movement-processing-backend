package service

import (
	"context"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"time"
)

type reportRepository interface {
	GetReportByDate(ctx context.Context, date time.Time) (*reportEntity.Report, error)
	CreateReport(ctx context.Context, data *reportEntity.Report) error
}

type movementRepository interface {
	GetByDateRange(ctx context.Context, from, to time.Time) ([]movementEntity.Movement, error)

	GetAggregatedByItem(ctx context.Context, from, to time.Time) (
		map[int32]int32, // item_id -> total quantity
		error,
	)

	GetSummaryByType(ctx context.Context, from, to time.Time) (
		*movementEntity.MovementSummary, // IN/OUT/ADJUST totals
		error,
	)
}

type itemRepository interface {
	GetItemByIDs(ctx context.Context, ids []int32) ([]itemEntity.Item, error)

	GetLowStockItems(ctx context.Context) ([]itemEntity.Item, error)
}

type reportService struct {
	reportRepository   reportRepository
	movementRepository movementRepository
	itemRepository     itemRepository
}

func NewReportService(
	reportRepository reportRepository,
	movemovementRepository movementRepository,
	itemRepository itemRepository,
) *reportService {
	return &reportService{
		reportRepository:   reportRepository,
		movementRepository: movemovementRepository,
		itemRepository:     itemRepository,
	}
}
