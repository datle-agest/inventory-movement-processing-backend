package redisc

import (
	"context"
	"encoding/json"
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

func (r *redisComponent) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (r *redisComponent) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisComponent) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

func (r *redisComponent) Del(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Del(ctx, keys...).Result()
}

func (r *redisComponent) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *redisComponent) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisComponent) IncrBy(ctx context.Context, key string, n int64) (int64, error) {
	return r.client.IncrBy(ctx, key, n).Result()
}

func (r *redisComponent) GetJSON(ctx context.Context, key string, dst interface{}) (bool, error) {
	raw, found, err := r.Get(ctx, key)
	if err != nil || !found {
		return found, err
	}
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		return true, fmt.Errorf("unmarshal redis key %q: %w", key, err)
	}
	return true, nil
}

func (r *redisComponent) SetJSON(ctx context.Context, key string, src interface{}, ttl time.Duration) error {
	b, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshal redis key %q: %w", key, err)
	}
	return r.Set(ctx, key, string(b), ttl)
}

func (r *redisComponent) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, ttl).Result()
}

func (r *redisComponent) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *redisComponent) HGet(ctx context.Context, key, field string) (string, bool, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (r *redisComponent) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

func (r *redisComponent) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *redisComponent) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return r.client.HDel(ctx, key, fields...).Result()
}

func (r *redisComponent) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.LPush(ctx, key, values...).Result()
}

func (r *redisComponent) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.RPush(ctx, key, values...).Result()
}

func (r *redisComponent) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

func (r *redisComponent) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

func (r *redisComponent) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(ctx, key, members...).Result()
}

func (r *redisComponent) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

func (r *redisComponent) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

func (r *redisComponent) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.SRem(ctx, key, members...).Result()
}
