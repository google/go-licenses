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
	"time"

	"golang.org/x/tools/go/packages"
)

// Module provides module information for a package.
type Module struct {
	// Differences from packages.Module:
	// * Replace field is removed, it's only an implementation detail in this package.
	//   If a module is replaced, we'll directly return the replaced module.
	// * ModuleError field is removed, it's only used in packages.Module.
	Path      string     // module path
	Version   string     // module version
	Time      *time.Time // time version was created
	Main      bool       // is this the main module?
	Indirect  bool       // is this module only an indirect dependency of main module?
	Dir       string     // directory holding files for this module, if any
	GoMod     string     // path to go.mod file used when loading this module, if any
	GoVersion string     // go version used in module
}

func newModule(mod *packages.Module) *Module {
	if mod == nil {
		return nil
	}
	// Example of a module with replace directive: 	k8s.io/kubernetes => k8s.io/kubernetes v1.11.1
	// {
	//         "Path": "k8s.io/kubernetes",
	//         "Version": "v0.17.9",
	//         "Replace": {
	//                 "Path": "k8s.io/kubernetes",
	//                 "Version": "v1.11.1",
	//                 "Time": "2018-07-17T04:20:29Z",
	//                 "Dir": "/home/gongyuan_kubeflow_org/go/pkg/mod/k8s.io/kubernetes@v1.11.1",
	//                 "GoMod": "/home/gongyuan_kubeflow_org/go/pkg/mod/cache/download/k8s.io/kubernetes/@v/v1.11.1.mod"
	//         },
	//         "Dir": "/home/gongyuan_kubeflow_org/go/pkg/mod/k8s.io/kubernetes@v1.11.1",
	//         "GoMod": "/home/gongyuan_kubeflow_org/go/pkg/mod/cache/download/k8s.io/kubernetes/@v/v1.11.1.mod"
	// }
	// handle replace directives
	// Note, we specifically want to replace version field.
	// Haven't confirmed, but we may also need to override the
	// entire struct when using replace directive with local folders.
	tmp := *mod
	if tmp.Replace != nil {
		tmp = *tmp.Replace
	}
	return &Module{
		Path:      tmp.Path,
		Version:   tmp.Version,
		Time:      tmp.Time,
		Main:      tmp.Main,
		Indirect:  tmp.Indirect,
		Dir:       tmp.Dir,
		GoMod:     tmp.GoMod,
		GoVersion: tmp.GoVersion,
	}
}
