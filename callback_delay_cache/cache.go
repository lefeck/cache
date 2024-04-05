package nodelay_cache

import (
	"fmt"
	"sync"
	"time"
)

// 3.  adding expiration time for memory cache implementation add callback function

const (
	DefaultExpiration time.Duration = 0
)

type Item struct {
	Obj        interface{}
	Expiration int64
}

func (item *Item) Expired() bool {
	if item.Expiration < 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

//type Cache struct {
//	c *cache
//}

type cache struct {
	items             map[string]Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	// 为使用这个框架的人提供一个接口使用
	onEvicted func(string, interface{})
}

func (c *cache) OnEvicted(f func(string, interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvicted = f
}

type Caches interface {
	Get(key string) (interface{}, bool)
	Delete(key string) error
	Set(key string, value interface{}, d time.Duration) error
	Flush()
	Replace(key string, value interface{}, d time.Duration) error
	Copy() map[string]Item
	Count() int
	DeleteExpired()
}

func NewCaches(m map[string]Item, d time.Duration) Caches {
	return &cache{
		items:             m,
		defaultExpiration: d,
	}
}

// record experation data to slice
type KeyAndValue struct {
	Key   string
	Value interface{}
}

func (c *cache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiredItems := []KeyAndValue{}
	for k, v := range c.items {
		if v.Expiration > 0 && time.Now().UnixNano() > v.Expiration {
			value, ok := c.delete(k)
			if ok {
				expiredItems = append(expiredItems, KeyAndValue{k, value})
			}
		}
	}
	return
}

func (c *cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.get(key)
}

func (c *cache) get(key string) (interface{}, bool) {
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if c.defaultExpiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Obj, true
}

func (c *cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.delete(key)
	if ok {
		c.onEvicted(key,v)
	}

	return nil
}

func (c *cache) delete(key string) (interface{}, bool) {
	if c.onEvicted !=nil {
		item, ok := c.items[key]
		if ok {
			return item.Obj, true
		}
		delete(c.items, key)
	}

	delete(c.items, key)

	return nil, false
}

func (c *cache) Set(key string, value interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var e int64

	_, ok := c.get(key)
	if ok {
		fmt.Errorf("key is exist %s", key)
	}

	if d == DefaultExpiration {
		c.defaultExpiration = DefaultExpiration
	}

	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}

	c.items[key] = Item{
		Obj:        value,
		Expiration: e,
	}
	return nil
}

func (c *cache) set(key string, value interface{}, d time.Duration) error {
	var e int64
	if d == DefaultExpiration {
		c.defaultExpiration = DefaultExpiration
	}

	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}

	c.items[key] = Item{
		Obj:        value,
		Expiration: e,
	}
	return nil
}

func (c *cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}

func (c *cache) Replace(key string, value interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.get(key)
	if !ok {
		fmt.Errorf("key is not exist %s", key)
	}
	c.set(key, value, d)
	return nil
}

func (c *cache) Copy() map[string]Item {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := make(map[string]Item, len(c.items))

	for key, value := range c.items {
		if value.Expiration > 0 {
			if time.Now().UnixNano() > value.Expiration {
				continue
			}
		}
		item[key] = value
	}
	return item
}
