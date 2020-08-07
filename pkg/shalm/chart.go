package shalm

import (
	"bytes"
	"encoding/base64"
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

// Chart -
type Chart interface {
	GetID() string
	GetName() string
	GetVersion() semver.Version
	GetNamespace() string
	Apply(thread *starlark.Thread, k K8s) error
	Delete(thread *starlark.Thread, k K8s) error
	Template(thread *starlark.Thread) Stream
	Package(writer io.Writer, helmFormat bool) error
}

// ChartValue -
type ChartValue interface {
	starlark.HasSetField
	Chart
}

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
	_ ChartValue         = (*chartImpl)(nil)
	_ starlark.HasSetKey = (*chartImpl)(nil)
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
	if err := c.init(thread, repo, hasChartYaml, co.args, co.kwargs); err != nil {
		return nil, err
	}
	if len(co.id) != 0 {
		c.clazz.ID = co.id
	}
	if co.version != nil {
		c.clazz.Version = co.version.String()
	}
	c.mergeValues(co.values)
	return c, nil

}

func (c *chartImpl) builtin(name string, fn func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) starlark.Callable {
	return starlark.NewBuiltin(name+" at "+path.Join(c.dir, "Chart.star"), fn)
}

func (c *chartImpl) GetID() string {
	if len(c.clazz.ID) == 0 {
		return c.clazz.Name
	}
	return c.clazz.ID
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

func (c *chartImpl) GetNamespace() string {
	return c.namespace
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
	if name == "namespace" {
		return starlark.String(c.namespace), nil
	}
	if name == "name" {
		return starlark.String(c.GetName()), nil
	}
	if name == "__class__" {
		return &c.clazz, nil
	}
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
	return WrapDict(value), nil
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
	c.values[name] = UnwrapDict(val)
	return nil
}

func (c *chartImpl) Get(name starlark.Value) (starlark.Value, bool, error) {
	value, found := c.values[name.(starlark.String).GoString()]
	if !found {
		return starlark.None, found, nil
	}
	return WrapDict(value), true, nil
}

func (c *chartImpl) SetKey(name, value starlark.Value) error {
	c.values[name.(starlark.String).GoString()] = UnwrapDict(value)
	return nil
}

func (c *chartImpl) applyFunction() starlark.Callable {
	return c.builtin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("apply", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}
		return starlark.None, c.apply(thread, k)
	})
}

func (c *chartImpl) Apply(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.methods["apply"], starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	k.Progress(100)
	return nil
}

func (c *chartImpl) apply(thread *starlark.Thread, k K8sValue) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := starlark.Call(thread, subChart.methods["apply"], starlark.Tuple{k}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.applyLocal(thread, k, &K8sOptions{}, "")
}

func (c *chartImpl) applyLocalFunction() starlark.Callable {
	return c.builtin("__apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		parser := &kwargsParser{kwargs: kwargs}
		var glob string
		k8sOptions := unpackK8sOptions(parser)
		if err := starlark.UnpackArgs("__apply", args, parser.Parse(), "k8s", &k, "glob?", &glob); err != nil {
			return nil, err
		}
		return starlark.None, c.applyLocal(thread, k, k8sOptions, glob)
	})
}

func (c *chartImpl) applyLocal(thread *starlark.Thread, k K8sValue, k8sOptions *K8sOptions, glob string) error {
	vault := &vaultK8s{k8s: k, namespace: c.namespace}
	err := c.eachJewel(func(v *jewel) error {
		return v.read(vault)
	})
	if err != nil {
		return err
	}
	if c.readOnly {
		return nil
	}
	k8sOptions.Namespaced = false
	return k.Apply(concat(decode(c.template(thread, glob, true)), c.nameSpaceObject(), c.packedChartObject()), k8sOptions)
}

func (c *chartImpl) nameSpaceObject() ObjectStream {
	return func(w ObjectWriter) error {
		return nil
		// return w(&Object{
		// 	APIVersion: corev1.SchemeGroupVersion.String(),
		// 	Kind:       "Namespace",
		// 	MetaData: MetaData{
		// 		Name: c.namespace,
		// 	},
		// })
	}
}

func (c *chartImpl) packedChartObject() ObjectStream {
	return func(w ObjectWriter) error {
		if c.skipChart {
			return nil
		}
		// values, err := json.Marshal(stringDictToGo(c.values))
		// if err != nil {
		// 	return err
		// }
		buffer := &bytes.Buffer{}
		if err := c.Package(buffer, false); err != nil {
			return err
		}
		data, err := json.Marshal(map[string]string{
			"id":      c.GetID(),
			"version": c.GetVersion().String(),
			"chart":   base64.StdEncoding.EncodeToString(buffer.Bytes()),
		})
		if err != nil {
			return err
		}
		w(&Object{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
			MetaData: MetaData{
				Name:      "shalm." + c.GetName(),
				Namespace: c.namespace,
				Labels: map[string]string{
					"shalm.wonderix.github.com/chart": "true",
					"shalm.wonderix.github.com/id":    c.GetID(),
				},
			},
			Additional: map[string]json.RawMessage{
				"data": data,
			},
		})
		return nil
	}
}

func (c *chartImpl) Delete(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.methods["delete"], starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	k.Progress(100)
	return nil
}

func (c *chartImpl) deleteFunction() starlark.Callable {
	return c.builtin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("delete", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}
		return starlark.None, c.delete(thread, k)
	})
}

func (c *chartImpl) delete(thread *starlark.Thread, k K8sValue) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := starlark.Call(thread, subChart.methods["delete"], starlark.Tuple{k}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.deleteLocal(thread, k, &K8sOptions{}, "")
}

func (c *chartImpl) deleteLocalFunction() starlark.Callable {
	return c.builtin("__delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		var glob string
		parser := &kwargsParser{kwargs: kwargs}
		k8sOptions := unpackK8sOptions(parser)
		if err := starlark.UnpackArgs("__delete", args, parser.Parse(), "k8s", &k, "glob?", &glob); err != nil {
			return nil, err
		}
		return starlark.None, c.deleteLocal(thread, k, k8sOptions, glob)
	})
}

func (c *chartImpl) deleteLocal(thread *starlark.Thread, k K8sValue, k8sOptions *K8sOptions, glob string) error {
	if c.readOnly {
		return nil
	}
	k8sOptions.Namespaced = false
	err := k.Delete(concat(decode(c.template(thread, glob, false)), c.packedChartObject()), k8sOptions)
	if err != nil {
		return err
	}
	return c.eachJewel(func(v *jewel) error {
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

func (c *chartImpl) eachJewel(block func(x *jewel) error) error {
	for _, val := range c.values {
		v, ok := val.(*jewel)
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
