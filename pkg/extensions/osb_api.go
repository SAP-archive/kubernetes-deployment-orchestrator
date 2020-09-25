package extensions

import (
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/wonderix/shalm/pkg/shalm"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

type osbBindingBackend struct {
	client  osb.Client
	service string
	plan    string
}

var _ shalm.JewelBackend = (*osbBindingBackend)(nil)

func (v *osbBindingBackend) Name() string {
	return "binding"
}

func (v *osbBindingBackend) Keys() map[string]string {
	return map[string]string{
		"credentials": "credentials",
	}
}

func (v *osbBindingBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	return map[string][]byte{
		"credentials": []byte("{}"),
	}, nil
}

func makeOsbBindung(client osb.Client) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		c := &osbBindingBackend{}
		var name string
		err := starlark.UnpackArgs("osbjewel", args, kwargs, "name", &name, "service", &c.service, "plan", &c.plan)
		if err != nil {
			return nil, err
		}
		return shalm.NewJewel(c, name)
	}
}

// OsbAPI -
func OsbAPI(client osb.Client) starlark.StringDict {
	return starlark.StringDict{
		"osb": &starlarkstruct.Module{
			Name: "osb",
			Members: starlark.StringDict{
				"binding": starlark.NewBuiltin("binding", makeOsbBindung(client)),
			},
		},
	}
}
