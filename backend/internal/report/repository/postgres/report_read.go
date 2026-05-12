package postgres

import (
	"context"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

func (repo *reportRepository) GetReportByDate(ctx context.Context, date time.Time) (*entity.Report, error) {
	var result entity.Report

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	err := repo.db.WithContext(ctx).
		Where("report_date >= ? AND report_date < ?", start, end).
		First(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}
