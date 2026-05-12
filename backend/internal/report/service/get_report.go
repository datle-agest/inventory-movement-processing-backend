package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

func (s *reportService) GetReport(ctx context.Context, date time.Time) (*entity.Report, error) {
	report, err := s.reportRepository.GetReportByDate(ctx, date)
	if err != nil {
		return nil, common.ErrNotFound(err.Error())
	}

	return report, nil
}
