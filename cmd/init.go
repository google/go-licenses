package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sulaiman-coder/gobouncer/bouncer"
	"github.com/sulaiman-coder/gobouncer/bouncer/presenter"
	"github.com/sulaiman-coder/gobouncer/internal/config"
	eventbus "github.com/sulaiman-coder/goeventbus"
)

var appConfig *config.Application
var eventBus *eventbus.Bus
var eventSubscription *eventbus.Subscription
var configPath string

func init() {
	setCliOptions()

	cobra.OnInitialize(
		initAppConfig,
		initEventBus,
	)
}

func setCliOptions() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "application config file")

	flag := "output"
	rootCmd.PersistentFlags().StringP(
		flag, "o", presenter.TextPresenter.String(),
		fmt.Sprintf("report output formatter, options=%v", presenter.Options),
	)
	if err := viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag)); err != nil {
		fmt.Printf("unable to bind flag '%s': %+v", flag, err)
		os.Exit(1)
	}

	flag = "verbose"
	rootCmd.PersistentFlags().CountP(
		flag, "v",
		"increase verbosity (-v = info, -vv = debug)",
	)
	if err := viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag)); err != nil {
		fmt.Printf("unable to bind flag '%s': %+v", flag, err)
		os.Exit(1)
	}
}

func initAppConfig() {
	cfg, err := config.LoadConfigFromFile(viper.GetViper(), configPath)
	if err != nil {
		fmt.Printf("failed to load application config: \n\t%+v\n", err)
		os.Exit(1)
	}
	appConfig = cfg
}

func initEventBus() {
	eventBus = eventbus.NewBus()
	eventSubscription = eventBus.Subscribe()

	bouncer.SetBus(eventBus)
}
