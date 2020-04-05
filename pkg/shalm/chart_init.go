package shalm

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func (c *chartImpl) loadChartYaml() error {

	err := readYamlFile(c.path("Chart.yaml"), &c.clazz)
	if err != nil {
		return err
	}
	if err := c.clazz.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *chartImpl) loadYaml(name string) error {
	var values map[string]interface{}
	err := readYamlFile(c.path(name), &values)
	if err != nil {
		return err
	}
	for k, v := range values {
		c.values[k] = toStarlark(v)
	}
	return nil
}

func (c *chartImpl) loadYamlFunction() starlark.Callable {
	return c.builtin("load_yaml", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var name string
		if err := starlark.UnpackArgs("load_yaml", args, kwargs, "name", &name); err != nil {
			return starlark.None, err
		}
		return starlark.None, c.loadYaml(name)
	})
}

// NewChartFunction -
func NewChartFunction(repo Repo, dir string, options ...ChartOption) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if len(args) == 0 {
			return starlark.None, fmt.Errorf("%s: got %d arguments, want at most %d", "chart", 0, 1)
		}
		url := args[0].(starlark.String).GoString()
		if !(filepath.IsAbs(url) || strings.HasPrefix(url, "http")) {
			url = path.Join(dir, url)
		}
		co := chartOptions(options)
		parser := &kwargsParser{kwargs: kwargs}
		parser.Arg("namespace", func(value starlark.Value) {
			co.namespace = value.(starlark.String).GoString()
		})
		parser.Arg("proxy", func(value starlark.Value) {
			_ = co.proxyMode.Set(value.(starlark.String).GoString())
		})
		parser.Arg("suffix", func(value starlark.Value) {
			co.suffix = value.(starlark.String).GoString()
		})
		co.kwargs = parser.Parse()
		return repo.Get(thread, url, co.Merge())
	}
}

func (c *chartImpl) init(thread *starlark.Thread, repo Repo, hasChartYaml bool, args starlark.Tuple, kwargs []starlark.Tuple) error {
	c.methods["apply"] = c.applyFunction()
	c.methods["delete"] = c.deleteFunction()
	c.methods["template"] = c.templateFunction()
	c.methods["__apply"] = c.applyLocalFunction()
	c.methods["__delete"] = c.deleteLocalFunction()
	c.methods["helm"] = c.helmTemplateFunction()
	c.methods["eytt"] = c.yttEmbeddedTemplateFunction()
	c.methods["ytt"] = c.yttTemplateFunction()
	c.methods["load_yaml"] = c.loadYamlFunction()

	file := c.path("Chart.star")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if !hasChartYaml {
			return fmt.Errorf("Neither Chart.star nor Chart.yaml nor values.yaml exists in %s", c.dir)
		}
		return nil
	}

	internal := starlark.StringDict{
		"version":         starlark.String(version),
		"kube_version":    starlark.String(kubeVersion),
		"chart":           c.builtin("chart", NewChartFunction(repo, c.dir, c.ChartOptions.Merge())),
		"user_credential": c.builtin("user_credential", makeUserCredential),
		"config_value":    c.builtin("config_value", makeConfigValue),
		"certificate":     c.builtin("certificate", makeCertificate),
		"struct":          starlark.NewBuiltin("struct", starlarkstruct.Make),
	}
	globals, err := starlark.ExecFile(thread, file, nil, internal)
	if err != nil {
		return err
	}

	for k, v := range globals {
		if k == "init" {
			c.initFunc = v.(*starlark.Function)
		}
		f, ok := v.(starlark.Callable)
		if ok {
			c.methods[k] = c.builtin("bind_"+f.Name(), func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				allArgs := make([]starlark.Value, args.Len()+1)
				allArgs[0] = c
				for i := 0; i < args.Len(); i++ {
					allArgs[i+1] = args.Index(i)
				}
				return starlark.Call(thread, f, allArgs, kwargs)
			})
		}
	}

	if c.initFunc != nil {
		_, err := starlark.Call(thread, c.initFunc, append([]starlark.Value{c}, args...), kwargs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *chartImpl) wrapNamespace(callable starlark.Callable, namespace string) starlark.Callable {
	return c.builtin("wrap_namespace", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("Missing first argument k8s")
		}
		k, ok := args[0].(K8sValue)
		if !ok {
			return nil, fmt.Errorf("Invalid first argument to %s", callable.Name())
		}
		subK8s := k.ForSubChart(namespace, c.GetName(), c.GetVersion())
		args[0] = &k8sValueImpl{subK8s}
		value, err := starlark.Call(thread, callable, args, kwargs)
		return value, err

	})
}
