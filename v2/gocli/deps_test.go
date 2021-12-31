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

package gocli_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-licenses/v2/gocli"
	"github.com/stretchr/testify/assert"
)

func TestListDeps(t *testing.T) {
	var tests = []struct {
		workdir    string
		mainModule string
		goflags    string
		modules    []string
	}{
		{
			workdir:    "../testdata/modules/hello01",
			mainModule: "github.com/google/go-licenses/v2/testdata/modules/hello01",
			modules: []string{
				"github.com/google/go-licenses/v2/testdata/modules/hello01@(devel)",
			},
		},
		{
			workdir:    "../testdata/modules/cli02",
			mainModule: "github.com/google/go-licenses/v2/testdata/modules/cli02",
			modules: []string{
				"github.com/google/go-licenses/v2/testdata/modules/cli02@(devel)",
				"github.com/fsnotify/fsnotify@v1.4.9",
				"github.com/hashicorp/hcl@v1.0.0",
				"github.com/magiconair/properties@v1.8.5",
				"github.com/mitchellh/go-homedir@v1.1.0",
				"github.com/mitchellh/mapstructure@v1.4.1",
				"github.com/pelletier/go-toml@v1.9.3",
				"github.com/spf13/afero@v1.6.0",
				"github.com/spf13/cast@v1.3.1",
				"github.com/spf13/cobra@v1.1.3",
				"github.com/spf13/jwalterweatherman@v1.1.0",
				"github.com/spf13/pflag@v1.0.5",
				"github.com/spf13/viper@v1.8.0",
				"github.com/subosito/gotenv@v1.2.0",
				"golang.org/x/sys@v0.0.0-20210510120138-977fb7262007",
				"golang.org/x/text@v0.3.5",
				"gopkg.in/ini.v1@v1.62.0",
				"gopkg.in/yaml.v2@v2.4.0",
			},
		},
		{
			// without tags, only the main module is included
			workdir:    "../testdata/modules/tags03",
			mainModule: "github.com/google/go-licenses/v2/testdata/modules/tags03",
			modules: []string{
				"github.com/google/go-licenses/v2/testdata/modules/tags03@(devel)",
			},
		},
		{
			// tags cause additional dependencies to be included
			workdir:    "../testdata/modules/tags03",
			mainModule: "github.com/google/go-licenses/v2/testdata/modules/tags03",
			goflags:    "-tags=tags",
			modules: []string{
				"github.com/google/go-licenses/v2/testdata/modules/tags03@(devel)",
				"github.com/mitchellh/go-homedir@v1.1.0",
			},
		},
	}
	originalWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range tests {
		os.Chdir(filepath.Join(originalWorkDir, tc.workdir))
		sort.Strings(tc.modules)
		normalize := func(mods []gocli.Module) []string {
			res := make([]string, 0, len(mods))
			for _, module := range mods {
				ver := module.Version
				if module.Main && ver == "" {
					// Main module may not have the version, normalize as develop version.
					ver = "(devel)"
				}
				assert.NotEmpty(t, module.Path)
				assert.NotEmpty(t, ver)
				res = append(res, fmt.Sprintf("%s@%s", module.Path, ver))
			}
			sort.Strings(res)
			return res
		}

		t.Run(fmt.Sprintf("gocli.ExtractBinaryMetadata(%s)", tc.workdir), func(t *testing.T) {
			if tc.goflags != "" {
				os.Setenv("GOFLAGS", tc.goflags)
				defer os.Unsetenv("GOFLAGS")
			}
			tempDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempDir)
			// This outputs the built binary as name "main".
			binaryName := path.Join(tempDir, "main")
			cmd := exec.Command("go", "build", "-o", binaryName)
			output, err := cmd.CombinedOutput()
			// defer remove before checking error, because the file
			// may be created even when there's an error.
			defer os.Remove(binaryName)
			if err != nil {
				t.Fatalf("go build: %v\n%s\n", err, string(output))
			}
			metadata, err := gocli.ExtractBinaryMetadata(binaryName)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.modules, normalize(append(metadata.Deps, metadata.Main)))
		})

		t.Run(fmt.Sprintf("gocli.ListDeps(%s)", tc.workdir), func(t *testing.T) {
			if tc.goflags != "" {
				os.Setenv("GOFLAGS", tc.goflags)
				defer os.Unsetenv("GOFLAGS")
			}
			mods, err := gocli.ListDeps(tc.mainModule)
			if err != nil {
				t.Fatalf("gocli.ListDeps: %v", err)
			}
			assert.Equal(t, tc.modules, normalize(mods), "gocli.Modules")
		})
	}
}
