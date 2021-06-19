// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gocli

import (
	"sort"

	"golang.org/x/tools/go/packages"
)

// ListDeps lists direct and transitive module dependencies of the import path packages.
// It leverages golang.org/x/tools/go/packages under the hood.
func ListDeps(importPaths ...string) ([]Module, error) {
	// TODO(Bobgy): wrap error messages
	rootPkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedModule | packages.NeedImports | packages.NeedName,
	}, importPaths...)
	if err != nil {
		return nil, err
	}
	mods := make(map[string]*Module)
	packages.Visit(rootPkgs, func(p *packages.Package) bool {
		mod := newModule(p.Module)
		if mod != nil && mods[mod.Path] == nil {
			mods[mod.Path] = mod
		}
		return true
	}, nil)
	res := make([]Module, 0, len(mods))
	for _, mod := range mods {
		res = append(res, *mod)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Path < res[j].Path
	})
	return res, nil
}
