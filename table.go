package cache

import (
	"log"
	"sort"
	"sync"
	"time"
)

type Caches struct {
	sync.RWMutex
	name string
	items map[interface{}]*Items
	cleanupTimer *time.Timer
	cleanupInterval   time.Duration
	logger *log.Logger
	loadData func(key interface{}, args ...interface{}) *Items
	addItem []func(item *Items)
	deleteItem []func(item *Items)
}

func (c *Caches) Count()int {
	c.RLock()
	defer c.RUnlock()
	return len(c.items)
}

func (c *Caches)Iterator(trans func(key interface{}, item *Items)) {
	c.RLock()
	defer c.RUnlock()

	for k, v :=range c.items {
		trans(k,v)
	}
}
// DataLoader 配置一个数据加载器回调，当试图访问一个不存在的键时会调用它。 键和 0...n 个附加参数被传递给回调函数。
func (c *Caches)LoaderData(f func(interface{}, ...interface{}) *Items)  {
	c.RLock()
	defer c.RUnlock()
	 c.loadData = f
}
//配置一个回调，每次有新的条目添加到缓存中时都会调用该回调。
func (c *Caches)AddItemCallback(f func( *Items ))  {
	if len(c.addItem) == 0 {
		c.RemoveAddedItemCallbacks()
	}
	c.addItem = append(c.addItem,f)
}

func (c *Caches)RemoveAddedItemCallbacks()  {
	c.RLock()
	defer c.RUnlock()
	c.deleteItem = nil
}

func (c *Caches)SetLogger(logger *log.Logger)  {
	c.RLock()
	defer c.RUnlock()
	c.logger = logger
}

func (c *Caches)expirationCheck()  {
	c.Lock()
	if c.cleanupTimer != nil {
		c.cleanupTimer.Stop()
	}
	if c.cleanupInterval > 0 {
		c.log("Expiration check triggered after", c.cleanupInterval, "for table", c.name)
	} else {
		c.log("Expiration check installed for table", c.name)
	}

	// To be more accurate with timers, we would need to update 'now' on every
	// loop iteration. Not sure it's really efficient though.
	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range c.items {
		// Cache values so we don't keep blocking the mutex.
		item.RLock()
		lifeSpan := item.lifePeriod
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifeSpan == 0 {
			continue
		}
		if now.Sub(accessedOn) >= lifeSpan {
			// Item has excessed its lifespan.
			c.deleteInternal(key)
		} else {
			// Find the item chronologically closest to its end-of-lifespan.
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	// Setup the interval for the next cleanup run.
	c.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		c.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go c.expirationCheck()
		})
	}
	c.Unlock()
}

func (c *Caches)addInternal(item *Items)  {
	//它会在运行回调和检查之前为调用者解锁它
	c.items[item.key] = item

	//缓存值，这样就不会一直阻塞互斥锁。
	expire := c.cleanupInterval
	addedTime := c.addItem
	//delete rwlock
	c.Unlock()

	//将项目添加到缓存后触发回调。
	if addedTime !=nil {
		for _, callback :=range addedTime {
			callback(item)
		}
	}
	if item.lifePeriod > 0 && (expire == 0 || item.lifePeriod < expire) {
		c.expirationCheck()
	}
}

// Add 将键/值对添加到缓存中。
// 参数key是item的cache-key。
// 参数 lifeSpan 确定在哪个时间段之后没有访问该项目
// 将从缓存中删除。
// 参数数据是itme的值。
func (c *Caches)Add(key interface{},lifeperiod time.Duration, data interface{}) *Items  {
	item := NewItems(key,lifeperiod,data)

	//add rwlock
	c.Lock()
	c.addInternal(item)
	return item
}

func (c *Caches)deleteInternal(key interface{})(*Items,error)  {
	//检查key是否存在
	r, ok := c.items[key]
	if !ok {
		return  nil, ErrKeyNotFound
	}
	//slice type item
	deleteitem := c.deleteItem
	c.Unlock()

	// delete key 存在， 就遍历 key对应的value， 删除value
	if deleteitem !=nil {
		for _, callback := range deleteitem {
			callback(r)
		}
	}
	r.RLock()
	defer r.RUnlock()
	// 检查lifecycle是否过期，获取过期时间，回调函数删除
	if r.expire !=nil {
		for _, callback := range r.expire {
			callback(key)
		}
	}
	c.Lock()
	// 删除缓存中的key
	delete(c.items,key)
	return r, nil
}
// Delete an item from the cache.
func (c *Caches)Delete(key interface{}) (*Items, error)  {
	c.RLock()
	defer c.RUnlock()
	return  c.deleteInternal(key)
}
// Exists returns whether an item exists in the cache. Unlike the Value method
// Exists neither tries to fetch data via the loadData callback nor does it
// keep the item alive in the cache.
func (c *Caches)Exist(key interface{}) bool  {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.items[key]
	return ok
}

// NotFoundAdd checks whether an item is not yet cached. Unlike the Exists
// method this also adds data if the key could not be found.
func (c *Caches)NotFoundAdd(key interface{}, lifePeriod time.Duration, data interface{}) bool  {
	c.Lock()
	if _, ok :=c.items[key]; ok {
		c.Unlock()
		return false
	}
	item := NewItems(key,lifePeriod,data)
	c.addInternal(item)
	return true
}


// 值从缓存中返回一个struct并将其标记为保持活动状态。 你可以将附加参数传递给您的 DataLoader 回调函数。
func (c *Caches)Value(key interface{}, args ...interface{})(*Items,error)  {
	c.RLock()
	defer c.RUnlock()
	r, ok := c.items[key]
	loadData := c.loadData

	if ok {
		// Update access counter and timestamp.
		r.KeepAlive()
		return r, nil
	}

	//缓存中不存在item。 尝试使用数据加载器获取它。
	if loadData !=nil {
		item := loadData(key, args...)
		if item !=nil {
			c.Add(key,item.lifePeriod,item.value)
			return item,nil
		}
		return nil, ErrKeyNotFoundOrLoadable
	}
	return nil, ErrKeyNotFound
}
// Flush deletes all items from this cache table.
func (c *Caches)Flush()  {
	c.Lock()
	defer c.Unlock()
	c.items = make(map[interface{}]*Items)
	c.cleanupInterval = 0
	if c.cleanupTimer !=nil {
		c.cleanupTimer.Stop()
	}
}

//对item key 进行统计排序
type ItemPair struct {
	Key interface{}
	Counter int64
}
type ItemPairList []ItemPair

func (c ItemPairList)Swap(i, j int)  {
	c[i],c[j] = c[j],c[i]
}
func (c ItemPairList)Len()int  {
	return len(c)
}
func (c ItemPairList)Less(i,j int) bool  {
	return c[i].Counter > c[j].Counter
}
//MostAccessed returns the most accessed items in this cache table
func (c *Caches)MostAccess(count int64) []*Items  {
	c.RLock()
	defer c.RUnlock()
	p := make(ItemPairList,len(c.items))
	i := 0
	for k, v := range  p {
		p[i]= ItemPair{k,v.Counter}
		i++
	}
	sort.Sort(p)
	var r []*Items
	x := int64(0)
	for _, v := range p {
		if x >= count {
			break
		}
		//?
		item, ok := c.items[v.Key]
		if ok {
			r = append(r,item)
		}
		x++
	}
	return r
}
// Internal logging method for convenience.
func (c *Caches)log(v ...interface{})  {
	if c.logger == nil {
		return
	}
	c.logger.Println(v)
}