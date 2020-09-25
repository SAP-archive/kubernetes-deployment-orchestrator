package cmd

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/template/core"
	"golang.org/x/crypto/bcrypt"
)

var (
	// BcryptAPI -
	BcryptAPI = starlark.StringDict{
		"bcrypt": &starlarkstruct.Module{
			Name: "bcrypt",
			Members: starlark.StringDict{
				"hash": starlark.NewBuiltin("bcrypt.hash", core.ErrWrapper(bcryptModule{}.Hash)),
			},
		},
	}
)

type bcryptModule struct{}

func (b bcryptModule) Hash(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	password, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("expected string as argument")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.GoString()), bcrypt.DefaultCost)
	if err != nil {
		return starlark.None, err
	}
	return starlark.String(hashedPassword), nil
}
