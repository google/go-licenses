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
	"fmt"

	"github.com/google/go-licenses/v2/third_party/go/runtime/debug"
)

// Module metadata extracted from binary and local go module workspace.
type BinaryMetadata struct {
	// The main module used to build the binary.
	// e.g. github.com//google/go-licenses/v2/tests/modules/cli02
	Main Module
	// Detailed metadata of all the module dependencies.
	// Does not include the main module.
	Deps []Module
}

// List dependencies from module metadata in a go binary.
// Modules with replace directives are returned as the replaced module instead.
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
func ExtractBinaryMetadata(path string) (*BinaryMetadata, error) {
	buildInfo, err := listModulesInBinary(path)
	if err != nil {
		return nil, err
	}
	main, deps, err := joinModulesMetadata(&buildInfo.Main, buildInfo.Deps)
	if err != nil {
		return nil, err
	}
	return &BinaryMetadata{
		Main: main,
		Deps: deps,
	}, nil
}

func listModulesInBinary(path string) (buildinfo *debug.BuildInfo, err error) {
	// TODO(Bobgy): replace with x/mod equivalent from https://github.com/golang/go/issues/39301
	// when it is available.
	buildinfo, err = version(path)
	if err != nil {
		err = fmt.Errorf("listModulesInGoBinary(path=%q): %w", path, err)
	}
	return buildinfo, nil
}

// joinModulesMetadata inner joins local go modules metadata with module ref
// extracted from the binary.
// The local go modules metadata is taken from calling `go list -m -json all`.
// Only those appeared in refs will be returned.
// An error is reported when we cannot find go module metadata for some refs,
// or when there's a version mismatch. These errors usually indicate your current
// working directory does not match exactly where the go binary is built.
func joinModulesMetadata(mainRef *debug.Module, refs []*debug.Module) (main Module, deps []Module, err error) {
	// Note, there was an attempt to use golang.org/x/tools/go/packages for
	// loading modules instead, but it fails for modules like golang.org/x/sys.
	// These modules only contains sub-packages, but no source code, so it
	// throws an error when using packages.Load.
	// More context: https://github.com/google/go-licenses/pull/71#issuecomment-890342154
	localModulesDict, err := ListModules()
	if err != nil {
		return main, nil, err
	}
	find := func(ref *debug.Module) (*Module, error) {
		if ref == nil {
			return nil, fmt.Errorf("ref is nil")
		}
		mod, ok := localModulesDict[ref.Path]
		if !ok {
			return nil, fmt.Errorf("Cannot find %v in current dir's go modules. Are you running this tool from the working dir to build the binary you are analyzing?", ref.Path)
		}
		if mod.Dir == "" {
			return nil, fmt.Errorf("Module %v's local directory is empty. Did you run `go mod download`?", ref.Path)
		}
		ver := ref.Version
		if ver == "(devel)" {
			// Main module's version will be (devel). We should expect an empty version when listing the module info.
			ver = ""
		}
		if ver != mod.Version {
			return nil, fmt.Errorf("Found %v@%v in go binary, but %v is downloaded in go modules. Are you running this tool from the working dir to build the binary you are analyzing?", ref.Path, ref.Version, mod.Version)
		}
		return &mod, nil
	}
	if mainRef != nil {
		found, err := find(mainRef)
		if err != nil {
			return main, nil, err
		}
		main = *found
	}

	for _, ref := range refs {
		found, err := find(ref)
		if err != nil {
			return main, nil, err
		}
		deps = append(deps, *found)
	}
	return main, deps, nil
}
