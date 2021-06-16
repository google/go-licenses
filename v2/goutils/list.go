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

package goutils

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

type Module struct {
	Path      string // go import path
	Version   string
	Time      string
	Indirect  bool
	Dir       string
	GoMod     string
	GoVersion string
}

func ListModules() ([]Module, error) {
	out, err := exec.Command("go", "list", "-m", "-json", "all").Output()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list go modules")
	}
	// reference: https://github.com/golang/go/issues/27655#issuecomment-420993215
	modules := make([]Module, 0)

	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var m Module
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrapf(err, "Failed to read go list output")
		}
		modules = append(modules, m)
	}
	return modules, nil
}

func BuildModuleDict(modules []Module) map[string]Module {
	dict := make(map[string]Module)
	for i := range modules {
		dict[modules[i].Path] = modules[i]
	}
	return dict
}
