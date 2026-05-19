package service

import (
	"context"
	"inventory-movement-processing/common"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"time"
)

func (s *service) AggregateDailyItemSummaryFromMovement(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	results, err := s.movementRepo.AggregateDailyItemSummaryFromMovement(ctx, start, end)
	if err != nil {
		s.logger.Errorf("[Service][AggregateDailyItemSummaryFromMovement] failed to aggregate summary for date %s: %v",
			date.Format("2006-01-02"), err)

		return nil, common.ErrInternal("failed to aggregate daily summary")
	}

	return results, nil
}
