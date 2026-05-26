package service

import (
	"context"
	"encoding/json"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

func (s *reportService) GetDailyReport(
	ctx context.Context,
	date time.Time,
	limit int,
) (*entity.TopActiveItemsResult, error) {

	if limit <= 0 {
		limit = 5
	}

	now := time.Now()
	isToday := isSameDay(now, date)
	dateStr := date.Format("2006-01-02")
	cacheKey := s.cacheConfig.topActiveCacheKey(dateStr)

	// 1. Try cache
	if items, ok, err := s.loadFromCache(ctx, cacheKey, isToday); err != nil {
		s.logger.Warnf("Cache load failed for key %s: %v", cacheKey, err)
	} else if ok {
		return s.buildResult(ctx, items, limit, isToday)
	}

	// 2. Single Flight wraps the rest
	// Key is dateStr so that requests are collected on the same day
	type sfResult struct {
		items []*entity.DailyItemSummary
		stale bool
	}

	v, err, _ := s.sfGroup.Do(dateStr, func() (interface{}, error) {
		// 2a. Check empty-day sentinel
		if empty, err := s.isKnownEmptyDay(ctx, dateStr); err != nil {
			s.logger.Warnf("Empty-day cache check failed for %s: %v", dateStr, err)
		} else if empty {
			s.logger.Infof("Known empty day (cache flag) for %s", dateStr)
			return &sfResult{items: nil, stale: false}, nil
		}

		// 2b. Acquire distributed lock
		lockKey := s.cacheConfig.generationLockKey(dateStr)
		acquired, err := s.cacheStore.SetNX(ctx, lockKey, "1", s.cacheConfig.GenerationLockTTL)
		if err != nil {
			s.logger.Warnf("Failed to acquire generation lock for %s: %v", dateStr, err)
		}
		if acquired {
			defer func() {
				if _, delErr := s.cacheStore.Del(ctx, lockKey); delErr != nil {
					s.logger.Warnf("Failed to release generation lock for %s: %v", dateStr, delErr)
				}
			}()
		} else {
			s.logger.Infof("Generation lock held by another instance for %s, querying DB directly", dateStr)
		}

		// 2c. Query DB
		topItems, err := s.reportRepository.
			ListTopActiveItemsByDate(ctx, date, s.cacheConfig.CacheLimit)
		if err != nil {
			s.logger.Errorf("ListTopActiveItemsByDate failed for %s: %v", dateStr, err)
			return nil, common.ErrInternal("cannot query top active items")
		}

		isStale := isDataStale(topItems, date)

		// 2d. Generate
		if acquired && (isToday || isStale) {
			s.logger.Infof("Generating daily summary for %s", dateStr)
			if err := s.GenerateDailySummary(ctx, date); err != nil {
				s.logger.Errorf("GenerateDailySummary failed for %s: %v", dateStr, err)
				return nil, common.ErrInternal("failed to process daily summary")
			}

			topItems, err = s.reportRepository.
				ListTopActiveItemsByDate(ctx, date, s.cacheConfig.CacheLimit)
			if err != nil {
				s.logger.Errorf("ListTopActiveItemsByDate failed after generation for %s: %v", dateStr, err)
				return nil, common.ErrInternal("cannot query top active items")
			}
			isStale = false
		}

		// 2e. Handle empty result
		if isEmptyResult(topItems) {
			s.logger.Infof("No data for %s, setting empty-day cache flag", dateStr)
			if err := s.cacheStore.Set(
				ctx,
				s.cacheConfig.emptyCacheKey(dateStr),
				"1",
				s.cacheConfig.EmptyDayPhysicalTTL,
			); err != nil {
				s.logger.Warnf("Failed to set empty-day cache flag for %s: %v", dateStr, err)
			}
			return &sfResult{items: nil, stale: false}, nil
		}

		// 2f. Populate cache
		if !isStale {
			s.cacheReport(ctx, cacheKey, topItems, now, isToday)
		}

		return &sfResult{items: topItems, stale: isStale}, nil
	})

	if err != nil {
		return nil, err
	}

	res := v.(*sfResult)
	return s.buildResult(ctx, res.items, limit, isToday)
}

// loadFromCache attempts to read and validate a cached report.
// Returns (items, true, nil) on a valid fresh hit; (nil, false, nil) on miss/stale; (nil, false, err) on error.
func (s *reportService) loadFromCache(
	ctx context.Context,
	cacheKey string,
	isToday bool,
) ([]*entity.DailyItemSummary, bool, error) {

	cached, found, err := s.cacheStore.Get(ctx, cacheKey)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	var cachedReport entity.CachedReport
	if err = json.Unmarshal([]byte(cached), &cachedReport); err != nil {
		s.logger.Errorf("Failed to unmarshal cache for key %s: %v", cacheKey, err)
		return nil, false, nil // treat corrupt cache as miss
	}

	isFresh := !isToday || time.Since(cachedReport.GeneratedAt) <= s.cacheConfig.TodayLogicalTTL
	if !isFresh {
		s.logger.Infof("Cache stale for key %s (age: %v)", cacheKey, time.Since(cachedReport.GeneratedAt))
		return nil, false, nil
	}

	s.logger.Infof("Cache hit (fresh) for key %s", cacheKey)
	return cachedReport.Items, true, nil
}

// isKnownEmptyDay checks the Redis flag set for days confirmed to have no movement data.
func (s *reportService) isKnownEmptyDay(ctx context.Context, dateStr string) (bool, error) {
	val, found, err := s.cacheStore.Get(ctx, s.cacheConfig.emptyCacheKey(dateStr))
	if err != nil {
		return false, err
	}
	return found && val == "1", nil
}

// cacheReport serializes and stores the report in Redis with appropriate TTL.
func (s *reportService) cacheReport(
	ctx context.Context,
	cacheKey string,
	items []*entity.DailyItemSummary,
	generatedAt time.Time,
	isToday bool,
) {
	cachedReport := entity.CachedReport{
		Items:       items,
		GeneratedAt: generatedAt,
	}
	raw, err := json.Marshal(cachedReport)
	if err != nil {
		s.logger.Errorf("Failed to marshal report for caching (key %s): %v", cacheKey, err)
		return
	}
	if err := s.cacheStore.Set(ctx, cacheKey, string(raw), s.cacheConfig.physicalTTL(isToday)); err != nil {
		s.logger.Warnf("Failed to set cache for key %s: %v", cacheKey, err)
	}
}

// isDataStale checks whether the DB result for a given date is outdated.
// It only checks items matching the requested date to avoid false positives
// from leftover records of other dates (e.g. unstable sort).
func isDataStale(items []*entity.DailyItemSummary, date time.Time) bool {
	startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	startOfNextDay := startOfDate.AddDate(0, 0, 1)

	now := time.Now()
	isToday := isSameDay(now, date)

	for _, item := range items {
		if !isSameDay(item.SummaryDate, date) {
			continue // skip records that don't belong to this date
		}
		if item.UpdatedAt == nil {
			return true
		}
		if isToday {
			// For today: stale if UpdatedAt predates today's midnight
			if item.UpdatedAt.Before(startOfDate) {
				return true
			}
		} else {
			// For past days: stale if UpdatedAt predates the next day's midnight
			// (i.e. was never updated after the day ended)
			if item.UpdatedAt.Before(startOfNextDay) {
				return true
			}
		}
	}

	return len(items) == 0
}

// isEmptyResult returns true when topItems contains only sentinel records (ItemID == 0)
// or is empty entirely, meaning the day had no real movement data.
func isEmptyResult(items []*entity.DailyItemSummary) bool {
	for _, item := range items {
		if item.ItemID != 0 {
			return false
		}
	}
	return true
}

func (s *reportService) buildResult(
	ctx context.Context,
	topItems []*entity.DailyItemSummary,
	limit int,
	isToday bool,
) (*entity.TopActiveItemsResult, error) {

	// Filter out sentinel records before slicing
	realItems := make([]*entity.DailyItemSummary, 0, len(topItems))
	for _, item := range topItems {
		if item.ItemID != 0 {
			realItems = append(realItems, item)
		}
	}

	if limit > len(realItems) {
		limit = len(realItems)
	}

	result := &entity.TopActiveItemsResult{
		TopItems: realItems[:limit],
	}

	if isToday {
		lowStock, err := s.fetchAllLowStockItems(ctx)
		if err != nil {
			s.logger.Errorf("Failed to fetch low stock items: %v", err)
			return nil, common.ErrInternal("failed to fetch low stock items")
		}
		result.LowStockItems = lowStock
	}

	return result, nil
}

func (s *reportService) fetchAllLowStockItems(ctx context.Context) ([]*itemEntity.Item, error) {
	items, err := s.itemService.ListLowStockItems(ctx)
	if err != nil {
		s.logger.Errorf("ItemRepository.ListLowStockItems failed: %v", err)
		return nil, err
	}
	return items, nil
}

func isSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
