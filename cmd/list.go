package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/khulnasoft/go-bouncer/bouncer"
	"github.com/khulnasoft/go-bouncer/bouncer/presenter"
)

var gitRemotes []string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered licenses for a project (including dependencies)",
	Run: func(cmd *cobra.Command, args []string) {
		err := doListCmd(cmd, args)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	listCmd.Flags().StringArrayVar(&gitRemotes, "git-remote", []string{"origin", "upstream"}, "Remote Git repositories to try")

	rootCmd.AddCommand(listCmd)
}

func doListCmd(_ *cobra.Command, args []string) error {
	var paths []string
	if len(args) > 0 {
		paths = args
	} else {
		paths = []string{"."}
	}
	licenseFinder := bouncer.NewLicenseFinder(paths, gitRemotes, 0.9)

	resultStream, err := licenseFinder.Find()
	if err != nil {
		return err
	}

	pres := presenter.GetPresenter(appConfig.PresenterOpt, resultStream)
	return pres.Present(os.Stdout)
}