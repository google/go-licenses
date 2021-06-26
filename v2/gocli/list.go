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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"golang.org/x/tools/go/packages"
)

// List go modules with metadata in workdir using go CLI list command.
func ListModules() (map[string]packages.Module, error) {
	out, err := exec.Command("go", "list", "-m", "-json", "all").Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to list go modules: %w", err)
	}
	// reference: https://github.com/golang/go/issues/27655#issuecomment-420993215
	modules := make([]packages.Module, 0)

	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var m packages.Module
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Failed to read go list output: %w", err)
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
		if m.Replace != nil {
			m = *m.Replace
		}
		modules = append(modules, m)
	}

	dict := make(map[string]packages.Module)
	for i := range modules {
		dict[modules[i].Path] = modules[i]
	}
	return dict, nil
}
