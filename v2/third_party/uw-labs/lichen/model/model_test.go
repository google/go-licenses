package model_test

import (
	"testing"

	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/model"
	"github.com/stretchr/testify/assert"
)

func TestModuleReference_IsLocal(t *testing.T) {
	testCases := []struct {
		name     string
		ref      model.ModuleReference
		expected bool
	}{
		{
			name: "with version",
			ref: model.ModuleReference{
				Version: "1.0.0",
			},
			expected: false,
		},
		{
			name: "current dir",
			ref: model.ModuleReference{
				Path: ".",
			},
			expected: true,
		},
		{
			name: "up one dir",
			ref: model.ModuleReference{
				Path: "..",
			},
			expected: true,
		},
		{
			name: "current dir with slash",
			ref: model.ModuleReference{
				Path: "./",
			},
			expected: true,
		},
		{
			name: "up one dir with slash",
			ref: model.ModuleReference{
				Path: "../",
			},
			expected: true,
		},
		{
			name: "dir relative to current",
			ref: model.ModuleReference{
				Path: "./test",
			},
			expected: true,
		},
		{
			name: "dir relative to up one",
			ref: model.ModuleReference{
				Path: "../test",
			},
			expected: true,
		},
		{
			name: "dir relative to current, up one",
			ref: model.ModuleReference{
				Path: "./../test",
			},
			expected: true,
		},
		{
			name: "absolute path, unix style",
			ref: model.ModuleReference{
				Path: "/test/abc",
			},
			expected: true,
		},
		{
			name: "absolute path, windows style",
			ref: model.ModuleReference{
				Path: "C:\\test\\abc",
			},
			expected: true,
		},
		{
			name: "github path",
			ref: model.ModuleReference{
				Path: "github.com/foo/bar",
			},
			expected: false,
		},
		{
			name: "ambiguous",
			ref: model.ModuleReference{
				Path: "github",
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			actual := tc.ref.IsLocal()
			assert.Equal(tt, tc.expected, actual)
		})
	}
}
