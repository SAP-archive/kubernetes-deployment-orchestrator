package shalm

import (
	"fmt"
	"regexp"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"go.starlark.net/starlark"
)

var nameRegexp = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9\\-\\./]*$")

type chartClass struct {
	APIVersion  string   `json:"apiVersion,omitempty"`
	Name        string   `json:"name,omitempty"`
	ID          string   `json:"id,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Home        string   `json:"home,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	Icon        string   `json:"icon,omitempty"`
}

var _ starlark.HasSetField = (*chartClass)(nil)

// String -
func (cc *chartClass) String() string { return cc.Name }

// Type -
func (cc *chartClass) Type() string { return "chart_class" }

// Freeze -
func (cc *chartClass) Freeze() {}

// Truth -
func (cc *chartClass) Truth() starlark.Bool { return true }

// Hash -
func (cc *chartClass) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (cc *chartClass) Attr(name string) (starlark.Value, error) {
	switch name {
	case "api_version":
		return starlark.String(cc.APIVersion), nil
	case "name":
		return starlark.String(cc.Name), nil
	case "version":
		return starlark.String(cc.Version), nil
	case "description":
		return starlark.String(cc.Description), nil
	case "keywords":
		return toStarlark(cc.Keywords), nil
	case "home":
		return starlark.String(cc.Home), nil
	case "sources":
		return toStarlark(cc.Sources), nil
	case "icon":
		return starlark.String(cc.Icon), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("chart_class has no .%s attribute", name))
}

// AttrNames -
func (cc *chartClass) AttrNames() []string {
	return []string{"api_version", "name", "version", "description", "keywords", "home", "sources", "icon"}
}

func validateName(name string) error {
	if !nameRegexp.MatchString(name) {
		return fmt.Errorf("Invalid name '%s' for chart", name)
	}
	return nil
}
func validateVersion(version string) error {
	if len(version) == 0 {
		return nil
	}
	_, err := semver.ParseTolerant(version)
	return err
}

func (cc *chartClass) Validate() error {
	if err := validateName(cc.Name); err != nil {
		return err
	}
	if err := validateVersion(cc.Version); err != nil {
		return err
	}
	return nil
}

func (cc *chartClass) GetVersion() semver.Version {
	if len(cc.Version) == 0 {
		return semver.Version{}
	}
	result, err := semver.ParseTolerant(cc.Version)
	if err != nil {
		panic(errors.Wrap(err, "Invalid version in helm chart"))
	}
	return result
}

// SetField -
func (cc *chartClass) SetField(name string, val starlark.Value) error {
	switch name {
	case "api_version":
		cc.APIVersion = val.(starlark.String).GoString()
		return nil
	case "name":
		name := val.(starlark.String).GoString()
		if err := validateName(name); err != nil {
			return err
		}
		cc.Name = name
		return nil
	case "version":
		version := val.(starlark.String).GoString()
		if err := validateVersion(version); err != nil {
			return err
		}
		cc.Version = version
		return nil
	case "description":
		cc.Description = val.(starlark.String).GoString()
		return nil
	case "home":
		cc.Home = val.(starlark.String).GoString()
		return nil
	case "icon":
		cc.Icon = val.(starlark.String).GoString()
		return nil
	}
	return starlark.NoSuchAttrError(fmt.Sprintf("chart_class has no .%s attribute", name))
}
