package nodelay_cache_test

import (
	nodelay_cache "cache/delay_cache"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := nodelay_cache.NewCaches(make(map[string]nodelay_cache.Item), 5*time.Second)

	// Test Set and Get
	err := cache.Set("key1", "value1", 2*time.Second)
	if err != nil {
		t.Error("Set error:", err)
	}
	value, ok := cache.Get("key1")
	if !ok || value != "value1" {
		t.Errorf("Get error: expected value1, got %v", value)
	}

	// Test key expiration
	time.Sleep(3 * time.Second)
	cache.DeleteExpired()
	value, ok = cache.Get("key1")
	if ok || value != nil {
		t.Errorf("Get error: expected nil after expiration, got %v", value)
	}

	// Test Delete
	cache.Set("key2", "value2", nodelay_cache.DefaultExpiration)
	err = cache.Delete("key2")
	if err != nil {
		t.Error("Delete error:", err)
	}

	// Test Flush
	cache.Set("key3", "value3", nodelay_cache.DefaultExpiration)
	cache.Flush()
	value, ok = cache.Get("key3")
	if ok || value != nil {
		t.Errorf("Get error: expected nil after flush, got %v", value)
	}

	// Test Replace
	cache.Set("key4", "value4", nodelay_cache.DefaultExpiration)
	err = cache.Replace("key4", "newValue4", nodelay_cache.DefaultExpiration)
	if err != nil {
		t.Error("Replace error:", err)
	}
	value, ok = cache.Get("key4")
	if !ok || value != "newValue4" {
		t.Errorf("Replace error: expected \"newValue4\", got %v", value)
	}

	// Test DeleteExpired
	cache.Set("key5", "value5", 2*time.Second)
	time.Sleep(3 * time.Second)
	cache.DeleteExpired()
	value, ok = cache.Get("key5")
	if ok || value != nil {
		t.Errorf("Get error: expected nil after deletion of expired item, got %v", value)
	}
}
