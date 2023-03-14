package golangunitedschoolcerts

import "sync"

type SafeCache[K comparable, V Cacheable] struct {
	c  Cache[K, V]
	mu sync.RWMutex
}

func NewSafeCache[K comparable, V Cacheable](c Cache[K, V]) *SafeCache[K, V] {
	return &SafeCache[K, V]{c: c}
}

func (s *SafeCache[K, V]) Add(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Add(key, value)
}

func (s *SafeCache[K, V]) Contains(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Contains(key)
}

func (s *SafeCache[K, V]) Get(key K) (*V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.c.Get(key)
}

func (s *SafeCache[K, V]) Remove(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Remove(key)
}

func (s *SafeCache[K, V]) RemoveOldest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.RemoveOldest()
}

func (s *SafeCache[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Keys()
}

func (s *SafeCache[K, V]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Size()
}

func (s *SafeCache[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Len()
}

func (s *SafeCache[K, V]) Capacity() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Capacity()
}

func (s *SafeCache[K, V]) Resize(capacity int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Resize(capacity)
}

func (s *SafeCache[K, V]) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Purge()
}

func (s *SafeCache[K, V]) Peek(key K) (*V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Peek(key)
}

func (s *SafeCache[K, V]) Touch(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c.Touch(key)
}
