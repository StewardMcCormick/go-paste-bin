package appcache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	"github.com/redis/go-redis/v9"
)

type pasteCacheValue struct {
	Content *domain.PasteContent `json:"content"`
}

type pasteCache struct {
	client *redis.Client
	wg     *sync.WaitGroup
	quite  chan struct{}
}

func NewPasteCache(client *redis.Client) *pasteCache {
	return &pasteCache{client: client, wg: &sync.WaitGroup{}}
}

func (c *pasteCache) Set(ctx context.Context, key string, value *domain.PasteContent) {
	log := appctx.GetLogger(ctx)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		jsonValue, err := json.Marshal(&pasteCacheValue{Content: value})
		if err != nil {
			log.Error(fmt.Sprintf("JSON parsing error - %v", err))
		}

		if c.client.Set(ctx, key, jsonValue, 0).Err() != nil {
			log.Error(fmt.Sprintf("Redis saving error - %v", err))
		}
	}()
}

func (c *pasteCache) Get(ctx context.Context, key string) *domain.PasteContent {
	log := appctx.GetLogger(ctx)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		log.Debug("cache miss")
		return nil
	}

	var result pasteCacheValue
	if err := json.Unmarshal(data, &result); err != nil {
		log.Error(fmt.Sprintf("JSON parsing error - %v", err))
		return nil
	}

	log.Debug("cache hit")
	return result.Content
}

func (c *pasteCache) Close(ctx context.Context) {
	log := appctx.GetLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	close(c.quite)

	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		log.Error(fmt.Sprintf("Cache closing error - %v", ctx.Err()))
		return
	}
}
