package service

import "time"

// ReportCacheConfig centralizes all cache-related parameters for the report service.
// This ensures cache key format, TTLs, and limits are always consistent
// and prevents mismatches between what is cached and what is served.
type ReportCacheConfig struct {
	// TopActiveCacheKeyPrefix is the Redis key prefix for daily top-active-item reports.
	// Full key format: "<TopActiveCacheKeyPrefix>:<date>" e.g. "report:top_active:2026-05-25"
	TopActiveCacheKeyPrefix string

	// EmptyCacheKeyPrefix is the Redis key prefix for sentinel flags marking dates with no movement data.
	// Full key format: "<EmptyCacheKeyPrefix>:<date>" e.g. "report:empty:2026-05-25"
	EmptyCacheKeyPrefix string

	// TodayLogicalTTL is the logical freshness window for today's report.
	// If the cached report's GeneratedAt is older than this, it is considered stale
	// and will be regenerated. This is independent of the Redis physical TTL.
	TodayLogicalTTL time.Duration

	// TodayPhysicalTTL is the Redis eviction TTL for today's report cache entry.
	// Should be >= TodayLogicalTTL to avoid unnecessary cache misses.
	TodayPhysicalTTL time.Duration

	// PastDayPhysicalTTL is the Redis eviction TTL for past-day report cache entries.
	// Past days do not change, so this can be set to a long duration (e.g. 24h).
	PastDayPhysicalTTL time.Duration

	// EmptyDayPhysicalTTL is the Redis eviction TTL for empty-day sentinel flags.
	// If evicted, the service will do one extra DB query before re-setting the flag —
	// the DB sentinel record acts as the permanent guard, so this is safe to be long.
	EmptyDayPhysicalTTL time.Duration

	// GenerationLockTTL is the TTL of the distributed lock used to prevent cache stampede.
	// Should be longer than the expected GenerateDailySummary execution time.
	GenerationLockTTL time.Duration

	// CacheLimit is the maximum number of top-active items stored in cache per day.
	// Must be >= the maximum `limit` value any caller can request, otherwise
	// responses will be silently truncated.
	CacheLimit int

	DisableCache        bool
	DisableSingleFlight bool
	DisableLock         bool
}

func DefaultReportCacheConfig() ReportCacheConfig {
	return ReportCacheConfig{
		TopActiveCacheKeyPrefix: "report:top_active",
		EmptyCacheKeyPrefix:     "report:empty",
		TodayLogicalTTL:         1 * time.Minute,
		TodayPhysicalTTL:        2 * time.Minute, // slightly longer than logical TTL
		PastDayPhysicalTTL:      24 * time.Hour,
		EmptyDayPhysicalTTL:     6 * time.Hour,
		GenerationLockTTL:       15 * time.Second,
		CacheLimit:              50,
	}
}

func (c ReportCacheConfig) topActiveCacheKey(dateStr string) string {
	return c.TopActiveCacheKeyPrefix + ":" + dateStr
}

func (c ReportCacheConfig) emptyCacheKey(dateStr string) string {
	return c.EmptyCacheKeyPrefix + ":" + dateStr
}

func (c ReportCacheConfig) generationLockKey(dateStr string) string {
	return "lock:report:gen:" + dateStr
}

func (c ReportCacheConfig) physicalTTL(isToday bool) time.Duration {
	if isToday {
		return c.TodayPhysicalTTL
	}
	return c.PastDayPhysicalTTL
}
