package golicenses

import (
	"github.com/khulnasoft/go-licenses/internal/bus"
	"github.com/khulnasoft/go-pulsebus"
)

func SetBus(b *pulsebus.Bus) {
	bus.SetPublisher(b)
}
