package ratelimit

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local cutoff = now - window
redis.call('ZREMRANGEBYSCORE', key, '-inf', cutoff)
local count = redis.call('ZCARD', key)

if count < limit then
	redis.call('ZADD', key, now, now)
	redis.call('EXPIRE', key, math.ceil(window / 1000))
	return 1
end

return 0
`)

type redisRateLimiter struct {
	c *redis.Client
}

func NewRedisRateLimiter(connStr string, password ...string) (RateLimiter, error) {
	pw := ""
	if len(password) > 0 {
		pw = password[0]
	}

	var c *redis.Client
	if strings.Contains(connStr, "://") {
		opt, err := redis.ParseURL(connStr)
		if err != nil {
			return nil, err
		}
		c = redis.NewClient(opt)
	} else {
		c = redis.NewClient(&redis.Options{
			Addr:     connStr,
			Password: pw,
		})
	}

	return &redisRateLimiter{c: c}, nil
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	now := time.Now()
	nowMs := now.UnixMilli()
	windowMs := window.Milliseconds()

	result, err := slidingWindowScript.Run(ctx, r.c, []string{key}, nowMs, windowMs, limit).Int()
	if err != nil {
		return false, 0, err
	}

	if result == 1 {
		return true, 0, nil
	}

	retryAfter := window
	return false, retryAfter, nil
}
