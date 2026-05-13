package http

import (
	"context"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

type reportService interface {
	GenerateDailySummary(
		ctx context.Context,
		date time.Time,
	) error

	ListTopActiveItems(
		ctx context.Context,
		date time.Time,
		limit int,
	) ([]*entity.DailyItemSummary, error)
}

type reportHandler struct {
	reportService reportService
}

func NewReportHandler(reportService reportService) *reportHandler {
	return &reportHandler{
		reportService: reportService,
	}
}
