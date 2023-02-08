/*
Package cache implements a users friendly cache fetching mechanism with support
for multiple storage backends, like Redis and in-memory. It prevents cache
stampedes by using probabilistic early expiration of cached values.

The package revolves around the Fetch method, which fetches data from a cache.
If the value doesn't exist in the cache, it is populated using the provided
recompute function, and the new value is returned.

The package supports multiple storage backends. It includes a Redis and an
in-memory backend, and will work with any backend that implements the
StorageBackend interface.

For more information on probabilistic early expiration, see Wikipedia
(https://en.wikipedia.org/wiki/Cache_stampede#Probabilistic_early_expiration) or
this paper (http://cseweb.ucsd.edu/~avattani/papers/cache_stampede.pdf).
*/
package cache

import (
	"bytes"
	"encoding/gob"
	"math"
	"math/rand"
	"reflect"
	"time"

	"github.com/deliveroo/cache-go/lockedrand"
)

const defaultBeta = 1

// Client is cache client.
type Client struct {
	backend StorageBackend
	ttl     time.Duration
	beta    float64
	rng     *lockedrand.Rand
	Notify  func(key string, isCacheHit bool)
}

type item struct {
	Value  interface{}
	Expiry time.Time
	Delta  time.Duration
}

// StorageBackend is an interface that allows for the cache to be used with any
// type of storage backend, as long as it implements this interface.
type StorageBackend interface {
	Get(key string) (val []byte, err error)
	Set(key string, val []byte, ttl time.Duration) (err error)
}

// New returns a new cache client.
func New(backend StorageBackend, ttl time.Duration) *Client {
	noopFunc := func(key string, isCacheHit bool) {}
	return &Client{
		backend: backend,
		ttl:     ttl,
		beta:    defaultBeta,
		rng:     lockedrand.New(rand.NewSource(time.Now().UnixNano())),
		Notify:  noopFunc,
	}
}

// Fetch fetches data from a cache. If the value doesn't exist in the cache, it
// is populated using the provided recompute function, and the new value is
// returned.
func (c *Client) Fetch(key string, result interface{}, recompute func() (interface{}, error)) error {
	gob.Register(result)
	cachedItem, err := c.getItem(key)
	if err != nil {
		return err
	}
	if cachedItem != nil && !c.shouldRecomputeEarly(cachedItem) {
		c.Notify(key, true)
		copyResult(cachedItem.Value, result)
		return nil
	}

	c.Notify(key, false)
	start := time.Now()
	newVal, err := recompute()
	if err != nil {
		return err
	}
	if reflect.ValueOf(newVal).IsNil() {
		// nil values cannot be encoded using gob, and it seems unlikely that
		// anyone would want to cache a nil value
		return nil
	}
	copyResult(newVal, result)
	newItem := &item{Value: newVal, Delta: time.Since(start), Expiry: time.Now().Add(c.ttl)}
	if err := c.setItem(key, newItem); err != nil {
		return err
	}
	return nil
}

// Set sets provided value for key in backend, allowing a configurable delta
// for stampede protection (time.Duration for early recalculation).
func (c *Client) Set(key string, val interface{}, delta time.Duration) error {
	gob.Register(val)
	newItem := &item{Value: val, Delta: delta, Expiry: time.Now().Add(c.ttl)}
	if err := c.setItem(key, newItem); err != nil {
		return err
	}
	return nil
}

func (c *Client) getItem(key string) (*item, error) {
	cachedVal, err := c.backend.Get(key)
	if err != nil {
		return nil, err
	}
	if cachedVal == nil {
		return nil, nil
	}
	cachedItem := &item{}
	if err := gob.NewDecoder(bytes.NewReader(cachedVal)).Decode(cachedItem); err != nil {
		return nil, err
	}
	return cachedItem, nil
}

func (c *Client) setItem(key string, newItem *item) error {
	var newVal bytes.Buffer
	if err := gob.NewEncoder(&newVal).Encode(newItem); err != nil {
		return err
	}
	return c.backend.Set(key, newVal.Bytes(), c.ttl)
}

func copyResult(src, dest interface{}) {
	srcValue := reflect.ValueOf(src)
	if srcValue.IsNil() {
		return
	}
	reflect.ValueOf(dest).Elem().Set(srcValue.Elem())
}

// shouldRecomputeEarly determines whether the cached value should be recomputed
// early, to prevent cache stampedes, using probabilistic early expiration. Each
// process that uses the cache might decide to recompute the cached value early,
// based on a probabilistic distribution, which becomes more likely the closer
// we get to the expiry time. For more information, check out Wikipedia
// (https://en.wikipedia.org/wiki/Cache_stampede#Probabilistic_early_expiration)
// or this paper (http://cseweb.ucsd.edu/~avattani/papers/cache_stampede.pdf).
func (c *Client) shouldRecomputeEarly(i *item) bool {
	return time.Now().Add(-time.Duration(float64(i.Delta) * c.beta * math.Log(c.rng.Float64()))).After(i.Expiry)
}

// WithTTL returns a copy of the Client with the provided ttl.
func (c Client) WithTTL(ttl time.Duration) *Client {
	c.ttl = ttl
	return &c
}

// WithBeta returns a copy of the Client with the provided beta.
func (c Client) WithBeta(beta float64) *Client {
	c.beta = beta
	return &c
}
