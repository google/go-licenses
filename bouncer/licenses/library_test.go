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
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-licenses/internal/third_party/pkgsite/source"
)

func TestLibraries(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get working directory: %v", err)
	}

	classifier := classifierStub{
		licenses: map[string][]License{
			"testdata/LICENSE": {
				{
					Name: "foo",
					Type: Notice,
				},
			},
			"testdata/direct/LICENSE": {
				{
					Name: "foo",
					Type: Notice,
				},
			},
			"testdata/indirect/LICENSE": {
				{
					Name: "foo",
					Type: Notice,
				},
			},
		},
	}

	type wantedLibInfo struct {
		Name        string
		LicenseFile string
		Licenses    []License
	}

	testdataLibInfo := wantedLibInfo{
		Name:        "github.com/google/go-licenses/licenses/testdata",
		LicenseFile: wd + "/testdata/LICENSE",
		Licenses:    classifier.licenses["testdata/LICENSE"],
	}
	directLibInfo := wantedLibInfo{
		Name:        "github.com/google/go-licenses/licenses/testdata/direct",
		LicenseFile: wd + "/testdata/direct/LICENSE",
		Licenses:    classifier.licenses["testdata/direct/LICENSE"],
	}
	indirectLibInfo := wantedLibInfo{
		Name:        "github.com/google/go-licenses/licenses/testdata/indirect",
		LicenseFile: wd + "/testdata/indirect/LICENSE",
		Licenses:    classifier.licenses["testdata/indirect/LICENSE"],
	}

	for _, test := range []struct {
		desc         string
		importPath   string
		goflags      string
		includeTests bool
		ignore       []string
		wantLibs     []wantedLibInfo
	}{
		{
			desc:       "Detects direct dependency",
			importPath: "github.com/google/go-licenses/licenses/testdata/direct",
			wantLibs:   []wantedLibInfo{directLibInfo, indirectLibInfo},
		},
		{
			desc:       "Detects transitive dependency",
			importPath: "github.com/google/go-licenses/licenses/testdata",
			wantLibs:   []wantedLibInfo{testdataLibInfo, directLibInfo, indirectLibInfo},
		},
		{
			desc:       "Ignores a package path",
			importPath: "github.com/google/go-licenses/licenses/testdata",
			ignore: []string{
				"github.com/google/go-licenses/licenses/testdata/direct",
			},
			wantLibs: []wantedLibInfo{testdataLibInfo, indirectLibInfo},
		},
		{
			desc:         "Detects the dependencies only imported in testing code",
			importPath:   "github.com/google/go-licenses/licenses/testdata/testlib",
			includeTests: true,
			wantLibs: []wantedLibInfo{
				indirectLibInfo,
				{
					Name:        "github.com/google/go-licenses/licenses/testdata/testlib",
					LicenseFile: wd + "/testdata/LICENSE",
					Licenses:    classifier.licenses["testdata/LICENSE"],
				},
			},
		},
		{
			desc:         "Should not detect the dependencies only imported in testing code",
			importPath:   "github.com/google/go-licenses/licenses/testdata/testlib",
			includeTests: false,
			wantLibs: []wantedLibInfo{
				{
					Name:        "github.com/google/go-licenses/licenses/testdata/testlib",
					LicenseFile: wd + "/testdata/LICENSE",
					Licenses:    classifier.licenses["testdata/LICENSE"],
				},
			},
		},
		{
			desc:       "Build tagged package",
			importPath: "github.com/google/go-licenses/licenses/testdata/tags",
			goflags:    "-tags=tags",
			wantLibs: []wantedLibInfo{
				indirectLibInfo,
				{
					Name:        "github.com/google/go-licenses/licenses/testdata/tags",
					LicenseFile: wd + "/testdata/LICENSE",
					Licenses:    classifier.licenses["testdata/LICENSE"],
				},
			},
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if test.goflags != "" {
				os.Setenv("GOFLAGS", test.goflags)
				defer os.Unsetenv("GOFLAGS")
			}
			gotLibs, err := Libraries(context.Background(), classifier, test.includeTests, test.ignore, test.importPath)
			if err != nil {
				t.Fatalf("Libraries(_, %q) = (_, %q), want (_, nil)", test.importPath, err)
			}

			for i, gotLib := range gotLibs {
				if diff := cmp.Diff(test.wantLibs[i].Name, gotLib.Name()); diff != "" {
					t.Errorf("Libraries(_, %q)[%d].Name() diff (-want +got): %s", test.importPath, i, diff)
				}
				if diff := cmp.Diff(test.wantLibs[i].LicenseFile, gotLib.LicenseFile); diff != "" {
					t.Errorf("Libraries(_, %q)[%d].LicenseFile diff (-want +got): %s", test.importPath, i, diff)
				}
				if diff := cmp.Diff(test.wantLibs[i].Licenses, gotLib.Licenses); diff != "" {
					t.Errorf("Libraries(_, %q)[%d].Licenses diff (-want +got): %s", test.importPath, i, diff)
				}
			}

			if diff := cmp.Diff(len(test.wantLibs), len(gotLibs)); diff != "" {
				t.Errorf("len(Libraries(_, %q)) diff (-want +got): %s", test.importPath, diff)
			}
		})
	}
}

