package rate

import (
	"context"
	"fmt"
	"time"

	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/redis/go-redis/v9"
)

const requestAllowScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local now = tonumber(ARGV[2]) -- in milliseconds
local rate = tonumber(ARGV[3])

local current = redis.call("GET", key)

if not current then
	local value = {
		bucket_size = capacity - 1,
		last_access = now
	}

	redis.call("SET", key, cjson.encode(value))
	redis.call("EXPIRE", key, 3600)
	return 1
end

local data = cjson.decode(current)

local diff = math.floor((now - data.last_access) / 1000)

if diff >= 0 then
	data.bucket_size = math.min(capacity, data.bucket_size + rate * diff)
end


if data.bucket_size < 1 then
	return 0
end

data.bucket_size = data.bucket_size - 1
data.last_access = now
redis.call("SET", key, cjson.encode(data))
redis.call("EXPIRE", key, 3600)

return 1
`

type bucketLimiter struct {
	client         *redis.Client
	prefix         string
	rate           int // tokens per second
	bucketCapacity int
	allowScript    *redis.Script
}

func NewBucketLimiter(client *redis.Client, prefix string, bucketCapacity int, rate int) *bucketLimiter {
	if bucketCapacity <= 0 {
		panic("Bucket capacity must be positive")
	}
	if rate <= 0 {
		panic("Rate must be positive")
	}

	return &bucketLimiter{
		client:         client,
		prefix:         prefix,
		bucketCapacity: bucketCapacity,
		rate:           rate,
		allowScript:    redis.NewScript(requestAllowScript),
	}
}

func (l *bucketLimiter) AllowRequest(ctx context.Context, key string) (bool, error) {
	log := appctx.GetLogger(ctx)

	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)

	result, err := l.allowScript.Run(ctx, l.client, []string{redisKey}, l.bucketCapacity, time.Now().UnixMilli(), l.rate).Result()

	if err != nil {
		log.Error(fmt.Sprintf("redis error - %v", err))
		return false, err
	}

	allowed := result.(int64) == 1

	if !allowed {
		log.Debug(fmt.Sprintf("request rejected - %s", redisKey))
	} else {
		log.Debug(fmt.Sprintf("request allowed - %s", redisKey))
	}

	return allowed, nil
}
