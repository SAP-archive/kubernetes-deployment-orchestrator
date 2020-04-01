package renderer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/k14s/ytt/pkg/template"

	"github.com/k14s/ytt/pkg/yttlibrary"
	"github.com/k14s/ytt/pkg/yttlibrary/overlay"

	"go.starlark.net/starlarkstruct"

	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
)

// YttFileRenderer -
func YttFileRenderer(value starlark.Value) func(filename string) func(writer io.Writer) error {
	return func(filename string) func(writer io.Writer) error {
		return func(writer io.Writer) error {
			return yttRenderFile(value, filename, writer)
		}
	}
}

func yttRenderFile(value starlark.Value, filename string, writer io.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return yttRender(value, f, filename, writer)
}

type templateLoader struct {
	template.NoopCompiledTemplateLoader
	template *template.CompiledTemplate
}

func (t templateLoader) FindCompiledTemplate(_ string) (*template.CompiledTemplate, error) {
	return t.template, nil
}

func yttRender(value starlark.Value, reader io.Reader, associatedName string, writer io.Writer) error {
	prefix := bytes.NewBuffer([]byte("#@ load(\"@shalm:self\", \"self\")\n"))
	content, err := ioutil.ReadAll(io.MultiReader(prefix, reader))
	if err != nil {
		return err
	}
	docSet, err := yamlmeta.NewDocumentSetFromBytes(content, yamlmeta.DocSetOpts{AssociatedName: associatedName})
	if err != nil {
		return err
	}
	loader := &templateLoader{}
	loader.template, err = yamltemplate.NewTemplate(associatedName, yamltemplate.TemplateOpts{}).Compile(docSet)
	if err != nil {
		return err
	}

	thread := &starlark.Thread{Name: "test", Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
		switch module {
		case "@shalm:self":
			return starlark.StringDict{
				"self": value,
			}, nil
		case "@ytt:data":
			return starlark.StringDict{
				"data": &starlarkstruct.Module{
					Name:    "data",
					Members: starlark.StringDict{"values": value},
				},
			}, nil
		case "@ytt:base64":
			return yttlibrary.Base64API, nil
		case "@ytt:json":
			return yttlibrary.JSONAPI, nil
		case "@ytt:md5":
			return yttlibrary.MD5API, nil
		case "@ytt:regexp":
			return yttlibrary.RegexpAPI, nil
		case "@ytt:sha256":
			return yttlibrary.SHA256API, nil
		case "@ytt:url":
			return yttlibrary.URLAPI, nil
		case "@ytt:yaml":
			return yttlibrary.YAMLAPI, nil
		case "@ytt:overlay":
			return overlay.API, nil
		case "@ytt:struct":
			return yttlibrary.StructAPI, nil
		case "@ytt:module":
			return yttlibrary.ModuleAPI, nil
		}

		return nil, fmt.Errorf("Unknown module '%s'", module)
	}}

	_, newVal, err := loader.template.Eval(thread, loader)
	if err != nil {
		return err
	}

	typedNewVal, ok := newVal.(interface{ AsBytes() ([]byte, error) })
	if !ok {
		return fmt.Errorf("Invalid return type of CompiledTemplate.Eval")
	}

	body, err := typedNewVal.AsBytes()
	if err != nil {
		return err
	}
	writer.Write(body)
	return nil
}
