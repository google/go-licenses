package bus

import "github.com/sulaiman-coder/goeventbus"

var publisher eventbus.Publisher
var active bool

// SetPublisher sets the singleton event bus publisher. This is optional; if no bus is provided, the library will
// behave no differently than if a bus had been provided.
func SetPublisher(p eventbus.Publisher) {
	publisher = p
	if p != nil {
		active = true
	}
}

// Publish an event onto the bus. If there is no bus set by the calling application, this does nothing.
func Publish(event eventbus.Event) {
	if active {
		publisher.Publish(event)
	}
}
