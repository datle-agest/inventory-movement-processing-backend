package service

import (
	"context"
	"time"
)

func (s *reportService) GenerateDailySummary(
	ctx context.Context,
	date time.Time,
) error {

	summaries, err := s.movementRepository.
		AggregateDailyItemSummaryFromMovement(ctx, date)
	if err != nil {
		return err
	}

	if len(summaries) > 0 {
		err = s.reportRepository.
			UpsertDailyItemSummary(ctx, summaries)
		if err != nil {
			return err
		}
	}

	return nil
}
