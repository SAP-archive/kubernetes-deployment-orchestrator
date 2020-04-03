package shalm

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
)

// K8sInMemory in memory implementation of K8s
type K8sInMemory struct {
	namespace string
	objects   map[string]Object
}

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}

// NewK8sInMemory creates a new K8sInMemory instance
func NewK8sInMemory(namespace string, objects ...Object) *K8sInMemory {
	result := &K8sInMemory{namespace: namespace, objects: map[string]Object{}}
	for _, obj := range objects {
		result.objects[result.key(obj.Kind, obj.MetaData.Name)] = obj
	}
	return result
}

var _ K8s = (*K8sInMemory)(nil)

func (k K8sInMemory) Host() string {
	return "memory.local"
}

// ForSubChart -
func (k K8sInMemory) ForSubChart(namespace string, app string, version semver.Version) K8s {
	return &K8sInMemory{namespace: namespace, objects: k.objects}
}

// Inspect -
func (k K8sInMemory) Inspect() string {
	return k.namespace
}

// Tool -
func (k K8sInMemory) Tool() Tool {
	return ToolKubectl
}

// Watch -
func (k K8sInMemory) Watch(kind string, name string, options *K8sOptions) ObjectStream {
	obj, err := k.GetObject(kind, name)
	if err != nil {
		return ObjectErrorStream(err)
	}
	return func(writer ObjectWriter) error {
		return writer(obj)
	}
}

// RolloutStatus -
func (k K8sInMemory) RolloutStatus(kind string, name string, options *K8sOptions) error {
	_, err := k.GetObject(kind, name)
	if err != nil {
		return err
	}
	return nil
}

// Wait -
func (k K8sInMemory) Wait(kind string, name string, condition string, options *K8sOptions) error {
	return k.RolloutStatus(kind, name, options)
}

// DeleteObject -
func (k K8sInMemory) DeleteObject(kind string, name string, options *K8sOptions) error {
	delete(k.objects, k.key(kind, name))
	return nil
}

// Apply -
func (k K8sInMemory) Apply(output ObjectStream, options *K8sOptions) error {
	return output(func(obj *Object) error {
		k.objects[k.key(obj.Kind, obj.MetaData.Name)] = *obj
		return nil
	})
}

// Delete -
func (k K8sInMemory) Delete(output ObjectStream, options *K8sOptions) error {
	return output(func(obj *Object) error {
		delete(k.objects, k.key(obj.Kind, obj.MetaData.Name))
		return nil
	})
}

// Get -
func (k K8sInMemory) Get(kind string, name string, options *K8sOptions) (*Object, error) {
	return k.GetObject(kind, name)
}

// IsNotExist -
func (k K8sInMemory) IsNotExist(err error) bool {
	_, ok := err.(notFoundError)
	return ok
}

// ConfigContent -
func (k K8sInMemory) ConfigContent() *string {
	return nil
}

// ForConfig -
func (k K8sInMemory) ForConfig(config string) (K8s, error) {
	return k, nil
}

func (k K8sInMemory) key(kind, name string) string {
	kind = strings.ToLower(kind)
	if isNameSpaced(kind) {
		return fmt.Sprintf("%s/%s/%s", k.namespace, kind, name)
	}
	return fmt.Sprintf("%s/%s", kind, name)
}

// GetObject -
func (k K8sInMemory) GetObject(kind string, name string) (*Object, error) {
	obj, ok := k.objects[k.key(kind, name)]
	if !ok {
		keys := []string{}
		for k := range k.objects {
			keys = append(keys, k)
		}
		return nil, notFoundError(fmt.Sprintf("NotFound: %s %s ", k.key(kind, name), strings.Join(keys, ", ")))
	}
	return &obj, nil
}

// Progress -
func (k K8sInMemory) Progress(progress int) {
}
