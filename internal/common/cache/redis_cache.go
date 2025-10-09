package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache provides Redis-based caching functionality
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Set stores a value in cache with expiration
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := rc.getFullKey(key)
	
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	if err := rc.client.Set(ctx, fullKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}
	
	return nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := rc.getFullKey(key)
	
	data, err := rc.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get cache value: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}
	
	return nil
}

// Delete removes a value from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := rc.getFullKey(key)
	
	if err := rc.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete cache value: %w", err)
	}
	
	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := rc.getFullKey(key)
	
	count, err := rc.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	
	return count > 0, nil
}

// SetHash stores a hash value in cache
func (rc *RedisCache) SetHash(ctx context.Context, key string, values map[string]interface{}, expiration time.Duration) error {
	fullKey := rc.getFullKey(key)
	
	hashValues := make(map[string]string)
	for k, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal hash value: %w", err)
		}
		hashValues[k] = string(data)
	}
	
	if err := rc.client.HMSet(ctx, fullKey, hashValues).Err(); err != nil {
		return fmt.Errorf("failed to set hash cache: %w", err)
	}
	
	if expiration > 0 {
		if err := rc.client.Expire(ctx, fullKey, expiration).Err(); err != nil {
			return fmt.Errorf("failed to set hash expiration: %w", err)
		}
	}
	
	return nil
}

// GetHash retrieves a hash value from cache
func (rc *RedisCache) GetHash(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	fullKey := rc.getFullKey(key)
	
	if len(fields) == 0 {
		return rc.client.HGetAll(ctx, fullKey).Result()
	}
	
	// Convert []interface{} to map[string]string
	values, err := rc.client.HMGet(ctx, fullKey, fields...).Result()
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	for i, field := range fields {
		if i < len(values) && values[i] != nil {
			if str, ok := values[i].(string); ok {
				result[field] = str
			}
		}
	}
	
	return result, nil
}

// GetHashField retrieves a specific field from a hash
func (rc *RedisCache) GetHashField(ctx context.Context, key, field string, dest interface{}) error {
	fullKey := rc.getFullKey(key)
	
	data, err := rc.client.HGet(ctx, fullKey, field).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get hash field: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal hash field: %w", err)
	}
	
	return nil
}

// SetHashField sets a specific field in a hash
func (rc *RedisCache) SetHashField(ctx context.Context, key, field string, value interface{}, expiration time.Duration) error {
	fullKey := rc.getFullKey(key)
	
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal hash field value: %w", err)
	}
	
	if err := rc.client.HSet(ctx, fullKey, field, data).Err(); err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}
	
	if expiration > 0 {
		if err := rc.client.Expire(ctx, fullKey, expiration).Err(); err != nil {
			return fmt.Errorf("failed to set hash field expiration: %w", err)
		}
	}
	
	return nil
}

// Increment increments a numeric value in cache
func (rc *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := rc.getFullKey(key)
	
	val, err := rc.client.IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache value: %w", err)
	}
	
	return val, nil
}

// SetExpiration sets expiration for a key
func (rc *RedisCache) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := rc.getFullKey(key)
	
	if err := rc.client.Expire(ctx, fullKey, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache expiration: %w", err)
	}
	
	return nil
}

// GetTTL gets the time to live for a key
func (rc *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := rc.getFullKey(key)
	
	ttl, err := rc.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get cache TTL: %w", err)
	}
	
	return ttl, nil
}

// Clear clears all keys with the prefix
func (rc *RedisCache) Clear(ctx context.Context) error {
	pattern := rc.prefix + "*"
	
	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}
	
	if len(keys) > 0 {
		if err := rc.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
	}
	
	return nil
}

// GetStats returns cache statistics
func (rc *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := rc.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}
	
	// Parse Redis info (simplified)
	stats := map[string]interface{}{
		"info": info,
	}
	
	return stats, nil
}

// getFullKey returns the full key with prefix
func (rc *RedisCache) getFullKey(key string) string {
	return fmt.Sprintf("%s:%s", rc.prefix, key)
}

// Cache key generators for common patterns
func (rc *RedisCache) VehicleKey(vehicleID string) string {
	return fmt.Sprintf("vehicle:%s", vehicleID)
}

func (rc *RedisCache) VehicleLocationKey(vehicleID string) string {
	return fmt.Sprintf("vehicle:location:%s", vehicleID)
}

func (rc *RedisCache) VehicleStatsKey(vehicleID string) string {
	return fmt.Sprintf("vehicle:stats:%s", vehicleID)
}

func (rc *RedisCache) DriverKey(driverID string) string {
	return fmt.Sprintf("driver:%s", driverID)
}

func (rc *RedisCache) CompanyStatsKey(companyID string) string {
	return fmt.Sprintf("company:stats:%s", companyID)
}

func (rc *RedisCache) GPSTrackKey(vehicleID string, timestamp time.Time) string {
	return fmt.Sprintf("gps:track:%s:%d", vehicleID, timestamp.Unix())
}

func (rc *RedisCache) TripKey(tripID string) string {
	return fmt.Sprintf("trip:%s", tripID)
}

func (rc *RedisCache) GeofenceKey(geofenceID string) string {
	return fmt.Sprintf("geofence:%s", geofenceID)
}

// Cache expiration constants
const (
	DefaultExpiration    = 5 * time.Minute
	ShortExpiration      = 1 * time.Minute
	MediumExpiration     = 15 * time.Minute
	LongExpiration       = 1 * time.Hour
	VeryLongExpiration   = 24 * time.Hour
	LocationExpiration   = 30 * time.Second
	StatsExpiration      = 10 * time.Minute
	VehicleExpiration    = 30 * time.Minute
	DriverExpiration     = 30 * time.Minute
	CompanyExpiration    = 1 * time.Hour
)
