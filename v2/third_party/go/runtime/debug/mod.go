// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"strings"
)

// ParseBuildInfo parses build info from output of go command `go version -m <go_binary_path>`.
// Caveat, `go version -m <folder_path>` can be used to get build info of all the go
// binaries inside a folder. The current implementation cannot correctly parse
// such outputs with multiple binaries yet.
func ParseBuildInfo(data string) (info *BuildInfo, ok bool) {
	return parseBuildInfo(data)
}

// BuildInfo represents the build information read from
// the binary.
type BuildInfo struct {
	Path string    // The main package path
	Main Module    // The module containing the main package
	Deps []*Module // Module dependencies
}

// Module represents a module.
type Module struct {
	Path    string  // module path
	Version string  // module version
	Sum     string  // checksum
	Replace *Module // replaced by this module
}

func parseBuildInfo(data string) (*BuildInfo, bool) {
	// Example data (note, separators are \t):
	//
	// tests/modules/cli02/main: go1.16.5
	//     path    github.com/google/go-licenses/v2/tests/modules/cli02
	//     mod     github.com/google/go-licenses/v2/tests/modules/cli02    (devel)
	//     dep     github.com/fsnotify/fsnotify    v1.4.9  h1:hsms1Qyu0jgnwNXIxa+/V/PDsU6CfLf6CNO8H7IWoS4=
	//     dep     github.com/hashicorp/hcl        v1.0.0  h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=
	//     dep     github.com/magiconair/properties        v1.8.5  h1:b6kJs+EmPFMYGkow9GiUyCyOvIwYetYJ3fSaWak/Gls=
	//     dep     golang.org/x/sys        v0.0.0-20210510120138-977fb7262007      h1:gG67DSER+11cZvqIMb8S8bt0vZtiN6xWYARwirrOSfE=
	const (
		pathLine = "\tpath\t"
		modLine  = "\tmod\t"
		depLine  = "\tdep\t"
		repLine  = "\t=>\t"
	)

	readEntryFirstLine := func(elem []string) (Module, bool) {
		if len(elem) != 2 && len(elem) != 3 {
			return Module{}, false
		}
		sum := ""
		if len(elem) == 3 {
			sum = elem[2]
		}
		return Module{
			Path:    elem[0],
			Version: elem[1],
			Sum:     sum,
		}, true
	}

	var (
		info = &BuildInfo{}
		last *Module
		line string
		ok   bool
	)
	// Reverse of cmd/go/internal/modload.PackageBuildInfo
	for len(data) > 0 {
		i := strings.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		line, data = data[:i], data[i+1:]
		switch {
		case strings.HasPrefix(line, pathLine):
			elem := line[len(pathLine):]
			info.Path = elem
		case strings.HasPrefix(line, modLine):
			elem := strings.Split(line[len(modLine):], "\t")
			last = &info.Main
			*last, ok = readEntryFirstLine(elem)
			if !ok {
				return nil, false
			}
		case strings.HasPrefix(line, depLine):
			elem := strings.Split(line[len(depLine):], "\t")
			last = new(Module)
			info.Deps = append(info.Deps, last)
			*last, ok = readEntryFirstLine(elem)
			if !ok {
				return nil, false
			}
		case strings.HasPrefix(line, repLine):
			elem := strings.Split(line[len(repLine):], "\t")
			if len(elem) != 3 {
				return nil, false
			}
			if last == nil {
				return nil, false
			}
			last.Replace = &Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     elem[2],
			}
			last = nil
		}
	}
	return info, true
}
