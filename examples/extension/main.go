package main

import (
	"fmt"
	"time"

	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/k14s/starlark-go/starlark"
	"github.com/wonderix/shalm/cmd"
)

type myJewelBackend struct {
	prefix string
}

var _ shalm.JewelBackend = (*myJewelBackend)(nil)

func (v *myJewelBackend) Name() string {
	return "myjewel"
}

func (v *myJewelBackend) Keys() map[string]shalm.JewelValue {
	return map[string]shalm.JewelValue{
		"username": {Name: "username"},
	}
}

func (v *myJewelBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	return map[string][]byte{
		"username": []byte(fmt.Sprintf("%s-%d", v.prefix, time.Now().Unix())),
	}, nil
}

func makeMyJewel(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c := &myJewelBackend{}
	var name string
	err := starlark.UnpackArgs("myjewel", args, kwargs, "name", &name, "prefix", &c.prefix)
	if err != nil {
		return nil, err
	}
	return shalm.NewJewel(c, name)
}

func myExtensions(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	switch module {
	case "@extension:message":
		return starlark.StringDict{
			"message": starlark.String("hello world"),
		}, nil
	case "@extension:myjewel":
		return starlark.StringDict{
			"myjewel": starlark.NewBuiltin("myjewel", makeMyJewel),
		}, nil
	}
	return nil, fmt.Errorf("Unknown module '%s'", module)
}

func main() {
	cmd.Execute(cmd.WithModules(myExtensions))
}
