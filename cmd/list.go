package cmd

import (
	"fmt"
	"os"

	"github.com/khulnasoft/go-licenses/golicenses/event"
	"github.com/khulnasoft/go-licenses/internal/bus"
	"github.com/khulnasoft/go-pulsebus"

	"github.com/khulnasoft/go-licenses/internal/ui"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/khulnasoft/go-licenses/golicenses/presenter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered licenses for a project (including dependencies)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := doListCmd(cmd, args)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func doListCmd(_ *cobra.Command, args []string) error {
	errs := startListWorker()
	ux := ui.Select()
	return ux(errs, eventSubscription)
}

func startListWorker() <-chan error {
	errs := make(chan error)
	go func() {
		defer close(errs)
		licenseFinder := golicenses.NewLicenseFinder("", 0.9)

		resultStream, err := licenseFinder.Find()
		if err != nil {
			errs <- err
			return
		}

		bus.Publish(pulsebus.Event{
			Type:  event.ModuleScanResult,
			Value: presenter.GetPresenter(appConfig.PresenterOpt, resultStream),
		})

	}()
	return errs
}
