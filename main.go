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
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "go-licenses",
		Short: "go-licenses helps you work with licenses of your go project's dependencies.",
		Long: `go-licenses helps you work with licenses of your go project's dependencies.

Prerequisites:
1. Go v1.16 or later.
2. Change directory to your go project.
3. Run "go mod download".`,
	}

	// Flags shared between subcommands
	confidenceThreshold float64
	ignore              []string
	packageHelp         = `

Typically, specify the Go package that builds your Go binary.
go-licenses expects the same package argument format as "go build".
For example:
* A rooted import path like "github.com/google/go-licenses" or "github.com/google/go-licenses/licenses".
* A relative path that denotes the package in that directory, like "." or "./cmd/some-command".
To learn more about Go package argument, run "go help packages".`
)

func init() {
	// Change glog default log level to INFO.
	// Note glog is not initialized yet, so we can only use fmt for printing
	// errors.
	err := flag.Set("logtostderr", "true")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = flag.Set("stderrthreshold", "INFO")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().Float64Var(&confidenceThreshold, "confidence_threshold", 0.9, "Minimum confidence required in order to positively identify a license.")
	rootCmd.PersistentFlags().StringSliceVar(&ignore, "ignore", nil, "Package path prefixes to be ignored. Dependencies from the ignored packages are still checked. Can be specified multiple times.")
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
