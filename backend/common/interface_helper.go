package common

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HTTPServer interface {
	GetPort() int
	GetRouter() *gin.Engine
	Run()
}

type Config interface {
	GetReportCacheLimit() int
}

type DBProvider interface {
	GetDB() *gorm.DB
}

type CacheProvider interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
