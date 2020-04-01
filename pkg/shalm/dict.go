package shalm

import (
	"fmt"

	"go.starlark.net/starlark"
)

type dict struct {
	*starlark.Dict
}

var _ starlark.HasSetField = &dict{}

func (d *dict) Attr(name string) (starlark.Value, error) {
	value, err := d.Dict.Attr(name)
	if value != nil {
		return value, nil
	}
	value, ok, err := d.Dict.Get(starlark.String(name))
	if err != nil {
		return starlark.None, err
	}
	if ok {
		return wrapDict(value), nil
	}
	return starlark.None, starlark.NoSuchAttrError(
		fmt.Sprintf("dict has no .%s attribute", name))
}

func (d *dict) AttrNames() []string {
	result := make([]string, 0)
	result = append(result, d.Dict.AttrNames()...)
	for _, k := range d.Dict.Keys() {
		result = append(result, k.(starlark.String).GoString())
	}
	return result
}

func (d *dict) SetField(name string, val starlark.Value) error {
	return d.Dict.SetKey(starlark.String(name), unwrapDict(val))
}

func wrapDict(value starlark.Value) starlark.Value {
	d, ok := value.(*starlark.Dict)
	if ok {
		return &dict{Dict: d}
	}
	return value
}

func unwrapDict(value starlark.Value) starlark.Value {
	d, ok := value.(*dict)
	if ok {
		return d.Dict
	}
	return value
}
