package cache

import (
	"sync"
	"time"
)

//
type Items struct {
	//加锁控制并发读写
	sync.RWMutex
	key interface{}
	value interface{}
	// 当不被访问保持活动时，缓存中数据存活多长时间。
	lifePeriod time.Duration
	//创建时间
	createdOn time.Time
	//访问时间
	accessedOn time.Time
	// value 被访问的次数
	count int64
	// 在从缓存中删除项之前触发回调方法
	expire []func(key interface{})
}

func NewItems(key interface{}, period time.Duration, value interface{} ) *Items  {
	time := time.Now()
	return &Items{
		key:key,
		value: value,
		lifePeriod: period,
		createdOn:  time,
		accessedOn: time,
		expire: nil,
		count: 0,
	}
}

func (item *Items) KeepAlive()   {
	item.RLock()
	defer item.RUnlock()
	item.accessedOn = time.Now()
	item.count++
}

func (item *Items) Period() time.Duration {
	return item.lifePeriod
}

func (item *Items) CreatedOn() time.Time   {
	return item.createdOn
}

func (item *Items) AccessedOn() time.Time   {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

func (item *Items) Key() interface{} {
	return item.key
}

func (item *Items) Value() interface{} {
	return item.value
}

type fe func(interface{})

//配置一个回调，它将在项目即将从缓存中删除之前调用。
func (item *Items) SetExpireCallBack(f fe)  {
	if len(item.expire) > 0 {
		item.RemoveExpireCallBack()
	}
	item.RLock()
	defer item.RUnlock()
	item.expire = append(item.expire,f)
}

//将新回调附加到Expire队列
func (item *Items) AddExpireCallBack(f fe)  {
	item.RLock()
	defer item.RUnlock()
	item.expire = append(item.expire,f)
}

//清空即将过期的回调队列
func (item *Items) RemoveExpireCallBack()  {
	item.RLock()
	defer item.RUnlock()
	item.expire = nil
}
