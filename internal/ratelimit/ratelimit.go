package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// tokenBucketScript implements a token bucket in Redis, executed atomically
// to avoid race conditions between concurrent requests from the same key.
// KEYS[1] = bucket key
// ARGV[1] = max tokens (bucket capacity / burst size)
// ARGV[2] = refill rate (tokens per second)
// ARGV[3] = current unix time (float, seconds)
// ARGV[4] = tokens requested (always 1 for our use case)
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local bucket = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    last_refill = now
end

local elapsed = math.max(0, now - last_refill)
tokens = math.min(capacity, tokens + (elapsed * refill_rate))

local allowed = 0
if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
redis.call("EXPIRE", key, 3600)

return {allowed, tostring(tokens)}
`)

type Limiter struct {
	client   *redis.Client
	capacity float64
	refill   float64
}

func NewLimiter(client *redis.Client, capacity, refillPerSecond float64) *Limiter {
	return &Limiter{client: client, capacity: capacity, refill: refillPerSecond}
}

// Allow checks whether the given key has capacity for one more request right now.
func (l *Limiter) Allow(ctx context.Context, key string) (allowed bool, remaining float64, err error) {
	now := float64(time.Now().UnixNano()) / 1e9

	result, err := tokenBucketScript.Run(ctx, l.client, []string{"ratelimit:" + key},
		l.capacity, l.refill, now, 1,
	).Result()
	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	vals, ok := result.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, fmt.Errorf("unexpected rate limit script result")
	}

	allowedInt, _ := vals[0].(int64)
	remainingStr, _ := vals[1].(string)

	var rem float64
	if _, err := fmt.Sscanf(remainingStr, "%f", &rem); err != nil {
		return false, 0, fmt.Errorf("failed to parse remaining tokens: %w", err)
	}

	return allowedInt == 1, rem, nil
}
