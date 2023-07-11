package bouncer

import (
	"github.com/sulaiman-coder/gobouncer/internal/bus"
	"github.com/sulaiman-coder/goeventbus"
)

func SetBus(b *eventbus.Bus) {
	bus.SetPublisher(b)
}
