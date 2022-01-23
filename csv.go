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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-licenses/licenses"
	"github.com/spf13/cobra"
)

var (
	csvCmd = &cobra.Command{
		Use:   "csv <package>",
		Short: "Prints all licenses that apply to a Go package and its dependencies",
		Args:  cobra.MinimumNArgs(1),
		RunE:  csvMain,
	}

	gitRemotes []string
)

func init() {
	csvCmd.Flags().StringArrayVar(&gitRemotes, "git_remote", []string{"origin", "upstream"}, "Remote Git repositories to try")

	rootCmd.AddCommand(csvCmd)
}

func csvMain(_ *cobra.Command, args []string) error {
	classifier, err := licenses.NewClassifier(confidenceThreshold)
	if err != nil {
		return err
	}

	libs, err := licenses.Libraries(context.Background(), classifier, args...)
	if err != nil {
		return err
	}
	for _, lib := range libs {
		licenseName := "Unknown"
		licenseURL := "Unknown"
		if lib.LicensePath != "" {
			name, _, err := classifier.Identify(lib.LicensePath)
			if err == nil {
				licenseName = name
			} else {
				glog.Errorf("Error identifying license in %q: %v", lib.LicensePath, err)
			}
			url, err := lib.FileURL(context.Background(), lib.LicensePath)
			if err == nil {
				licenseURL = url
			} else {
				glog.Warningf("Error discovering license URL: %s", err)
			}
		}
		// Using ", " to join words makes vscode/terminal recognize the
		// correct license URL. Otherwise, if there's no space after
		// comma, vscode interprets the URL as concatenated with the
		// license name after it.
		// Also, the extra spaces does not affect csv syntax much, we
		// can still copy the csv text and paste into Excel / Google
		// Sheets.
		if _, err := fmt.Fprintln(os.Stdout, strings.Join([]string{lib.Name(), licenseURL, licenseName}, ", ")); err != nil {
			return err
		}
	}
	return nil
}
