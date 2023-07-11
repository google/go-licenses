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

package bouncer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLicensePath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get working directory: %v", err)
	}

	classifier := classifierStub{
		licenseNames: map[string]string{
			"testdata/LICENSE":               "foo",
			"testdata/lower_license/license": "foo",
			"testdata/MIT/LICENSE.MIT":       "MIT",
			"testdata/licence/LICENCE":       "foo",
			"testdata/copying/COPYING":       "foo",
			"testdata/notice/NOTICE.txt":     "foo",
			"testdata/readme/README.md":      "foo",
		},
		licenseTypes: map[string]Type{
			"testdata/LICENSE":               Notice,
			"testdata/lower_license/license": Notice,
			"testdata/MIT/LICENSE.MIT":       Notice,
			"testdata/licence/LICENCE":       Notice,
			"testdata/copying/COPYING":       Notice,
			"testdata/notice/NOTICE.txt":     Notice,
			"testdata/readme/README.md":      Notice,
		},
	}

	for _, test := range []struct {
		desc            string
		dir             string
		wantLicensePath string
	}{
		{
			desc:            "licenSe",
			dir:             "testdata",
			wantLicensePath: filepath.Join(wd, "testdata/LICENSE"),
		},
		{
			desc:            "licenCe",
			dir:             "testdata/licence",
			wantLicensePath: filepath.Join(wd, "testdata/licence/LICENCE"),
		},
		{
			desc:            "license",
			dir:             "testdata/lower_license",
			wantLicensePath: filepath.Join(wd, "testdata/lower_license/license"),
		},
		{
			desc:            "LICENSE.MIT",
			dir:             "testdata/MIT",
			wantLicensePath: filepath.Join(wd, "testdata/MIT/LICENSE.MIT"),
		},
		{
			desc:            "COPYING",
			dir:             "testdata/copying",
			wantLicensePath: filepath.Join(wd, "testdata/copying/COPYING"),
		},
		{
			desc:            "NOTICE",
			dir:             "testdata/notice",
			wantLicensePath: filepath.Join(wd, "testdata/notice/NOTICE.txt"),
		},
		{
			desc:            "README",
			dir:             "testdata/readme",
			wantLicensePath: filepath.Join(wd, "testdata/readme/README.md"),
		},
		{
			desc:            "parent dir",
			dir:             "testdata/internal",
			wantLicensePath: filepath.Join(wd, "testdata/LICENSE"),
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			licensePath, err := findLicensePath(test.dir, classifier)
			if err != nil || licensePath != test.wantLicensePath {
				t.Fatalf("findLicensePath(%q) = (%#v, %v), want (%v, nil)", test.dir, licensePath, err, test.wantLicensePath)
			}
		})
	}
}
