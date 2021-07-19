package test

import (
	"github.com/yjp19871013/event-bus/eventbus"
	"strconv"
	"sync"
	"testing"
)

const (
	dealEventRoutingCount = 5
	startBlockEventCount  = 50
	maxSubscribersCount   = 1000000
)

const (
	eventContent1 = "aaa"
	eventContent2 = "bbb"
	eventContent3 = "ccc"
)

type Event1 struct {
	content string
}

type Event2 struct {
	content string
}

type Event3 struct {
	content string
}

type Subscriber struct {
	id string
	t  *testing.T
	*sync.WaitGroup
}

func (s *Subscriber) OnEvent(event *eventbus.Event) {
	defer s.WaitGroup.Done()

	switch event.PublishTypeName {
	case "Event1":
		e := event.PublishEvent.(*Event1)

		if e.content != eventContent1 {
			s.t.Fatal(s.id, "Event1", "事件数据错误")
		}
	case "Event2":
		e := event.PublishEvent.(*Event2)

		if e.content != eventContent2 {
			s.t.Fatal(s.id, "Event2", "事件数据错误")
		}
	case "Event3":
		e := event.PublishEvent.(*Event3)

		if e.content != eventContent3 {
			s.t.Fatal(s.id, "Event3", "事件数据错误")
		}
	default:
		s.t.Fatal(s.id, "事件label不存在")
	}
}

func TestOnePublisherAndOneSubscriber(t *testing.T) {
	subscriber := Subscriber{
		id:        "1",
		t:         t,
		WaitGroup: &sync.WaitGroup{},
	}

	subscriber.WaitGroup.Add(3)

	eventbus.InitEventBus(startBlockEventCount, dealEventRoutingCount)

	err := eventbus.GetBus().RegisterSubscriber(&subscriber, &Event1{}, &Event2{}, &Event3{})
	if err != nil {
		t.Fatal("bus.RegisterSubscriber", err)
	}

	eventbus.GetBus().Publish(&Event1{
		content: eventContent1,
	})

	eventbus.GetBus().Publish(&Event2{
		content: eventContent2,
	})

	eventbus.GetBus().Publish(&Event3{
		content: eventContent3,
	})

	subscriber.WaitGroup.Wait()

	err = eventbus.GetBus().UnRegisterSubscriber(&subscriber, &Event1{}, &Event2{}, &Event3{})
	if err != nil {
		t.Fatal("bus.UnRegisterSubscriber", err)
	}

	eventbus.DestroyEventBus()
}

func TestUnRegisterAllEvents(t *testing.T) {
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

	eventbus.InitEventBus(startBlockEventCount, dealEventRoutingCount)

	err := eventbus.GetBus().RegisterSubscriber(&subscriber1, &Event1{}, &Event2{}, &Event3{})
	if err != nil {
		t.Fatal("subscriber1 bus.RegisterSubscriber", err)
	}

	err = eventbus.GetBus().RegisterAllExistEvents(&subscriber2)
	if err != nil {
		t.Fatal("subscriber2 bus.RegisterAllExistEvents", err)
	}

	eventbus.GetBus().Publish(&Event1{
		content: eventContent1,
	})

	eventbus.GetBus().Publish(&Event2{
		content: eventContent2,
	})

	eventbus.GetBus().Publish(&Event3{
		content: eventContent3,
	})

	subscriber1.WaitGroup.Wait()
	subscriber2.WaitGroup.Wait()

	err = eventbus.GetBus().UnRegisterAllEvents(&subscriber2)
	if err != nil {
		t.Fatal("subscriber2 bus.UnRegisterAllEvents", err)
	}

	err = eventbus.GetBus().UnRegisterAllEvents(&subscriber1)
	if err != nil {
		t.Fatal("subscriber1 bus.UnRegisterAllEvents", err)
	}

	eventbus.DestroyEventBus()
}

func TestPublishToEvents(t *testing.T) {
	subscriber := Subscriber{
		id:        "1",
		t:         t,
		WaitGroup: &sync.WaitGroup{},
	}

	subscriber.WaitGroup.Add(3)

	eventbus.InitEventBus(startBlockEventCount, dealEventRoutingCount)

	err := eventbus.GetBus().RegisterSubscriber(&subscriber, &Event1{}, &Event2{}, &Event3{})
	if err != nil {
		t.Fatal("bus.RegisterSubscriber", err)
	}

	eventbus.GetBus().Publish(
		&Event1{
			content: eventContent1,
		},
		&Event2{
			content: eventContent2,
		},
		&Event3{
			content: eventContent3,
		},
	)

	subscriber.WaitGroup.Wait()

	err = eventbus.GetBus().UnRegisterSubscriber(&subscriber, &Event1{}, &Event2{}, &Event3{})
	if err != nil {
		t.Fatal("bus.UnRegisterSubscriber", err)
	}

	eventbus.DestroyEventBus()
}

func TestManySubscribers(t *testing.T) {
	eventbus.InitEventBus(startBlockEventCount, dealEventRoutingCount)

	subscribers := make([]*Subscriber, 0)
	for i := 0; i < maxSubscribersCount; i++ {
		subscriber := Subscriber{
			id:        strconv.Itoa(i),
			t:         t,
			WaitGroup: &sync.WaitGroup{},
		}

		subscriber.WaitGroup.Add(3)

		err := eventbus.GetBus().RegisterSubscriber(&subscriber, &Event1{}, &Event2{}, &Event3{})
		if err != nil {
			t.Fatal(subscriber.id, "bus.RegisterSubscriber", err)
		}

		subscribers = append(subscribers, &subscriber)
	}

	eventbus.GetBus().Publish(
		&Event1{
			content: eventContent1,
		},
		&Event2{
			content: eventContent2,
		},
		&Event3{
			content: eventContent3,
		},
	)

	for _, subscriber := range subscribers {
		subscriber.WaitGroup.Wait()

		err := eventbus.GetBus().UnRegisterSubscriber(subscriber, &Event1{}, &Event2{}, &Event3{})
		if err != nil {
			t.Fatal(subscriber.id, "bus.UnRegisterSubscriber", err)
		}
	}

	eventbus.DestroyEventBus()
}
