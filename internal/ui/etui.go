package ui

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sulaiman-coder/gobouncer/bouncer/event/parser"

	"github.com/sulaiman-coder/goprogress"

	"github.com/mitchellh/go-glint"
	gc "github.com/mitchellh/go-glint/components"
	"github.com/sulaiman-coder/gobouncer/bouncer/event"
	"github.com/sulaiman-coder/goeventbus"
)

type etuiStatus struct {
	progress progress.StagedProgressable
	lock     *sync.Mutex
}

func (e *etuiStatus) update(p progress.StagedProgressable) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.progress = p
	return nil
}

func (e *etuiStatus) render(rows, cols uint) string {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.progress != nil {
		return e.progress.Stage()
	}
	return "starting..."
}

func etui(workerErrs <-chan error, subscription *partybus.Subscription) error {
	status := &etuiStatus{lock: &sync.Mutex{}}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		events := subscription.Events()

	eventLoop:
		for {
			select {
			case err := <-workerErrs:
				// TODO: we should show errors more explicitly in the ETUI
				if err != nil {
					panic(err)
				}
			case e, ok := <-events:
				if !ok {
					break eventLoop
				}
				switch e.Type {
				case event.ModuleScanStarted:
					p, err := parser.ParseModuleScanStarted(e)
					if err != nil {
						panic(err)
					}
					if err := status.update(p); err != nil {
						if err != nil {
							fmt.Printf("could not update (%+v)\n", err)
						}
						break eventLoop
					}
				case event.ModuleScanResult:
					cancel() // cancel UI
					p, err := parser.ParseModuleScanResult(e)
					if err != nil {
						panic(err)
					}
					if err = p.Present(os.Stdout); err != nil {
						panic(err)
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
	}()

	d := glint.New()
	d.Append(
		glint.Style(
			glint.Layout(
				gc.Spinner(),
				glint.Layout(glint.TextFunc(status.render)).MarginLeft(1),
			).Row(),
			glint.Color("green"),
		),
	)
	d.Render(ctx)
	return nil
}
