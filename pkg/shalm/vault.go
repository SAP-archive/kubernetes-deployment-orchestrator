package shalm

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.starlark.net/starlark"
	corev1 "k8s.io/api/core/v1"
)

// VaultBackend -
type VaultBackend interface {
	Name() string
	Keys() map[string]string
	Apply(map[string][]byte) (map[string][]byte, error)
}

// ComplexVaultBackend -
type ComplexVaultBackend interface {
	VaultBackend
	Template() (map[string][]byte, error)
	Delete() error
}

const stateInit = 0
const stateLoaded = 1
const stateReady = 2

type vault struct {
	backend VaultBackend
	state   int
	name    string
	data    map[string][]byte
}

var (
	_ starlark.Value = (*vault)(nil)
)

// NewVault -
func NewVault(backend VaultBackend, name string) (starlark.Value, error) {
	return &vault{
		backend: backend,
		name:    name,
		data:    map[string][]byte{},
	}, nil
}

func (c *vault) delete() error {
	complex, ok := c.backend.(ComplexVaultBackend)
	if ok {
		return complex.Delete()
	}
	return nil
}

// String -
func (c *vault) String() string {
	buf := new(strings.Builder)
	buf.WriteString(c.backend.Name())
	buf.WriteByte('(')
	buf.WriteString("name = ")
	buf.WriteString(c.name)
	buf.WriteByte(')')
	return buf.String()
}

func (c *vault) read(k8s K8sReader) error {
	obj, err := k8s.Get("secret", c.name, &K8sOptions{Namespaced: true})
	if err != nil {
		if !k8s.IsNotExist(err) {
			return err
		}
	} else {
		if err := json.Unmarshal(obj.Additional["data"], &c.data); err != nil {
			return err
		}
	}
	c.state = stateLoaded
	return nil
}

func (c *vault) ensure() (err error) {
	var data map[string][]byte
	switch c.state {
	case stateLoaded:
		data, err = c.backend.Apply(c.data)
		if err != nil {
			return
		}
	case stateInit:
		complex, ok := c.backend.(ComplexVaultBackend)
		if ok {
			data, err = complex.Template()
		} else {
			data, err = c.backend.Apply(make(map[string][]byte))
		}
		if err != nil {
			return
		}
	case stateReady:
		return nil
	}
	c.data = data
	c.state = stateReady
	return nil
}

func (c *vault) object(namespace string) (*Object, error) {
	err := c.ensure()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(c.data)
	if err != nil {
		return nil, err
	}
	return &Object{
		APIVersion: corev1.SchemeGroupVersion.String(),
		Kind:       "Secret",
		MetaData: MetaData{
			Name:      c.name,
			Namespace: namespace,
		},
		Additional: map[string]json.RawMessage{
			"type": json.RawMessage([]byte(`"Opaque"`)),
			"data": json.RawMessage(data),
		},
	}, nil
}

func (c *vault) templateValues() map[string]string {
	_ = c.ensure()
	result := map[string]string{}
	for k, v := range c.backend.Keys() {
		result[k] = dataValue(c.data[v])
	}
	return result
}

func dataValue(data []byte) string {
	if data != nil {
		return string(data)
	} else {
		return ""
	}
}

// Type -
func (c *vault) Type() string { return c.backend.Name() }

// Freeze -
func (c *vault) Freeze() {}

// Truth -
func (c *vault) Truth() starlark.Bool { return false }

// Hash -
func (c *vault) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (c *vault) Attr(name string) (starlark.Value, error) {
	if name == "name" {
		return starlark.String(c.name), nil
	}
	key, ok := c.backend.Keys()[name]
	if !ok {
		return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("%s has no .%s attribute", c.backend.Name(), name))
	}
	err := c.ensure()
	if err != nil {
		return starlark.None, err
	}
	return starlark.String(dataValue(c.data[key])), nil
}

// AttrNames -
func (c *vault) AttrNames() []string {
	result := []string{"name"}
	for k := range c.backend.Keys() {
		result = append(result, k)
	}
	return result
}
