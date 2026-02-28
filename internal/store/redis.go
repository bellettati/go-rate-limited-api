package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

type RedisConfig struct {
	Addr string
	Password string
	DB int

	DialTimeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

func NewRedisStore(cfg RedisConfig) (*RedisStore, error) {
	opts := &redis.Options{
		Addr: cfg.Addr,
		Password: cfg.Password,
		DB: cfg.DB,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	if opts.Addr == "" {
		opts.Addr = "localhost:6379"
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 2 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 2 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 2 * time.Second
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisStore{client: client}, nil
}

var incrWithTTLLua = redis.NewScript(`
local v = redis.call('INCR', KEYS[1])
if v == 1 then
	redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
return {v, ttl}
`)

func (r *RedisStore) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, time.Duration, error) {
	if ttl < 0 {
		ttl = 0
	}

	ttlMs := ttl.Milliseconds()

	res, err := incrWithTTLLua.Run(ctx, r.client, []string{key}, ttlMs).Result()
	if err != nil {
		return 0, 0, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) != 2 {
		return 0, 0, fmt.Errorf("unexpected lua result type=%T value=%v", res, res)
	}

	val, ok := arr[0].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected counter type=%T value=%v", arr[0], arr[0])
	}

	ttlRaw, ok := arr[1].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected ttl type=%T value=%v", arr[1], arr[1])
	}

	var ttlRemaining time.Duration
	switch {
	case ttlRaw <= 0:
		ttlRemaining = 0
	default:
		ttlRemaining = time.Duration(ttlRaw) * time.Millisecond
	}

	return val, ttlRemaining, nil
}

func (r *RedisStore) Close() error {
	if r.client == nil {
		return nil
	}

	return r.client.Close()
}