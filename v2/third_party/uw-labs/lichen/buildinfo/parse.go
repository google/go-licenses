package buildinfo

import (
	"fmt"
	"strings"

	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/model"
)

// Parse parses build info details as returned by `go version -m [bin ...]`
func Parse(info string) ([]model.BuildInfo, error) {
	var (
		lines       = strings.Split(info, "\n")
		results     = make([]model.BuildInfo, 0)
		current     model.BuildInfo
		replacement bool
	)
	for _, l := range lines {
		// ignore blank lines
		if l == "" {
			continue
		}

		// start of new build info output
		if !strings.HasPrefix(l, "\t") {
			parts := strings.Split(l, ":")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid version line: %s", l)
			}
			version := strings.TrimSpace(parts[len(parts)-1])
			path := strings.Join(parts[:len(parts)-1], ":")
			switch {
			case version == "not executable file":
				return nil, fmt.Errorf("%s is not an executable", parts[0])
			case version == "unrecognized executable format":
				return nil, fmt.Errorf("%s has an unrecognized executable format", parts[0])
			case version == "go version not found":
				return nil, fmt.Errorf("%s does not appear to be a Go compiled binary", parts[0])
			case strings.HasPrefix(version, "go"):
				// sensible looking
			default:
				return nil, fmt.Errorf("unrecognised version line: %s", l)
			}
			if current.Path != "" {
				results = append(results, current)
			}
			current = model.BuildInfo{Path: path}
			continue
		}

		// inside build info output
		parts := strings.Split(l, "\t")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid build info line: %s", l)
		}
		if replacement {
			if parts[1] != "=>" {
				return nil, fmt.Errorf("expected path replacement, received: %s", l)
			}
			replacement = false
		}
		switch parts[1] {
		case "path":
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid path line: %s", l)
			}
			current.PackagePath = parts[2]
		case "mod":
			if len(parts) != 5 {
				return nil, fmt.Errorf("invalid mod line: %s", l)
			}
			current.ModulePath = parts[2]
		case "dep", "=>":
			switch len(parts) {
			case 5:
				current.ModuleRefs = append(current.ModuleRefs, model.ModuleReference{
					Path:    parts[2],
					Version: parts[3],
				})
			case 4:
				replacement = true
			default:
				return nil, fmt.Errorf("invalid dep line: %s", l)
			}
		default:
			return nil, fmt.Errorf("unrecognised line: %s", l)
		}
	}
	if current.Path != "" {
		results = append(results, current)
	}
	return results, nil
}