func TestLibraryName(t *testing.T) {
	for _, test := range []struct {
		desc     string
		lib      *Library
		wantName string
	}{
		{
			desc:     "Library with no packages",
			lib:      &Library{},
			wantName: "",
		},
		{
			desc: "Library with 1 package",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian/crypto",
				},
			},
			wantName: "github.com/google/trillian/crypto",
		},
		{
			desc: "Library with 2 packages",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian/crypto",
					"github.com/google/trillian/server",
				},
			},
			wantName: "github.com/google/trillian",
		},
		{
			desc: "Vendored library",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian/vendor/coreos/etcd",
				},
			},
			wantName: "github.com/google/trillian/vendor/coreos/etcd",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if got, want := test.lib.Name(), test.wantName; got != want {
				t.Fatalf("Name() = %q, want %q", got, want)
			}
		})
	}
}

func TestLibraryFileURL(t *testing.T) {
	for _, test := range []struct {
		desc    string
		lib     *Library
		path    string
		wantURL string
		wantErr bool
	}{
		{
			desc: "Library on github.com",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian",
					"github.com/google/trillian/crypto",
				},
				LicenseFile: "/go/src/github.com/google/trillian/LICENSE",
				module: &Module{
					Path:    "github.com/google/trillian",
					Dir:     "/go/src/github.com/google/trillian",
					Version: "v1.2.3",
				},
			},
			path:    "/go/src/github.com/google/trillian/foo/README.md",
			wantURL: "https://github.com/google/trillian/blob/v1.2.3/foo/README.md",
		},
		{
			desc: "Library on bitbucket.org",
			lib: &Library{
				Packages: []string{
					"bitbucket.org/user/project/pkg",
					"bitbucket.org/user/project/pkg2",
				},
				LicenseFile: "/foo/bar/bitbucket.org/user/project/LICENSE",
				module: &Module{
					Path:    "bitbucket.org/user/project",
					Dir:     "/foo/bar/bitbucket.org/user/project",
					Version: "v1.2.3",
				},
			},
			path:    "/foo/bar/bitbucket.org/user/project/foo/README.md",
			wantURL: "https://bitbucket.org/user/project/src/v1.2.3/foo/README.md",
		},
		{
			desc: "Library on example.com",
			lib: &Library{
				Packages: []string{
					"example.com/user/project/pkg",
					"example.com/user/project/pkg2",
				},
				LicenseFile: "/foo/bar/example.com/user/project/LICENSE",
				module: &Module{
					Path:    "example.com/user/project",
					Dir:     "/foo/bar/example.com/user/project",
					Version: "v1.2.3",
				},
			},
			path:    "/foo/bar/example.com/user/project/foo/README.md",
			wantURL: "https://example.com/user/project/blob/v1.2.3/foo/README.md",
		},
		{
			desc: "Library without version defaults to remote HEAD",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian",
					"github.com/google/trillian/crypto",
				},
				LicenseFile: "/go/src/github.com/google/trillian/LICENSE",
				module: &Module{
					Path: "github.com/google/trillian",
					Dir:  "/go/src/github.com/google/trillian",
				},
			},
			path:    "/go/src/github.com/google/trillian/foo/README.md",
			wantURL: "https://github.com/google/trillian/blob/HEAD/foo/README.md",
		},
		{
			desc: "Library on k8s.io",
			lib: &Library{
				Packages: []string{
					"k8s.io/api/core/v1",
				},
				LicenseFile: "/go/modcache/k8s.io/api/LICENSE",
				module: &Module{
					Path:    "k8s.io/api",
					Dir:     "/go/modcache/k8s.io/api",
					Version: "v0.23.1",
				},
			},
			path:    "/go/modcache/k8s.io/api/LICENSE",
			wantURL: "https://github.com/kubernetes/api/blob/v0.23.1/LICENSE",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			client := source.NewClient(time.Second * 20)
			fileURL, err := test.lib.FileURL(context.Background(), client, test.path)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Fatalf("FileURL(%q) = (_, %q), want err? %t", test.path, err, test.wantErr)
			} else if gotErr {
				return
			}
			if got, want := fileURL, test.wantURL; got != want {
				t.Fatalf("FileURL(%q) = %q, want %q", test.path, got, want)
			}
		})
	}
}
