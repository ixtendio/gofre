package cache

import "time"

// The Cache interface defines the methods that a cache implementation should have
type Cache interface {
	// Add adds a key in the cache with an expiration time
	Add(key string, expiresIn time.Duration) error
	// Remove removes a key from the cache
	Remove(key string)
	// Contains returns true if the key is in the cache and was not expired
	Contains(key string) bool
}

// InMemory is a simple Cache implementation used for testing and examples
type InMemory struct {
	cache map[string]time.Time
}

func NewInMemory() *InMemory {
	return &InMemory{cache: make(map[string]time.Time)}
}

func (i *InMemory) Add(key string, expiresIn time.Duration) error {
	i.cache[key] = time.Now().Add(expiresIn)
	return nil
}

func (i *InMemory) Remove(key string) {
	delete(i.cache, key)
}

func (i *InMemory) Contains(key string) bool {
	t, found := i.cache[key]
	return found && time.Now().Before(t)
}
