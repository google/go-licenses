// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-licenses/v2/config"
	"github.com/google/go-licenses/v2/dict"
	"github.com/google/go-licenses/v2/ghutils"
	"github.com/google/go-licenses/v2/goutils"
	"github.com/google/licenseclassifier"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save full license text locally",
	Long:  `Save full license text and source code locally to be compliant.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.Load("")
		if err != nil {
			klog.ErrorS(err, "Failed: load config")
			os.Exit(1)
		}
		info, err := loadInfo(config.Module.Csv.Path)
		if err != nil {
			klog.ErrorS(err, "Failed: load license info csv")
			os.Exit(1)
		}
		err = complyWithLicenses(info, *config)
		if err != nil {
			klog.ErrorS(err, "Failed: comply with licenses")
			os.Exit(1)
		}
		klog.Flush()
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// saveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// saveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const defaultNoticesPath = "NOTICES"
const defaultLicenseSubPath = "licenses.txt"
const defaultSrcPath = "src"

// Dir permission needs execute bit for `cd` or `ls` commands
// ref: https://www.tutorialspoint.com/unix/unix-file-permission.htm
const permDirCurrentUser = 0700
const permFileCurrentUser = 0600

// The following code is largely inspired by
// https://github.com/google/go-licenses/blob/master/save.go, which is licensed under:
//
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

// license compliance requirement type
type ComplianceReq string

const (
	// We do not allow unknown licenses.
	Unknown ComplianceReq = "Unknown"
	// We need to redistribute the entire source directory to be compliant,
	// example licenses: GPL, MPL, etc.
	RedistributeSource ComplianceReq = "DistributeSource"
	// We need to redistribute full text license and a copyright notice to be
	// compliant: most other licenses.
	RedistributeNotice ComplianceReq = "DistributeNotice"
)

// Determines compliance requirement type of a license, returns ComplianceReq.
// license can be a list of licenses like "Apache-2.0 / MIT", this method returns
// strictest ComplianceReq type. The license names should be SPDX ID format.
func requirementType(license string) (ComplianceReq, error) {
	// By default, we distribute notice for any licenses.
	requirement := RedistributeNotice
	for _, part := range strings.Split(license, "/") {
		spdxId := strings.TrimSpace(part)
		if spdxId == "" {
			return Unknown, fmt.Errorf("Empty SPDX ID in %q", license)
		}

		licenseType := licenseclassifier.LicenseType(spdxId)
		switch licenseType {
		case "restricted", "reciprocal":
			requirement = RedistributeSource
		case "notice", "permissive", "unencumbered":
			// No special handling.
		default:
			// Any unknown license type is not allowed, so we return unknown.
			// TODO: allow user configurable license type dictionary.
			return Unknown, nil
		}
	}
	return requirement, nil
}

func complyWithLicenses(info []*dict.LicenseRecord, config config.GoModLicensesConfig) error {
	noticesPath := config.Module.Notices.Path
	if noticesPath == "" {
		noticesPath = defaultNoticesPath
	}
	licensePath := filepath.Join(noticesPath, defaultLicenseSubPath)
	srcPath := filepath.Join(noticesPath, defaultSrcPath)
	modules, err := goutils.ListModules()
	if err != nil {
		return errors.Wrap(err, "Failed to list modules")
	}
	moduleDict := goutils.BuildModuleDict(modules)

	err = os.RemoveAll(srcPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to remove all in %s", srcPath)
	}
	err = os.MkdirAll(path.Dir(licensePath), permDirCurrentUser)
	if err != nil {
		return errors.Wrapf(err, "Failed to mkdir %s", path.Dir(licensePath))
	}
	f, err := os.Create(licensePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to create %s", licensePath)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	modulesWithBadLicenses := make([]*dict.LicenseRecord, 0)
	for _, record := range info {
		reqType, err := requirementType(record.Type)
		if err != nil {
			return fmt.Errorf("%s: license=%q: %w", record.Module, record.Type, err)
		}
		switch reqType {
		case RedistributeSource:
			// Copy the entire source directory for the library.
			moduleRecord, exists := moduleDict[record.Module]
			if !exists {
				// TODO: try if any parent module exists in moduleDict.
				return errors.Errorf("%s: Cannot find module in `go list -m all`", record.Module)
			}
			if moduleRecord.Dir == "" {
				return errors.Errorf(
					"%s: Module Dir is empty in `go list -m -json %s`. Please run `go mod download` before running `go-licenses save`.",
					record.Module, record.Module,
				)
			}
			if err := copySrc(moduleRecord.Dir, filepath.Join(srcPath, record.Module)); err != nil {
				return errors.Wrapf(err, "%s: Failed to copy source dir from %s to %s", record.Module, moduleRecord.Dir, srcPath)
			}
		case RedistributeNotice:
			// No special handling.
		default:
			modulesWithBadLicenses = append(modulesWithBadLicenses, record)
		}
		if len(modulesWithBadLicenses) > 0 {
			// if we find bad licenses, we only need to report all moodules with
			// bad licenses.
			continue
		}
		licenseContent, err := ghutils.SmartDownload(record.DownaloadUrl)
		if err != nil {
			return errors.Wrapf(err, "%s", record.Module)
		}
		mustWrite := func(text string) {
			_, err := w.WriteString(text)
			if err != nil {
				klog.Exit(fmt.Errorf("Failed to write license to %q: %w", licensePath, err))
			}
		}
		// Despite license type, we always put its notice and license in a single licenses.txt file.
		mustWrite(fmt.Sprintf("============= %s =============\n", record.Module))
		mustWrite(fmt.Sprintf("%s\n\n", record.DownaloadUrl))
		mustWrite(string(licenseContent))
		mustWrite("\n\n")
		klog.Infof("%s: Downloaded %s", record.Module, record.DownaloadUrl)
	}
	if len(modulesWithBadLicenses) > 0 {
		for _, module := range modulesWithBadLicenses {
			klog.ErrorS(fmt.Errorf("unknown license type"), "module", module.Module, "license", module.Type)
		}
		return fmt.Errorf("%v modules has rejected licenses", len(modulesWithBadLicenses))
	}
	err = w.Flush()
	if err != nil {
		return errors.Wrapf(err, "Failed to flush %s", defaultLicenseInfoLocation)
	}
	return nil
}

func loadInfo(path string) ([]*dict.LicenseRecord, error) {
	if path == "" {
		path = defaultLicenseInfoLocation
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read license info, path=%q: %w", path, err)
	}
	return dict.LoadLicenseRecords(bytes.NewReader(content))
}

func copySrc(src, dest string) error {
	opt := copy.Options{
		// Go module files are by default read-only, so we need to change perm on copy.
		// Reference: https://github.com/golang/go/issues/31481.
		AddPermission: permFileCurrentUser,
		// Skip the .git directory for copying, if it exists, since we don't want to save the user's
		// local Git config along with the source code.
		Skip: func(src string) (bool, error) { return strings.HasSuffix(src, ".git"), nil },
	}
	if err := copy.Copy(src, dest, opt); err != nil {
		return err
	}
	return nil
}
