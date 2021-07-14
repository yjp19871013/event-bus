package test

import (
	"event-bus/eventbus"
	"strconv"
	"sync"
	"testing"
)

const (
	dealEventRoutingCount = 5
	startBlockEventCount  = 50
	maxSubscribersCount   = 10000000
)

const (
	eventLabel1 = "event1"
	eventLabel2 = "event2"
	eventLabel3 = "event3"
)

const (
	eventContent1 = "aaa"
	eventContent2 = "bbb"
	eventContent3 = "ccc"
)

type Subscriber struct {
	id string
	t  *testing.T
	*sync.WaitGroup
}

func (s *Subscriber) OnEvent(label string, event interface{}) {
	defer s.WaitGroup.Done()

	e, ok := event.(*Event)
	if !ok {
		s.t.Fatal(s.id, "事件结构类型错误")
	}

	switch label {
	case eventLabel1:
		if e.Content != eventContent1 {
			s.t.Fatal(s.id, eventLabel1, "事件数据错误")
		}
	case eventLabel2:
		if e.Content != eventContent2 {
			s.t.Fatal(s.id, eventLabel2, "事件数据错误")
		}
	case eventLabel3:
		if e.Content != eventContent3 {
			s.t.Fatal(s.id, eventLabel3, "事件数据错误")
		}
	default:
		s.t.Fatal(s.id, "事件label不存在")
	}
}

type Event struct {
	Content string
}

func TestOnePublisherAndOneSubscriber(t *testing.T) {
	subscriber := Subscriber{
		id:        "1",
		t:         t,
		WaitGroup: &sync.WaitGroup{},
	}

	subscriber.WaitGroup.Add(3)

	bus := eventbus.NewEventBus(startBlockEventCount, dealEventRoutingCount)

	err := bus.RegisterSubscriber(&subscriber, eventLabel1, eventLabel2, eventLabel3)
	if err != nil {
		t.Fatal("bus.RegisterSubscriber", err)
	}

	bus.Publish(eventLabel1, &Event{
		Content: eventContent1,
	})

	bus.Publish(eventLabel2, &Event{
		Content: eventContent2,
	})

	bus.Publish(eventLabel3, &Event{
		Content: eventContent3,
	})

	subscriber.WaitGroup.Wait()

	err = bus.UnRegisterSubscriber(&subscriber, eventLabel1, eventLabel2, eventLabel3)
	if err != nil {
		t.Fatal("bus.UnRegisterSubscriber", err)
	}

	eventbus.DestroyEventBus(bus)
}

func TestUnRegisterAllLabel(t *testing.T) {
	subscriber1 := Subscriber{
		id:        "1",
		t:         t,
		WaitGroup: &sync.WaitGroup{},
	}

	subscriber1.WaitGroup.Add(3)

	subscriber2 := Subscriber{
		id:        "1",
		t:         t,
		WaitGroup: &sync.WaitGroup{},
	}

	subscriber2.WaitGroup.Add(3)

	bus := eventbus.NewEventBus(startBlockEventCount, dealEventRoutingCount)

	err := bus.RegisterSubscriber(&subscriber1, eventLabel1, eventLabel2, eventLabel3)
	if err != nil {
		t.Fatal("subscriber1 bus.RegisterSubscriber", err)
	}

	err = bus.RegisterAllExistLabel(&subscriber2)
	if err != nil {
		t.Fatal("subscriber2 bus.RegisterAllExistLabel", err)
	}

	bus.Publish(eventLabel1, &Event{
		Content: eventContent1,
	})

	bus.Publish(eventLabel2, &Event{
		Content: eventContent2,
	})

	bus.Publish(eventLabel3, &Event{
		Content: eventContent3,
	})

	subscriber1.WaitGroup.Wait()
	subscriber2.WaitGroup.Wait()

	err = bus.UnRegisterAllLabel(&subscriber2)
	if err != nil {
		t.Fatal("subscriber2 bus.UnRegisterAllLabel", err)
	}

	err = bus.UnRegisterAllLabel(&subscriber1)
	if err != nil {
		t.Fatal("subscriber1 bus.UnRegisterAllLabel", err)
	}

	eventbus.DestroyEventBus(bus)
}

func TestPublishToLabels(t *testing.T) {
	subscriber := Subscriber{
		id:        "1",
		t:         t,
		WaitGroup: &sync.WaitGroup{},
	}

	subscriber.WaitGroup.Add(3)

	bus := eventbus.NewEventBus(startBlockEventCount, dealEventRoutingCount)

	err := bus.RegisterSubscriber(&subscriber, eventLabel1, eventLabel2, eventLabel3)
	if err != nil {
		t.Fatal("bus.RegisterSubscriber", err)
	}

	bus.Publish(eventLabel1, &Event{
		Content: eventContent1,
	})

	bus.PublishToLabels(map[string]interface{}{
		eventLabel2: &Event{
			Content: eventContent2,
		},
		eventLabel3: &Event{
			Content: eventContent3,
		},
	})

	subscriber.WaitGroup.Wait()

	err = bus.UnRegisterSubscriber(&subscriber, eventLabel1, eventLabel2, eventLabel3)
	if err != nil {
		t.Fatal("bus.UnRegisterSubscriber", err)
	}

	eventbus.DestroyEventBus(bus)
}

func TestManySubscribers(t *testing.T) {
	bus := eventbus.NewEventBus(startBlockEventCount, dealEventRoutingCount)

	subscribers := make([]*Subscriber, 0)
	for i := 0; i < maxSubscribersCount; i++ {
		subscriber := Subscriber{
			id:        strconv.Itoa(i),
			t:         t,
			WaitGroup: &sync.WaitGroup{},
		}

		subscriber.WaitGroup.Add(3)

		err := bus.RegisterSubscriber(&subscriber, eventLabel1, eventLabel2, eventLabel3)
		if err != nil {
			t.Fatal(subscriber.id, "bus.RegisterSubscriber", err)
		}

		subscribers = append(subscribers, &subscriber)
	}

	bus.PublishToLabels(map[string]interface{}{
		eventLabel1: &Event{
			Content: eventContent1,
		},
		eventLabel2: &Event{
			Content: eventContent2,
		},
		eventLabel3: &Event{
			Content: eventContent3,
		},
	})

	for _, subscriber := range subscribers {
		subscriber.WaitGroup.Wait()

		err := bus.UnRegisterSubscriber(subscriber, eventLabel1, eventLabel2, eventLabel3)
		if err != nil {
			t.Fatal(subscriber.id, "bus.UnRegisterSubscriber", err)
		}
	}

	eventbus.DestroyEventBus(bus)
}
