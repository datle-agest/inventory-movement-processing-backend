package service

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

// TODO: Format error response
// TODO: What if cronjob dead and yesterday data stale?
func (s *reportService) GetDailyReport(
	ctx context.Context,
	date time.Time,
	limit int,
) (*entity.TopActiveItemsResult, error) {

	if limit <= 0 {
		limit = 5
	}

	now := time.Now()
	isToday :=
		now.Year() == date.Year() &&
			now.Month() == date.Month() &&
			now.Day() == date.Day()

	cacheKey := fmt.Sprintf(
		"report:top_active:%s",
		date.Format("2006-01-02"),
	)

	// 1. Try cache
	cached, found, err := s.cacheStore.Get(ctx, cacheKey)
	if err == nil && found {
		var cachedTop []*entity.DailyItemSummary

		if err = json.Unmarshal([]byte(cached), &cachedTop); err == nil {
			if limit > len(cachedTop) {
				limit = len(cachedTop)
			}

			result := &entity.TopActiveItemsResult{
				TopItems: cachedTop[:limit],
			}

			if isToday {
				result.LowStockItems, err = s.fetchAllLowStockItems(ctx)
				if err != nil {
					return nil, err
				}
			}

			return result, nil
		}
	}

	// 2. If today -> regenerate summary
	if isToday {
		if err := s.GenerateDailySummary(ctx, date); err != nil {
			return nil, common.ErrInternal(err.Error())
		}
	}

	// 3. Query full ranking
	topItems, err := s.reportRepository.
		ListTopActiveItemsByDate(ctx, date, s.config.GetReportCacheLimit())
	if err != nil {
		return nil, err
	}

	// 4. Cache full top list
	if raw, err := json.Marshal(topItems); err == nil {
		ttl := 24 * time.Hour
		if isToday {
			ttl = 1 * time.Minute
		}
		_ = s.cacheStore.Set(ctx, cacheKey, string(raw), ttl)
	}

	// 5. Slice limit
	if limit > len(topItems) {
		limit = len(topItems)
	}

	result := &entity.TopActiveItemsResult{
		TopItems: topItems[:limit],
	}

	if isToday {
		result.LowStockItems, err = s.fetchAllLowStockItems(ctx)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *reportService) fetchAllLowStockItems(ctx context.Context) ([]*itemEntity.Item, error) {
	items, err := s.itemRepository.ListLowStockItems(ctx)
	if err != nil {
		return nil, common.ErrInternal(err.Error())
	}

	return items, nil
}
