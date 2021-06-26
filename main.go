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
	"flag"
	"github.com/google/go-licenses/licenses"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "licenses",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setupArchivePath()
		},
	}

	// Flags shared between subcommands
	confidenceThreshold float64
	licenseArchivePath string
)

func init() {
	rootCmd.PersistentFlags().Float64Var(&confidenceThreshold, "confidence_threshold", 0.9, "Minimum confidence required in order to positively identify a license.")
	rootCmd.PersistentFlags().StringVar(&licenseArchivePath, "license_db", "", "Path to the license archive (uses archive in source checkout if not specificed)")
}

func main() {
	flag.Parse()
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	if err := rootCmd.Execute(); err != nil {
		glog.Exit(err)
	}
}

// Unvendor removes the "*/vendor/" prefix from the given import path, if present.
func unvendor(importPath string) string {
	if vendorerAndVendoree := strings.SplitN(importPath, "/vendor/", 2); len(vendorerAndVendoree) == 2 {
		return vendorerAndVendoree[1]
	}
	return importPath
}

// setupArchivePath sets the path to the license archive if specified
func setupArchivePath() {
	if licenseArchivePath == "" {
		return
	}

	if err := licenses.SetArchiveLocation(licenseArchivePath); err != nil {
		glog.Exit(err)
	}
}
