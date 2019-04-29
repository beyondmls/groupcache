/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lru

import "container/list"

// Cache是LRU缓存的实现，不是并发安全的
type Cache struct {
	// 缓存元素的最大数量限制，0 代表没有限制
	MaxEntries int

	// 缓存元素被移除的时候触发的回调函数
	OnEvicted func(key Key, value interface{})

	// 缓存元素存储的数据结构：双向链表+哈希表
	ll    *list.List
	cache map[interface{}]*list.Element
}

// 键值可以是任何可比较的数据类型
type Key interface{}

// 键值对的数据结构，存储到哈希表
type entry struct {
	key   Key
	value interface{}
}

// Cache结构的构造函数
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// 添加键值到缓存
func (c *Cache) Add(key Key, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	// 如果键值已缓存，将元素移动到双向链表的最前面，更新value
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}

	// 如果键值未缓存，将元素添加到双向链表的最前面
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		// 如果元素个数已经达到最大限制，移除最近没有使用的键值
		c.RemoveOldest()
	}
}

// 从缓存中获取键值
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		// 如果键值已缓存，将元素移动到双向链表的最前面，返回value
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// 从缓存中移除键值
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// 从缓存中移除最老的键值
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}

	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// 从缓存中移除键值
func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// 获取缓存的元素数量
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// 重置缓存，清除所有元素
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}
