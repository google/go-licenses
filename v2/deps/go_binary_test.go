package deps_test

import (
	"testing"

	"github.com/google/go-licenses/v2/deps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModulesInGoBinary(t *testing.T) {
	actual, err := deps.ListModulesInGoBinary("testdata/binary-1")
	require.Nil(t, err)
	modulesActual := make([]string, 0)
	for _, module := range actual {
		assert.NotEmpty(t, module.ImportPath)
		assert.NotEmpty(t, module.Version)
		modulesActual = append(modulesActual, module.ImportPath)
	}
	expected := []string{
		"github.com/PuerkitoBio/goquery",
		"github.com/andybalholm/cascadia",
		"github.com/go-logr/logr",
		"github.com/google/go-github/v33",
		"github.com/google/go-querystring",
		"github.com/google/licenseclassifier",
		"github.com/hashicorp/errwrap",
		"github.com/hashicorp/go-multierror",
		"github.com/otiai10/copy",
		"github.com/pkg/errors",
		"github.com/sergi/go-diff",
		"github.com/spf13/cobra",
		"github.com/spf13/pflag",
		"golang.org/x/crypto",
		"golang.org/x/net",
		"golang.org/x/oauth2",
		"gopkg.in/yaml.v2",
		"k8s.io/klog/v2",
	}
	assert.Equal(t, expected, modulesActual)
}
