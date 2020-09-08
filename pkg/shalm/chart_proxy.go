package shalm

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	shalmv1a2 "github.com/wonderix/shalm/api/v1alpha2"

	"github.com/k14s/starlark-go/starlark"
)

type chartProxy struct {
	*chartImpl
	args      []intstr.IntOrString
	kwArgs    map[string]interface{}
	url       string
	proxyMode ProxyMode
}

var (
	_ ChartValue = (*chartProxy)(nil)
)

func newChartProxy(delegate *chartImpl, url string, proxyMode ProxyMode, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error) {
	result := &chartProxy{
		chartImpl: delegate,
		url:       url,
		proxyMode: proxyMode,
	}
	var err error
	result.args, err = toIntOrStringArray(args)
	if err != nil {
		return nil, errors.Wrap(err, "proxy mode only supports int and string values")
	}
	result.kwArgs = kwargsToGo(kwargs)
	return result, nil
}

// Attr returns the value of the specified field.
func (c *chartProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "apply":
		return c.applyFunction(), nil
	case "delete":
		return c.deleteFunction(), nil
	}
	return c.chartImpl.Attr(name)
}

func (c *chartProxy) applyFunction() starlark.Callable {
	return c.chartImpl.builtin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("apply", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}
		var k8s K8s = k.(K8s)
		namespace := &corev1.Namespace{
			TypeMeta: v1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name: c.namespace,
			},
		}
		shalmSpec := shalmv1a2.ChartSpec{
			Args:      c.args,
			Namespace: c.namespace,
			Suffix:    c.suffix,
			Tool:      k8s.Tool().String(),
		}
		shalmSpec.SetValues(stringDictToGo(c.chartImpl.values))
		shalmSpec.SetKwArgs(c.kwArgs)
		if c.proxyMode == ProxyModeLocal {
			kubeConfig := k.ConfigContent()
			if kubeConfig != nil {
				shalmSpec.KubeConfig = *kubeConfig
			}
			var err error
			k8s, err = k8s.ForConfig("")
			if err != nil {
				return starlark.None, err
			}
		}
		if isValidShalmURL(c.url) {
			shalmSpec.ChartURL = c.url
		} else {
			buffer := &bytes.Buffer{}
			if err := c.chartImpl.Package(buffer, false); err != nil {
				return nil, err
			}
			shalmSpec.ChartTgz = buffer.Bytes()
		}
		shalmChart := &shalmv1a2.ShalmChart{
			TypeMeta: v1.TypeMeta{
				Kind:       "ShalmChart",
				APIVersion: shalmv1a2.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      c.GetName(),
				Namespace: c.namespace,
			},
			Spec: shalmSpec,
		}

		return starlark.None, k8s.Apply(objectStreamOf(namespace, shalmChart), &K8sOptions{})
	})
}

func (c *chartProxy) Apply(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.applyFunction(), starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *chartProxy) deleteFunction() starlark.Callable {
	return c.chartImpl.builtin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("delete", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}

		return starlark.None, k.DeleteObject("ShalmChart", c.GetName(), &K8sOptions{})
	})
}

func (c *chartProxy) Delete(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.deleteFunction(), starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	return nil

}

func isValidShalmURL(u string) bool {
	return strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")
}
