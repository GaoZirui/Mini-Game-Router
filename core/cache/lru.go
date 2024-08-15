package cache

import (
	"container/list"
	"sync"
	"ziruigao/mini-game-router/core/router"
)

type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type entry struct {
	key   string
	value *router.Endpoint
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache) Get(key string) *router.Endpoint {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*entry).value
	}
	return nil
}

func (c *LRUCache) Put(key string, value *router.Endpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
	} else {
		if c.list.Len() == c.capacity {
			last := c.list.Back()
			delete(c.cache, last.Value.(*entry).key)
			c.list.Remove(last)
		}
		newEntry := &entry{key, value}
		newElem := c.list.PushFront(newEntry)
		c.cache[key] = newElem
	}
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		delete(c.cache, key)
		c.list.Remove(elem)
	}
}

func (c *LRUCache) Reset() {
	c.cache = map[string]*list.Element{}
	c.list = list.New()
}
