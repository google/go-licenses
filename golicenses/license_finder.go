package golicenses

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

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/khulnasoft/go-licenses/golicenses/event"
	"github.com/khulnasoft/go-licenses/internal/bus"
	"github.com/khulnasoft/go-progress"
	"github.com/khulnasoft/go-pulsebus"

	"github.com/google/licenseclassifier"
	"github.com/hashicorp/go-multierror"
	"github.com/markbates/pkger"
)

type LicenseFinder struct {
	Path                string
	ConfidenceThreshold float64
}

func NewLicenseFinder(path string, threshold float64) LicenseFinder {
	return LicenseFinder{
		Path:                path,
		ConfidenceThreshold: threshold,
	}
}

func licenseDbArchiveFetcher() ([]byte, error) {
	f, err := pkger.Open("/assets/licenses.db")
	if err != nil {
		return nil, fmt.Errorf("unable to open license.db: %w", err)
	}

	defer f.Close()
	return ioutil.ReadAll(f)
}

func (r LicenseFinder) Find() ([]LicenseResult, error) {

	dbFetcherOpt := licenseclassifier.ArchiveFunc(licenseDbArchiveFetcher)
	classifier, err := NewLicenseClassifier(r.ConfidenceThreshold, dbFetcherOpt)
	if err != nil {
		return nil, err
	}

	modules, err := ListModules(r.Path)
	if err != nil {
		return nil, err
	}

	stage := &progress.Stage{}
	prog := &progress.Manual{
		// this is the number of stages to expect; start + individual endpoints + stop
		Total: int64(len(modules)),
	}
	bus.Publish(pulsebus.Event{
		Type: event.ModuleScanStarted,
		//Source: source,
		Value: progress.StagedProgressable(&struct {
			progress.Stager
			progress.Progressable
		}{
			Stager:       progress.Stager(stage),
			Progressable: prog,
		}),
	})

	var results []LicenseResult

	defer prog.SetCompleted()
	for _, modInfo := range modules {
		prog.N++
		stage.Current = modInfo.Path
		var licenseName string
		var classification Type
		var errs error

		licensePath, err := findLicensePath(modInfo.Dir, classifier)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to find license (%s): %w", modInfo.Dir, err))
			licensePath = ""
		}

		if licensePath != "" {
			licenseName, classification, err = classifier.Identify(licensePath)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("failed to identify license (%s): %w", licensePath, err))
				licenseName = ""
			}
		}

		results = append(results, LicenseResult{
			ModulePath:  modInfo.Path,
			LicensePath: licensePath,
			License:     licenseName,
			Type:        classification.String(),
			Errs:        errs,
		})
	}

	return results, nil
}

var (
	licenseRegexp = regexp.MustCompile(`^((L|l)icen(s|c)e|LICEN(S|C)E|COPYING|README|NOTICE)(\..+)?$`)
	srcDirRegexps = func() []*regexp.Regexp {
		var rs []*regexp.Regexp
		for _, s := range build.Default.SrcDirs() {
			rs = append(rs, regexp.MustCompile("^"+regexp.QuoteMeta(s)+"$"))
		}
		return rs
	}()
	vendorRegexp = regexp.MustCompile(`.+/vendor(/)?$`)
)

// findLicensePath returns the file path of the license for this package.
func findLicensePath(dir string, classifier Classifier) (string, error) {
	var stopAt []*regexp.Regexp
	stopAt = append(stopAt, srcDirRegexps...)
	stopAt = append(stopAt, vendorRegexp)
	return findUpwards(dir, licenseRegexp, stopAt, func(path string) bool {
		// TODO(RJPercival): Return license details
		if _, _, err := classifier.Identify(path); err != nil {
			return false
		}
		return true
	})
}

func findUpwards(dir string, r *regexp.Regexp, stopAt []*regexp.Regexp, predicate func(path string) bool) (string, error) {
	// Dir must be made absolute for reliable matching with stopAt regexps
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	start := dir
	// Stop once dir matches a stopAt regexp or dir is the filesystem root
	for !matchAny(stopAt, dir) {
		dirContents, err := ioutil.ReadDir(dir)
		if err != nil {
			return "", err
		}
		for _, f := range dirContents {
			if r.MatchString(f.Name()) {
				path := filepath.Join(dir, f.Name())
				if predicate != nil && !predicate(path) {
					continue
				}
				return path, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Can't go any higher up the directory tree.
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no file/directory matching regexp %q found for %s", r, start)
}

func matchAny(patterns []*regexp.Regexp, s string) bool {
	for _, p := range patterns {
		if p.MatchString(s) {
			return true
		}
	}
	return false
}
