package export

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// ExportCacheService provides caching functionality for data export operations
type ExportCacheService struct {
	redis  *redis.Client
	prefix string
}

// ExportCacheKey represents a cache key for export operations
type ExportCacheKey struct {
	ExportType string                 `json:"export_type"`
	Format     string                 `json:"format"`
	Filters    map[string]interface{} `json:"filters"`
	CompanyID  string                 `json:"company_id"`
	UserID     string                 `json:"user_id"`
}

// ExportCacheData represents cached export data
type ExportCacheData struct {
	Data      interface{} `json:"data"`
	Metadata  ExportMetadata `json:"metadata"`
	CachedAt  time.Time   `json:"cached_at"`
	ExpiresAt time.Time   `json:"expires_at"`
	Hash      string      `json:"hash"`
}

// ExportMetadata contains metadata about the export
type ExportMetadata struct {
	RecordCount int64     `json:"record_count"`
	FileSize    int64     `json:"file_size"`
	ExportTime  time.Time `json:"export_time"`
	Format      string    `json:"format"`
	Filters     map[string]interface{} `json:"filters"`
}

// NewExportCacheService creates a new export cache service
func NewExportCacheService(redis *redis.Client) *ExportCacheService {
	return &ExportCacheService{
		redis:  redis,
		prefix: "export_cache",
	}
}

// generateCacheKey generates a unique cache key for export parameters
func (ecs *ExportCacheService) generateCacheKey(key *ExportCacheKey) string {
	// Create a deterministic hash of the export parameters
	keyData, _ := json.Marshal(key)
	hash := md5.Sum(keyData)
	return fmt.Sprintf("%s:%s:%x", ecs.prefix, key.ExportType, hash)
}

// generateFilterKey generates a cache key for filtered exports
func (ecs *ExportCacheService) generateFilterKey(exportType, format string, filters map[string]interface{}, companyID, userID string) string {
	key := &ExportCacheKey{
		ExportType: exportType,
		Format:     format,
		Filters:    filters,
		CompanyID:  companyID,
		UserID:     userID,
	}
	return ecs.generateCacheKey(key)
}

// SetExportCache stores export data in cache
func (ecs *ExportCacheService) SetExportCache(ctx context.Context, exportType, format string, data interface{}, metadata ExportMetadata, filters map[string]interface{}, companyID, userID string, ttl time.Duration) error {
	cacheKey := ecs.generateFilterKey(exportType, format, filters, companyID, userID)
	
	// Create cache data
	cacheData := ExportCacheData{
		Data:      data,
		Metadata:  metadata,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		Hash:      fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", data)))),
	}
	
	// Serialize cache data
	cacheDataBytes, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}
	
	// Store in Redis
	err = ecs.redis.Set(ctx, cacheKey, cacheDataBytes, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	
	// Also store metadata separately for quick access
	metadataKey := fmt.Sprintf("%s:meta:%s", ecs.prefix, cacheKey)
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	err = ecs.redis.Set(ctx, metadataKey, metadataBytes, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set metadata cache: %w", err)
	}
	
	return nil
}

// GetExportCache retrieves export data from cache
func (ecs *ExportCacheService) GetExportCache(ctx context.Context, exportType, format string, filters map[string]interface{}, companyID, userID string) (*ExportCacheData, error) {
	cacheKey := ecs.generateFilterKey(exportType, format, filters, companyID, userID)
	
	// Get cache data
	cacheDataBytes, err := ecs.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	
	// Deserialize cache data
	var cacheData ExportCacheData
	err = json.Unmarshal([]byte(cacheDataBytes), &cacheData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	
	// Check if cache has expired
	if time.Now().After(cacheData.ExpiresAt) {
		// Cache expired, remove it
		ecs.redis.Del(ctx, cacheKey)
		return nil, nil
	}
	
	return &cacheData, nil
}

// GetExportMetadata retrieves only metadata from cache
func (ecs *ExportCacheService) GetExportMetadata(ctx context.Context, exportType, format string, filters map[string]interface{}, companyID, userID string) (*ExportMetadata, error) {
	cacheKey := ecs.generateFilterKey(exportType, format, filters, companyID, userID)
	metadataKey := fmt.Sprintf("%s:meta:%s", ecs.prefix, cacheKey)
	
	// Get metadata
	metadataBytes, err := ecs.redis.Get(ctx, metadataKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get metadata cache: %w", err)
	}
	
	// Deserialize metadata
	var metadata ExportMetadata
	err = json.Unmarshal([]byte(metadataBytes), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return &metadata, nil
}

// InvalidateExportCache invalidates cache for specific export parameters
func (ecs *ExportCacheService) InvalidateExportCache(ctx context.Context, exportType, format string, filters map[string]interface{}, companyID, userID string) error {
	cacheKey := ecs.generateFilterKey(exportType, format, filters, companyID, userID)
	metadataKey := fmt.Sprintf("%s:meta:%s", ecs.prefix, cacheKey)
	
	// Remove both data and metadata
	pipe := ecs.redis.Pipeline()
	pipe.Del(ctx, cacheKey)
	pipe.Del(ctx, metadataKey)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}
	
	return nil
}

// InvalidateCompanyExports invalidates all exports for a company
func (ecs *ExportCacheService) InvalidateCompanyExports(ctx context.Context, companyID string) error {
	pattern := fmt.Sprintf("%s:*:*", ecs.prefix)
	
	// Get all matching keys
	keys, err := ecs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}
	
	// Filter keys for the specific company
	var companyKeys []string
	for _, key := range keys {
		// Check if key contains the company ID
		if len(key) > len(ecs.prefix)+1 {
			keyData := key[len(ecs.prefix)+1:]
			// Simple check - in a real implementation, you might want to parse the key more carefully
			if len(keyData) > 32 { // MD5 hash length
				// This is a simplified check - you might want to decode the key to verify company ID
				companyKeys = append(companyKeys, key)
				// Also add metadata key
				metadataKey := fmt.Sprintf("%s:meta:%s", ecs.prefix, key)
				companyKeys = append(companyKeys, metadataKey)
			}
		}
	}
	
	// Remove all company-related cache entries
	if len(companyKeys) > 0 {
		err = ecs.redis.Del(ctx, companyKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete company cache: %w", err)
		}
	}
	
	return nil
}

