package eventbus

import (
	"math/rand"
	"reflect"
	"sync"
	"time"
)

type eventQueueItem struct {
	eventType  string
	subscriber Subscriber
	event      interface{}
}

type Event struct {
	PublishTypeName string
	PublishEvent    interface{}
}

type Subscriber interface {
	OnEvent(event *Event)
}

type EventBus struct {
	dealEventChannels []chan *eventQueueItem
	dealEventDoneChan chan interface{}

	eventMapMutex sync.Mutex
	eventMap      map[string]map[Subscriber]interface{}
}

var busInstance *EventBus
var busInstanceMutex sync.Mutex

func GetBus() *EventBus {
	busInstanceMutex.Lock()
	busInstanceMutex.Unlock()

	return busInstance
}

func InitEventBus(startBlockEventCount uint, dealEventRoutingCount uint8) {
	busInstanceMutex.Lock()
	defer busInstanceMutex.Unlock()

	if busInstance != nil {
		panic("repeated call of InitEventBus")
	}

	busInstance = new(EventBus)
	busInstance.eventMap = make(map[string]map[Subscriber]interface{})
	busInstance.dealEventChannels = make([]chan *eventQueueItem, 0)
	busInstance.dealEventDoneChan = make(chan interface{})

	rand.Seed(time.Now().UnixNano())

	if dealEventRoutingCount == 0 {
		dealEventRoutingCount = 1
	}

	for i := 0; i < int(dealEventRoutingCount); i++ {
		dealEventChan := make(chan *eventQueueItem, startBlockEventCount)
		busInstance.dealEventChannels = append(busInstance.dealEventChannels, dealEventChan)

		go busInstance.doDealEvent(dealEventChan)
	}
}

func DestroyEventBus() {
	busInstanceMutex.Lock()
	defer busInstanceMutex.Unlock()

	if busInstance == nil {
		return
	}

	close(busInstance.dealEventDoneChan)
	busInstance.dealEventDoneChan = nil

	for _, dealEventChan := range busInstance.dealEventChannels {
		close(dealEventChan)
		dealEventChan = nil
	}

	busInstance.dealEventChannels = make([]chan *eventQueueItem, 0)

	busInstance.destroyEventMap()

	busInstance = nil
}

func (bus *EventBus) RegisterSubscriber(subscriber Subscriber, events ...interface{}) error {
	if subscriber == nil || events == nil || len(events) == 0 {
		return ErrParam
	}

	bus.addEvents(events, subscriber)

	return nil
}

func (bus *EventBus) UnRegisterSubscriber(subscriber Subscriber, events ...interface{}) error {
	if subscriber == nil || events == nil || len(events) == 0 {
		return ErrParam
	}

	bus.deleteEvents(events, subscriber)

	return nil
}

func (bus *EventBus) RegisterAllExistEvents(subscriber Subscriber) error {
	if subscriber == nil {
		return ErrParam
	}

	bus.addToAllEvents(subscriber)

	return nil
}

func (bus *EventBus) UnRegisterAllEvents(subscriber Subscriber) error {
	if subscriber == nil {
		return ErrParam
	}

	bus.deleteFromAllEvents(subscriber)

	return nil
}

func (bus *EventBus) Publish(events ...interface{}) {
	for _, event := range events {
		t := reflect.TypeOf(event).Elem()

		subscribers := bus.getEventSubscribers(t.Name())
		for subscriber := range subscribers {
			item := &eventQueueItem{
				eventType:  t.Name(),
				subscriber: subscriber,
				event:      event,
			}

			bus.selectDealEventChan() <- item
		}
	}
}

func (bus *EventBus) doDealEvent(dealEventChan chan *eventQueueItem) {
	for {
		select {
		case <-bus.dealEventDoneChan:
			return
		case item := <-dealEventChan:
			if item == nil {
				continue
			}

			go item.subscriber.OnEvent(&Event{
				PublishTypeName: item.eventType,
				PublishEvent:    item.event,
			})
		}
	}
}

func (bus *EventBus) selectDealEventChan() chan *eventQueueItem {
	return bus.dealEventChannels[rand.Intn(len(bus.dealEventChannels))]
}

func (bus *EventBus) destroyEventMap() {
	bus.eventMapMutex.Lock()
	defer bus.eventMapMutex.Unlock()

	bus.eventMap = make(map[string]map[Subscriber]interface{})
}

func (bus *EventBus) addEvents(events []interface{}, subscriber Subscriber) {
	bus.eventMapMutex.Lock()
	defer bus.eventMapMutex.Unlock()

	if bus.eventMap == nil {
		return
	}

	for _, event := range events {
		t := reflect.TypeOf(event).Elem()

		subscribers, ok := bus.eventMap[t.Name()]
		if !ok {
			subscribers = make(map[Subscriber]interface{}, 0)
		}

		subscribers[subscriber] = nil

		bus.eventMap[t.Name()] = subscribers
	}
}

func (bus *EventBus) deleteEvents(events []interface{}, subscriber Subscriber) {
	bus.eventMapMutex.Lock()
	defer bus.eventMapMutex.Unlock()

	if bus.eventMap == nil || len(bus.eventMap) == 0 {
		return
	}

	for _, event := range events {
		t := reflect.TypeOf(event).Elem()

		subscribers, ok := bus.eventMap[t.Name()]
		if !ok {
			continue
		}

		bus.deleteEventWithoutLock(subscribers, t.Name(), subscriber)
	}
}

func (bus *EventBus) addToAllEvents(subscriber Subscriber) {
	bus.eventMapMutex.Lock()
	defer bus.eventMapMutex.Unlock()

	if bus.eventMap == nil || len(bus.eventMap) == 0 {
		return
	}

	for eventType, subscribers := range bus.eventMap {
		subscribers[subscriber] = nil
		bus.eventMap[eventType] = subscribers
	}
}

func (bus *EventBus) deleteFromAllEvents(subscriber Subscriber) {
	bus.eventMapMutex.Lock()
	defer bus.eventMapMutex.Unlock()

	if bus.eventMap == nil || len(bus.eventMap) == 0 {
		return
	}

	for eventType, subscribers := range bus.eventMap {
		bus.deleteEventWithoutLock(subscribers, eventType, subscriber)
	}
}

func (bus *EventBus) deleteEventWithoutLock(subscribers map[Subscriber]interface{},
	eventType string, subscriber Subscriber) {
	_, ok := subscribers[subscriber]
	if !ok {
		return
	}

	delete(subscribers, subscriber)

	if len(subscribers) == 0 {
		delete(bus.eventMap, eventType)
	} else {
		bus.eventMap[eventType] = subscribers
	}
}

func (bus *EventBus) getEventSubscribers(eventType string) map[Subscriber]interface{} {
	bus.eventMapMutex.Lock()
	defer bus.eventMapMutex.Unlock()

	if bus.eventMap == nil || len(bus.eventMap) == 0 {
		return make(map[Subscriber]interface{})
	}

	subscribers, ok := bus.eventMap[eventType]
	if !ok {
		return make(map[Subscriber]interface{})
	}

	return subscribers
}
