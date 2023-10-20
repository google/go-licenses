package testdata

import (
	// This import should be detected by library_test.go. It has to be a
	// package that isn't in the standard library and has a separate license
	// file to the one covering the Trillian repository, so that it's detected
	// as being an external dependency.
	_ "github.com/khulnasoft/go-bouncer/bouncer/licenses/testdata/direct"

	// This import should be ignored, since it's an internal dependency.
	_ "github.com/khulnasoft/go-bouncer/bouncer/licenses/testdata/internal"

	// This import should be ignored, since it's an standard library package.
	_ "strings"
)