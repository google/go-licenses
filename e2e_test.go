// Copyright 2022 Google LLC
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

package main_test

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update golden files")

func TestReportCommandE2E(t *testing.T) {
	tests := []struct {
		workdir        string
		args           []string // additional arguments to pass to report command.
		goldenFilePath string
	}{
		{"testdata/modules/hello01", nil, "licenses.csv"},
		{"testdata/modules/cli02", nil, "licenses.csv"},
		{"testdata/modules/vendored03", nil, "licenses.csv"},
		{"testdata/modules/replace04", nil, "licenses.csv"},

		{"testdata/modules/hello01", []string{"--template", "licenses.tpl"}, "licenses.md"},
	}

	originalWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWorkDir) })

	// This builds go-licenses CLI to temporary dir.
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	goLicensesPath := filepath.Join(tempDir, "go-licenses")
	cmd := exec.Command("go", "build", "-o", goLicensesPath)
	_, err = cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Built go-licenses binary in %s.", goLicensesPath)

	for _, tt := range tests {
		t.Run(tt.workdir, func(t *testing.T) {
			err := os.Chdir(filepath.Join(originalWorkDir, tt.workdir))
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "mod", "download")
			log, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("downloading go modules:\n%s", string(log))
			}
			args := append([]string{"report", "."}, tt.args...)
			cmd = exec.Command(goLicensesPath, args...)
			// Capture stderr to buffer.
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			t.Logf("%s $ go-licenses report .", tt.workdir)
			output, err := cmd.Output()
			if err != nil {
				t.Logf("\n=== start of log ===\n%s=== end of log ===\n\n\n", stderr.String())
				t.Fatalf("running go-licenses report: %s. Full log shown above.", err)
			}
			got := string(output)
			if *update {
				err := os.WriteFile(tt.goldenFilePath, output, 0600)
				if err != nil {
					t.Fatalf("writing golden file: %s", err)
				}
			}
			goldenBytes, err := os.ReadFile(tt.goldenFilePath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					t.Fatalf("reading golden file: %s. Create a golden file by running `go test --update .`", err)
				}
				t.Fatalf("reading golden file: %s", err)
			}
			golden := string(goldenBytes)
			if got != golden {
				t.Logf("\n=== start of log ===\n%s=== end of log ===\n\n\n", stderr.String())
				t.Fatalf("result of go-licenses report does not match the golden file.\n"+
					"Diff -golden +got:\n%s\n"+
					"Update the golden by running `go test --update .`",
					cmp.Diff(golden, got))
			}
		})
	}
}

func TestCheckCommandE2E(t *testing.T) {
	tests := []struct {
		workdir        string
		args           []string // additional arguments to pass to report command.
		goldenFilePath string
		wantExitCode   int
	}{
		{"testdata/modules/hello01", nil, "output-check-forbidden.txt", 0},
		{"testdata/modules/hello01", []string{"--exclude-notice=true"}, "output-check-notice-forbidden.txt", 1},
		{"testdata/modules/cli02", nil, "output-check-forbidden.txt", 0},
		{"testdata/modules/cli02", []string{"--exclude-notice=true"}, "output-check-notice-forbidden.txt", 1},
		{"testdata/modules/cli02", []string{"--allowed-license-names=Apache-2.0"}, "output-check-license-names-1.txt", 1},
		{"testdata/modules/cli02", []string{"--allowed-license-names=Apache-2.0,MIT"}, "output-check-license-names-2.txt", 1},
	}

	originalWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWorkDir) })

	// This builds go-licenses CLI to temporary dir.
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	goLicensesPath := filepath.Join(tempDir, "go-licenses")
	cmd := exec.Command("go", "build", "-o", goLicensesPath)
	_, err = cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Built go-licenses binary in %s.", goLicensesPath)

	for _, tt := range tests {
		t.Run(tt.workdir, func(t *testing.T) {
			err := os.Chdir(filepath.Join(originalWorkDir, tt.workdir))
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "mod", "download")
			log, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("downloading go modules:\n%s", string(log))
			}
			args := append([]string{"check", "."}, tt.args...)
			cmd = exec.Command(goLicensesPath, args...)
			// Capture stderr to buffer.
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			exitCode := 0
			t.Logf("%s $ go-licenses check .", tt.workdir)
			output, err := cmd.Output()
			if err != nil {
				exitCode = -1

				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if len(output) != 0 {
				t.Fatalf("unexpected output running go-licenses check: %s.", string(output))
			}

			if exitCode != tt.wantExitCode {
				t.Fatalf("unexpected exit code running go-licenses check, expected %d but got %d", tt.wantExitCode, exitCode)
			}

			got := filterOutput(stderr.String())
			if *update {
				err := os.WriteFile(tt.goldenFilePath, []byte(got), 0600)
				if err != nil {
					t.Fatalf("writing golden file: %s", err)
				}
			}
			goldenBytes, err := os.ReadFile(tt.goldenFilePath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					t.Fatalf("reading golden file: %s. Create a golden file by running `go test --update .`", err)
				}
				t.Fatalf("reading golden file: %s", err)
			}
			golden := string(goldenBytes)
			if got != golden {
				t.Logf("\n=== start of log ===\n%s=== end of log ===\n\n\n", stderr.String())
				t.Fatalf("result of go-licenses check does not match the golden file.\n"+
					"Diff -golden +got:\n%s\n"+
					"Update the golden by running `go test --update .`",
					cmp.Diff(golden, got))
			}
		})
	}
}

func filterOutput(output string) string {
	output = regexp.MustCompile(`(?m)W\d+.*\n`).
		ReplaceAllString(output, "")

	output = regexp.MustCompile(`(?m)^/.*\n`).
		ReplaceAllString(output, "")

	return output
}
