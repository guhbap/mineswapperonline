package cache

import (
	"sync"
	"time"
)

// CacheEntry представляет запись в кеше с временем истечения
type CacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
}

// Cache представляет in-memory кеш с TTL
type Cache struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry
	ttl   time.Duration
}

// NewCache создает новый кеш с указанным TTL
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
	// Запускаем горутину для очистки устаревших записей
	go c.cleanup()
	return c
}

// Get получает значение из кеша по ключу
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Проверяем, не истекла ли запись
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set сохраняет значение в кеш с TTL
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete удаляет значение из кеша
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Invalidate удаляет все записи, начинающиеся с префикса
func (c *Cache) Invalidate(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}

// Clear очищает весь кеш
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheEntry)
}

// cleanup периодически очищает устаревшие записи
func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2) // Проверяем каждые TTL/2
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.items {
			if now.After(entry.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

