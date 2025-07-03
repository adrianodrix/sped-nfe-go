// Package certificate provides certificate caching functionality for improved performance.
package certificate

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// CertificateCache provides thread-safe caching of loaded certificates
type CertificateCache struct {
	cache   map[string]*CacheEntry
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// CacheEntry represents a cached certificate with metadata
type CacheEntry struct {
	Certificate Certificate
	LoadedAt    time.Time
	AccessCount int64
	LastAccess  time.Time
	Key         string
}

// CacheConfig holds configuration for certificate cache
type CacheConfig struct {
	// MaxSize is the maximum number of certificates to cache (0 = unlimited)
	MaxSize int `json:"maxSize"`
	
	// TTL is the time-to-live for cached certificates
	TTL time.Duration `json:"ttl"`
	
	// CleanupInterval is how often to run cache cleanup
	CleanupInterval time.Duration `json:"cleanupInterval"`
}

// DefaultCacheConfig returns a default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:         100,
		TTL:             1 * time.Hour,
		CleanupInterval: 10 * time.Minute,
	}
}

// NewCertificateCache creates a new certificate cache with the given configuration
func NewCertificateCache(config *CacheConfig) *CertificateCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cache := &CertificateCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: config.MaxSize,
		ttl:     config.TTL,
	}

	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		go cache.startCleanup(config.CleanupInterval)
	}

	return cache
}

// Get retrieves a certificate from cache by key
func (c *CertificateCache) Get(key string) (Certificate, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if c.ttl > 0 && time.Since(entry.LoadedAt) > c.ttl {
		delete(c.cache, key)
		return nil, false
	}

	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()

	return entry.Certificate, true
}

// Put stores a certificate in cache with the given key
func (c *CertificateCache) Put(key string, cert Certificate) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict entries
	if c.maxSize > 0 && len(c.cache) >= c.maxSize {
		c.evictLRU()
	}

	entry := &CacheEntry{
		Certificate: cert,
		LoadedAt:    time.Now(),
		AccessCount: 1,
		LastAccess:  time.Now(),
		Key:         key,
	}

	c.cache[key] = entry
}

// Remove removes a certificate from cache
func (c *CertificateCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entry, exists := c.cache[key]; exists {
		// Close the certificate if it implements Close()
		if entry.Certificate != nil {
			entry.Certificate.Close()
		}
		delete(c.cache, key)
	}
}

// Clear removes all certificates from cache
func (c *CertificateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Close all certificates
	for _, entry := range c.cache {
		if entry.Certificate != nil {
			entry.Certificate.Close()
		}
	}

	c.cache = make(map[string]*CacheEntry)
}

// Size returns the current number of cached certificates
func (c *CertificateCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// Stats returns cache statistics
func (c *CertificateCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := CacheStats{
		Size:    len(c.cache),
		MaxSize: c.maxSize,
		TTL:     c.ttl,
	}

	for _, entry := range c.cache {
		stats.TotalAccesses += entry.AccessCount
	}

	return stats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size          int           `json:"size"`
	MaxSize       int           `json:"maxSize"`
	TTL           time.Duration `json:"ttl"`
	TotalAccesses int64         `json:"totalAccesses"`
}

// evictLRU evicts the least recently used entry from cache
func (c *CertificateCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		if entry := c.cache[oldestKey]; entry != nil && entry.Certificate != nil {
			entry.Certificate.Close()
		}
		delete(c.cache, oldestKey)
	}
}

// startCleanup starts the periodic cache cleanup goroutine
func (c *CertificateCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries from cache
func (c *CertificateCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.ttl <= 0 {
		return
	}

	now := time.Now()
	var keysToDelete []string

	for key, entry := range c.cache {
		if now.Sub(entry.LoadedAt) > c.ttl {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		if entry := c.cache[key]; entry != nil && entry.Certificate != nil {
			entry.Certificate.Close()
		}
		delete(c.cache, key)
	}
}

// GenerateCacheKey generates a cache key for a certificate file
func GenerateCacheKey(filePath string, password string) string {
	// Create a hash of the file path and password for cache key
	// Don't include the actual password for security
	h := sha256.New()
	h.Write([]byte(filePath))
	h.Write([]byte("::"))
	h.Write([]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password)))))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GeneratePKCS11CacheKey generates a cache key for PKCS#11 certificates
func GeneratePKCS11CacheKey(config *PKCS11Config) string {
	h := sha256.New()
	h.Write([]byte(config.LibraryPath))
	h.Write([]byte("::"))
	h.Write([]byte(config.TokenLabel))
	h.Write([]byte("::"))
	h.Write([]byte(config.CertificateLabel))
	if len(config.CertificateID) > 0 {
		h.Write([]byte("::"))
		h.Write(config.CertificateID)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Global certificate cache instance
var globalCache *CertificateCache
var globalCacheOnce sync.Once

// GetGlobalCache returns the global certificate cache instance
func GetGlobalCache() *CertificateCache {
	globalCacheOnce.Do(func() {
		globalCache = NewCertificateCache(DefaultCacheConfig())
	})
	return globalCache
}

// SetGlobalCache sets a custom global certificate cache
func SetGlobalCache(cache *CertificateCache) {
	globalCache = cache
}

// LoadA1FromFileWithCache loads an A1 certificate with caching
func LoadA1FromFileWithCache(filePath, password string) (Certificate, error) {
	cache := GetGlobalCache()
	key := GenerateCacheKey(filePath, password)

	// Try to get from cache first
	if cert, found := cache.Get(key); found {
		return cert, nil
	}

	// Load from file
	loader := NewA1CertificateLoader(DefaultConfig())
	cert, err := loader.LoadFromFile(filePath, password)
	if err != nil {
		return nil, err
	}

	// Store in cache
	cache.Put(key, cert)

	return cert, nil
}

// LoadA3FromTokenWithCache loads an A3 certificate with caching
func LoadA3FromTokenWithCache(config *PKCS11Config) (Certificate, error) {
	cache := GetGlobalCache()
	key := GeneratePKCS11CacheKey(config)

	// Try to get from cache first
	if cert, found := cache.Get(key); found {
		return cert, nil
	}

	// Load from token
	loader := NewA3CertificateLoader(DefaultConfig())
	cert, err := loader.LoadFromToken(config)
	if err != nil {
		return nil, err
	}

	// Store in cache
	cache.Put(key, cert)

	return cert, nil
}