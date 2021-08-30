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
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/go-licenses/v2/third_party/go/runtime/debug"
)

// version runs `go version -m <path>` to get and parse build info of a go binary.
// The go binary must be built using go modules, otherwise it won't contain go
// modules information.
func version(path string) (*debug.BuildInfo, error) {
	out, err := exec.Command("go", "version", "-m", path).Output()
	if err != nil {
		cmd := strings.Join([]string{"go", "version", "-m", path}, " ")
		return nil, fmt.Errorf("%s failed: %w", cmd, err)
	}
	buildInfo, ok := debug.ParseBuildInfo(string(out))
	if !ok {
		return nil, fmt.Errorf("invalid build info in %s. Is it a go binary built using go modules?", path)
	}
	return buildInfo, nil
}
