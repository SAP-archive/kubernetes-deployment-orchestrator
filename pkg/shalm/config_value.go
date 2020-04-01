package shalm

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"go.starlark.net/starlark"
)

type configType int

const (
	configTypeString = iota
	configTypeBool
	configTypePassword
	configTypeSelection
)

func (p configType) String() string {
	return [...]string{"string", "password", "selection"}[p]
}

func (p *configType) set(val string) error {
	switch val {
	case "":
		fallthrough
	case "string":
		*p = configTypeString
	case "bool":
		*p = configTypeBool
	case "password":
		*p = configTypePassword
	case "selection":
		*p = configTypeSelection
	default:
		return fmt.Errorf("invalid type mode %s", val)
	}
	return nil
}

type configValueBackend struct {
	name        string
	description string
	dflt        string
	typ         configType
	options     []string
}

var _ VaultBackend = (*configValueBackend)(nil)

func (v *configValueBackend) Name() string {
	return "config_value"
}

func (v *configValueBackend) Keys() map[string]string {
	return map[string]string{
		"value": "value",
	}
}

func (v *configValueBackend) read() (string, error) {
	switch v.typ {
	case configTypeString:
		p := promptui.Prompt{
			Label:   v.name,
			Default: v.dflt,
		}
		return p.Run()
	case configTypePassword:
		p := promptui.Prompt{
			Label:   v.name,
			Default: v.dflt,
			Mask:    '*',
		}
		return p.Run()
	case configTypeBool:
		sel := promptui.Select{
			Label: v.name,
			Items: []string{"yes", "no"},
		}
		_, s, err := sel.Run()
		return s, err
	case configTypeSelection:
		sel := promptui.Select{
			Label: v.name,
			Items: v.options,
		}
		_, s, err := sel.Run()
		return s, err
	}
	return "", nil
}
func (v *configValueBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	if len(m["value"]) == 0 {
		fmt.Println("\n------------------------------------")
		fmt.Println(v.description)
		result, err := v.read()
		if err != nil {
			return nil, err
		}
		m["value"] = []byte(result)
	}
	return m, nil
}

func makeConfigValue(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c := &configValueBackend{}
	var typ string
	options := &starlark.List{}
	err := starlark.UnpackArgs("config_value", args, kwargs, "name", &c.name, "description?", &c.description, "type?", &typ, "options", &options,
		"default", &c.dflt)
	if err != nil {
		return starlark.None, err
	}
	for i := 0; i < options.Len(); i++ {
		c.options = append(c.options, options.Index(i).(starlark.String).GoString())
	}
	if err = c.typ.set(typ); err != nil {
		return starlark.None, err
	}
	return NewVault(c, c.name)
}
