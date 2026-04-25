package lsp

import (
	"sync"
	"time"
)

type cacheEntry[T any] struct {
	data      T
	expiresAt time.Time
}

type TTLMap[T any] struct {
	mu    sync.RWMutex
	items map[string]cacheEntry[T]
	ttl   time.Duration
}

func NewTTLMap[T any](ttl time.Duration) *TTLMap[T] {
	m := &TTLMap[T]{
		items: make(map[string]cacheEntry[T]),
		ttl:   ttl,
	}
	go m.cleanup()
	return m
}

func (m *TTLMap[T]) Set(key string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = cacheEntry[T]{
		data:      value,
		expiresAt: time.Now().Add(m.ttl),
	}
}

func (m *TTLMap[T]) Get(key string) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, found := m.items[key]
	if !found {
		var zero T
		return zero, false
	}
	if time.Now().After(entry.expiresAt) {
		var zero T
		return zero, false
	}
	return entry.data, true
}

func (m *TTLMap[T]) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for k, v := range m.items {
			if now.After(v.expiresAt) {
				delete(m.items, k)
			}
		}
		m.mu.Unlock()
	}
}
