package cache

import "errors"

// Cache errors
var (
	ErrCacheMiss = errors.New("cache miss")
	ErrCacheKeyNotFound = errors.New("cache key not found")
	ErrCacheSerialization = errors.New("cache serialization error")
	ErrCacheDeserialization = errors.New("cache deserialization error")
	ErrCacheConnection = errors.New("cache connection error")
	ErrCacheTimeout = errors.New("cache timeout")
)
