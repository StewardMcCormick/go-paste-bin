package appcache

import (
	"sort"
)

type cacheEntry[K comparable, V domainTypes] struct {
	key   K
	value *cacheValue[V]
}

func (c *inMemoryCache[K, V]) getSortEntries() []cacheEntry[K, V] {
	entries := make([]cacheEntry[K, V], 0, len(c.storage))

	for k, v := range c.storage {
		entries = append(entries, cacheEntry[K, V]{k, v})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].value.usageCount < entries[j].value.usageCount
	})

	return entries
}
