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

	"github.com/google/go-licenses/licenses"
	"github.com/google/licenseclassifier"
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
	supported := getSupportedLicenseNames()

	for _, licenseName := range allowedLicenseNames {
		if isAllowedLicenseName(licenseName, supported) {
			allowed = append(allowed, licenseName)

			continue
		}

		fmt.Fprintf(os.Stderr, "Unsupported license name %s provided, it will be ignored\n", licenseName)
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

func getSupportedLicenseNames() []string {
	return []string{
		licenseclassifier.AFL11,
		licenseclassifier.AFL12,
		licenseclassifier.AFL20,
		licenseclassifier.AFL21,
		licenseclassifier.AFL30,
		licenseclassifier.AGPL10,
		licenseclassifier.AGPL30,
		licenseclassifier.Apache10,
		licenseclassifier.Apache11,
		licenseclassifier.Apache20,
		licenseclassifier.APSL10,
		licenseclassifier.APSL11,
		licenseclassifier.APSL12,
		licenseclassifier.APSL20,
		licenseclassifier.Artistic10cl8,
		licenseclassifier.Artistic10Perl,
		licenseclassifier.Artistic10,
		licenseclassifier.Artistic20,
		licenseclassifier.BCL,
		licenseclassifier.Beerware,
		licenseclassifier.BSD2ClauseFreeBSD,
		licenseclassifier.BSD2ClauseNetBSD,
		licenseclassifier.BSD2Clause,
		licenseclassifier.BSD3ClauseAttribution,
		licenseclassifier.BSD3ClauseClear,
		licenseclassifier.BSD3ClauseLBNL,
		licenseclassifier.BSD3Clause,
		licenseclassifier.BSD4Clause,
		licenseclassifier.BSD4ClauseUC,
		licenseclassifier.BSDProtection,
		licenseclassifier.BSL10,
		licenseclassifier.CC010,
		licenseclassifier.CCBY10,
		licenseclassifier.CCBY20,
		licenseclassifier.CCBY25,
		licenseclassifier.CCBY30,
		licenseclassifier.CCBY40,
		licenseclassifier.CCBYNC10,
		licenseclassifier.CCBYNC20,
		licenseclassifier.CCBYNC25,
		licenseclassifier.CCBYNC30,
		licenseclassifier.CCBYNC40,
		licenseclassifier.CCBYNCND10,
		licenseclassifier.CCBYNCND20,
		licenseclassifier.CCBYNCND25,
		licenseclassifier.CCBYNCND30,
		licenseclassifier.CCBYNCND40,
		licenseclassifier.CCBYNCSA10,
		licenseclassifier.CCBYNCSA20,
		licenseclassifier.CCBYNCSA25,
		licenseclassifier.CCBYNCSA30,
		licenseclassifier.CCBYNCSA40,
		licenseclassifier.CCBYND10,
		licenseclassifier.CCBYND20,
		licenseclassifier.CCBYND25,
		licenseclassifier.CCBYND30,
		licenseclassifier.CCBYND40,
		licenseclassifier.CCBYSA10,
		licenseclassifier.CCBYSA20,
		licenseclassifier.CCBYSA25,
		licenseclassifier.CCBYSA30,
		licenseclassifier.CCBYSA40,
		licenseclassifier.CDDL10,
		licenseclassifier.CDDL11,
		licenseclassifier.CommonsClause,
		licenseclassifier.CPAL10,
		licenseclassifier.CPL10,
		"eGenix", // licenseclassifier.eGenix is not exported
		licenseclassifier.EPL10,
		licenseclassifier.EPL20,
		licenseclassifier.EUPL10,
		licenseclassifier.EUPL11,
		licenseclassifier.Facebook2Clause,
		licenseclassifier.Facebook3Clause,
		licenseclassifier.FacebookExamples,
		licenseclassifier.FreeImage,
		licenseclassifier.FTL,
		licenseclassifier.GPL10,
		licenseclassifier.GPL20,
		licenseclassifier.GPL20withautoconfexception,
		licenseclassifier.GPL20withbisonexception,
		licenseclassifier.GPL20withclasspathexception,
		licenseclassifier.GPL20withfontexception,
		licenseclassifier.GPL20withGCCexception,
		licenseclassifier.GPL30,
		licenseclassifier.GPL30withautoconfexception,
		licenseclassifier.GPL30withGCCexception,
		licenseclassifier.GUSTFont,
		licenseclassifier.ImageMagick,
		licenseclassifier.IPL10,
		licenseclassifier.ISC,
		licenseclassifier.LGPL20,
		licenseclassifier.LGPL21,
		licenseclassifier.LGPL30,
		licenseclassifier.LGPLLR,
		licenseclassifier.Libpng,
		licenseclassifier.Lil10,
		licenseclassifier.LinuxOpenIB,
		licenseclassifier.LPL102,
		licenseclassifier.LPL10,
		licenseclassifier.LPPL13c,
		licenseclassifier.MIT,
		licenseclassifier.MPL10,
		licenseclassifier.MPL11,
		licenseclassifier.MPL20,
		licenseclassifier.MSPL,
		licenseclassifier.NCSA,
		licenseclassifier.NPL10,
		licenseclassifier.NPL11,
		licenseclassifier.OFL11,
		licenseclassifier.OpenSSL,
		licenseclassifier.OpenVision,
		licenseclassifier.OSL10,
		licenseclassifier.OSL11,
		licenseclassifier.OSL20,
		licenseclassifier.OSL21,
		licenseclassifier.OSL30,
		licenseclassifier.PHP301,
		licenseclassifier.PHP30,
		licenseclassifier.PIL,
		licenseclassifier.PostgreSQL,
		licenseclassifier.Python20complete,
		licenseclassifier.Python20,
		licenseclassifier.QPL10,
		licenseclassifier.Ruby,
		licenseclassifier.SGIB10,
		licenseclassifier.SGIB11,
		licenseclassifier.SGIB20,
		licenseclassifier.SISSL12,
		licenseclassifier.SISSL,
		licenseclassifier.Sleepycat,
		licenseclassifier.UnicodeTOU,
		licenseclassifier.UnicodeDFS2015,
		licenseclassifier.UnicodeDFS2016,
		licenseclassifier.Unlicense,
		licenseclassifier.UPL10,
		licenseclassifier.W3C19980720,
		licenseclassifier.W3C20150513,
		licenseclassifier.W3C,
		licenseclassifier.WTFPL,
		licenseclassifier.X11,
		licenseclassifier.Xnet,
		licenseclassifier.Zend20,
		licenseclassifier.ZeroBSD,
		licenseclassifier.ZlibAcknowledgement,
		licenseclassifier.Zlib,
		licenseclassifier.ZPL11,
		licenseclassifier.ZPL20,
		licenseclassifier.ZPL21,
	}
}
