package main

import (
	"fmt"
	"time"

	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/wonderix/shalm/cmd"
	"go.starlark.net/starlark"
)

type myVaultBackend struct {
	prefix string
}

var _ shalm.VaultBackend = (*myVaultBackend)(nil)

func (v *myVaultBackend) Name() string {
	return "myvault"
}

func (v *myVaultBackend) Keys() map[string]string {
	return map[string]string{
		"username": "username",
	}
}

func (v *myVaultBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	return map[string][]byte{
		"username": []byte(fmt.Sprintf("%s-%d", v.prefix, time.Now().Unix())),
	}, nil
}

func makeMyVault(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c := &myVaultBackend{}
	var name string
	err := starlark.UnpackArgs("myvault", args, kwargs, "name", &name, "prefix", &c.prefix)
	if err != nil {
		return nil, err
	}
	return shalm.NewVault(c, name)
}

func myExtensions(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	switch module {
	case "@extension:message":
		return starlark.StringDict{
			"message": starlark.String("hello world"),
		}, nil
	case "@extension:myvault":
		return starlark.StringDict{
			"myvault": starlark.NewBuiltin("myvault", makeMyVault),
		}, nil
	}
	return nil, fmt.Errorf("Unknown module '%s'", module)
}

func main() {
	cmd.Execute(cmd.WithModules(myExtensions))
}
