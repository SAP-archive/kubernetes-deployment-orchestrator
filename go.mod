module github.com/wonderix/shalm

go 1.13

require (
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/fatih/color v1.9.0
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/k14s/ytt v0.22.0
	github.com/manifoldco/promptui v0.7.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3-0.20200303032533-a447b6683e1c
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/procfs v0.0.0-20190522114515-bc1a522cf7b1 // indirect
	github.com/rickb777/date v1.12.4
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	go.starlark.net v0.0.0-20191021185836-28350e608555
	golang.org/x/sys v0.0.0-20200120151820-655fe14d7479 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0

)

replace go.starlark.net => github.com/wonderix/starlark-go v0.0.0-20200331102949-46a1d2522494

replace github.com/k14s/ytt => github.com/wonderix/ytt v0.26.1-0.20200331105310-71f8f1d9e7c8
