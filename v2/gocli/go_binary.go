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

package gocli

import (
	"context"
	"fmt"

	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/model"
	lichenmodule "github.com/google/go-licenses/v2/third_party/uw-labs/lichen/module"
	"golang.org/x/tools/go/packages"
)

// Module metadata extracted from binary and local go module workspace.
type BinaryMetadata struct {
	// The main module used to build the binary.
	// e.g. github.com//google/go-licenses/v2/tests/modules/cli02
	MainModule string
	// Import path of the main package to build the binary.
	// e.g. github.com//google/go-licenses/v2/tests/modules/cli02/cmd
	Path string
	// Detailed metadata of all the module dependencies.
	// Does not include the main module.
	Modules []packages.Module
}

// List dependencies from module metadata in a go binary.
//
// Prerequisites:
// * The go binary must be built with go modules without any further modifications.
// * The command must run with working directory same as to build the analyzed
// go binary, because we need the exact go modules info used to build it.
//
// Here, I am using [1] as a short term solution. It runs [4] go version -m and parses
// output. This is preferred over [2], because [2] is an alternative implemention
// for go version -m, and I expect better long term compatibility for go version -m.
//
// The parsing command output hack is still unfavorable in the long term. As
// dicussed in [3], golang community will move go version parsing into an individual
// module in golang.org/x. We can use that module instead after it is built.
//
// References of similar implementations or dicussions:
// 1. https://github.com/uw-labs/lichen/blob/be9752894a5958f6ba7be9e05dc370b7a73b58db/internal/module/extract.go#L16
// 2. https://github.com/mitchellh/golicense/blob/8c09a94a11ac73299a72a68a7b41e3a737119f91/module/module.go#L27
// 3. https://github.com/golang/go/issues/39301
// 4. https://golang.org/pkg/cmd/go/internal/version/
func ExtractBinaryMetadata(ctx context.Context, path string) (*BinaryMetadata, error) {
	buildInfo, err := listModulesInBinary(ctx, path)
	if err != nil {
		return nil, err
	}
	mods, err := joinModulesMetadata(buildInfo.ModuleRefs)
	if err != nil {
		return nil, err
	}
	return &BinaryMetadata{
		MainModule: buildInfo.ModulePath,
		Path:       buildInfo.PackagePath,
		Modules:    mods,
	}, nil
}

func listModulesInBinary(ctx context.Context, path string) (buildinfo *model.BuildInfo, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("listModulesInGoBinary(path=%q): %w", path, err)
		}
	}()
	depsBuildInfo, err := lichenmodule.Extract(ctx, path)
	if err != nil {
		return nil, err
	}
	if len(depsBuildInfo) != 1 {
		return nil, fmt.Errorf("len(depsBuildInfo) should be 1, but found %v", len(depsBuildInfo))
	}
	return &depsBuildInfo[0], nil
}

func joinModulesMetadata(refs []model.ModuleReference) (modules []packages.Module, err error) {
	localModulesDict, err := ListModules()
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		localModule, ok := localModulesDict[ref.Path]
		if !ok {
			return nil, fmt.Errorf("Cannot find %v in current dir's go modules. Are you running this tool from the working dir to build the binary you are analyzing?", ref.Path)
		}
		if localModule.Dir == "" {
			return nil, fmt.Errorf("Module %v's local directory is empty. Did you run go mod download?", ref.Path)
		}
		if localModule.Version != ref.Version {
			return nil, fmt.Errorf("Found %v %v in go binary, but %v is downloaded in go modules. Are you running this tool from the working dir to build the binary you are analyzing?", ref.Path, ref.Version, localModule.Version)
		}
		modules = append(modules, localModule)
	}
	return modules, nil
}
