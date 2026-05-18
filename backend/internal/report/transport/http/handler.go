package http

import (
	"context"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

type reportService interface {
	GetDailyReport(
		ctx context.Context,
		date time.Time,
		limit int,
	) (*entity.TopActiveItemsResult, error)
}

type reportHandler struct {
	reportService reportService
}

func NewReportHandler(reportService reportService) *reportHandler {
	return &reportHandler{
		reportService: reportService,
	}
}
