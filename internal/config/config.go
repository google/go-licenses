package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/khulnasoft/go-bouncer/bouncer/presenter"
	"github.com/khulnasoft/go-bouncer/internal"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Application struct {
	ConfigPath   string
	PresenterOpt presenter.Option
	Output       string      `mapstructure:"output"`
	Verbose      int         `mapstructure:"verbose"`
	Forbid       StringArray `mapstructure:"forbid,deny"`
	Permit       StringArray `mapstructure:"permit,allow"`
	IgnorePkg    StringArray `mapstructure:"ignore-packages"`
}

type StringArray []string

func (a *StringArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}

func setNonCliDefaultValues(v *viper.Viper) {
	// TODO
}

func LoadConfigFromFile(v *viper.Viper, configPath string) (*Application, error) {
	// the user may not have a config, and this is OK, we can use the default config + default cobra cli values instead
	setNonCliDefaultValues(v)
	if configPath != "" {
		_ = readConfig(v, configPath)
	} else {
		_ = readConfig(v, "")
	}

	config := &Application{
		ConfigPath: configPath,
	}
	err := v.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}
	config.ConfigPath = v.ConfigFileUsed()

	err = config.Build()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

func (cfg *Application) Build() error {
	// validate rule input
	if len(cfg.Forbid) > 0 && len(cfg.Permit) > 0 {
		return fmt.Errorf("'forbid'/'deny' and 'permit'/'allow' options are mutually exclusive")
	}

	// set the presenter
	presenterOption := presenter.ParseOption(cfg.Output)
	if presenterOption == presenter.UnknownPresenter {
		return fmt.Errorf("bad --output value '%s'", cfg.Output)
	}
	cfg.PresenterOpt = presenterOption

	return nil
}

func readConfig(v *viper.Viper, configPath string) error {
	v.AutomaticEnv()
	v.SetEnvPrefix(internal.ApplicationName)
	// allow for nested options to be specified via environment variables
	// e.g. pod.context = APPNAME_POD_CONTEXT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// use explicitly the given user config
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err == nil {
			return nil
		}
		// don't fall through to other options if this fails
		return fmt.Errorf("unable to read config: %v", configPath)
	}

	// start searching for valid configs in order...

	// 1. look for .<appname>.yaml (in the current directory)
	v.AddConfigPath(".")
	v.SetConfigName(internal.ApplicationName)
	if err := v.ReadInConfig(); err == nil {
		return nil
	}

	// 2. look for .<appname>/config.yaml (in the current directory)
	v.AddConfigPath("." + internal.ApplicationName)
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err == nil {
		return nil
	}

	// 3. look for ~/.<appname>.yaml
	home, err := homedir.Dir()
	if err == nil {
		v.AddConfigPath(home)
		v.SetConfigName("." + internal.ApplicationName)
		if err := v.ReadInConfig(); err == nil {
			return nil
		}
	}

	// 4. look for <appname>/config.yaml in xdg locations (starting with xdg home config dir, then moving upwards)
	v.AddConfigPath(path.Join(xdg.ConfigHome, internal.ApplicationName))
	for _, dir := range xdg.ConfigDirs {
		v.AddConfigPath(path.Join(dir, internal.ApplicationName))
	}
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err == nil {
		return nil
	}

	return fmt.Errorf("application config not found")
}