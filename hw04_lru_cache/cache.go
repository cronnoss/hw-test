package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheItem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	item, ok := c.items[key]

	if !ok {
		if (c.queue.Len() + 1) > c.capacity {
			lastItem := c.queue.Back()
			c.queue.Remove(lastItem)
			delete(c.items, lastItem.Value.(cacheItem).key)
		}

		c.items[key] = c.queue.PushFront(cacheItem{
			key:   key,
			value: value,
		})
	} else {
		item.Value = cacheItem{
			key:   key,
			value: value,
		}

		c.queue.MoveToFront(item)
	}

	return ok
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	item, ok := c.items[key]

	if !ok {
		return nil, false
	}

	c.queue.MoveToFront(item)

	return item.Value.(cacheItem).value, true
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
