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

package deps

import (
	"context"
	"fmt"

	"github.com/google/go-licenses/v2/goutils"
	lichenmodule "github.com/google/go-licenses/v2/third_party/uw-labs/lichen/module"
	"golang.org/x/mod/module"
)

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
func ListGoBinary(path string) ([]goutils.Module, error) {
	versions, err := listModulesInBinary(path)
	if err != nil {
		return nil, err
	}
	return joinModulesMetadata(versions)
}

func listModulesInBinary(Path string) (versions []module.Version, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("ListModulesInGoBinary(Path='%s'): %w", Path, err)
		}
	}()
	depsBuildInfo, err := lichenmodule.Extract(context.Background(), Path)
	if err != nil {
		return nil, err
	}
	if len(depsBuildInfo) != 1 {
		return nil, fmt.Errorf("len(depsBuildInfo) should be 1, but found %v", len(depsBuildInfo))
	}
	versions = make([]module.Version, 0)
	for _, buildInfo := range depsBuildInfo {
		for _, ref := range buildInfo.ModuleRefs {
			versions = append(versions, module.Version{
				Path:    ref.Path,
				Version: ref.Version,
			})
		}
	}
	return versions, nil
}

func joinModulesMetadata(refs []module.Version) (modules []goutils.Module, err error) {
	localModules, err := goutils.ListModules()
	if err != nil {
		return
	}
	localModulesDict := goutils.BuildModuleDict(localModules)

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
