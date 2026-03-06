package appcache

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

type domainTypes interface {
	*domain.PasteContent | *domain.APIKey
}

type cacheValue[T domainTypes] struct {
	expireAt   time.Time
	lastAccess time.Time
	value      T
}

type inMemoryCache[K comparable, T domainTypes] struct {
	storage  map[K]cacheValue[T]
	mu       *sync.RWMutex
	wg       *sync.WaitGroup
	capacity int
	quite    chan struct{}
}

func NewInMemoryCache[K comparable, T domainTypes](
	ctx context.Context,
	capacity int,
) *inMemoryCache[K, T] {
	cache := &inMemoryCache[K, T]{
		storage:  make(map[K]cacheValue[T], capacity),
		mu:       &sync.RWMutex{},
		wg:       &sync.WaitGroup{},
		capacity: capacity,
		quite:    make(chan struct{}),
	}

	cache.startAutoCleanUp(ctx)

	return cache
}

func (c *inMemoryCache[K, T]) Set(ctx context.Context, key K, value T) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		c.mu.Lock()
		defer c.mu.Unlock()

		if len(c.storage) == c.capacity {
			c.eviction()
		}

		c.storage[key] = cacheValue[T]{
			value:      value,
			expireAt:   time.Now().Add(ExpirationTime),
			lastAccess: time.Now(),
		}
	}()
}

// eviction use LRU algorithm
func (c *inMemoryCache[K, T]) eviction() {
	var (
		oldestKey  K
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

	delete(c.storage, oldestKey)
}

func (c *inMemoryCache[K, T]) Get(ctx context.Context, key K) T {
	log := appctx.GetLogger(ctx)
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.storage[key]; ok {
		log.Debug("cache hit")
		return v.value
	}

	log.Debug("cache miss")
	return nil
}

func (c *inMemoryCache[K, T]) startAutoCleanUp(ctx context.Context) {
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

func (c *inMemoryCache[K, T]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.storage {
		if !v.expireAt.Before(time.Now()) {
			delete(c.storage, k)
		}
	}
}

func (c *inMemoryCache[K, T]) Close(ctx context.Context) {
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
