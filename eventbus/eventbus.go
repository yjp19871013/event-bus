package eventbus

import (
	"math/rand"
	"sync"
	"time"
)

type eventQueueItem struct {
	label      string
	subscriber Subscriber
	event      interface{}
}

type Subscriber interface {
	OnEvent(label string, event interface{})
}

type EventBus struct {
	sync.Mutex

	dealEventChannels []chan *eventQueueItem
	dealEventDoneChan chan interface{}

	labelMapMutex sync.Mutex
	labelMap      map[string]map[Subscriber]interface{}
}

var busInstance *EventBus

func GetBus() *EventBus {
	return busInstance
}

func InitEventBus(startBlockEventCount uint, dealEventRoutingCount uint8) {
	busInstance = new(EventBus)
	busInstance.labelMap = make(map[string]map[Subscriber]interface{})
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
	if busInstance == nil {
		return
	}

	busInstance.Lock()
	defer busInstance.Unlock()

	close(busInstance.dealEventDoneChan)
	busInstance.dealEventDoneChan = nil

	for _, dealEventChan := range busInstance.dealEventChannels {
		close(dealEventChan)
		dealEventChan = nil
	}

	busInstance.dealEventChannels = make([]chan *eventQueueItem, 0)

	busInstance.destroyLabelMap()

	busInstance = nil
}

func (bus *EventBus) RegisterSubscriber(subscriber Subscriber, labels ...string) error {
	if subscriber == nil || labels == nil || len(labels) == 0 {
		return ErrParam
	}

	bus.Lock()
	defer bus.Unlock()

	bus.addLabels(labels, subscriber)

	return nil
}

func (bus *EventBus) UnRegisterSubscriber(subscriber Subscriber, labels ...string) error {
	if subscriber == nil || labels == nil || len(labels) == 0 {
		return ErrParam
	}

	bus.Lock()
	defer bus.Unlock()

	bus.deleteLabels(labels, subscriber)

	return nil
}

func (bus *EventBus) RegisterAllExistLabel(subscriber Subscriber) error {
	if subscriber == nil {
		return ErrParam
	}

	bus.Lock()
	defer bus.Unlock()

	bus.addToAllLabels(subscriber)

	return nil
}

func (bus *EventBus) UnRegisterAllLabel(subscriber Subscriber) error {
	if subscriber == nil {
		return ErrParam
	}

	bus.Lock()
	defer bus.Unlock()

	bus.deleteFromAllLabels(subscriber)

	return nil
}

func (bus *EventBus) Publish(label string, event interface{}) {
	bus.Lock()
	defer bus.Unlock()

	subscribers := bus.getLabelSubscribers(label)
	for subscriber := range subscribers {
		item := &eventQueueItem{
			label:      label,
			subscriber: subscriber,
			event:      event,
		}

		bus.selectDealEventChan() <- item
	}
}

func (bus *EventBus) PublishToLabels(labelEventMap map[string]interface{}) {
	bus.Lock()
	defer bus.Unlock()

	for label, event := range labelEventMap {
		subscribers := bus.getLabelSubscribers(label)
		for subscriber := range subscribers {
			item := &eventQueueItem{
				label:      label,
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

			go item.subscriber.OnEvent(item.label, item.event)
		}
	}
}

func (bus *EventBus) selectDealEventChan() chan *eventQueueItem {
	return bus.dealEventChannels[rand.Intn(len(bus.dealEventChannels))]
}

func (bus *EventBus) destroyLabelMap() {
	bus.labelMapMutex.Lock()
	defer bus.labelMapMutex.Unlock()

	bus.labelMap = make(map[string]map[Subscriber]interface{})
}

func (bus *EventBus) addLabels(labels []string, subscriber Subscriber) {
	bus.labelMapMutex.Lock()
	defer bus.labelMapMutex.Unlock()

	if bus.labelMap == nil {
		return
	}

	for _, label := range labels {
		subscribers, ok := bus.labelMap[label]
		if !ok {
			subscribers = make(map[Subscriber]interface{}, 0)
		}

		subscribers[subscriber] = nil

		bus.labelMap[label] = subscribers
	}
}

func (bus *EventBus) deleteLabels(labels []string, subscriber Subscriber) {
	bus.labelMapMutex.Lock()
	defer bus.labelMapMutex.Unlock()

	if bus.labelMap == nil || len(bus.labelMap) == 0 {
		return
	}

	for _, label := range labels {
		subscribers, ok := bus.labelMap[label]
		if !ok {
			continue
		}

		bus.deleteLabelWithoutLock(subscribers, label, subscriber)
	}
}

func (bus *EventBus) addToAllLabels(subscriber Subscriber) {
	bus.labelMapMutex.Lock()
	defer bus.labelMapMutex.Unlock()

	if bus.labelMap == nil || len(bus.labelMap) == 0 {
		return
	}

	for label, subscribers := range bus.labelMap {
		subscribers[subscriber] = nil
		bus.labelMap[label] = subscribers
	}
}

func (bus *EventBus) deleteFromAllLabels(subscriber Subscriber) {
	bus.labelMapMutex.Lock()
	defer bus.labelMapMutex.Unlock()

	if bus.labelMap == nil || len(bus.labelMap) == 0 {
		return
	}

	for label, subscribers := range bus.labelMap {
		bus.deleteLabelWithoutLock(subscribers, label, subscriber)
	}
}

func (bus *EventBus) deleteLabelWithoutLock(subscribers map[Subscriber]interface{}, label string, subscriber Subscriber) {
	_, ok := subscribers[subscriber]
	if !ok {
		return
	}

	delete(subscribers, subscriber)

	if len(subscribers) == 0 {
		delete(bus.labelMap, label)
	} else {
		bus.labelMap[label] = subscribers
	}
}

func (bus *EventBus) getLabelSubscribers(label string) map[Subscriber]interface{} {
	bus.labelMapMutex.Lock()
	defer bus.labelMapMutex.Unlock()

	if bus.labelMap == nil || len(bus.labelMap) == 0 {
		return make(map[Subscriber]interface{})
	}

	subscribers, ok := bus.labelMap[label]
	if !ok {
		return make(map[Subscriber]interface{})
	}

	return subscribers
}
