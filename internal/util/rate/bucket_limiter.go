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
		BucketSize = capacity - 1,
		LastAccess = now
	}

	redis.call("SET", key, cjson.encode(value))
	redis.call("EXPIRE", key, math.ceil(capacity / rate) * 2)
	return 1
end

local data = cjson.decode(current)

local diff = (now - data.LastAccess) / 1000

if diff >= 0 then
	data.BucketSize = math.min(capacity, data.BucketSize + rate * diff)
end


if data.BucketSize < 1 then
	return 0
end

data.BucketSize = data.BucketSize - 1
data.LastAccess = now
redis.call("SET", key, cjson.encode(data))
redis.call("EXPIRE", key, math.ceil(capacity / rate) * 2)

return 1
`

type bucketLimiterValue struct {
	BucketSize int       `json:"bucket_size"`
	LastAccess time.Time `json:"last_access"`
}

type bucketLimiter struct {
	client         *redis.Client
	prefix         string
	rate           int // tokens per second
	bucketCapacity int
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
	}
}

func (l *bucketLimiter) AllowRequest(ctx context.Context, key string) (bool, error) {
	log := appctx.GetLogger(ctx)

	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)

	allowScript := redis.NewScript(requestAllowScript)

	result, err := allowScript.Run(ctx, l.client, []string{redisKey}, l.bucketCapacity, time.Now().UnixMilli(), l.rate).Result()

	if err != nil {
		log.Error(fmt.Sprintf("redis error - %v", err))
		return false, err
	}

	allowed := result.(int64) == 1

	return allowed, nil
}
