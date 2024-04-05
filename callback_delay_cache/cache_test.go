package callback_delay_cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCaches(map[string]Item{}, DefaultExpiration)

	err := cache.Set("key1", "value1", 1*time.Hour)
	assert.Nil(t, err)

	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	time.Sleep(2 * time.Hour)
	value, ok = cache.Get("key1")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestCacheDelete(t *testing.T) {
	cache :=  NewCaches(map[string] Item{},  DefaultExpiration)

	err := cache.Set("key1", "value1", 1*time.Hour)
	assert.Nil(t, err)

	err = cache.Delete("key1")
	assert.Nil(t, err)

	value, ok := cache.Get("key1")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestCacheReplace(t *testing.T) {
	cache :=  NewCaches(map[string] Item{},  DefaultExpiration)

	err := cache.Set("key1", "value1", 1*time.Hour)
	assert.Nil(t, err)

	err = cache.Replace("key1", "value1_updated", 2*time.Hour)
	assert.Nil(t, err)

	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1_updated", value)
}

func TestCacheFlush(t *testing.T) {
	cache :=  NewCaches(map[string] Item{},  DefaultExpiration)

	_ = cache.Set("key1", "value1",  DefaultExpiration)
	_ = cache.Set("key2", "value2",  DefaultExpiration)

	cache.Flush()

	value, ok := cache.Get("key1")
	assert.False(t, ok)
	assert.Nil(t, value)

	value, ok = cache.Get("key2")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestCacheCount(t *testing.T) {
	cache :=  NewCaches(map[string] Item{},  DefaultExpiration)

	_ = cache.Set("key1", "value1",  DefaultExpiration)
	_ = cache.Set("key2", "value2",  DefaultExpiration)

	count := cache.Count()
	assert.Equal(t, 2, count)
}

var deletedKey string
var deletedValue interface{}

func onEvictedTest(key string, value interface{}) {
	deletedKey = key
	deletedValue = value
}


func TestOnEvicted(t *testing.T) {
	cache :=  NewCaches(map[string] Item{},  DefaultExpiration)

	cache.OnEvicted(onEvictedTest)

	_ = cache.Set("key1", "value1",  DefaultExpiration)
	_ = cache.Set("key2", "value2",  DefaultExpiration)

	_ = cache.Delete("key1")

	assert.Equal(t, "key1", deletedKey)
	assert.Equal(t, "value1", deletedValue)

	_ = cache.Delete("key2")

	assert.Equal(t, "key2", deletedKey)
	assert.Equal(t, "value2", deletedValue)
}