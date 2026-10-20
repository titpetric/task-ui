// Package event is the bus the pages follow: a publisher names a topic
// and an event, and every subscriber of the topic gets it. It holds
// everything in memory, for one instance, and never waits on a subscriber.
package event

import (
	"sync"
)

// Buffer is how many events a subscriber may fall behind before it
// loses the next ones.
const Buffer = 64

// Event is what a subscriber receives.
type Event struct {
	Topic string
	// Name says what happened: run, step, done.
	Name string
	// Data is the event's payload, JSON on the wire.
	Data any
}

// Bus delivers events by topic.
type Bus struct {
	mu          sync.Mutex
	subscribers map[string]map[int]chan Event
	next        int
}

// NewBus returns an empty bus.
func NewBus() *Bus {
	return &Bus{subscribers: map[string]map[int]chan Event{}}
}

// Subscribe returns a channel of the topic's events and the function
// that ends the subscription, which closes the channel.
func (b *Bus) Subscribe(topic string) (<-chan Event, func()) {
	ch := make(chan Event, Buffer)
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.next
	b.next++
	if b.subscribers[topic] == nil {
		b.subscribers[topic] = map[int]chan Event{}
	}
	b.subscribers[topic][id] = ch
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			delete(b.subscribers[topic], id)
			if len(b.subscribers[topic]) == 0 {
				delete(b.subscribers, topic)
			}
			close(ch)
		})
	}
}

// Publish sends an event to the topic's subscribers. One whose buffer
// is full loses the event: the publisher never waits.
func (b *Bus) Publish(topic, name string, data any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers[topic] {
		select {
		case ch <- Event{Topic: topic, Name: name, Data: data}:
		default:
		}
	}
}
