package cache

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Item struct {
	Object     interface{}
	Expiration int64
}

type Cache struct {
	*cache
}

const (
	NotExpiration     time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type cache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex

	// 驱逐过期的key and value， 定义了回调函数, 如果我不定义这种回调函数执行，代码会存在什么问题？
	onEvicted func(string, interface{})

	//控制时间过期,
	janitor *janitor
}

// 实时探测时间是否超过设定的过期时间，如果，过期channel就会返回true，表示过期，可以clear
type janitor struct {
	Internel time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *cache) {
	ticker := time.NewTicker(j.Internel)
	for {
		select {
		// 从ticker中读取间隔d时间
		case <-ticker.C:
			//删除已经过期的cache
			c.DeleteExpired()
			//
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

// 初始化janitor，
func runJanitor(d time.Duration, cache *cache) {
	//初始化
	j := &janitor{
		Internel: d,
		stop:     make(chan bool),
	}
	cache.janitor = j
	//调用方法实时探测，过期时间
	go j.Run(cache)
}

// 初始化cache中，过期时间和传入的value
func newCache(de time.Duration, m map[string]Item) *cache {
	if de == 0 {
		de = -1
	}
	c := &cache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

// 实例化Cache， 调用cache，de是标识过期时间，ci 是标识清除
func newCacheWithJanitor(de time.Duration, ci time.Duration, m map[string]Item) *Cache {
	c := newCache(de, m)

	C := &Cache{c}

	if ci > 0 {
		//调用定时器计时
		runJanitor(ci, c)
		// 关键点，通过runtime控制超时时间过期。
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

func New(de time.Duration, ci time.Duration) *Cache {
	m := make(map[string]Item)
	return newCacheWithJanitor(de, ci, m)
}

func NewFrom(defaultExpiration time.Duration, clearupInternal time.Duration, items map[string]Item) *Cache {
	return newCacheWithJanitor(defaultExpiration, clearupInternal, items)
}

type KeyAndValue struct {
	Key   string
	Value interface{}
}

func (c *cache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	var tmpitem []KeyAndValue
	now := time.Now().UnixNano()
	for key, value := range c.items {
		//cache过期
		if value.Expiration > 0 && now > value.Expiration {
			//删除cache
			v, ok := c.delete(key)
			//记录过期的cache
			if ok {
				tmpitem = append(tmpitem, KeyAndValue{Key: key, Value: v})
			}
		}
	}
	//写入到新的struct中
	for _, v := range tmpitem {
		c.onEvicted(v.Key, v.Value)
	}
}

// 驱逐过期缓存的方法, 驱逐在cache中存在的key value
func (c *cache) OnEvicted(f func(string, interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvicted = f
}

func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration //超时cache过期
}

func (c *cache) Set(key string, value interface{}, d time.Duration) {
	var e int64
	c.mu.Lock()
	defer c.mu.Unlock()

	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano() // d时间后的纳秒值
	}

	c.items[key] = Item{
		Object:     value,
		Expiration: e, // 写入到items
	}
	return
}

func (c *cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if c.defaultExpiration > 0 {
		//cache已经过期
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Object, true
}

func (c *cache) Add(key string, value interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 获取是否存在
	_, ok := c.get(key)
	if ok {
		return fmt.Errorf("key is exist")
	}
	// 不存在就添加
	c.Set(key, value, d)
	return nil
}

func (c *cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.delete(key)

	if ok {
		c.onEvicted(key, value)
	}

	return nil
}

func (c *cache) delete(key string) (interface{}, bool) {
	if c.onEvicted != nil {
		item, ok := c.items[key]
		if ok {
			delete(c.items, key)
			return item.Object, true
		}
	}
	delete(c.items, key)
	return nil, false
}

func (c *cache) Replace(key string, value interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.get(key)
	if !ok {
		return fmt.Errorf("key is not exist")
	}
	c.Set(key, value, d)
	return nil
}

func (c *cache) get(key string) (interface{}, bool) {
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if c.defaultExpiration > 0 {
		//cache已经过期
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Object, true
}

func (c *cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}

func (c *cache) ItemCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// increment会增加一个n值，这个增加的值会根据传入值的类型，为标准上增加n大小，因为golang是强类型，所以要这样做
func (c *cache) Increment(key string, n int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.items[key]
	if !ok || value.Expired() {
		fmt.Errorf("key is not exist")
	}
	switch value.Object.(type) {
	case int:
		value.Object = value.Object.(int) + int(n)
	case int32:
		value.Object = value.Object.(int32) + int32(n)
	case int64:
		value.Object = value.Object.(int64) + n
	case uint:
		value.Object = value.Object.(uint) + uint(n)
	case uintptr:
		value.Object = value.Object.(uintptr) + uintptr(n)
	case uint8:
		value.Object = value.Object.(uint8) + uint8(n)
	case uint16:
		value.Object = value.Object.(uint16) + uint16(n)
	case uint32:
		value.Object = value.Object.(uint32) + uint32(n)
	case uint64:
		value.Object = value.Object.(uint64) + uint64(n)
	case float32:
		value.Object = value.Object.(float32) + float32(n)
	case float64:
		value.Object = value.Object.(float64) + float64(n)
	default:
		fmt.Errorf("the value for %s is not an integer", key)
	}
	c.items[key] = value
	return nil
}

// Copies all unexpired items in the cache into a new map and returns it.
func (c *cache) Items() map[string]Item {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := make(map[string]Item, len(c.items))
	for key, value := range c.items {
		now := time.Now().UnixNano()

		if value.Expiration > 0 {
			//cache已经过期
			if now > value.Expiration {
				continue
			}
			m[key] = value
		}
	}
	return m
}

func (c *cache) SetDefault(key string, value interface{}) {
	c.Set(key, value, DefaultExpiration)
}
