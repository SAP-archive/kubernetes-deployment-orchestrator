package shalm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/wonderix/shalm/pkg/shalm/renderer"

	"github.com/k14s/starlark-go/starlark"
	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"

	"github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/workspace"
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
		s := c.helmTemplate(thread, "templates", glob)
		yttTemplateDir := path.Join(c.dir, "ytt-templates")
		if _, err := os.Stat(yttTemplateDir); err == nil {
			s = yamlConcat(s, c.yttTemplate(thread, starlark.Tuple{
				&injectedFiles{
					dir:    c.dir,
					files:  []string{"ytt-templates"},
					kwargs: starlark.StringDict{"self": c},
				}}))
		}
		return &stream{Stream: s}, nil
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

func (c *chartImpl) yttTemplate(thread *starlark.Thread, fileTuple starlark.Tuple) Stream {
	return func(writer io.Writer) error {
		context := injectedContext{}
		o := &template.TemplateOptions{Extender: func(l workspace.ModuleLoader) workspace.ModuleLoader {
			return func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				if module == "@shalm:context" {
					return context.module(), nil
				}
				return l(thread, module)

			}
		}}
		filesToProcess := []*files.File{}
		var tempFiles []string
		defer func() {
			for _, f := range tempFiles {
				_ = os.Remove(f)
			}
		}()
		for _, arg := range fileTuple {
			switch arg := arg.(type) {
			case *stream:
				f, err := ioutil.TempFile("", "shalm*.yml")
				tempFiles = append(tempFiles, f.Name())
				if err != nil {
					return errors.Wrapf(err, "Error saving stream to file in ytt")
				}
				err = arg.Stream(f)
				if err != nil {
					return errors.Wrapf(err, "Error saving stream to file in ytt")
				}
				f.Close()
				fs, err := files.NewFileFromSource(files.NewLocalSource(f.Name(), ""))
				if err != nil {
					return err
				}
				filesToProcess = append(filesToProcess, fs)
			case *injectedFiles:
				prefix := context.add(arg.kwargs)
				for _, file := range arg.files {
					fn := path.Join(arg.dir, file)
					stat, err := os.Stat(fn)
					if err != nil {
						return err
					}
					if stat.IsDir() {
						err = filepath.Walk(fn, func(file string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if !info.IsDir() {
								filesToProcess = append(filesToProcess, files.MustNewFileFromSource(&chartSource{path: file, prefix: prefix}))
							}
							return nil
						})
					} else {
						filesToProcess = append(filesToProcess, files.MustNewFileFromSource(&chartSource{path: fn, prefix: prefix}))
					}
				}
			case starlark.String:
				fs, err := files.NewSortedFilesFromPaths([]string{path.Join(c.dir, arg.GoString())}, files.SymlinkAllowOpts{})
				if err != nil {
					return err
				}
				filesToProcess = append(filesToProcess, fs...)
			default:
				return fmt.Errorf("Invalid type passed to ytt")
			}

		}
		filesToProcess = files.NewSortedFiles(filesToProcess)
		fmt.Printf("ytt ")
		for _, f := range filesToProcess {
			fmt.Printf("-f %s ", f.RelativePath())
		}
		fmt.Println("")
		out := o.RunWithFiles(template.TemplateInput{Files: files.NewSortedFiles(filesToProcess)}, cmdcore.NewPlainUI(o.Debug))

		if out.Err != nil {
			return out.Err
		}

		body, err := out.DocSet.AsBytes()
		if err != nil {
			return err
		}
		_, err = writer.Write(body)
		return err
	}
}

type chartSource struct {
	path   string
	prefix string
}

func (s *chartSource) Description() string { return fmt.Sprintf("file '%s'", s.path) }

func (s *chartSource) RelativePath() (string, error) {
	return s.path, nil
}

func (s *chartSource) Bytes() ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteString("#@ ")
	buffer.WriteString(s.prefix)
	buffer.WriteString("\n")
	f, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = io.Copy(buffer, f)
	return buffer.Bytes(), err
}

func (c *chartImpl) yttTemplateFunction() starlark.Callable {
	return c.builtin("ytt", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return &stream{Stream: c.yttTemplate(thread, args)}, nil
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
