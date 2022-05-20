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

package licenses

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestFind(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get working directory: %v", err)
	}

	classifier := classifierStub{
		licenseNames: map[string]string{
			"testdata/LICENSE":                                   "foo",
			"testdata/MIT/LICENSE.MIT":                           "MIT",
			"testdata/licence/LICENCE":                           "foo",
			"testdata/copying/COPYING":                           "foo",
			"testdata/notice/NOTICE.txt":                         "foo",
			"testdata/readme/README.md":                          "foo",
			"testdata/lowercase/license":                         "foo",
			"testdata/license-apache-2.0/LICENSE-APACHE-2.0.txt": "foo",
		},
		licenseTypes: map[string]Type{
			"testdata/LICENSE":                                   Notice,
			"testdata/MIT/LICENSE.MIT":                           Notice,
			"testdata/licence/LICENCE":                           Notice,
			"testdata/copying/COPYING":                           Notice,
			"testdata/notice/NOTICE.txt":                         Notice,
			"testdata/readme/README.md":                          Notice,
			"testdata/lowercase/license":                         Notice,
			"testdata/license-apache-2.0/LICENSE-APACHE-2.0.txt": Notice,
		},
	}

	for _, test := range []struct {
		desc            string
		dir             string
		rootDir         string
		wantLicensePath string
		wantErr         *regexp.Regexp
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
		{
			desc:            "lowercase",
			dir:             "testdata/lowercase",
			wantLicensePath: filepath.Join(wd, "testdata/lowercase/license"),
		},
		{
			desc:            "license-apache-2.0.txt",
			dir:             "testdata/license-apache-2.0",
			wantLicensePath: filepath.Join(wd, "testdata/license-apache-2.0/LICENSE-APACHE-2.0.txt"),
		},
		{
			desc:    "proprietary-license",
			dir:     "testdata/proprietary-license",
			rootDir: "testdata/proprietary-license",
			wantErr: regexp.MustCompile(`cannot find a known open source license for.*testdata/proprietary-license.*whose name matches regexp.*and locates up until.*testdata/proprietary-license`),
		},
		{
			desc:            "UNLICENSE",
			dir:             "testdata/unlicense",
			wantLicensePath: filepath.Join(wd, "testdata/unlicense/UNLICENSE"),
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if test.rootDir == "" {
				test.rootDir = "./testdata"
			}
			licensePath, err := Find(test.dir, test.rootDir, classifier)
			if test.wantErr != nil {
				if err == nil || !test.wantErr.Match([]byte(err.Error())) {
					t.Fatalf("Find(%q) = %q, %q, want (%q, %q)", test.dir, licensePath, err, "", test.wantErr)
				}
				return
			}
			if err != nil || licensePath != test.wantLicensePath {
				t.Fatalf("Find(%q) = (%#v, %q), want (%q, nil)", test.dir, licensePath, err, test.wantLicensePath)
			}
		})
	}
}
