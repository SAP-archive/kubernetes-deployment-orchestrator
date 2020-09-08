package shalm

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/pkg/errors"
)

type injectedFiles struct {
	dir    string
	files  []string
	kwargs starlark.StringDict
}

type injectedContext []starlark.StringDict

var _ starlark.Value = (*injectedFiles)(nil)

func (context injectedContext) module() starlark.StringDict {
	return starlark.StringDict{"context": starlark.NewBuiltin("context", context.injectedValue)}
}

func (context injectedContext) injectedValue(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	var index int
	var name string
	if err := starlark.UnpackArgs("context", args, kwargs, "index", &index, "name?", &name); err != nil {
		return nil, err
	}
	if index > len(context) {
		return starlark.None, errors.New("Index out of bounds")
	}
	return context[index][name], nil
}

func (context *injectedContext) add(kwargs starlark.StringDict) string {
	index := len(*context)
	*context = append(*context, kwargs)
	buf := &bytes.Buffer{}
	_, _ = buf.WriteString("load('@shalm:context','context'); ")
	for k := range kwargs {
		_, _ = buf.WriteString(fmt.Sprintf(" %s = context(%d,'%s');", k, index, k))
	}
	return buf.String()
}

func makeInjectedFiles(dir string) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		result := &injectedFiles{dir: dir, kwargs: kwargsToStringDict(kwargs)}
		for i := 0; i < args.Len(); i++ {
			result.files = append(result.files, args.Index(i).(starlark.String).GoString())
		}
		return result, nil
	}
}

// String -
func (c *injectedFiles) String() string {
	return strings.Join(c.files, ", ")
}

// Type -
func (c *injectedFiles) Type() string { return "injected_files" }

// Freeze -
func (c *injectedFiles) Freeze() {}

// Truth -
func (c *injectedFiles) Truth() starlark.Bool { return false }

// Hash -
func (c *injectedFiles) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (c *injectedFiles) Attr(name string) (starlark.Value, error) {
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("stream has no .%s attribute", name))
}

// AttrNames -
func (c *injectedFiles) AttrNames() []string {
	return []string{}
}
