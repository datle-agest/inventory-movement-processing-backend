package redisc

import (
	"context"
	"flag"
	"fmt"
	"inventory-movement-processing/pkg/logger"
	sctx "inventory-movement-processing/pkg/service_context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultHost     = "localhost"
	defaultPort     = 6379
	defaultPassword = ""
	defaultDB       = 0
	defaultPoolSize = 10
)

type Config struct {
	host     string
	port     int
	password string
	db       int
	poolSize int
}

type redisComponent struct {
	*Config
	id     string
	logger logger.Logger
	client *redis.Client
}

func NewRedis(id string) *redisComponent {
	return &redisComponent{
		Config: new(Config),
		id:     id,
	}
}

func (r *redisComponent) ID() string {
	return r.id
}

func (r *redisComponent) InitFlags() {
	flag.StringVar(&r.host, "redis-host", defaultHost, "Redis server host. Default localhost")
	flag.IntVar(&r.port, "redis-port", defaultPort, "Redis server port. Default 6379")
	flag.StringVar(&r.password, "redis-password", defaultPassword, "Redis server password. Default empty")
	flag.IntVar(&r.db, "redis-db", defaultDB, "Redis database index. Default 0")
	flag.IntVar(&r.poolSize, "redis-pool-size", defaultPoolSize, "Redis connection pool size. Default 10")
}

func (r *redisComponent) Activate(serviceContext sctx.ServiceContext) error {
	r.logger = serviceContext.Logger(r.id)
	r.logger.Info("init redis client...")

	addr := fmt.Sprintf("%s:%d", r.host, r.port)

	r.client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     r.password,
		DB:           r.db,
		PoolSize:     r.poolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("connect to redis at %s: %w", addr, err)
	}

	r.logger.Infof("connected to redis at %s (db=%d, pool_size=%d)", addr, r.db, r.poolSize)
	return nil
}

func (r *redisComponent) Stop() error {
	if r.client != nil {
		r.logger.Info("closing redis connection...")
		if err := r.client.Close(); err != nil {
			r.logger.Errorf("close redis client: %v", err)
			return err
		}
		r.logger.Info("redis connection closed")
	}
	return nil
}

func (r *redisComponent) GetClient() *redis.Client {
	return r.client
}
