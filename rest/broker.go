package rest

import "sync"

type Event struct {
	Action string   `json:"action"`
	ID     string   `json:"id"`
	Data   Resource `json:"data"`
}

type Broker struct {
	channels map[string]map[chan Event]struct{} // resource -> channels（空结构体优化内存）
	mu       sync.RWMutex
}

// Subscribe to a resource.
func (b *Broker) Subscribe(resource string, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.channels[resource] == nil {
		b.channels[resource] = make(map[chan Event]struct{})
	}
	b.channels[resource][ch] = struct{}{} // 空结构体无内存开销
}

// Unsubscribe from a resource.
func (b *Broker) Unsubscribe(resource string, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs := b.channels[resource]; subs != nil {
		delete(subs, ch)
	}
}

// Publish an event to a resource.
func (b *Broker) Publish(resource string, evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs := b.channels[resource]; subs != nil {
		for ch := range subs {
			select {
			case ch <- evt:
			default:
			}
		}
	}
}
