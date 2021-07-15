package main

import (
	"github.com/yjp19871013/event-bus/eventbus"
	DEATH "gopkg.in/vrecan/death.v3"
	"syscall"
)

func main() {
	bus := eventbus.NewEventBus()

	death := DEATH.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	_ = death.WaitForDeath()
}
