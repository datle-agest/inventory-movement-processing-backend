package service

import (
	"context"
	"time"
)

func (s *reportService) GenerateDailySummary(
	ctx context.Context,
	date time.Time,
) error {
	dateStr := date.Format("2006-01-02")
	s.logger.Infof("Start generating daily summary for date: %s", dateStr)

	summaries, err := s.movementService.
		AggregateDailyItemSummaryFromMovement(ctx, date)

	if err != nil {
		s.logger.Errorf("Failed to aggregate movement data for %s: %v", dateStr, err)
		return err
	}

	if len(summaries) > 0 {
		s.logger.Infof("Upserting %d daily item summaries for %s", len(summaries), dateStr)
		err = s.reportRepository.
			UpsertDailyItemSummary(ctx, summaries)
		if err != nil {
			s.logger.Errorf("Failed to upsert daily item summaries for %s: %v", dateStr, err)
			return err
		}
	} else {
		s.logger.Infof("No movement data found to aggregate for %s", dateStr)
	}

	s.logger.Infof("Successfully generated daily summary for date: %s", dateStr)
	return nil
}
