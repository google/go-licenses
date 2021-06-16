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

package ghutils_test

import (
	"testing"

	"github.com/google/go-licenses/v2/ghutils"
)

func TestGithubRepoRemoteUrl(t *testing.T) {
	repo := ghutils.GitHubRepo{
		Owner: "googleapis",
		Name:  "google-cloud-go",
	}
	cases := []struct {
		args     ghutils.RemoteUrlArgs
		expected string
	}{
		{
			args: ghutils.RemoteUrlArgs{
				Path:      "cmd/go-cloud-debug-agent/internal/debug/elf/elf.go",
				Version:   "v0.72.0",
				LineStart: 1,
				LineEnd:   43,
			},
			expected: "https://github.com/googleapis/google-cloud-go/blob/v0.72.0/cmd/go-cloud-debug-agent/internal/debug/elf/elf.go#L1-L43",
		},
		{
			args: ghutils.RemoteUrlArgs{
				Path:    "LICENSE",
				Version: "v0.0.0-20210108172934-dcfadaf1a8b1",
			},
			expected: "https://github.com/googleapis/google-cloud-go/blob/dcfadaf1a8b1/LICENSE",
		},
		{
			args: ghutils.RemoteUrlArgs{
				Path:    "LICENSE",
				Version: "v0.0.0-pre.0.20210108172934-dcfadaf1a8b1",
			},
			expected: "https://github.com/googleapis/google-cloud-go/blob/dcfadaf1a8b1/LICENSE",
		},
	}
	for _, tt := range cases {
		got, err := repo.RemoteUrl(tt.args)
		if err != nil {
			t.Errorf("repo.RemoteUrl(%+v) failed: %w", tt.args, err)
		}
		if got != tt.expected {
			t.Errorf("repo.RemoteUrl(%+v) got %q, expected %q", tt.args, got, tt.expected)
		}
	}
}

func TestGithubDownloadUrl(t *testing.T) {
	cases := []struct {
		url         string
		downloadUrl string
		lineStart   int
		lineEnd     int
	}{
		{
			url:         "https://github.com/sergi/go-diff/blob/v1.1.0/LICENSE",
			downloadUrl: "https://github.com/sergi/go-diff/raw/v1.1.0/LICENSE",
		},
		{
			url:         "https://github.com/sergi/go-diff/blob/v1.1.0/LICENSE#L3-L8",
			downloadUrl: "https://github.com/sergi/go-diff/raw/v1.1.0/LICENSE",
			lineStart:   3,
			lineEnd:     8,
		},
	}
	for _, tt := range cases {
		downloadUrl, lineStart, lineEnd, err := ghutils.GithubDownloadUrl(tt.url)
		if err != nil {
			t.Errorf("GithubDownloadUrl(%q) failed: %w", tt.url, err)
		}
		if downloadUrl != tt.downloadUrl || lineStart != tt.lineStart || lineEnd != tt.lineEnd {
			t.Errorf("GithubDownloadUrl(%q) got downloadUrl=%q lineStart=%v lineEnd=%v, expected %+v", tt.url, downloadUrl, lineStart, lineEnd, tt)
		}
	}
}
