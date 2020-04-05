package shalm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"go.starlark.net/starlark"
	corev1 "k8s.io/api/core/v1"
)

type chartImpl struct {
	ChartOptions
	clazz    chartClass
	Version  semver.Version
	values   starlark.StringDict
	methods  map[string]starlark.Callable
	dir      string
	initFunc *starlark.Function
}

var (
	_ ChartValue = (*chartImpl)(nil)
)

func newChart(thread *starlark.Thread, repo Repo, dir string, opts ...ChartOption) (*chartImpl, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	name := strings.Split(filepath.Base(abs), ":")[0]
	co := chartOptions(opts)
	c := &chartImpl{dir: dir, ChartOptions: *co, clazz: chartClass{Name: name}}
	c.values = make(map[string]starlark.Value)
	c.methods = make(map[string]starlark.Callable)
	hasChartYaml := false
	if err := c.loadChartYaml(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		hasChartYaml = true
	}
	if err := c.loadYaml("values.yaml"); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		hasChartYaml = true
	}
	if co.k8s == nil {
		c.k8s = NewK8sInMemory("default")
	} else {
		c.k8s = co.k8s.ForSubChart(c.namespace, c.clazz.Name, c.clazz.GetVersion())
	}
	if err := c.init(thread, repo, hasChartYaml, co.args, co.kwargs); err != nil {
		return nil, err
	}
	c.mergeValues(co.values)
	return c, nil

}

func (c *chartImpl) builtin(name string, fn func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) starlark.Callable {
	return starlark.NewBuiltin(name+" at "+path.Join(c.dir, "Chart.star"), fn)
}

func (c *chartImpl) GetName() string {
	if c.suffix == "" {
		return c.clazz.Name
	}
	return fmt.Sprintf("%s-%s", c.clazz.Name, c.suffix)
}

func (c *chartImpl) GetVersion() semver.Version {
	return c.clazz.GetVersion()
}

func (c *chartImpl) GetVersionString() string {
	return c.clazz.Version
}

func (c *chartImpl) walk(cb func(name string, size int64, body io.Reader, err error) error) error {
	return filepath.Walk(c.dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(c.dir, file)
		if err != nil {
			return err
		}
		body, err := os.Open(file)
		if err != nil {
			return err
		}
		defer body.Close()
		return cb(rel, info.Size(), body, nil)
	})
}

func (c *chartImpl) path(part ...string) string {
	return filepath.Join(append([]string{c.dir}, part...)...)
}

func (c *chartImpl) String() string {
	buf := new(strings.Builder)
	buf.WriteString("chart")
	buf.WriteByte('(')
	s := 0
	for i, e := range c.values {
		if s > 0 {
			buf.WriteString(", ")
		}
		s++
		buf.WriteString(i)
		buf.WriteString(" = ")
		buf.WriteString(e.String())
	}
	buf.WriteByte(')')
	return buf.String()
}

// Type -
func (c *chartImpl) Type() string { return "chart" }

// Truth -
func (c *chartImpl) Truth() starlark.Bool { return true } // even when empty

// Hash -
func (c *chartImpl) Hash() (uint32, error) {
	var x, m uint32 = 8731, 9839
	for k, e := range c.values {
		namehash, _ := starlark.String(k).Hash()
		x = x ^ 3*namehash
		y, err := e.Hash()
		if err != nil {
			return 0, err
		}
		x = x ^ y*m
		m += 7349
	}
	return x, nil
}

// Freeze -
func (c *chartImpl) Freeze() {
}

// Attr returns the value of the specified field.
func (c *chartImpl) Attr(name string) (starlark.Value, error) {
	switch name {
	case "namespace":
		return starlark.String(c.namespace), nil
	case "k8s":
		return NewK8sValue(c.k8s), nil
	case "name":
		return starlark.String(c.GetName()), nil
	case "__class__":
		return &c.clazz, nil
	default:
		value, ok := c.values[name]
		if !ok {
			var m starlark.Value
			m, ok = c.methods[name]
			if !ok {
				m = nil
			}
			if m == nil {
				return nil, starlark.NoSuchAttrError(
					fmt.Sprintf("chart has no .%s attribute", name))

			}
			return m, nil
		}
		return wrapDict(value), nil
	}
}

// AttrNames returns a new sorted list of the struct fields.
func (c *chartImpl) AttrNames() []string {
	names := make([]string, 0)
	for k := range c.values {
		names = append(names, k)
	}
	names = append(names, "template")
	return names
}

