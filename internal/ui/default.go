package ui

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sulaiman-coder/gobouncer/bouncer/event"
	"github.com/sulaiman-coder/gobouncer/bouncer/event/parser"
	"github.com/sulaiman-coder/goeventbus"
)

func nop(workerErrs <-chan error, subscription *eventbus.Subscription) error {
	status := &etuiStatus{lock: &sync.Mutex{}}
	ctx := context.Background()
	events := subscription.Events()

eventLoop:
	for {
		select {
		case err := <-workerErrs:
			if err != nil {
				return err
			}
		case e, ok := <-events:
			if !ok {
				break eventLoop
			}
			switch e.Type {
			case event.ModuleScanStarted:
				p, err := parser.ParseModuleScanStarted(e)
				if err != nil {
					return err
				}
				if err := status.update(p); err != nil {
					if err != nil {
						return fmt.Errorf("could not update status: %w", err)
					}
					break eventLoop
				}
			case event.ModuleScanResult:
				p, err := parser.ParseModuleScanResult(e)
				if err != nil {
					return err
				}
				if err = p.Present(os.Stdout); err != nil {
					return err
				}
				break eventLoop
			}
		case <-ctx.Done():
			if ctx.Err() != nil {
				fmt.Printf("cancelled (%+v)\n", ctx.Err())
			}
			break eventLoop
		}
	}
	return nil
}
