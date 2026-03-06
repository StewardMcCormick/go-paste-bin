package paste

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
)

const (
	ExpirationTime = time.Hour
)

type pasteContentCacheValue struct {
	expireAt   time.Time
	lastAccess time.Time
	value      *domain.PasteContent
}

type inMemoryPasteContentCache struct {
	storage  map[string]pasteContentCacheValue
	mu       *sync.RWMutex
	wg       *sync.WaitGroup
	capacity int
	quite    chan struct{}
}

func NewPasteInMemoryCache(ctx context.Context, capacity int) *inMemoryPasteContentCache {
	cache := &inMemoryPasteContentCache{
		storage:  make(map[string]pasteContentCacheValue, capacity),
		mu:       &sync.RWMutex{},
		wg:       &sync.WaitGroup{},
		capacity: capacity,
		quite:    make(chan struct{}),
	}

	cache.startAutoCleanUp(ctx)

	return cache
}

func (c *inMemoryPasteContentCache) Set(ctx context.Context, hash string, content *domain.PasteContent) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		c.mu.Lock()
		defer c.mu.Unlock()

		if len(c.storage) == c.capacity {
			c.eviction()
		}

		c.storage[hash] = pasteContentCacheValue{
			value:      content,
			expireAt:   time.Now().Add(ExpirationTime),
			lastAccess: time.Now(),
		}
	}()
}

// eviction use LRU algorithm
func (c *inMemoryPasteContentCache) eviction() {
	var (
		oldestKey  string
		oldestTime time.Time
	)
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.storage {
		if v.lastAccess.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.lastAccess
		}
	}

	if oldestKey != "" {
		delete(c.storage, oldestKey)
	}
}

func (c *inMemoryPasteContentCache) Get(ctx context.Context, hash string) *domain.PasteContent {
	log := appctx.GetLogger(ctx)
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.storage[hash]; ok {
		log.Debug("cache hit")
		return v.value
	}

	log.Debug("cache miss")
	return nil
}

func (c *inMemoryPasteContentCache) startAutoCleanUp(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-ctx.Done():
				return
			case <-c.quite:
				return
			}
		}
	}()
}

func (c *inMemoryPasteContentCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.storage {
		if !v.expireAt.Before(time.Now()) {
			delete(c.storage, k)
		}
	}
}

func (c *inMemoryPasteContentCache) Close(ctx context.Context) {
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