// SetField -
func (c *chartImpl) SetField(name string, val starlark.Value) error {
	val = unwrapDict(val)
	vaultVal, ok := val.(*vault)
	if ok {
		vaultVal.read(c.k8s)
	}
	c.values[name] = val
	return nil
}

func (c *chartImpl) applyFunction() starlark.Callable {
	return c.builtin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if err := starlark.UnpackArgs("apply", args, kwargs); err != nil {
			return nil, err
		}
		return starlark.None, c.apply(thread)
	})
}

func (c *chartImpl) Apply(thread *starlark.Thread) error {
	_, err := starlark.Call(thread, c.methods["apply"], nil, nil)
	if err != nil {
		return err
	}
	c.k8s.Progress(100)
	return nil
}

func (c *chartImpl) apply(thread *starlark.Thread) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := starlark.Call(thread, subChart.methods["apply"], nil, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.applyLocal(thread, &K8sOptions{}, "")
}

func (c *chartImpl) applyLocalFunction() starlark.Callable {
	return c.builtin("__apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		parser := &kwargsParser{kwargs: kwargs}
		var glob string
		k8sOptions := unpackK8sOptions(parser)
		if err := starlark.UnpackArgs("__apply", args, parser.Parse(), "glob?", &glob); err != nil {
			return nil, err
		}
		return starlark.None, c.applyLocal(thread, k8sOptions, glob)
	})
}

func (c *chartImpl) applyLocal(thread *starlark.Thread, k8sOptions *K8sOptions, glob string) error {
	if c.readOnly {
		return nil
	}
	k8sOptions.Namespaced = false
	return c.k8s.Apply(concat(decode(c.template(thread, glob, true)), c.packedChart()), k8sOptions)
}

func (c *chartImpl) packedChart() ObjectStream {
	return func(w ObjectWriter) error {
		if c.skipChart {
			return nil
		}
		values, err := json.Marshal(stringDictToGo(c.values))
		if err != nil {
			return err
		}
		buffer := &bytes.Buffer{}
		if err := c.Package(buffer, false); err != nil {
			return err
		}
		data, err := json.Marshal(map[string][]byte{
			"values": values,
			"chart":  buffer.Bytes(),
		})
		if err != nil {
			return err
		}
		w(&Object{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
			MetaData: MetaData{
				Name: "wonderix.chart." + c.GetName(),
			},
			Additional: map[string]json.RawMessage{
				"type": json.RawMessage([]byte(`"github.com/wonderix/shalm"`)),
				"data": json.RawMessage(data),
			},
		})
		return nil
	}
}

func (c *chartImpl) Delete(thread *starlark.Thread) error {
	_, err := starlark.Call(thread, c.methods["delete"], nil, nil)
	if err != nil {
		return err
	}
	c.k8s.Progress(100)
	return nil
}

func (c *chartImpl) deleteFunction() starlark.Callable {
	return c.builtin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if err := starlark.UnpackArgs("delete", args, kwargs); err != nil {
			return nil, err
		}
		return starlark.None, c.delete(thread)
	})
}

func (c *chartImpl) delete(thread *starlark.Thread) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := starlark.Call(thread, subChart.methods["delete"], nil, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.deleteLocal(thread, &K8sOptions{}, "")
}

func (c *chartImpl) deleteLocalFunction() starlark.Callable {
	return c.builtin("__delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var glob string
		parser := &kwargsParser{kwargs: kwargs}
		k8sOptions := unpackK8sOptions(parser)
		if err := starlark.UnpackArgs("__delete", args, parser.Parse(), "glob?", &glob); err != nil {
			return nil, err
		}
		return starlark.None, c.deleteLocal(thread, k8sOptions, glob)
	})
}

func (c *chartImpl) deleteLocal(thread *starlark.Thread, k8sOptions *K8sOptions, glob string) error {
	if c.readOnly {
		return nil
	}
	k8sOptions.Namespaced = false
	err := c.k8s.Delete(concat(decode(c.template(thread, glob, false)), c.packedChart()), k8sOptions)
	if err != nil {
		return err
	}
	return c.eachVault(func(v *vault) error {
		return v.delete()
	})
}

func (c *chartImpl) eachSubChart(block func(subChart *chartImpl) error) error {
	for _, v := range c.values {
		subChart, ok := v.(*chartImpl)
		if ok {
			err := block(subChart)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *chartImpl) eachVault(block func(x *vault) error) error {
	for _, val := range c.values {
		v, ok := val.(*vault)
		if ok {
			err := block(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *chartImpl) mergeValues(values map[string]interface{}) {
	for k, v := range values {
		c.values[k] = merge(c.values[k], toStarlark(v))
	}
}
