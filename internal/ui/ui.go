package ui

import (
	"github.com/sulaiman-coder/goeventbus"
)

type UI func(<-chan error, *eventbus.Subscription) error

// Select is responsible for determining the specific UI function given select user option, the current platform
// config values, and environment status (such as a TTY being present).
func Select() UI {
	// verbose, quiet bool
	var ui UI

	//isStdoutATty := terminal.IsTerminal(int(os.Stdout.Fd()))
	//isStderrATty := terminal.IsTerminal(int(os.Stderr.Fd()))
	//notATerminal := !isStderrATty && !isStdoutATty

	switch {
	//case runtime.GOOS == "windows" || verbose || quiet || notATerminal || !isStderrATty:
	//	ui = logger
	default:
		ui = etui
		//ui = nop
	}

	return ui
}
