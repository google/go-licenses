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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-licenses/v2/gocli"
	"github.com/stretchr/testify/assert"
)

func TestListModulesInGoBinary(t *testing.T) {
	var tests = []struct {
		workdir     string
		mainModule  string
		packagePath string
		modules     []string
	}{
		{
			workdir:     "../tests/modules/hello01",
			mainModule:  "github.com/google/go-licenses/v2/tests/modules/hello01",
			packagePath: "github.com/google/go-licenses/v2/tests/modules/hello01",
			modules:     []string{},
		},
		{
			workdir:     "../tests/modules/cli02",
			mainModule:  "github.com/google/go-licenses/v2/tests/modules/cli02",
			packagePath: "github.com/google/go-licenses/v2/tests/modules/cli02",
			modules: []string{
				"github.com/fsnotify/fsnotify",
				"github.com/hashicorp/hcl",
				"github.com/magiconair/properties",
				"github.com/mitchellh/go-homedir",
				"github.com/mitchellh/mapstructure",
				"github.com/pelletier/go-toml",
				"github.com/spf13/afero",
				"github.com/spf13/cast",
				"github.com/spf13/cobra",
				"github.com/spf13/jwalterweatherman",
				"github.com/spf13/pflag",
				"github.com/spf13/viper",
				"github.com/subosito/gotenv",
				"golang.org/x/sys",
				"golang.org/x/text",
				"gopkg.in/ini.v1",
				"gopkg.in/yaml.v2",
			},
		},
	}
	originalWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("List modules in go binary built from %s", tc.workdir), func(t *testing.T) {
			os.Chdir(filepath.Join(originalWorkDir, tc.workdir))
			// This outputs the built binary as name "main".
			binaryName := "main"
			cmd := exec.Command("go", "build", "-o", binaryName)
			_, err := cmd.Output()
			if err != nil {
				t.Fatalf("Failed to build binary: %v", err)
			}
			metadata, err := gocli.ExtractBinaryMetadata(binaryName)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.mainModule, metadata.MainModule)
			assert.Equal(t, tc.packagePath, metadata.Path)
			modulesActual := make([]string, 0)
			for _, module := range metadata.Modules {
				assert.NotEmpty(t, module.Path)
				assert.NotEmpty(t, module.Version)
				modulesActual = append(modulesActual, module.Path)
			}
			assert.Equal(t, tc.modules, modulesActual)
		})
	}
}
