package shalm

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/spf13/pflag"
	"go.starlark.net/starlark"
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

// KwArgsYamlVar -
type KwArgsYamlVar struct {
	kwargs *KwArgsVar
}

func (p KwArgsYamlVar) String() string {
	return p.kwargs.String()
}

// Set -
func (p *KwArgsYamlVar) Set(val string) error {
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
func (p KwArgsYamlVar) Type() string {
	return "kwargs-yaml"
}

// KwArgsFileVar -
type KwArgsFileVar struct {
	kwargs *KwArgsVar
}

func (p KwArgsFileVar) String() string {
	return p.kwargs.String()
}

// Set -
func (p *KwArgsFileVar) Set(val string) error {
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
func (p KwArgsFileVar) Type() string {
	return "kwargs-file"
}

// ChartOptions -
type ChartOptions struct {
	namespace string
	suffix    string
	proxyMode ProxyMode
	args      starlark.Tuple
	kwargs    KwArgsVar
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

// AddFlags -
func (v *ChartOptions) AddFlags(flagsSet *pflag.FlagSet) {
	defaultNamespace := os.Getenv("SHALM_NAMESPACE")
	if defaultNamespace == "" {
		defaultNamespace = "default"
	}
	flagsSet.Var(&v.kwargs, "set", "Set values (key=val).")
	flagsSet.Var(&KwArgsYamlVar{kwargs: &v.kwargs}, "set-yaml", "Set values from respective YAML files (key=path).")
	flagsSet.Var(&KwArgsFileVar{kwargs: &v.kwargs}, "set-file", "Set values from respective files (key=path).")
	flagsSet.VarP(&v.proxyMode, "proxy", "p", "Install helm chart using a combination of CR and operator. Possible values off, local and remote")
	flagsSet.StringVarP(&v.namespace, "namespace", "n", defaultNamespace, "Namespace for installation")
	flagsSet.StringVarP(&v.suffix, "suffix", "s", "", "Suffix which is used to build the chart name")
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
