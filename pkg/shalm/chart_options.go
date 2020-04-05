package shalm

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/spf13/pflag"
	"go.starlark.net/starlark"
	"sigs.k8s.io/yaml"
)

// ProxyMode -
type ProxyMode int

const (
	// ProxyModeOff -
	ProxyModeOff = iota
	// ProxyModeLocal -
	ProxyModeLocal
	// ProxyModeRemote -
	ProxyModeRemote
)

func (p ProxyMode) String() string {
	return [...]string{"off", "local", "remote"}[p]
}

// Set -
func (p *ProxyMode) Set(val string) error {
	switch val {
	case "off":
		*p = ProxyModeOff
	case "":
		*p = ProxyModeOff
	case "local":
		*p = ProxyModeLocal
	case "remote":
		*p = ProxyModeRemote
	default:
		return fmt.Errorf("invalid proxy mode %s", val)
	}
	return nil
}

// Type -
func (p *ProxyMode) Type() string {
	return "proxy-mode"
}

// KwArgsVar -
type KwArgsVar []starlark.Tuple

func (p KwArgsVar) String() string {
	return `""`
}

var setVarRegexp = regexp.MustCompile("^([a-z][a-z0-9_]*)=(.*)")

func parseSet(val string, cb func(key string, value string) error) error {
	match := setVarRegexp.FindStringSubmatch(val)
	if match == nil {
		return fmt.Errorf("Invalid argument %s", val)
	}
	if err := cb(match[1], match[2]); err != nil {
		return err
	}
	return nil
}

// Set -
func (p *KwArgsVar) Set(val string) error {
	return parseSet(val, func(key string, value string) error {
		*p = append(*p, starlark.Tuple{starlark.String(key), starlark.String(value)})
		return nil
	})
}

// Type -
func (p KwArgsVar) Type() string {
	return "kwargs"
}

type kwArgsYamlVar struct {
	kwargs *KwArgsVar
}

func (p kwArgsYamlVar) String() string {
	return p.kwargs.String()
}

// Set -
func (p *kwArgsYamlVar) Set(val string) error {
	return parseSet(val, func(key string, value string) error {
		var v map[string]interface{}
		err := readYamlFile(value, &v)
		if err != nil {
			return err
		}
		*p.kwargs = append(*p.kwargs, starlark.Tuple{starlark.String(key), toStarlark(v)})
		return nil
	})
}

// Type -
func (p kwArgsYamlVar) Type() string {
	return "kwargs-yaml"
}

// kwArgsFileVar -
type kwArgsFileVar struct {
	kwargs *KwArgsVar
}

func (p kwArgsFileVar) String() string {
	return p.kwargs.String()
}

// Set -
func (p *kwArgsFileVar) Set(val string) error {
	return parseSet(val, func(key string, value string) error {
		data, err := ioutil.ReadFile(value)
		if err != nil {
			return err
		}
		*p.kwargs = append(*p.kwargs, starlark.Tuple{starlark.String(key), starlark.String(string(data))})
		return nil
	})
}

// Type -
func (p kwArgsFileVar) Type() string {
	return "kwargs-file"
}

type kwArgsEnvVar struct {
	kwargs *KwArgsVar
}

func (p kwArgsEnvVar) String() string {
	return p.kwargs.String()
}

// Set -
func (p *kwArgsEnvVar) Set(val string) error {
	return parseSet(val, func(key string, value string) error {
		*p.kwargs = append(*p.kwargs, starlark.Tuple{starlark.String(key), starlark.String(os.Getenv(value))})
		return nil
	})
}

// Type -
func (p kwArgsEnvVar) Type() string {
	return "kwargs-env"
}

type valuesFile map[string]interface{}

func (p valuesFile) String() string {
	if p == nil {
		return ""
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// Set -
func (p *valuesFile) Set(val string) error {
	return readYamlFile(val, p)
}

// Type -
func (p valuesFile) Type() string {
	return "values"
}

// ChartOptions -
type ChartOptions struct {
	namespace string
	suffix    string
	proxyMode ProxyMode
	args      starlark.Tuple
	kwargs    KwArgsVar
	values    valuesFile
	skipChart bool
}

// ChartOption -
type ChartOption func(options *ChartOptions)

// WithNamespace -
func WithNamespace(namespace string) ChartOption {
	return func(options *ChartOptions) { options.namespace = namespace }
}

// WithSuffix -
func WithSuffix(suffix string) ChartOption {
	return func(options *ChartOptions) { options.suffix = suffix }
}

// WithProxy -
func WithProxy(proxy ProxyMode) ChartOption {
	return func(options *ChartOptions) { options.proxyMode = proxy }
}

// WithArgs -
func WithArgs(args starlark.Tuple) ChartOption {
	return func(options *ChartOptions) { options.args = args }
}

// WithKwArgs -
func WithKwArgs(kwargs []starlark.Tuple) ChartOption {
	return func(options *ChartOptions) { options.kwargs = kwargs }
}

// WithValues -
func WithValues(values map[string]interface{}) ChartOption {
	return func(options *ChartOptions) { options.values = valuesFile(values) }
}

// WithSkipChart -
func WithSkipChart(value bool) ChartOption {
	return func(options *ChartOptions) { options.skipChart = value }
}

// AddFlags -
func (v *ChartOptions) AddFlags(flagsSet *pflag.FlagSet) {
	defaultNamespace := os.Getenv("SHALM_NAMESPACE")
	if defaultNamespace == "" {
		defaultNamespace = "default"
	}
	flagsSet.Var(&v.kwargs, "set", "Set values (key=val).")
	flagsSet.Var(&kwArgsYamlVar{kwargs: &v.kwargs}, "set-yaml", "Set values from respective YAML files (key=path).")
	flagsSet.Var(&kwArgsFileVar{kwargs: &v.kwargs}, "set-file", "Set values from respective files (key=path).")
	flagsSet.Var(&kwArgsEnvVar{kwargs: &v.kwargs}, "set-env", "Set values from respective environment variable (key=env).")
	flagsSet.VarP(&v.proxyMode, "proxy", "p", "Install helm chart using a combination of CR and operator. Possible values off, local and remote")
	flagsSet.StringVarP(&v.namespace, "namespace", "n", defaultNamespace, "Namespace for installation")
	flagsSet.StringVarP(&v.suffix, "suffix", "s", "", "Suffix which is used to build the chart name")
	flagsSet.VarP(&v.values, "values", "f", "Load additional values from a file")
}

// Merge -
func (v *ChartOptions) Merge() ChartOption {
	return func(o *ChartOptions) {
		*o = *v
	}
}

func chartOptions(opts []ChartOption) *ChartOptions {
	co := ChartOptions{}
	for _, option := range opts {
		option(&co)
	}
	if co.namespace == "" {
		co.namespace = "default"
	}
	return &co
}
