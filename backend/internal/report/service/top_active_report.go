package service

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

// TODO: Format error response
func (s *reportService) ListTopActiveItems(
	ctx context.Context,
	date time.Time,
	limit int,
) ([]*entity.DailyItemSummary, error) {

	if limit <= 0 {
		limit = 5
	}

	cacheKey := fmt.Sprintf(
		"report:top_active:%s",
		date.Format("2006-01-02"),
	)

	// 1. Try cache
	cached, found, err := s.cacheStore.Get(ctx, cacheKey)
	if err == nil && found {

		var cachedData []*entity.DailyItemSummary

		err = json.Unmarshal([]byte(cached), &cachedData)
		if err == nil {

			if limit > len(cachedData) {
				limit = len(cachedData)
			}

			return cachedData[:limit], nil
		}
	}

	// 2. Nếu hôm nay -> regenerate summary
	now := time.Now()

	isToday :=
		now.Year() == date.Year() &&
			now.Month() == date.Month() &&
			now.Day() == date.Day()

	if isToday {

		summaries, err := s.movementRepository.
			AggregateDailyItemSummaryFromMovement(ctx, date)
		if err != nil {
			return nil, err
		}

		if len(summaries) > 0 {
			err = s.reportRepository.
				UpsertDailyItemSummary(ctx, summaries)
			if err != nil {
				return nil, err
			}
		}
	}

	// 3. Query full ranking
	result, err := s.reportRepository.
		ListTopActiveItemsByDate(ctx, date, 100)
	if err != nil {
		return nil, err
	}

	// 4. Cache full list
	raw, err := json.Marshal(result)
	if err == nil {

		ttl := 24 * time.Hour

		if isToday {
			ttl = 1 * time.Minute
		}

		_ = s.cacheStore.Set(
			ctx,
			cacheKey,
			string(raw),
			ttl,
		)
	}

	// 5. Slice theo limit
	if limit > len(result) {
		limit = len(result)
	}

	return result[:limit], nil
}