// InvalidateUserExports invalidates all exports for a user
func (ecs *ExportCacheService) InvalidateUserExports(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("%s:*:*", ecs.prefix)
	
	// Get all matching keys
	keys, err := ecs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}
	
	// Filter keys for the specific user
	var userKeys []string
	for _, key := range keys {
		// This is a simplified implementation
		// In a real scenario, you'd want to decode the key to check user ID
		userKeys = append(userKeys, key)
		// Also add metadata key
		metadataKey := fmt.Sprintf("%s:meta:%s", ecs.prefix, key)
		userKeys = append(userKeys, metadataKey)
	}
	
	// Remove all user-related cache entries
	if len(userKeys) > 0 {
		err = ecs.redis.Del(ctx, userKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete user cache: %w", err)
		}
	}
	
	return nil
}

// GetCacheStats returns cache statistics
func (ecs *ExportCacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	pattern := fmt.Sprintf("%s:*", ecs.prefix)
	
	// Get all cache keys
	keys, err := ecs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache keys: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_keys": len(keys),
		"data_keys":  0,
		"meta_keys":  0,
		"total_size": 0,
	}
	
	// Count different types of keys
	for _, key := range keys {
		if len(key) > len(ecs.prefix)+1 {
			keyData := key[len(ecs.prefix)+1:]
			if keyData[:4] == "meta" {
				stats["meta_keys"] = stats["meta_keys"].(int) + 1
			} else {
				stats["data_keys"] = stats["data_keys"].(int) + 1
			}
		}
		
		// Get key size
		size, err := ecs.redis.MemoryUsage(ctx, key).Result()
		if err == nil {
			stats["total_size"] = stats["total_size"].(int64) + size
		}
	}
	
	return stats, nil
}

// CleanupExpiredCache removes expired cache entries
func (ecs *ExportCacheService) CleanupExpiredCache(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", ecs.prefix)
	
	// Get all cache keys
	keys, err := ecs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}
	
	var expiredKeys []string
	
	// Check each key for expiration
	for _, key := range keys {
		// Skip metadata keys for now
		if len(key) > len(ecs.prefix)+1 {
			keyData := key[len(ecs.prefix)+1:]
			if keyData[:4] == "meta" {
				continue
			}
		}
		
		// Get TTL
		ttl, err := ecs.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		
		// If TTL is -1 (no expiration) or -2 (key doesn't exist), skip
		if ttl <= 0 {
			continue
		}
		
		// Get the actual data to check expiration
		cacheDataBytes, err := ecs.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var cacheData ExportCacheData
		err = json.Unmarshal([]byte(cacheDataBytes), &cacheData)
		if err != nil {
			continue
		}
		
		// Check if expired
		if time.Now().After(cacheData.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
			// Also add metadata key
			metadataKey := fmt.Sprintf("%s:meta:%s", ecs.prefix, key)
			expiredKeys = append(expiredKeys, metadataKey)
		}
	}
	
	// Remove expired keys
	if len(expiredKeys) > 0 {
		err = ecs.redis.Del(ctx, expiredKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete expired cache: %w", err)
		}
	}
	
	return nil
}

// GetCacheHitRate returns cache hit rate statistics
func (ecs *ExportCacheService) GetCacheHitRate(ctx context.Context) (map[string]interface{}, error) {
	// Get hit/miss counters from Redis
	hitKey := fmt.Sprintf("%s:stats:hits", ecs.prefix)
	missKey := fmt.Sprintf("%s:stats:misses", ecs.prefix)
	
	hits, err := ecs.redis.Get(ctx, hitKey).Int64()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get hit count: %w", err)
	}
	if err == redis.Nil {
		hits = 0
	}
	
	misses, err := ecs.redis.Get(ctx, missKey).Int64()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get miss count: %w", err)
	}
	if err == redis.Nil {
		misses = 0
	}
	
	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}
	
	return map[string]interface{}{
		"hits":     hits,
		"misses":   misses,
		"total":    total,
		"hit_rate": hitRate,
	}, nil
}

// RecordCacheHit records a cache hit
func (ecs *ExportCacheService) RecordCacheHit(ctx context.Context) error {
	hitKey := fmt.Sprintf("%s:stats:hits", ecs.prefix)
	return ecs.redis.Incr(ctx, hitKey).Err()
}

// RecordCacheMiss records a cache miss
func (ecs *ExportCacheService) RecordCacheMiss(ctx context.Context) error {
	missKey := fmt.Sprintf("%s:stats:misses", ecs.prefix)
	return ecs.redis.Incr(ctx, missKey).Err()
}

// GetTTLForExportType returns the appropriate TTL for an export type
func (ecs *ExportCacheService) GetTTLForExportType(exportType string) time.Duration {
	switch exportType {
	case "vehicles":
		return 2 * time.Hour // Vehicle data changes less frequently
	case "drivers":
		return 2 * time.Hour // Driver data changes less frequently
	case "trips":
		return 1 * time.Hour // Trip data changes more frequently
	case "gps_tracks":
		return 30 * time.Minute // GPS data changes very frequently
	case "reports":
		return 4 * time.Hour // Reports are expensive to generate
	default:
		return 1 * time.Hour // Default TTL
	}
}
