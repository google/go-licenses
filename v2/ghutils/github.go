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

package ghutils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type GitHubRepo struct {
	Owner string
	Name  string
}

const (
	githubBase = "github.com/"
	protocol   = "https://"
	gitSuffix  = ".git"
)

func ParseGitHubUrl(githubUrl string) (*GitHubRepo, error) {
	if strings.HasPrefix(githubUrl, protocol) {
		githubUrl = githubUrl[len(protocol):]
	}
	if !strings.HasPrefix(githubUrl, githubBase) {
		return nil, errors.Errorf("%s is not github url", githubUrl)
	}
	githubUrl = githubUrl[len(githubBase):]
	if strings.HasSuffix(githubUrl, gitSuffix) {
		githubUrl = githubUrl[:len(githubUrl)-len(gitSuffix)]
	}
	segments := strings.Split(githubUrl, "/")
	if len(segments) < 2 {
		return nil, errors.Errorf("Too few segments in github url: %s", githubUrl)
	}
	return &GitHubRepo{Owner: segments[0], Name: segments[1]}, nil
}

type RemoteUrlArgs struct {
	Path      string
	Version   string
	Raw       bool
	LineStart int
	LineEnd   int
}

func (repo *GitHubRepo) RemoteUrl(args RemoteUrlArgs) (string, error) {
	if (args.LineStart != 0) != (args.LineEnd != 0) {
		return "", fmt.Errorf("GitHubRepo.RemoteUrl(%+v): LineStart and LineEnd must be specified at the same time.", args)
	}
	if args.LineStart < 0 || args.LineEnd < 0 {
		return "", fmt.Errorf("GitHubRepo.RemoteUrl(%+v): LineStart and LineEnd must be positive integers when specified, 1 means first line.", args)
	}
	if repo == nil {
		// return local path when repo not available
		return args.Path, nil
	}
	template := "https://github.com/%s/%s/blob/%s/%s"
	if args.Raw {
		template = "https://github.com/%s/%s/raw/%s/%s"
	}
	url := fmt.Sprintf(
		template,
		repo.Owner,
		repo.Name,
		parseGoModulePseudoVersion(args.Version),
		args.Path)
	if args.LineStart > 0 {
		if args.Raw {
			return "", fmt.Errorf("GitHubRepo.RemoteUrl(%+v): LineStart and LineEnd not supported for url to raw content.", args)
		}
		url = url + fmt.Sprintf("#L%v-L%v", args.LineStart, args.LineEnd)
	}
	return url, nil
}

// Reference: https://golang.org/ref/mod#pseudo-versions
// vX.0.0-yyyymmddhhmmss-abcdefabcdef is used when there is no known base version. As with all versions, the major version X must match the module's major version suffix.
// vX.Y.Z-pre.0.yyyymmddhhmmss-abcdefabcdef is used when the base version is a pre-release version like vX.Y.Z-pre.
// vX.Y.(Z+1)-0.yyyymmddhhmmss-abcdefabcdef is used when the base version is a release version like vX.Y.Z. For example, if the base version is v1.2.3, a pseudo-version might be v1.2.4-0.20191109021931-daa7c04131f5.
var psuedoVersionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+-.*[0-9]{14}-(?P<commit>[a-f0-9]{12})$`)

func parseGoModulePseudoVersion(version string) string {
	if version == "" {
		return "main" // default to content in main branch
	}
	// Parse version like v0.0.0-20210108172934-df6aa8a2788b to commit hash:
	// df6aa8a2788b.
	matches := psuedoVersionPattern.FindStringSubmatch(version)
	if len(matches) == 2 {
		// matches[0] is regex match, matches[1] is result of the capture group.
		return matches[1]
	}
	return version
}

var githubUrlPattern = regexp.MustCompile(`^(https://)?(www\.)?github.com/(?P<repo>[^/]+/[^/]+)/blob/(?P<path>[^#]*)(?P<hash>#.*)?$`)
var githubLinePattern = regexp.MustCompile(`^#L(?P<linestart>[0-9]+)-L(?P<lineend>[0-9]+)$`)

func GithubDownloadUrl(url string) (downloadUrl string, lineStart int, lineEnd int, err error) {
	matches := githubUrlPattern.FindStringSubmatch(url)
	if len(matches) > 0 {
		repo := matches[3]
		path := matches[4]
		hash := matches[5]
		if hash == "" {
			return fmt.Sprintf("https://github.com/%s/raw/%s", repo, path), 0, 0, nil
		}
		lineMatches := githubLinePattern.FindStringSubmatch(hash)
		if len(lineMatches) == 0 {
			return "", 0, 0, fmt.Errorf("getGithubDownloadUrl(%q): cannot find line numbers in hash", url)
		}
		// line start and line end included
		lineStart, err := strconv.ParseInt(lineMatches[1], 10, 0)
		if err != nil {
			return "", 0, 0, err
		}
		lineEnd, err := strconv.ParseInt(lineMatches[2], 10, 0)
		if err != nil {
			return "", 0, 0, err
		}
		return fmt.Sprintf("https://github.com/%s/raw/%s", repo, path), int(lineStart), int(lineEnd), nil
	}
	return "", 0, 0, nil
}

// TODO: this downloads url content in memory.
// We might need optimization in the future.
func SmartDownload(url string) (string, error) {
	wrap := func(err error) error {
		return fmt.Errorf("SmartDownload(%q): %w", url, err)
	}
	downloadUrl, lineStart, lineEnd, err := GithubDownloadUrl(url)
	if err != nil {
		return "", wrap(err)
	}
	if downloadUrl == "" {
		// if not detected, use original url to download
		downloadUrl = url
	}
	content, err := download(downloadUrl)
	if err != nil {
		return "", wrap(err)
	}
	if content == "" {
		return "", wrap(fmt.Errorf("downloaded content is empty"))
	}
	if lineStart == 0 {
		return content, nil
	}
	if lineEnd == 0 {
		return "", wrap(fmt.Errorf("lineEnd must be non zero when lineStart isn't"))
	}
	lines := strings.Split(content, "\n")
	if lineEnd >= len(lines) {
		return "", wrap(fmt.Errorf("total %v lines, but lineEnd=%v", len(lines), lineEnd))
	}
	// lineStart start from 1, so we convert to start from 0.
	return strings.Join(lines[lineStart-1:lineEnd], "\n"), nil
}

func download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download(%q): %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("download(%q) response status code %v not OK", url, resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("download(%q) failed to read from response body: %w", url, err)
	}
	return string(bodyBytes), nil
}
