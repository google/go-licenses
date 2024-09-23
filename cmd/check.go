package cmd

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/khulnasoft/go-licenses/golicenses"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "ensure only select licenses or types are used",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := doCheckCmd(cmd, args)
		if err != nil {
			color.Style{color.Red, color.Bold}.Println(err.Error())
			os.Exit(1)
		}
		color.Style{color.Green, color.Bold}.Println("Passed!")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

// TODO: add to check the ability to check for 3rd party notices are in the repo

func doCheckCmd(_ *cobra.Command, args []string) error {
	var rules golicenses.Rules
	var err error
	switch {
	case len(appConfig.Permit) > 0:
		rules, err = golicenses.NewRules(golicenses.AllowAction, appConfig.Permit, appConfig.IgnorePkg...)
		fmt.Printf("Allow Rules: %+v\n", appConfig.Permit)
	case len(appConfig.Forbid) > 0:
		rules, err = golicenses.NewRules(golicenses.DenyAction, appConfig.Forbid, appConfig.IgnorePkg...)
		fmt.Printf("Deny Rules: %+v\n", appConfig.Forbid)
	default:
		return fmt.Errorf("no rules configured")
	}
	if err != nil {
		return fmt.Errorf("could not parse rules: %+v", err)
	}

	licenseFinder := golicenses.NewLicenseFinder("", 0.9)

	resultStream, err := licenseFinder.Find()
	if err != nil {
		return err
	}

	failed := false
	for _, result := range resultStream {
		allowable, _, err := rules.Evaluate(result)
		if err != nil {
			return err
		}

		if !allowable {
			failed = true
			fmt.Printf("Unallowable license (%s) from %q\n", result.License, result.ModulePath)
		}
	}

	if failed {
		return fmt.Errorf("failed validation")
	}

	return nil
}
