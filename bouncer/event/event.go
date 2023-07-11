package event

import eventbus "github.com/sulaiman-coder/goeventbus"

const (
	ModuleScanStarted eventbus.EventType = "bouncer-module-scan-started"
	ModuleScanResult  eventbus.EventType = "bouncer-module-scan-result"
)
