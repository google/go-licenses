package buildinfo_test

import (
	"testing"

	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/buildinfo"
	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    []model.BuildInfo
		expectedErr string
	}{
		{
			name: "basic single binary input",
			input: `/tmp/lichen: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
			expected: []model.BuildInfo{
				{
					Path:        "/tmp/lichen",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
			},
		},
		{
			name: "single binary input with dep replace",
			input: `/tmp/lichen: go1.14
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d
	=>	github.com/uw-labs/go-md2man/v2	v0.4.16-0.20200608113539-44d3cd590db7	h1:7JSMFy7v19QNuP77yBMWawhzb9xD82oPmrlda5yrBkE=
`,
			expected: []model.BuildInfo{
				{
					Path:        "/tmp/lichen",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/uw-labs/go-md2man/v2",
							Version: "v0.4.16-0.20200608113539-44d3cd590db7",
						},
					},
				},
			},
		},
		{
			name: "basic multi binary input",
			input: `/tmp/lichen: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
/tmp/lichen2: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/google/goterm	v0.0.0-20190703233501-fc88cf888a3f	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
			expected: []model.BuildInfo{
				{
					Path:        "/tmp/lichen",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
				{
					Path:        "/tmp/lichen2",
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/google/goterm",
							Version: "v0.0.0-20190703233501-fc88cf888a3f",
						},
					},
				},
			},
		},
		{
			name: "windows basic single binary input",
			input: `C:\lichen.exe: go1.14.4
	path	github.com/uw-labs/lichen
	mod	github.com/uw-labs/lichen	(devel)	
	dep	github.com/cpuguy83/go-md2man/v2	v2.0.0-20190314233015-f79a8a8ca69d	h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
			expected: []model.BuildInfo{
				{
					Path:        `C:\lichen.exe`,
					PackagePath: "github.com/uw-labs/lichen",
					ModulePath:  "github.com/uw-labs/lichen",
					ModuleRefs: []model.ModuleReference{
						{
							Path:    "github.com/cpuguy83/go-md2man/v2",
							Version: "v2.0.0-20190314233015-f79a8a8ca69d",
						},
					},
				},
			},
		},
		{
			name:        "not executable file",
			input:       `/tmp/lichen: not executable file`,
			expectedErr: "/tmp/lichen is not an executable",
		},
		{
			name:        "unrecognised exe file",
			input:       `/tmp/lichen: unrecognized executable format`,
			expectedErr: "/tmp/lichen has an unrecognized executable format",
		},
		{
			name:        "go version not found",
			input:       `/tmp/lichen: go version not found`,
			expectedErr: "/tmp/lichen does not appear to be a Go compiled binary",
		},
		{
			name:        "invalid",
			input:       `/tmp/lichen: invalid`,
			expectedErr: "unrecognised version line: /tmp/lichen: invalid",
		},
		{
			name: "partial path line",
			input: `lichen: go1.14.4
	path
`,
			expectedErr: "invalid path line: \tpath",
		},
		{
			name: "path line unexpectedly long",
			input: `lichen: go1.14.4
	path	foo	bar
`,
			expectedErr: "invalid path line: \tpath\tfoo\tbar",
		},
		{
			name: "partial mod line",
			input: `lichen: go1.14.4
	mod	foo	(devel)
`,
			expectedErr: "invalid mod line: \tmod\tfoo\t(devel)",
		},
		{
			name: "mod line unexpectedly long",
			input: `lichen: go1.14.4
	mod	foo	(devel)	x	
`,
			expectedErr: "invalid mod line: \tmod\tfoo\t(devel)\tx\t",
		},
		{
			name: "partial dep line",
			input: `lichen: go1.14.4
	dep	foo
`,
			expectedErr: "invalid dep line: \tdep\tfoo",
		},
		{
			name: "dep line unexpectedly long",
			input: `lichen: go1.14.4
	dep	foo	v0	h1:x	x
`,
			expectedErr: "invalid dep line: \tdep\tfoo\tv0\th1:x\tx",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			actual, err := buildinfo.Parse(tc.input)
			if tc.expectedErr == "" {
				require.NoError(tt, err)
				assert.Equal(tt, tc.expected, actual)
			} else {
				assert.EqualError(tt, err, tc.expectedErr)
			}
		})
	}
}
