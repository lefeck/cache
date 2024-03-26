package nodelay_cache

import "testing"

func TestCacheOperations(t *testing.T) {
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

//func TestCache(t *testing.T) {
//	// 创建一个新的 Cache 实例
//	cache := NewCaches(make(map[string]Item))
//
//	// 设置键值对
//	err := cache.Set("key1", "value1")
//	if err != nil {
//		t.Errorf("Set failed: %v", err)
//	}
//
//	// 获取键值对并检查是否正确
//	val, ok := cache.Get("key1")
//	if !ok {
//		t.Error("Get failed: key not found")
//	} else if val != "value1" {
//		t.Errorf("Get failed: expected 'value1', got '%v'", val)
//	}
//
//	// 替换键值对
//	//err = cache.Replace("key1", "value2")
//	//if err != nil {
//	//	t.Errorf("Replace failed: %v", err)
//	//}
//
//	// 获取替换后的值并检查是否正确
//	val, ok = cache.Get("key1")
//	if !ok {
//		t.Error("Get failed: key not found after Replace")
//	} else if val != "value1" {
//		t.Errorf("Get failed: expected 'value2', got '%v'", val)
//	}
//
//	// 删除键值对
//	deleted := cache.Delete("key1")
//	if !deleted {
//		t.Error("Delete failed: key not found")
//	}
//
//	// 确保键值对已删除
//	_, ok = cache.Get("key1")
//	if ok {
//		t.Error("Get after Delete failed: key still exists")
//	}
//}
