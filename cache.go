package golangunitedschoolcerts

import "errors"

type Cache[K comparable, V Cacheable] interface {
	// Adds value to the cache
	Add(K, V)
	// Returns value for key and updates it's recent status
	Get(K) (*V, bool)
	// Returns value without changing it recent status
	Peek(K) (*V, bool)
	// Update recent status
	Touch(K)
	// Checks is key is in the cache without updating it recent status
	Contains(K) bool
	// Return keys sorted from old to new
	Keys() []K
	// Remove value, invoking onEviction callback if it was provided
	Remove(K)
	// Remove oldest value, invoking onEviction callback if it was provided
	RemoveOldest()
	// Remove all values, invoking onEviction callback if it was provided
	Purge()
	// Returns cache capacity
	Capacity() int
	// Returns size of values in cache
	Size() int
	// Returns number of values in cache
	Len() int
	// Changes cache capacity, invoking onEviction callback if it was provided
	Resize(int)
}
type Cacheable interface {
	Size() int
}
type LRUCache[K comparable, V Cacheable] struct {
	capacity   int
	used       int
	head       *node[K, V]
	tail       *node[K, V]
	items      map[K]*node[K, V]
	onEviction EvictionCallback[K, V]
}

const (
	B = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
)

type EvictionCallback[K comparable, V Cacheable] func(key *K, value *V)

type node[K comparable, V Cacheable] struct {
	next  *node[K, V]
	prev  *node[K, V]
	key   K
	value V
}

func NewLRUCache[K comparable, V Cacheable](capacity int, onEviction EvictionCallback[K, V]) (*LRUCache[K, V], error) {
	if capacity < 0 {
		return nil, errors.New("capacity can't be negative")
	}
	return &LRUCache[K, V]{
		capacity:   capacity,
		used:       0,
		head:       nil,
		tail:       nil,
		items:      make(map[K]*node[K, V]),
		onEviction: onEviction,
	}, nil
}

func (c *LRUCache[K, V]) Add(key K, value V) {
	if c.Contains(key) {
		c.Remove(key)
	}
	n := node[K, V]{
		key:   key,
		value: value,
	}
	c.items[key] = &n
	c.addToHead(&n)
	c.used += value.Size()
	c.checkSize()
}

func (c *LRUCache[K, V]) addToHead(n *node[K, V]) {
	n.prev = c.head
	n.next = nil
	if c.head != nil {
		c.head.next = n
	}
	if c.tail == nil {
		c.tail = n
	}
	c.head = n
}

func (c *LRUCache[K, V]) removeFromList(n *node[K, V]) {
	if n == c.head {
		c.head = n.prev
		if n.prev != nil {
			n.prev.next = nil
		}
	}
	if n == c.tail {
		c.tail = n.next
		if n.next != nil {
			n.next.prev = nil
		}
	}
	if n.next != nil && n.prev != nil {
		n.next.prev = n.prev
		n.prev.next = n.next
	}
	n.prev = nil
	n.next = nil
}

func (c *LRUCache[K, V]) checkSize() {
	for c.capacity != 0 && c.used > c.capacity {
		c.RemoveOldest()
	}
}

func (c *LRUCache[K, V]) Contains(key K) bool {
	_, ok := c.items[key]
	return ok
}

func (c *LRUCache[K, V]) Get(key K) (*V, bool) {
	if n, ok := c.items[key]; ok {
		c.removeFromList(n)
		c.addToHead(n)
		return &n.value, true
	}
	return nil, false
}

func (c *LRUCache[K, V]) Remove(key K) {
	if n, ok := c.items[key]; ok {
		delete(c.items, key)
		c.removeFromList(n)
		c.used -= n.value.Size()
		if c.used < 0 {
			c.used = 0
		}
		if c.onEviction != nil {
			c.onEviction(&n.key, &n.value)
		}
	}
}

func (c *LRUCache[K, V]) RemoveOldest() {
	if c.tail != nil {
		c.Remove(c.tail.key)
	}
}

func (c *LRUCache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	for n := c.tail; n != nil; n = n.next {
		keys = append(keys, n.key)
	}
	return keys
}

func (c *LRUCache[K, V]) Size() int {
	return c.used
}

func (c *LRUCache[K, V]) Len() int {
	return len(c.items)
}

func (c *LRUCache[K, V]) Capacity() int {
	return c.capacity
}

func (c *LRUCache[K, V]) Resize(capacity int) {
	if capacity >= c.capacity {
		c.capacity = capacity
		return
	}
	c.capacity = capacity
	c.checkSize()
}

func (c *LRUCache[K, V]) Purge() {
	for _, k := range c.Keys() {
		c.Remove(k)
	}
}

func (c *LRUCache[K, V]) Peek(key K) (*V, bool) {
	if n, ok := c.items[key]; ok {
		return &n.value, true
	}
	return nil, false
}

func (c *LRUCache[K, V]) Touch(key K) {
	if n, ok := c.items[key]; ok {
		c.removeFromList(n)
		c.addToHead(n)
	}
}
