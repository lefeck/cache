package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := New(5*time.Second, 1*time.Second)
	cache.Set("a", 1, DefaultExpiration)
	cache.SetDefault("b", 2)

	// Testing Get method
	v, ok := cache.Get("a")
	if !ok || v != 1 {
		t.Errorf("Expected value 1, got %v", v)
	}

	v, ok = cache.Get("b")
	if !ok || v != 2 {
		t.Errorf("Expected value 2, got %v", v)
	}

	// Testing Add method
	err := cache.Add("a", 3, DefaultExpiration)
	if err == nil {
		t.Errorf("Expected error as 'a' already exists, got nil")
	}

	// Testing Replace method
	err = cache.Replace("a", 3, DefaultExpiration)
	if err != nil {
		t.Error("Failed to replace existing key 'a'")
	}
	err = cache.Replace("c", 3, DefaultExpiration)
	if err == nil {
		t.Error("Replacing non-existent key should raise error")
	}

	// Testing ItemCount method
	count := cache.ItemCount()
	if count != 2 {
		t.Errorf("Expected item count 2, got %d", count)
	}

	// Testing Increment method
	cache.SetDefault("number", 42)
	cache.Increment("number", 1)
	v, ok = cache.Get("number")
	if !ok || v != 43 {
		t.Errorf("Expected value 43, got %v", v)
	}

	// Testing Delete method
	deleted := make(chan struct{})
	cache.OnEvicted(func(key string, value interface{}) {
		deleted <- struct{}{}
	})
	cache.Delete("a")

	select {
	case <-deleted:
		// Delete and onEvicted worked correctly
	case <-time.After(time.Second):
		t.Error("OnEvicted didn't trigger after Delete")
	}

	// Testing expired items clean up
	cache.Set("c", 3, 1*time.Second)
	time.Sleep(2 * time.Second)
	<-cache.janitor.stop // Waiting for janitor to remove the item

	v, ok = cache.Get("c")
	if ok {
		t.Error("Item 'c' should have been deleted due to expiration")
	}

	// Testing Flush method
	cache.Flush()
	_, ok = cache.Get("b")
	if ok {
		t.Error("Item 'b' should have been deleted after Flush")
	}
}
