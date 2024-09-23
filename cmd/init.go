package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/khulnasoft/go-licenses/golicenses/presenter"
	"github.com/khulnasoft/go-licenses/internal/config"
)

var appConfig *config.Application
var configPath string

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
