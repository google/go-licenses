package bouncer

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type ModuleInfo struct {
	Path    string
	Dir     string
	Version string
}

func ListModules(path string) ([]ModuleInfo, error) {
	var results []ModuleInfo
	cmd := exec.Command("go", "list", "-f", "{{.Path}} {{.Dir}} {{.Version}}", "-m", "all")
	if path != "" {
		cmd.Path = path
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unable to list modules: %q : %w", out, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		text := scanner.Text()
		entries := strings.Split(text, " ")
		if len(entries) != 3 {
			return nil, fmt.Errorf("bad listing output: %q", text)
		}
		results = append(results, ModuleInfo{
			Path:    entries[0],
			Dir:     entries[1],
			Version: entries[2],
		})
	}
	return results, nil
}
