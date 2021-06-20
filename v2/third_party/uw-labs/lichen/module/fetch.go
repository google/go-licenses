package module

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/model"
	"github.com/hashicorp/go-multierror"
)

func Fetch(ctx context.Context, refs []model.ModuleReference) ([]model.Module, error) {
	if len(refs) == 0 {
		return []model.Module{}, nil
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	// tempDir, err := ioutil.TempDir("", "lichen")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create temp directory: %w", err)
	// }
	// defer os.Remove(tempDir)

	args := []string{"mod", "download", "-json"}
	for _, ref := range refs {
		if !ref.IsLocal() {
			args = append(args, ref.String())
		}
	}

	cmd := exec.CommandContext(ctx, goBin, args...)
	// cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w (output: %s)", err, string(out))
	}

	// parse JSON output from `go mod download`
	modules := make([]model.Module, 0)
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var m model.Module
		if err := dec.Decode(&m); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		modules = append(modules, m)
	}

	// add local modules, as these won't be included in the set returned by `go mod download`
	for _, ref := range refs {
		if ref.IsLocal() {
			modules = append(modules, model.Module{
				ModuleReference: ref,
			})
		}
	}

	// sanity check: all modules should have been covered in the output from `go mod download`
	if err := verifyFetched(modules, refs); err != nil {
		return nil, fmt.Errorf("failed to fetch all modules: %w", err)
	}

	return modules, nil
}

func verifyFetched(fetched []model.Module, requested []model.ModuleReference) (err error) {
	fetchedRefs := make(map[model.ModuleReference]struct{}, len(fetched))
	for _, module := range fetched {
		fetchedRefs[module.ModuleReference] = struct{}{}
	}
	for _, ref := range requested {
		if _, found := fetchedRefs[ref]; !found {
			err = multierror.Append(err, fmt.Errorf("module %s could not be resolved", ref))
		}
	}
	return
}
