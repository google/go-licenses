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

package licenses

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	licenseclassifier "github.com/google/licenseclassifier/v2"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

const DefaultConfidenceThreshold = 0.80

var ErrorEmptyDir = errors.New("Invalid Argument: dir is empty")

type LicenseFound struct {
	SpdxId     string // https://spdx.org/licenses/
	Path       string
	StartLine  int
	EndLine    int
	Confidence float64
}

type ScanDirOptions struct {
	ExcludePaths []string
	DbPath       string
}

type matchType string

const (
	matchTypeHeader  matchType = "Header"
	matchTypeLicense matchType = "License"
)

var ignoredDir map[string]bool = make(map[string]bool)

func init() {
	ignoredDir[".git"] = true
	ignoredDir["node_modules"] = true
}

// Scan a directory for licenses.
func ScanDir(dir string, options ScanDirOptions) ([]LicenseFound, error) {
	var wrap = func(cause error, extra string) error {
		extraMessage := ""
		if extra != "" {
			extraMessage = fmt.Sprintf(": %s", extra)
		}
		return errors.Wrapf(cause, "Failed to scan dir %s%s", dir, extraMessage)
	}
	if dir == "" {
		return nil, ErrorEmptyDir
	}

	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
	}
	excludeAbsPaths := make(map[string]bool)
	for _, path := range options.ExcludePaths {
		absPath, err := filepath.Abs(filepath.Join(dir, path))
		if err != nil {
			return nil, wrap(err, fmt.Sprintf("Invalid exclude path %s", path))
		}
		excludeAbsPaths[absPath] = true
	}
	classifier := licenseclassifier.NewClassifier(DefaultConfidenceThreshold)
	classifier.LoadLicenses(options.DbPath)
	foundLicenses := make([]LicenseFound, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return wrap(err, "walk error")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			// skip symbolic links
			return nil
		}
		if info.IsDir() {
			// TODO: move this to config
			if ignoredDir[info.Name()] {
				return filepath.SkipDir
			}
			_, excluded := excludeAbsPaths[path]
			if excluded {
				return filepath.SkipDir
			}
			return nil
		}
		klog.V(5).Infof(path)
		_, excluded := excludeAbsPaths[path]
		if excluded {
			return nil
		}
		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return wrap(err, fmt.Sprintf("reading file %s", path))
		}
		matches := classifier.Match(fileBytes)
		for _, match := range matches {
			if match.MatchType == string(matchTypeHeader) {
				// ignore headers
				// TODO: verify detected header licenses are included by top level license file
				continue
			}
			foundLicenses = append(foundLicenses, LicenseFound{
				SpdxId:     match.Name,
				Path:       path[len(dir)+1:], // relative path from module.Dir
				StartLine:  match.StartLine,
				EndLine:    match.EndLine,
				Confidence: match.Confidence,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return foundLicenses, nil
}

// Temporarily disabled
// func GetLicenseFullText(module goutils.Module, license LicenseFound) (string, error) {
// 	errorContext := func() string {
// 		return fmt.Sprintf("Failed to get full text license for %s by %+v", module.Path, license)
// 	}
// 	fileBytes, err := ioutil.ReadFile(filepath.Join(module.Dir, license.Path))
// 	if err != nil {
// 		return "", errors.Wrap(err, errorContext())
// 	}
// 	fullText := string(fileBytes)
// 	if license.Offset < 0 || license.Offset >= len(fullText) {
// 		return "", errors.Errorf("%s: offset invalid for full text length %v", errorContext(), len(fullText))
// 	}
// 	if license.Extent <= 0 || license.Offset+license.Extent > len(fullText) {
// 		return "", errors.Errorf("%s: extent invalid for full text length %v", errorContext(), len(fullText))
// 	}
// 	return fullText[license.Offset:(license.Offset + license.Extent)], nil
// }
