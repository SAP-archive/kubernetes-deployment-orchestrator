package shalm

import (
	"github.com/blang/semver"
)

var version = "latest"

var kubeVersion = "v1.17.0"

var kubeSemver, _ = semver.ParseTolerant(kubeVersion)

// Version -
func Version() string {
	return version
}

// DockerTag -
func DockerTag() string {
	v, err := semver.ParseTolerant(version)
	if err != nil || len(v.Pre) > 0 {
		return "latest"
	}
	return "v" + v.String()
}
