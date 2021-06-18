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

package config

import (
	"fmt"
	"io/ioutil"

	"github.com/google/licenseclassifier"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type GoModLicensesConfig struct {
	Module struct {
		LicenseDB LicenseDB        `yaml:"licenseDB"`
		Csv       CsvConfig        `yaml:"csv"`
		Notices   NoticesConfig    `yaml:"notices"`
		Go        GoModuleConfig   `yaml:"go"`
		Overrides []ModuleOverride `yaml:"overrides"`
	} `yaml:"module"`
	Licenses LicensesConfig `yaml:"licenses"`
}

type LicenseDB struct {
	Path string `yaml:"path"`
}

type GoModuleConfig struct {
	Module  string `yaml:"module"`  // module name, e.g. github.com/google/go-licenses/v2
	Version string `yaml:"version"` // main module version, e.g. master (defaults to main)
	Path    string `yaml:"path"`    // local path where the go module lives in
	Binary  struct {
		Path string `yaml:"path"` // local path where the go binary lives in
	} `yaml:"binary"`
}

type CsvConfig struct {
	Path string `yaml:"path"` // local path where the csv lives in, optional. Defaults to license_info.csv.
}

type NoticesConfig struct {
	Path string `yaml:"path"`
}

type ModuleOverride struct {
	Name string `yaml:"name"`
	// optional, if specified, the override is pinned to a version. After an
	// upgrade, you need to confirm the module again and pin to the new version.
	Version      string          `yaml:"version"`
	Skip         bool            `yaml:"skip"`
	License      LicenseOverride `yaml:"license"`    // required, license of root module
	SubModules   []SubModule     `yaml:"subModules"` // optional, specify if sub modules have a different license
	ExcludePaths []string        `yaml:"excludePaths"`
}

type LicenseOverride struct {
	Path      string `yaml:"path"`      // required, a license must map to a local file
	SpdxId    string `yaml:"spdxId"`    // required, TODO: make this optional
	Url       string `yaml:"url"`       // optional, license file public url (recommend using url for raw file)
	LineStart int    `yaml:"lineStart"` // optional, start line of license in the file. The first line is 1.
	LineEnd   int    `yaml:"lineEnd"`   // optional, end line of license in the file. The first line is 1.
}

type SubModule struct {
	Path    string          `yaml:"path"`    // required, path of sub module
	License LicenseOverride `yaml:"license"` // required, path of license in sub module
}

type LicenseTypeOverride struct {
	SpdxId string `yaml:"spdxId"` // required, SPDX ID of the license. Refer to https://spdx.org/licenses/.
	// required, should be one of https://github.com/google/licenseclassifier/blob/df6aa8a2788bdf5ac382148c2453a407a29819b8/license_type.go#L367-L374
	Type string `yaml:"type"`
}

type LicensesConfig struct {
	Types LicenseTypes `yaml:"types"`
}

type LicenseTypes struct {
	Overrides []LicenseTypeOverride `yaml:"overrides"`
}

const (
	DefaultConfigPath = "go-licenses.yaml"
)

func Load(path string) (config *GoModLicensesConfig, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "Failed to load config from %s", path)
		}
	}()
	if path == "" {
		path = DefaultConfigPath
	}
	// set defaults
	config = &GoModLicensesConfig{}

	// load config from file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.UnmarshalStrict(data, config)
	if err != nil {
		return nil, err
	}
	if config.Module.Go.Binary.Path == "" {
		return nil, errors.Errorf("goBinary.path is required")
	}
	if config.Module.Go.Version == "" {
		config.Module.Go.Version = "main"
	}
	for i, licenseOverride := range config.Licenses.Types.Overrides {
		if licenseOverride.SpdxId == "" {
			return nil, fmt.Errorf("config.licenses.types.overrides[%v]: license override's spdxId must be non empty", i)
		}
		if !licenseclassifier.LicenseTypes.Contains(licenseOverride.Type) {
			return nil, fmt.Errorf("license override spdxId=%q type=%q is invalid: type must be one of %v", licenseOverride.SpdxId, licenseOverride.Type, licenseclassifier.LicenseTypes.String())
		}
	}
	return config, nil
}
