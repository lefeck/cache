package cache1

import (
	"sync"
	"fmt"
)

// 不带过期时间的cache实现

type Item struct {
	Obj interface{}
}

//type Cache struct {
//	c *cache
//}

type cache struct {
	items map[string]Item
	mu sync.RWMutex
}

type Caches interface {
	Get(key string) (interface{}, bool)
	Delete(key string) bool
	Set(key string, value interface{}) error
	Flush()
	Replace(key string, value interface{}) error
	Copy(key string, value interface{}) map[string]Item
}

func NewCaches(m map[string]Item) Caches  {
	return &cache{
		items:m,
	}
}

func (c *cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]

	if !ok {
		return nil, false
	}

	return item.Obj, true

}

func (c *cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.items[key]
	if !ok {
		return false
	}

	delete(c.items,key)

	return  true
}

func (c *cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range c.items {
		_, ok := c.items[key]
		if ok {
			fmt.Errorf("key is exist %s",key)
		}
		c.items[key]=value
	}
	return nil
}

func (c *cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}

func (c *cache) Replace(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range c.items {
		_, ok := c.items[key]
		if ok {
			c.items[key]=value

		}
	}
	return nil
}

func (c *cache) Copy(key string, value interface{}) map[string]Item {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := make(map[string]Item,len(c.items))

	for key,value :=range c.items {
		item[key]=value
	}
	return item
}
