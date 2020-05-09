package modx

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveModule(t *testing.T) {
	gomod, err := Open(".fixtures/example1")
	require.Equal(t, err, nil, "unable to open go.mod: %v", err)

	// should succeed when the module doesn't exist.
	err = RemoveModule(gomod, "")
	require.Equal(t, err, nil, "failed to remove blank module path: %v", err)

	// remove all defined modules
	for _, v := range gomod.Require {
		err = RemoveModule(gomod, v.Mod.Path)
		require.Equal(t, err, nil, "failed to remove module: %s - %v", v.Mod.Path, err)
	}

	output, err := Print(gomod)
	require.Equal(t, err, nil, "failed to print go.mod %v", err)

	expected, err := ioutil.ReadFile(".fixtures/example1/empty.go.mod")
	require.Equal(t, err, nil, "failed to read fixture %v", err)
	require.Equal(t, output, string(expected))
}

func TestRemoveModuleIdempotent(t *testing.T) {
	const module = "github.com/pkg/errors"
	gomod, err := Open(".fixtures/example1")
	require.Equal(t, err, nil, "unable to open go.mod: %v", err)

	err = RemoveModule(gomod, module)
	require.Equal(t, err, nil, "failed to remove import %s", module)

	err = RemoveModule(gomod, module)
	require.Equal(t, err, nil, "failed to remove import %s", module)

	output, err := Print(gomod)
	require.Equal(t, err, nil, "failed to print go.mod %v", err)

	expected, err := ioutil.ReadFile(".fixtures/example1/output1.go.mod")
	require.Equal(t, err, nil, "failed to read fixture %v", err)
	require.Equal(t, output, string(expected))
}
