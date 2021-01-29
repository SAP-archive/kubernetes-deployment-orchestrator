package renderer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
)

type helmRenderer struct {
	helpers string
	root    *template.Template
}

// HelmFileRenderer -
func HelmFileRenderer(dir string, value interface{}) func(filename string) func(writer io.Writer) error {
	h, err := newHelmRenderer(dir)
	if err != nil {
		return errorFileRenderer(err)
	}
	if h.helpers != "" {
		_, err = h.root.Parse(h.helpers)
		if err != nil {
			return errorFileRenderer(err)
		}
	}
	return h.fileTemplater(value)
}

func newHelmRenderer(dir string) (*helmRenderer, error) {
	h := &helmRenderer{root: template.New("root")}
	content, err := ioutil.ReadFile(path.Join(dir, "templates", "_helpers.tpl"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		h.helpers = ""
	} else {
		h.helpers = string(content)
	}
	h.root.Funcs(sprig.TxtFuncMap())
	h.root.Funcs(map[string]interface{}{
		"toToml":   notImplemented("toToml"),
		"toYaml":   toYAML,
		"fromYaml": notImplemented("fromYaml"),
		"toJson":   toJSON,
		"fromJson": notImplemented("fromJson"),
		"tpl":      h.tpl(),
		"required": required,
		"include": func(name string, data interface{}) (string, error) {
			var buf strings.Builder
			err := h.root.ExecuteTemplate(&buf, name, data)
			return buf.String(), err
		},
	})
	return h, nil
}

func errorFileRenderer(err error) func(filename string) func(writer io.Writer) error {
	return func(filename string) func(writer io.Writer) error {
		return errorWriter(err)
	}
}

func errorWriter(err error) func(writer io.Writer) error {
	return func(writer io.Writer) error {
		return err
	}
}
func (h *helmRenderer) fileTemplater(value interface{}) func(filename string) func(writer io.Writer) error {
	return func(filename string) func(writer io.Writer) error {
		tpl, err := h.loadTemplate("./" + filepath.Base(filename))
		if err != nil {
			return errorWriter(err)
		}
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return errorWriter(err)
		}
		tpl, err = tpl.Parse(string(content))
		if err != nil {
			return errorWriter(err)
		}
		return func(writer io.Writer) error {
			return tpl.Execute(writer, value)
		}
	}

}

func notImplemented(name string) func(_ interface{}) (string, error) {
	return func(_ interface{}) (string, error) {
		return "", fmt.Errorf("Function %s is not implemented in kdo", name)
	}
}

func (h *helmRenderer) loadTemplate(name string) (result *template.Template, err error) {
	return h.root.New(name), nil
}

func (h *helmRenderer) tpl() func(stringTemplate string, values interface{}) (interface{}, error) {
	return func(stringTemplate string, values interface{}) (interface{}, error) {
		tpl, err := h.loadTemplate("internal template")
		if err != nil {
			return nil, err
		}
		tpl, err = tpl.Parse(stringTemplate)
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		err = tpl.Execute(&buffer, values)
		if err != nil {
			return nil, err
		}
		return buffer.String(), nil
	}
}

func toYAML(v interface{}) (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}

func toJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}

func required(msg string, c interface{}) (interface{}, error) {
	if c == nil || (reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()) {
		return nil, fmt.Errorf("%s", msg)
	}
	return c, nil
}
