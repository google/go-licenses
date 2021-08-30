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
// Modules with replace directive are returned as the replaced module instead.
func ListModules() (map[string]Module, error) {
	out, err := exec.Command("go", "list", "-m", "-json", "all").Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to list go modules: %w", err)
	}
	// reference: https://github.com/golang/go/issues/27655#issuecomment-420993215
	modules := make([]Module, 0)

	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var tmp packages.Module
		if err := dec.Decode(&tmp); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Failed to read go list output: %w", err)
		}
		modules = append(modules, *newModule(&tmp))
	}

	dict := make(map[string]Module)
	for i := range modules {
		dict[modules[i].Path] = modules[i]
	}
	return dict, nil
}
