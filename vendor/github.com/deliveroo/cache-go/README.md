# cache-go 

[![CircleCI](https://circleci.com/gh/deliveroo/cache-go.svg?style=svg&circle-token=2cc427c76320033adbf7bdff7ea3c868dc73275c)](https://circleci.com/gh/deliveroo/cache-go)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](http://godoc.deliveroo.net/github.com/deliveroo/cache-go)

The cache-go package implements a user friendly cache fetching mechanism using
Redis, an in-memory LRU cache, or your own storage backend. 

The package revolves around the `Fetch` method, which fetches data from a cache.
If the value does not exist in the cache, it is populated using the provided
`recompute` function, and the new value is returned.

## Usage

The following snippet instantiates an in-memory cache with a 5-minute TTL and
performs a fetch:

```go
size := 10
client := cache.New(memorycache.LRUBackend(size), 5 * time.Minute)

result := &OperationResult{}
// If the key "operation" is in the cache, the cached value is returned.
// Otherwise, the value is computed with the given function, and then cached.
err := client.Fetch("operation", result, func() (interface{}, error) {
	time.Sleep(10 * time.Second)
	return &OperationResult{ ID: 1 }, nil
})
fmt.Println(result.ID) // This will print 1
```

Alternatively you can instantiate a Redis-backed cache using
[`go-redis/redis`](https://github.com/go-redis/redis):

```go
redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 1})
client := cache.New(rediscache.Backend(redisClient), 5 * time.Minute)
```

You can register a callback with the cache client to get notified on cache hits
and misses. A common use case is to gather metrics about the effectiveness of
the cache.

```go
cache.Notify = func(key string, isCacheHit bool) {
  // report the cache metrics
}
```

You can also use a different storage backend by implementing the following
interface:

```go
type StorageBackend interface {
	Get(key string) (val []byte, err error)
	Set(key string, val []byte, ttl time.Duration) (err error)
}
```

## Preventing cache stampedes 

To prevent [cache stampedes](https://en.wikipedia.org/wiki/Cache_stampede), the
cache implemented in this package uses [probabilistic early
expiration](http://cseweb.ucsd.edu/~avattani/papers/cache_stampede.pdf).

We ran a small experiment to demonstrate the effectiveness of cache stampede
prevention. The experiment runs 100 goroutines that access the cache
concurrently. The cache has a 10 second TTL. The operation that produces the
value that is being cached takes 1 second, and the experiment runs for 2000
iterations.

The results:

<p align="center"> 
    <img src="https://user-images.githubusercontent.com/697118/56369557-7e91ae80-61f1-11e9-91ea-9af5bfda47d5.png" width="600">
</p>

As you can see, without any cache stampede protection, every 10 seconds all 100
goroutines have a cache miss.

<p align="center">
    <img src="https://user-images.githubusercontent.com/697118/56369678-b6005b00-61f1-11e9-91f2-ee3e035b8d0a.png" width="600">
</p>

With cache stampede protection, cache hits are much less frequent, although
there are many cache hits when the cache is first being used.

<p align="center">
    <img src="https://user-images.githubusercontent.com/697118/56369702-c31d4a00-61f1-11e9-979c-f612bf8c19fb.png" width="600">
</p>

By warming the cache we get rid of the initial cache stampede.

<p align="center">
    <img src="https://user-images.githubusercontent.com/697118/56369726-cf090c00-61f1-11e9-98ac-6e026071fc83.png" width="600">
</p>

We can tweak the beta value (see the reference paper for more info) to get even
fewer cache misses. In practice this won't be necessary as the ttl will most
likely be much higher than 10 seconds, which will give better performance.


