package test

import (
	"event-bus/eventbus"
	"sync"
	"testing"
)

const (
	eventLabel1 = "event1"
	eventLabel2 = "event2"
	eventLabel3 = "event3"
)

type Subscriber struct {
	sync.WaitGroup
}

func (s *Subscriber) OnEvent(label string, event interface{}) {
	defer s.WaitGroup.Done()

	switch label {

	}
}

type Event struct {
	Content string
}

func TestOnePublisherAndOneSubscriber(t *testing.T) {
	subscriber := new(Subscriber)
	subscriber.WaitGroup.Add(3)

	bus := eventbus.NewEventBus()
	bus.RegisterSubscriber(subscriber, eventLabel1, eventLabel2, eventLabel3)

	bus.Publish(eventLabel1, &Event{
		Content: "aaaa",
	})

	bus.Publish(eventLabel2, &Event{
		Content: "bbbb",
	})

	bus.Publish(eventLabel3, &Event{
		Content: "cccc",
	})

	subscriber.WaitGroup.Wait()

	bus.UnRegisterSubscriber(subscriber)
	eventbus.DestroyEventBus(bus)
}
