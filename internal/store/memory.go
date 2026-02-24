package store

import (
	"context"
	"sync"
	"time"
)

const defaultCleanupInterval = time.Minute

type memEntry struct {
	value int64
	expiresAt time.Time
}

type MemoryStore struct {
	mu sync.Mutex
	items map[string]memEntry

	stopOnce sync.Once
	stopCh chan struct{}
}

func NewMemoryStore() *MemoryStore {
	return NewMemoryStoreWithCleanupInterval(defaultCleanupInterval)
}

func NewMemoryStoreWithCleanupInterval(interval time.Duration) *MemoryStore{
	if interval <= 0 {
		interval = defaultCleanupInterval
	}

	m := &MemoryStore{
		items: make(map[string]memEntry),
		stopCh: make(chan struct{}),
	}

	go m.startCleanup(interval)

	return m
}

func (m *MemoryStore) IncrWithTTL(_ context.Context, key string, ttl time.Duration) (value int64, ttlRemaining time.Duration, err error) {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.items[key]; ok {
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			delete(m.items, key)
		}
	}

	e, ok := m.items[key]
	if !ok {
		expiresAt := now.Add(ttl)
		e = memEntry{value: 1, expiresAt: expiresAt}
		m.items[key] = e
		return e.value, time.Until(expiresAt), nil
	}

	e.value++
	m.items[key] = e
	return e.value, time.Until(e.expiresAt), nil
}

func (m *MemoryStore) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <- ticker.C:
			m.cleanupExpired()
		case <- m.stopCh:
			return
		}
	}
}

func (m *MemoryStore) cleanupExpired() {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	for k, e := range m.items {
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			delete(m.items, k)
		}
	}
}

func (m *MemoryStore) Close() error {
	m.stopOnce.Do(func() {
		close(m.stopCh)
	})

	return nil
}