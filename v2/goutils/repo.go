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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-licenses/v2/ghutils"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

const (
	githubBase = "github.com/"
)

func GetGithubRepo(importPath string) (*ghutils.GitHubRepo, error) {
	if strings.HasPrefix(importPath, githubBase) {
		repo, err := ghutils.ParseGitHubUrl(importPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse repo: importPath=%q", importPath)
		}
		return repo, nil
	}

	repo, err := parseGoGet(importPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse using go-get: importPath=%q", importPath)
	}
	return repo, nil
}

func parseGoGet(module string) (*ghutils.GitHubRepo, error) {
	request := fmt.Sprintf("https://%s?go-get=1", module)
	resp, err := http.Get(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed sending request %s", request)
	}
	defer resp.Body.Close()
	// Some go modules like gonum.org/v1/gonum return a 404 as response, but it also has the meta tags.
	// if resp.StatusCode >= 400 {
	// 	return nil, errors.Errorf("Response status code %v not OK for request %s", resp.StatusCode, request)
	// }
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed reading response body for request %s", request)
	}
	bodyString := string(bodyBytes)
	// fmt.Println(bodyString)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader((bodyString)))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed creating document")
	}
	var vcs, repoRoot, sourceHome, sourceDir string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		content, _ := s.Attr("content")
		if name == "go-import" {
			// fmt.Println("go import content", goImportContent)
			segments := strings.Fields(content)
			// documentation: https://golang.org/cmd/go/#hdr-Remote_import_paths
			if len(segments) != 3 {
				klog.Warningf("Ignore invalid go-import - content wrong number of segments: %s", content)
				return
			}
			importPrefix := segments[0]
			thisVcs := segments[1]
			thisRepoRoot := segments[2]
			// Some go modules like gonum.org/v1/gonum returns meta tag for all modules in the org in one response,
			// so we have to find the correct go-import meta tag.
			if !strings.HasPrefix(module, importPrefix) {
				klog.Infof("Ignore go-import for a different module: %s", content)
				return
			}
			vcs = thisVcs
			repoRoot = thisRepoRoot
		} else if name == "go-source" {
			segments := strings.Fields(content)
			// documentation: https://github.com/golang/gddo/wiki/Source-Code-Links
			if len(segments) != 4 {
				klog.Warningf("Ignore invalid go-source - content wrong number of segments: %s", content)
				return
			}
			importPrefix := segments[0]
			// Some go modules like golang.org/x/net includes github repo in go-source home.
			home := segments[1]
			// Some go modules like gopkg.in/jcmturner/dnsutils.v1 includes github repo in go-source directory.
			dir := segments[2]
			if !strings.HasPrefix(module, importPrefix) {
				klog.Infof("Ignore go-source for a different module: %s", content)
				return
			}
			sourceHome = home
			sourceDir = dir
		}
	})
	if vcs == "" && repoRoot == "" {
		return nil, errors.Errorf("Cannot find go-import meta tag for %s in html: %s", module, bodyString)
	}
	if vcs != "git" {
		return nil, errors.Errorf("go-import vcs %s not supported", vcs)
	}
	repo, err := ghutils.ParseGitHubUrl(repoRoot)
	if err == nil {
		return repo, nil
	}
	repo, err2 := ghutils.ParseGitHubUrl(sourceHome)
	if err2 == nil {
		return repo, nil
	}
	repo, err3 := ghutils.ParseGitHubUrl(sourceDir)
	if err3 == nil {
		return repo, nil
	}
	return nil, errors.Errorf("Failed to parse github url from repoRoot '%s', sourceHome '%s', sourceDir '%s'", repoRoot, sourceHome, sourceDir)
}
