package service

import (
	"context"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"time"
)

func (s *service) AggregateDailyItemSummaryFromMovement(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	return s.movementRepo.AggregateDailyItemSummaryFromMovement(ctx, start, end)
}
