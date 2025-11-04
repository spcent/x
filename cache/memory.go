package cache

import (
	"sync"
	"time"
)

// Memory struct contains *memcache.Client
type Memory struct {
	sync.Mutex

	data map[string]*data
}

type data struct {
	Data    any       // store data
	Expired time.Time // expire time
}

// NewMemory create new memcache
func NewMemory() *Memory {
	return &Memory{
		data: map[string]*data{},
	}
}

// Get return cached value
func (m *Memory) Get(key string) any {
	if ret, ok := m.data[key]; ok {
		if ret.Expired.Before(time.Now()) {
			m.deleteKey(key)
			return nil
		}
		return ret.Data
	}
	return nil
}

// IsExist check value exists in memcache.
func (m *Memory) IsExist(key string) bool {
	if ret, ok := m.data[key]; ok {
		if ret.Expired.Before(time.Now()) {
			m.deleteKey(key)
			return false
		}
		return true
	}
	return false
}

// Set cached value with key and expire time.
func (m *Memory) Set(key string, val any, timeout time.Duration) error {
	m.Lock()
	defer m.Unlock()

	m.data[key] = &data{
		Data:    val,
		Expired: time.Now().Add(timeout),
	}
	return nil
}

// Delete delete value in memcache.
func (m *Memory) Delete(key string) error {
	m.deleteKey(key)
	return nil
}

// deleteKey
func (m *Memory) deleteKey(key string) {
	m.Lock()
	defer m.Unlock()

	delete(m.data, key)
}
