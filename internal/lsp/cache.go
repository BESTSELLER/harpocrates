package lsp

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

type TTLMap struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

func NewTTLMap(ttl time.Duration) *TTLMap {
	m := &TTLMap{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
	}
	go m.cleanup()
	return m
}

func (m *TTLMap) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = cacheEntry{
		data:      value,
		expiresAt: time.Now().Add(m.ttl),
	}
}

func (m *TTLMap) Get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, found := m.items[key]
	if !found {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (m *TTLMap) cleanup() {
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
