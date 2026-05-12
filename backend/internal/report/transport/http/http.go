package http

import (
	"context"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

type reportService interface {
	CreateReport(ctx context.Context, date time.Time) (*entity.Report, error)
	GetReport(ctx context.Context, date time.Time) (*entity.Report, error)
}

type reportHandler struct {
	reportService reportService
}

func NewReportHandler(reportService reportService) *reportHandler {
	return &reportHandler{
		reportService: reportService,
	}
}
