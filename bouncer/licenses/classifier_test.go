// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package licenses

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Useful in other tests in this package
type classifierStub struct {
	licenses map[string][]License
	errors   map[string]error
}

func (c classifierStub) Identify(licensePath string) ([]License, error) {
	// Convert licensePath to relative path for tests.
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	relPath, err := filepath.Rel(wd, licensePath)
	if err != nil {
		return nil, err
	}
	if licenses, ok := c.licenses[relPath]; ok {
		return licenses, c.errors[relPath]
	}
	if err := c.errors[relPath]; err != nil {
		return nil, c.errors[relPath]
	}
	return nil, fmt.Errorf("classifierStub has no programmed response for %q", relPath)
}

func TestIdentify(t *testing.T) {
	for _, test := range []struct {
		desc         string
		file         string
		confidence   float64
		wantLicenses []License
		wantType     Type
		wantErr      bool
	}{
		{
			desc:       "Apache 2.0 license",
			file:       "testdata/LICENSE",
			confidence: 1,
			wantLicenses: []License{
				{
					Name: "Apache-2.0",
					Type: Notice,
				},
			},
		},
		{
			desc:       "MIT license",
			file:       "testdata/MIT/LICENSE.MIT",
			confidence: 1,
			wantLicenses: []License{
				{
					Name: "MIT",
					Type: Notice,
				},
			},
		},
		{
			desc:       "non-existent file",
			file:       "non-existent-file",
			confidence: 1,
			wantErr:    true,
		},
		{
			desc:         "empty file path",
			file:         "",
			confidence:   1,
			wantLicenses: nil,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			c, err := NewClassifier()
			if err != nil {
				t.Fatalf("NewClassifier(%v) = (_, %q), want (_, nil)", test.confidence, err)
			}

			gotLicenses, err := c.Identify(test.file)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Fatalf("c.Identify(%q) = (_, _, %q), want err? %t", test.file, err, test.wantErr)
			} else if gotErr {
				return
			}

			if !reflect.DeepEqual(gotLicenses, test.wantLicenses) {
				t.Fatalf("c.Identify(%q) = %q, want %q", test.file, gotLicenses, test.wantLicenses)
			}
		})
	}
}
