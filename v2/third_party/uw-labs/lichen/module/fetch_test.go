package module_test

import (
	"context"
	"testing"

	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/model"
	"github.com/google/go-licenses/v2/third_party/uw-labs/lichen/module"
	"github.com/stretchr/testify/assert"
)

func TestModuleFetchNoModules(test *testing.T) {
	modules, err := module.Fetch(context.Background(), []model.ModuleReference{})

	assert.NoError(test, err)
	assert.Empty(test, modules)
}
