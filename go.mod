module github.com/sap/kubernetes-deployment-orchestrator

go 1.13

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/Masterminds/sprig/v3 v3.0.2
	github.com/drewolson/testflight v1.0.0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/fatih/color v1.9.0
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/k14s/starlark-go v0.0.0-20200720175618-3a5c849cc368
	github.com/k14s/ytt v0.26.1-0.20200402233022-1aaca8db2e6a
	github.com/manifoldco/promptui v0.7.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3-0.20200303032533-a447b6683e1c
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pivotal-cf/brokerapi v6.4.2+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/rickb777/date v1.12.4
	github.com/sabhiram/go-gitignore v0.0.0-20180611051255-d3107576ba94
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/go-open-service-broker-client/v2 v2.0.0-20200911103215-9787cad28392

)

replace github.com/k14s/ytt => github.com/wonderix/ytt v0.28.1-0.20200908051131-36914082e903
