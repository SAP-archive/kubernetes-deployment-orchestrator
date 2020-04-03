package shalm

import (
	"io"
	"time"

	"github.com/blang/semver"
	"go.starlark.net/starlark"

	shalmv1a2 "github.com/wonderix/shalm/api/v1alpha2"
)

//go:generate ./generate_fake.sh

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

// Stream -
type Stream = func(io.Writer) error

// Chart -
type Chart interface {
	GetName() string
	GetVersion() semver.Version
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

// K8sOptions common options for calls to k8s
type K8sOptions struct {
	Namespaced     bool
	Namespace      string
	Timeout        time.Duration
	IgnoreNotFound bool
}

// K8sReader kubernetes reader API
type K8sReader interface {
	Host() string
	Get(kind string, name string, options *K8sOptions) (*Object, error)
	IsNotExist(err error) bool
}

// K8s kubernetes API
type K8s interface {
	K8sReader
	ForSubChart(namespace string, app string, version semver.Version) K8s
	Inspect() string
	Watch(kind string, name string, options *K8sOptions) ObjectStream
	RolloutStatus(kind string, name string, options *K8sOptions) error
	Wait(kind string, name string, condition string, options *K8sOptions) error
	DeleteObject(kind string, name string, options *K8sOptions) error
	Apply(output ObjectStream, options *K8sOptions) error
	Delete(output ObjectStream, options *K8sOptions) error
	ConfigContent() *string
	ForConfig(config string) (K8s, error)
	Progress(progress int)
	Tool() Tool
}

// K8sValue -
type K8sValue interface {
	starlark.Value
	K8s
}

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, url string, options ...ChartOption) (ChartValue, error)
	// GetFromSpec -
	GetFromSpec(thread *starlark.Thread, spec *shalmv1a2.ChartSpec) (ChartValue, error)
}
