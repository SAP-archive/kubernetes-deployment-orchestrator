package shalm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/wonderix/shalm/pkg/shalm/renderer"

	"go.starlark.net/starlark"
)

type release struct {
	Name      string
	Namespace string
	Service   string
	Revision  int
	IsUpgrade bool
	IsInstall bool
}

type chart struct {
	Name       string
	Version    string
	AppVersion string
	APIVersion string
}

type templateSpec struct {
	BasePath string
}

type kubeVersions struct {
	GitVersion string
	Major      int
	Minor      int
	Version    string
}

type capabilities struct {
	KubeVersion kubeVersions
}

func (c *chartImpl) Template(thread *starlark.Thread) Stream {
	streams := []Stream{}
	err := c.eachSubChart(func(subChart *chartImpl) error {
		streams = append(streams, subChart.template(thread, "", false))
		return nil
	})
	if err != nil {
		return ErrorStream(err)
	}
	streams = append(streams, c.template(thread, "", false))
	return yamlConcat(streams...)
}

func (c *chartImpl) template(thread *starlark.Thread, glob string, reconcile bool) Stream {
	kwargs := []starlark.Tuple{}
	if glob != "" {
		kwargs = append(kwargs, starlark.Tuple{starlark.String("glob"), starlark.String(glob)})
	}
	return yamlConcat(c.jewelStream().Encode(), toStream(starlark.Call(thread, c.methods["template"], nil, kwargs)))
}

func (c *chartImpl) helmTemplateFunction() starlark.Callable {
	return c.builtin("helm", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var glob string
		var dir string
		if err := starlark.UnpackArgs("helm", args, kwargs, "dir", &dir, "glob?", &glob); err != nil {
			return nil, err
		}
		s := c.helmTemplate(thread, dir, glob)
		return &stream{Stream: s}, nil
	})
}

func (c *chartImpl) templateFunction() starlark.Callable {
	return c.builtin("template", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var glob string
		if err := starlark.UnpackArgs("template", args, kwargs, "glob?", &glob); err != nil {
			return nil, err
		}
		return &stream{Stream: yamlConcat(c.helmTemplate(thread, "templates", glob), c.yttEmbeddedTemplate(thread, "ytt-templates", glob))}, nil
	})
}

func (c *chartImpl) helmTemplate(thread *starlark.Thread, dir string, glob string) Stream {
	values := stringDictToGo(c.values)
	methods := make(map[string]interface{})
	for k, f := range c.methods {
		method := f
		methods[k] = func() (interface{}, error) {
			value, err := starlark.Call(thread, method, nil, nil)
			return value, err
		}
	}
	helmFileRenderer := renderer.HelmFileRenderer(c.path(), struct {
		Values       interface{}
		Methods      map[string]interface{}
		Chart        chart
		Release      release
		Files        renderer.Files
		Template     templateSpec
		Capabilities capabilities
	}{
		Values:  values,
		Methods: methods,
		Chart: chart{
			Name:       c.clazz.Name,
			AppVersion: c.GetVersionString(),
			Version:    c.GetVersionString(),
		},
		Release: release{
			Name:      c.GetName(),
			Namespace: c.namespace,
			Service:   c.GetName(),
			Revision:  1,
			IsInstall: false,
			IsUpgrade: true,
		},
		Template: templateSpec{
			BasePath: ".",
		},
		Capabilities: capabilities{
			KubeVersion: kubeVersions{
				GitVersion: kubeSemver.String(),
				Version:    kubeSemver.String(),
				Major:      int(kubeSemver.Major),
				Minor:      int(kubeSemver.Minor),
			},
		},
		Files: renderer.Files{Dir: c.dir},
	})

	return func(writer io.Writer) error {

		return renderer.DirRender(glob,
			renderer.DirSpec{
				Dir:          path.Join(c.dir, dir),
				FileRenderer: helmFileRenderer,
			})(writer)

	}
}

func (c *chartImpl) yttEmbeddedTemplateFunction() starlark.Callable {
	return c.builtin("ytt", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var glob string
		var dir string
		if err := starlark.UnpackArgs("ytt", args, kwargs, "dir", &dir, "glob?", &glob); err != nil {
			return nil, err
		}
		s := c.yttEmbeddedTemplate(thread, dir, glob)
		return &stream{Stream: s}, nil
	})
}

func (c *chartImpl) yttEmbeddedTemplate(thread *starlark.Thread, dir string, glob string) Stream {

	return func(writer io.Writer) error {

		if strings.HasPrefix(dir, ".") {
			return fmt.Errorf("Invalid template directory '%s'", dir)
		}

		return renderer.DirRender(glob,
			renderer.DirSpec{
				Dir:          path.Join(c.dir, dir),
				FileRenderer: renderer.YttFileRenderer(c),
			})(writer)

	}
}

func (c *chartImpl) yttTemplateFunction() starlark.Callable {
	return c.builtin("ytt", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		s := func(writer io.Writer) error {
			flags := []string{}
			var tempFiles []string
			defer func() {
				for _, f := range tempFiles {
					_ = os.Remove(f)
				}
			}()
			for _, arg := range args {
				switch arg := arg.(type) {
				case *stream:
					f, err := ioutil.TempFile("", "shalm*.yml")
					tempFiles = append(tempFiles, f.Name())
					flags = append(flags, "-f", f.Name())
					if err != nil {
						return errors.Wrapf(err, "Error saving stream to file in ytt")
					}
					err = arg.Stream(f)
					if err != nil {
						return errors.Wrapf(err, "Error saving stream to file in ytt")
					}
					f.Close()
				case starlark.String:
					fmt.Printf("%v\n", arg)
					flags = append(flags, "-f", path.Join(c.dir, arg.GoString()))
				default:
					return fmt.Errorf("Invalid type passed to ytt")
				}

			}
			cmd := exec.Command("ytt", flags...)
			fmt.Println(cmd.String())
			cmd.Stdout = writer
			buffer := bytes.Buffer{}
			cmd.Stderr = &buffer
			err := cmd.Run()
			if err != nil {
				return errors.Wrap(err, string(buffer.Bytes()))
			}
			return nil
		}
		return &stream{Stream: s}, nil
	})
}

func (c *chartImpl) jewelStream() ObjectStream {
	return func(w ObjectWriter) error {
		vault := &vaultK8s{objectWriter: w, namespace: c.namespace}
		return c.eachJewel(func(v *jewel) error {
			return v.write(vault)
		})
	}
}
