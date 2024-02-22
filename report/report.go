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

package report

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/google/go-licenses/internal/third_party/pkgsite/source"
	"github.com/google/go-licenses/licenses"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
)

const (
	UNKNOWN = "Unknown"
)

type libraryData struct {
	Name         string
	Version      string
	LicensePath  string
	LicenseURL   string
	LicenseNames []string
}

type LibraryData struct {
	Name        string
	Version     string
	LicensePath string
	LicenseURL  string
	LicenseName string
}

// LicenseText reads and returns the contents of LicensePath, if set
// or an empty string if not.
func (lib LibraryData) LicenseText() (string, error) {
	if lib.LicensePath == "" {
		return "", nil
	}
	data, err := os.ReadFile(lib.LicensePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Report(includeTests bool, ignore []string, importPaths []string) ([]LibraryData, error) {
	classifier, err := licenses.NewClassifier()
	if err != nil {
		return nil, err
	}

	libs, err := licenses.Libraries(context.Background(), classifier, includeTests, ignore, importPaths...)
	if err != nil {
		return nil, err
	}

	reportData := make([]libraryData, len(libs))
	client := source.NewClient(time.Second * 20)
	group, gctx := errgroup.WithContext(context.Background())
	for idx, lib := range libs {
		idx := idx
		lib := lib

		reportData[idx] = libraryData{
			Name:         lib.Name(),
			Version:      UNKNOWN,
			LicensePath:  UNKNOWN,
			LicenseURL:   UNKNOWN,
			LicenseNames: nil,
		}

		if version := lib.Version(); version != "" {
			reportData[idx].Version = version
		}

		if lib.LicenseFile != "" {
			reportData[idx].LicensePath = lib.LicenseFile
		}

		for _, license := range lib.Licenses {
			reportData[idx].LicenseNames = append(reportData[idx].LicenseNames, license.Name)
		}

		if lib.LicenseFile != "" {
			group.Go(func() error {
				url, err := lib.FileURL(gctx, client, lib.LicenseFile)
				if err == nil {
					reportData[idx].LicenseURL = url
				} else {
					klog.Warningf("Error discovering license URL: %s", err)
				}
				return nil
			})
		}
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return flatteReportData(reportData), nil
}

func flatteReportData(reportData []libraryData) []LibraryData {
	// Flatten the report data
	reportDataFlat := make([]LibraryData, 0, len(reportData))
	for _, lib := range reportData {
		if len(lib.LicenseNames) == 0 {
			if lib.LicensePath != UNKNOWN {
				klog.Errorf("Error identifying license in %q: %v", lib.LicensePath, fmt.Errorf("no license found"))
			} else if lib.Version != UNKNOWN {
				klog.Errorf("Error identifying license for version %q of %q: %v", lib.Version, lib.Name, fmt.Errorf("no license found"))
			} else {
				klog.Errorf("Error identifying license for %q: %v", lib.Name, fmt.Errorf("no license found"))
			}
			reportDataFlat = append(reportDataFlat, LibraryData{
				Name:        lib.Name,
				Version:     lib.Version,
				LicensePath: lib.LicensePath,
				LicenseURL:  lib.LicenseURL,
				LicenseName: UNKNOWN,
			})
		} else {
			for _, licenseName := range lib.LicenseNames {
				reportDataFlat = append(reportDataFlat, LibraryData{
					Name:        lib.Name,
					Version:     lib.Version,
					LicensePath: lib.LicensePath,
					LicenseURL:  lib.LicenseURL,
					LicenseName: licenseName,
				})
			}
		}
	}
	return reportDataFlat
}

func ReportCSV(w io.Writer, libs []LibraryData) error {
	writer := csv.NewWriter(w)
	for _, lib := range libs {
		if err := writer.Write([]string{lib.Name, lib.LicenseURL, lib.LicenseName}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func ReportTemplate(w io.Writer, libs []LibraryData, templateFile string) error {
	templateBytes, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}
	tmpl, err := template.New("").Parse(string(templateBytes))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, libs)
}
