package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes int64                    //允许使用的最大内存
	nbytes   int64                    //当前已使用的内存
	ll       *list.List               //双向链表
	cache    map[string]*list.Element //字典，映射每一个链表元素，加快查询速度

	OnEvicted func(key string, value Value)
}

type entry struct { //双向链表存储的数据类型
	key   string
	value Value
} //保存key是为了根据key从字典中删除数据

type Value interface {
	Len() int //返回所占用内存大小
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { //查找
		c.ll.MoveToFront(ele) //根据LRU策略，移至队首
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele) //从链表中删除
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                //从字典中删除
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) //更新内存使用情况
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { //键存在
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { //不存在
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	} //达到最大内存则进行淘汰
}
