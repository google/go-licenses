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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-licenses/licenses"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	checkHelp = "Checks whether licenses for a package are not Forbidden."
	checkCmd  = &cobra.Command{
		Use:   "check <package> [package...]",
		Short: checkHelp,
		Long:  checkHelp + packageHelp,
		Args:  cobra.MinimumNArgs(1),
		RunE:  checkMain,
	}

	allowedLicenseNames []string

	excludeForbidden    bool
	excludeNotice       bool
	excludePermissive   bool
	excludeReciprocal   bool
	excludeRestricted   bool
	excludeUnencumbered bool
)

func init() {
	checkCmd.Flags().StringSliceVar(&allowedLicenseNames, "allowed_license_names", []string{}, "list of allowed license names")

	checkCmd.Flags().BoolVarP(&excludeForbidden, "exclude_forbidden", "", true, "exclude forbidden licenses")
	checkCmd.Flags().BoolVarP(&excludeNotice, "exclude_notice", "", false, "exclude notice licenses")
	checkCmd.Flags().BoolVarP(&excludePermissive, "exclude_permissive", "", false, "exclude permissive licenses")
	checkCmd.Flags().BoolVarP(&excludeReciprocal, "exclude_reciprocal", "", false, "exclude reciprocal licenses")
	checkCmd.Flags().BoolVarP(&excludeRestricted, "exclude_restricted", "", false, "exclude restricted licenses")
	checkCmd.Flags().BoolVarP(&excludeUnencumbered, "exclude_unencumbered", "", false, "exclude unencumbered licenses")

	rootCmd.AddCommand(checkCmd)
}

func checkMain(_ *cobra.Command, args []string) error {
	excludedLicenseTypes := make([]licenses.Type, 0)

	checkLicenseName, checkLicenseType := true, false
	allowedLicenseNames := getAllowedLicenseNames()

	if len(allowedLicenseNames) == 0 {
		checkLicenseName, checkLicenseType = false, true
		excludedLicenseTypes = getExcludedLicenseTypes()

		if len(excludedLicenseTypes) == 0 {
			return errors.New("nothing configured to check")
		}
	}

	classifier, err := licenses.NewClassifier(confidenceThreshold)
	if err != nil {
		return err
	}

	libs, err := licenses.Libraries(context.Background(), classifier, ignore, args...)
	if err != nil {
		return err
	}

	// indicate that a forbidden license was found
	found := false

	for _, lib := range libs {
		licenseName, licenseType, err := classifier.Identify(lib.LicensePath)
		if err != nil {
			return err
		}

		if checkLicenseName && !isAllowedLicenseName(licenseName, allowedLicenseNames) {
			fmt.Fprintf(os.Stderr, "Not allowed license %s found for library %v\n", licenseName, lib)
			found = true
		}

		if checkLicenseType && isExcludedLicenseType(licenseType, excludedLicenseTypes) {
			fmt.Fprintf(os.Stderr, "%s license type %s found for library %v\n", cases.Title(language.English).String(licenseType.String()), licenseName, lib)
			found = true
		}
	}

	if found {
		os.Exit(1)
	}

	return nil
}

func getExcludedLicenseTypes() []licenses.Type {
	excludedLicenseTypes := make([]licenses.Type, 0)

	if excludeForbidden {
		excludedLicenseTypes = append(excludedLicenseTypes, licenses.Forbidden)
	}

	if excludeNotice {
		excludedLicenseTypes = append(excludedLicenseTypes, licenses.Notice)
	}

	if excludePermissive {
		excludedLicenseTypes = append(excludedLicenseTypes, licenses.Permissive)
	}

	if excludeReciprocal {
		excludedLicenseTypes = append(excludedLicenseTypes, licenses.Reciprocal)
	}

	if excludeRestricted {
		excludedLicenseTypes = append(excludedLicenseTypes, licenses.Restricted)
	}

	if excludeUnencumbered {
		excludedLicenseTypes = append(excludedLicenseTypes, licenses.Unencumbered)
	}

	return excludedLicenseTypes
}

func isExcludedLicenseType(licenseType licenses.Type, excludedLicenseTypes []licenses.Type) bool {
	for _, excluded := range excludedLicenseTypes {
		if excluded == licenseType {
			return true
		}
	}

	return false
}

func getAllowedLicenseNames() []string {
	if len(allowedLicenseNames) == 0 {
		return []string{}
	}

	var allowed []string

	for _, licenseName := range allowedLicenseNames {
		allowed = append(allowed, strings.TrimSpace(licenseName))
	}

	return allowed
}

func isAllowedLicenseName(licenseName string, allowedLicenseNames []string) bool {
	for _, allowed := range allowedLicenseNames {
		if allowed == licenseName {
			return true
		}
	}

	return false
}
