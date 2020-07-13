module github.com/wonderix/shalm

go 1.13

require (
	github.com/Masterminds/sprig/v3 v3.0.2
	github.com/blang/semver v3.5.1+incompatible
	github.com/fatih/color v1.9.0
	github.com/go-logr/logr v0.1.0
	github.com/k14s/ytt v0.26.1-0.20200402233022-1aaca8db2e6a
	github.com/manifoldco/promptui v0.7.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3-0.20200303032533-a447b6683e1c
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/rickb777/date v1.12.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.starlark.net v0.0.0-20191021185836-28350e608555
	golang.org/x/tools v0.0.0-20200407041343-bf15fae40dea // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/yaml v1.1.0

)

replace go.starlark.net => github.com/k14s/starlark-go v0.0.0-20200402152745-409c85f3828d // ytt branch

replace github.com/k14s/ytt => github.com/wonderix/ytt  v0.26.1-0.20200713050030-d986378ccb4d
