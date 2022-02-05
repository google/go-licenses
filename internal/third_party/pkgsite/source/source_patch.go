// Copyright 2022 Google LLC
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

package source

import (
	"path"
	"strings"
)

// This file includes all local additions to source package for google/go-licenses use-cases.

// SetCommit overrides commit to a specified commit. Usually, you should pass your version to
// ModuleInfo(). However, when you do not know the version and just wants a link that points to
// a known commit/branch/tag. You can use this method to directly override the commit like
// info.SetCommit("master").
//
// Note this is different from directly passing "master" as version to ModuleInfo(), because for
// modules not at the root of a repo, there are conventions that add a module's relative dir in
// front of the version as the actual git tag. For example, for a sub module at ./submod whose
// version is v1.0.1, the actual git tag should be submod/v1.0.1.
func (i *Info) SetCommit(commit string) {
	if i == nil {
		return
	}
	i.commit = commit
}

// RepoFileURL returns a URL for a file whose pathname is relative to the repo's home directory instead of the module's.
func (i *Info) RepoFileURL(pathname string) string {
	if i == nil {
		return ""
	}
	dir, base := path.Split(pathname)
	return expand(i.templates.File, map[string]string{
		"repo":       i.repoURL,
		"importPath": path.Join(strings.TrimPrefix(i.repoURL, "https://"), dir),
		"commit":     i.commit,
		"dir":        dir,
		"file":       pathname,
		"base":       base,
	})
}

// RepoRawURL returns a URL referring to the raw contents of a file relative to the
// repo's home directory instead of the module's.
func (i *Info) RepoRawURL(pathname string) string {
	if i == nil {
		return ""
	}
	// Some templates don't support raw content serving.
	if i.templates.Raw == "" {
		return ""
	}
	return expand(i.templates.Raw, map[string]string{
		"repo":   i.repoURL,
		"commit": i.commit,
		"file":   pathname,
	})
}
