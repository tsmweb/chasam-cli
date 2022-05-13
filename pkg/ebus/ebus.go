/*
Package ebus implements the event bus design pattern, being an alternative to component
communication while maintaining loose coupling and separation of interests principles.
Publish event:
	const topic = "user"
	type user struct {
		id   int
		name string
	}
	u := &user {...}
	Instance().Publish(topic, u)
Subscribe to the topic:
	sub := Instance().Subscribe(topic)
	defer sub.Unsubscribe()
	for event := range sub.Event {
		go func(e DataEvent) {
			fmt.Printf("Topic: %s", e.Topic)
			fmt.Printf("Data: %v", e.Data.(user))
		}(event)
	}
*/

package ebus

import "sync"

// DataEvent represents an event posted to a topic.
type DataEvent struct {
	Data  interface{}
	Topic string
}

// dataChannel is a channel which can accept a DataEvent.
type dataChannel chan DataEvent

// dataChannelSet is a set of DataChannels.
type dataChannelSet map[dataChannel]bool

// Subscription represents a subscription to a topic.
type Subscription struct {
	Event       <-chan DataEvent
	Unsubscribe func()
}

// EventBus stores the information about subscribers interested for a particular topic.
type EventBus struct {
	subscribers map[string]dataChannelSet
	mu          sync.RWMutex
}

// NewEventBus creates an EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: map[string]dataChannelSet{},
	}
}

var eventBus *EventBus

// Instance creates an single EventBus instance.
func Instance() *EventBus {
	if eventBus == nil {
		eventBus = NewEventBus()
	}
	return eventBus
}

// Publish publishes data in the topic provided.
func (eb *EventBus) Publish(topic string, data interface{}) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if chans, found := eb.subscribers[topic]; found {
		// go func(data DataEvent, dataChannelSet dataChannelSet) {
		// 	for ch := range dataChannelSet {
		// 		ch <- data
		// 	}
		// }(DataEvent{Data: data, Topic: topic}, chans)

		for ch := range chans {
			ch <- DataEvent{Data: data, Topic: topic}
		}
	}
}

// Subscribe to the topic provided to receive data events.
func (eb *EventBus) Subscribe(topic string) *Subscription {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan DataEvent, 1)

	if prev, found := eb.subscribers[topic]; found {
		prev[ch] = true
	} else {
		eb.subscribers[topic] = dataChannelSet{ch: true}
	}

	return &Subscription{
		Event: ch,
		Unsubscribe: func() {
			eb.unsubscribe(topic, ch)
		},
	}
}

func (eb *EventBus) unsubscribe(topic string, ch chan DataEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if chans, found := eb.subscribers[topic]; found {
		delete(chans, ch)
	}
}
