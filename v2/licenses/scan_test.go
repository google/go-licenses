package licenses_test

import (
	"testing"

	"github.com/google/go-licenses/v2/licenses"
	"github.com/stretchr/testify/assert"
)

const DbPath = "../third_party/google/licenseclassifier/licenses"

func TestScan_ThisRepo(t *testing.T) {
	found, err := licenses.ScanDir(
		"..", // repo root
		licenses.ScanDirOptions{
			DbPath: DbPath,
			ExcludePaths: []string{
				// binaries
				"go-licenses",
				"deps/testdata",
				// testdata
				"licenses/testdata",
				// distribution
				"dist",
				// license db
				"third_party/google/licenseclassifier/licenses",
				// notices
				"third_party/NOTICES",
			},
		},
	)
	if err != nil {
		t.Error(err)
	}
	expected := []licenses.LicenseFound{
		{SpdxId: "Apache-2.0", Path: "LICENSE", StartLine: 2, EndLine: 175, Confidence: 1},
		{SpdxId: "Apache-2.0", Path: "third_party/google/licenseclassifier/LICENSE", StartLine: 2, EndLine: 175, Confidence: 1},
		{SpdxId: "MIT", Path: "third_party/uw-labs/lichen/LICENSE", StartLine: 5, EndLine: 21, Confidence: 1},
	}
	assert.Equal(t, expected, found)
}

func TestScan_DirWithSymlink(t *testing.T) {
	found, err := licenses.ScanDir(
		"testdata/folder-with-symlink",
		licenses.ScanDirOptions{
			DbPath: DbPath,
		},
	)
	if err != nil {
		t.Error(err)
	}
	expected := []licenses.LicenseFound{}
	assert.Equal(t, expected, found)
}
