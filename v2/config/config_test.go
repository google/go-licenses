package config_test

import (
	"testing"

	"github.com/google/go-licenses/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_DefaultPath(t *testing.T) {
	_, err := config.Load("")
	// default path is current folder, so it doesn't exist
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestLoadConfig_SpecifiedPath(t *testing.T) {
	loaded, err := config.Load("testdata/1.yaml")
	require.Nil(t, err)
	assert.Equal(t, ".cache/licenses", loaded.Module.LicenseDB.Path)
	assert.Equal(t, "go-licenses", loaded.Module.Go.Binary.Path)
	assert.Equal(t, ".", loaded.Module.Go.Path)
	assert.Equal(t, "github.com/google/go-licenses/v2", loaded.Module.Go.Module)
	assert.Equal(t, "license_info.csv", loaded.Module.Csv.Path)
	assert.Equal(t, "third_party/NOTICES", loaded.Module.Notices.Path)
	expected := []config.ModuleOverride{
		{
			Name:         "github.com/google/go-licenses/v2",
			Version:      "",
			License:      config.LicenseOverride{Path: "LICENSE", SpdxId: "Apache-2.0", Url: "https://github.com/google/go-licenses/v2/dummy-url"},
			ExcludePaths: []string{"go-licenses"},
		}, {
			Name:    "github.com/aws/aws-sdk-go",
			Version: "v1.36.1",
			License: config.LicenseOverride{Path: "LICENSE.txt", SpdxId: "Apache-2.0"},
			SubModules: []config.SubModule{
				{
					Path:    "internal/sync/singleflight",
					License: config.LicenseOverride{Path: "LICENSE", SpdxId: "BSD-3-Clause"},
				},
			},
		},
		{
			Name:         "github.com/google/licenseclassifier",
			ExcludePaths: []string{"licenses"},
		},
		{
			Name:    "cloud.google.com/go",
			Version: "v0.72.0",
			License: config.LicenseOverride{
				Path:   "LICENSE",
				SpdxId: "Apache-2.0",
			},
			SubModules: []config.SubModule{
				{
					Path: "cmd/go-cloud-debug-agent/internal/debug/elf",
					License: config.LicenseOverride{
						Path:      "elf.go",
						SpdxId:    "BSD-2-Clause",
						LineStart: 1,
						LineEnd:   43,
					},
				}, {
					Path:    "third_party/pkgsite",
					License: config.LicenseOverride{Path: "LICENSE", SpdxId: "BSD-3-Clause"},
				},
			},
		},
	}
	assert.Equal(t, expected, loaded.Module.Overrides)
}

func TestLoadConfig_PathNotExist(t *testing.T) {
	_, err := config.Load("file-not-exist")
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestLoadConfig_ErrorOnTypo(t *testing.T) {
	// there is a typo in the config yaml, so we have unknown fields
	_, err := config.Load("testdata/typo.yaml")
	require.NotNil(t, err, "should report error when config has unknown fields")
}
