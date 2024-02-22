// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/google/go-licenses/report"
	"github.com/spf13/cobra"
)

var (
	reportHelp = "Prints report of all licenses that apply to one or more Go packages and their dependencies."
	reportCmd  = &cobra.Command{
		Use:   "report <package> [package...]",
		Short: reportHelp,
		Long:  reportHelp + packageHelp,
		Args:  cobra.MinimumNArgs(1),
		RunE:  templateMain,
	}

	templateFile string
)

func init() {
	reportCmd.Flags().StringVar(&templateFile, "template", "", "Custom Go template file to use for report")

	rootCmd.AddCommand(reportCmd)
}

func templateMain(_ *cobra.Command, args []string) error {
	libraryData, err := report.Report(includeTests, ignore, args)
	if err != nil {
		return err
	}

	return report.ReportTemplate(os.Stdout, libraryData, templateFile)
}
