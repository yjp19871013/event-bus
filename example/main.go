package main

import (
	"fmt"
	DEATH "github.com/vrecan/death"
	"github.com/yjp19871013/event-bus/eventbus"
	"log"
	"syscall"
	"time"
)

const (
	beatHeartEventLabel = "receive beat heart"
	tickerDurationSec   = 3
)

type BeatHeartEvent struct {
	Time string
}

type Subscriber struct {
}

func (s *Subscriber) OnEvent(label string, event interface{}) {
	if label == beatHeartEventLabel {
		e, ok := event.(*BeatHeartEvent)
		if !ok {
			return
		}

		fmt.Println("receive: ", e.Time)
	}
}

func main() {
	subscriber := &Subscriber{}

	bus := eventbus.NewEventBus(10, 5)
	err := bus.RegisterSubscriber(subscriber, beatHeartEventLabel)
	if err != nil {
		panic(err)
	}

	done := make(chan interface{})
	go beatHeart(bus, done)

	death := DEATH.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	_ = death.WaitForDeath()
	log.Println("Shutdown Server ...")

	close(done)
}

func beatHeart(bus *eventbus.EventBus, done chan interface{}) {
	ticker := time.NewTicker(tickerDurationSec * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			bus.Publish(beatHeartEventLabel, &BeatHeartEvent{Time: time.Now().Format("2006-01-02 15:04:05")})
		}
	}
}
