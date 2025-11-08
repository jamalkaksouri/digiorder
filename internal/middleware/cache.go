package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
	Timestamp  time.Time
}

// Cache manages cached responses
type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

// NewCache creates a new cache with specified TTL
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a cached entry
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	
	// Check if entry is expired
	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}
	
	return entry, true
}

// Set stores a cache entry
func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.entries[key] = entry
}

// Delete removes a cache entry
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.entries, key)
}

// Clear removes all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.entries = make(map[string]*CacheEntry)
}

// cleanup removes expired entries periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if time.Since(entry.Timestamp) > c.ttl {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// generateCacheKey creates a unique key for the request
func generateCacheKey(c echo.Context) string {
	req := c.Request()
	
	// Include method, path, and query string
	base := fmt.Sprintf("%s:%s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
	
	// Add user context if available
	if userID, ok := c.Get("user_id").(string); ok {
		base += fmt.Sprintf(":user:%s", userID)
	}
	
	// Create hash
	hash := md5.Sum([]byte(base))
	return hex.EncodeToString(hash[:])
}

// CacheMiddleware creates a caching middleware
func CacheMiddleware(ttl time.Duration, cachableStatuses ...int) echo.MiddlewareFunc {
	cache := NewCache(ttl)
	
	// Default cachable statuses
	if len(cachableStatuses) == 0 {
		cachableStatuses = []int{http.StatusOK}
	}
	
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only cache GET requests
			if c.Request().Method != http.MethodGet {
				return next(c)
			}
			
			// Generate cache key
			key := generateCacheKey(c)
			
			// Check cache
			if entry, found := cache.Get(key); found {
				// Serve from cache
				c.Response().Header().Set("X-Cache", "HIT")
				c.Response().Header().Set("X-Cache-Age", fmt.Sprintf("%d", int(time.Since(entry.Timestamp).Seconds())))
				
				// Copy headers
				for k, v := range entry.Headers {
					for _, vv := range v {
						c.Response().Header().Set(k, vv)
					}
				}
				
				return c.Blob(entry.StatusCode, echo.MIMEApplicationJSON, entry.Body)
			}
			
			// Create custom response writer to capture response
			rec := &responseRecorder{
				ResponseWriter: c.Response().Writer,
				body:          []byte{},
			}
			c.Response().Writer = rec
			
			// Call next handler
			err := next(c)
			
			// Check if response should be cached
			shouldCache := false
			for _, status := range cachableStatuses {
				if rec.status == status {
					shouldCache = true
					break
				}
			}
			
			if err == nil && shouldCache {
				// Store in cache
				entry := &CacheEntry{
					Body:       rec.body,
					StatusCode: rec.status,
					Headers:    c.Response().Header().Clone(),
					Timestamp:  time.Now(),
				}
				cache.Set(key, entry)
				c.Response().Header().Set("X-Cache", "MISS")
			}
			
			return err
		}
	}
}

// responseRecorder captures the response for caching
type responseRecorder struct {
	http.ResponseWriter
	body   []byte
	status int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// CacheInvalidationMiddleware invalidates cache on write operations
func CacheInvalidationMiddleware(cache *Cache) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Clear cache on write operations
			method := c.Request().Method
			if method == http.MethodPost || method == http.MethodPut || 
			   method == http.MethodPatch || method == http.MethodDelete {
				cache.Clear()
			}
			
			return next(c)
		}
	}
}