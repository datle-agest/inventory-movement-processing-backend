package postgres

import (
	"context"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

func (repo *reportRepository) ListTopActiveItemsByDate(ctx context.Context, date time.Time, limit int) ([]*entity.DailyItemSummary, error) {
	var results []*entity.DailyItemSummary

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	err := repo.db.WithContext(ctx).
		Where("summary_date >= ? AND summary_date < ?", start, end).
		Order("(total_in + total_out + total_adjust) DESC").
		Limit(limit).
		Preload("Item").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
