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

package dict

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/pkg/errors"
)

type LicenseDict map[string]*LicenseRecord

type LicenseRecord struct {
	Module       string
	DownaloadUrl string
	Type         string
	ShouldIgnore bool
}

const defaultDictLocation = "license_dict.csv"

func LoadLicenseRecords(r io.Reader) ([]*LicenseRecord, error) {
	reader := csv.NewReader(r)
	reader.Comment = '#'
	reader.FieldsPerRecord = 3
	rawRecords, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrapf(err, "Error when reading %s", defaultDictLocation)
	}
	records := make([]*LicenseRecord, 0)
	for index, raw := range rawRecords {
		record, err := parseRawRecord(raw)
		if err != nil {
			return nil, errors.Wrapf(err, "Record #%v with content '%s' is invalid ", index+1, strings.Join(raw, ","))
		}
		records = append(records, record)
	}
	return records, nil
}

func LoadLicenseDict(r io.Reader) (LicenseDict, error) {
	records, err := LoadLicenseRecords(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to load license records")
	}
	dict := make(LicenseDict)
	for _, record := range records {
		dict[record.Module] = record
	}
	return dict, nil
}

func parseRawRecord(raw []string) (*LicenseRecord, error) {
	if len(raw) != 3 {
		return nil, errors.Errorf("Invalid license record: 3 segments expected")
	}
	var record LicenseRecord
	record.Module = strings.TrimSpace(raw[0])
	if record.Module == "" {
		return nil, errors.Errorf("Empty module")
	}
	record.DownaloadUrl = strings.TrimSpace(raw[1])
	record.Type = strings.TrimSpace(raw[2])
	if record.Type == "Ignore" {
		record.ShouldIgnore = true
	}
	if !record.ShouldIgnore {
		if record.DownaloadUrl == "" {
			return nil, errors.Errorf("Empty download url")
		}
	}
	return &record, nil
}
