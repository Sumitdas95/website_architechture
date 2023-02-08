package determinator

import (
	"time"

	"github.com/deliveroo/cache-go"
	"github.com/deliveroo/cache-go/memorycache"
)

const defaultCacheItemCount = 10000

// CachedRetriever wraps a retriever to cache the results
type CachedRetriever struct {
	wrapped       Retriever
	cacheDuration time.Duration
	cache         *cache.Client
}

// CachedOpt is an option passed to NewCachedRetriever.
type CachedOpt func(*CachedRetriever)

// WithCache allows setting a custom cache to the CachedRetriever. By passing in
// a custom cache, a consumer can share the same cache between different
// CachedRetriever instances (which can be useful if the wrapped HTTPRetriever
// holds session information).
func WithCache(c *cache.Client) CachedOpt {
	return func(r *CachedRetriever) {
		r.cache = c
	}
}

// NewCachedRetriever creates a new NewCachedRetriever object.
func NewCachedRetriever(r Retriever, cacheDuration time.Duration, opts ...CachedOpt) *CachedRetriever {
	retriever := &CachedRetriever{
		wrapped:       r,
		cacheDuration: cacheDuration,
		cache:         cache.New(memorycache.LRUBackend(defaultCacheItemCount), cacheDuration),
	}
	for _, o := range opts {
		o(retriever)
	}
	return retriever
}

// Retrieve will attempt to retrieve from cache or pull from the wrapped
// retriever if the cache misses.
func (c *CachedRetriever) Retrieve(featureID string) (Feature, error) {
	// Dummy feature that will be returned in case of errors or misses
	result := &FeatureData{
		ID:         featureID,
		Name:       featureID,
		Identifier: featureID,
	}
	err := c.cache.Fetch(featureID, result, func() (interface{}, error) {
		f, e := c.wrapped.Retrieve(featureID)
		if f == nil {
			return result, e
		}
		return f, e
	})
	return result, err
}
