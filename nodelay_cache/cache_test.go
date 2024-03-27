package nodelay_cache

import (
	"fmt"
	"testing"
)

func TestCache(t *testing.T) {
	m := make(map[string]Item)
	cache := NewCaches(m)

	// 测试 Set
	err := cache.Set("key1", 123)
	if err != nil {
		t.Errorf("Set failed: %s", err)
	}

	// 测试 Get
	value, ok := cache.Get("key1")
	if !ok || value.(int) != 123 {
		t.Errorf("Get failed: expected 123, got %v", value)
	}

	// 测试 Replace
	err = cache.Replace("key1", 456)
	if err != nil {
		t.Errorf("Replace failed: %s", err)
	}

	value, ok = cache.Get("key1")
	if !ok || value.(int) != 456 {
		t.Errorf("Replace failed: expected 456, got %v", value)
	}

	// testing count
	count := cache.Count()
	fmt.Printf("the item of count is %d\n", count)
	if count == 0 {
		t.Errorf("the item of count is not zero ")
	}

	// 测试 Copy
	copied := cache.Copy()
	if len(copied) != 1 || copied["key1"].Obj.(int) != 456 {
		t.Errorf("Copy failed: expected map[key1]: 456, got %v", copied)
	}

	// 测试 Delete
	deleted := cache.Delete("key1")
	if !deleted {
		t.Errorf("Delete failed")
	}

	_, ok = cache.Get("key1")
	if ok {
		t.Errorf("Delete failed: key1 still exists in cache")
	}

	// 测试 Flush
	err = cache.Set("key1", 123)
	cache.Flush()

	_, ok = cache.Get("key1")
	if ok {
		t.Errorf("Flush failed: key1 still exists in cache")
	}
}
