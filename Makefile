
# Image URL to use all building/pushing image targets
OS := $(shell uname )
VERSION := $(shell git describe --tags --always --dirty)
REPOSITORY := ulrichsap/kdo
IMG ?= ${REPOSITORY}:${VERSION}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

KUBERNETES_VERSION := v0.18.13

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GO_FILES:=$(shell git ls-files '*.go')

all: bin/kdo manager 

# Run tests
test: generate fmt vet 
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet 
	go run ./main.go

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: 
	go run ./main.go apply charts/kdo


# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

test-e2e : test-watch test-certificate test-controller 

test-controller: docker-build bin/kdo
#	bin/kdo apply charts/kdo --set local=True
#	bin/kdo apply --proxy local charts/example/simple/hello
#	while [ "$$(kubectl get kdochart hello -o 'jsonpath={.status.lastOp.progress}')" != "100" ] ; do sleep 5 ; done
#	kubectl get secret secret
#	bin/kdo delete --proxy local charts/example/simple/hello
#	bin/kdo delete charts/kdo --set local=True

test-watch: bin/kdo 
	bin/kdo apply charts/test/watch
	bin/kdo delete charts/test/watch

test-certificate: bin/kdo 
	bin/kdo apply charts/test/certificate
	bin/kdo apply charts/test/certificate
	bin/kdo delete charts/test/certificate

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	cp config/crd/bases/sap.github.com_kdocharts.yaml charts/kdo/templates/crd.yaml

docker-context/kubectl: Makefile
	curl -SsL https://storage.googleapis.com/kubernetes-release/release/$(KUBERNETES_VERSION)/bin/linux/amd64/kubectl -o docker-context/kubectl
	chmod +x docker-context/kubectl

docker-context/kapp:  Makefile
	curl -SsL https://github.com/k14s/kapp/releases/download/v0.34.0/kapp-linux-amd64 -o docker-context/kapp
	chmod +x docker-context/kapp


docker-prepare:: docker-context/kdo docker-context/kubectl docker-context/kapp

docker-build:: docker-context/build

# Build the docker image
docker-context/build:  docker-context/kdo docker-context/kubectl docker-context/kapp
	docker build docker-context -f Dockerfile -t ${IMG}
	docker tag ${IMG} ${REPOSITORY}:latest
	touch docker-context/build

# Push the docker image
docker-push: docker-context/build
	docker push ${IMG}
	docker push ${REPOSITORY}:latest

chart:
	rm -rf /tmp/kdo
	cp -r charts/kdo /tmp/kdo
ifeq ($(OS),Darwin)
	sed -i '' -e 's|version:.*|version: ${VERSION}|g' /tmp/kdo/Chart.yaml
	sed -i '' -e 's|image: ulrichsap/kdo:.*|image: ulrichsap/kdo:${VERSION}|g' /tmp/kdo/templates/deployment.yaml
else
	sed -i -e 's|version:.*|version: ${VERSION}|g' /tmp/kdo/Chart.yaml
	sed -i -e 's|image: ulrichsap/kdo:.*|image: ulrichsap/kdo:${VERSION}|g' /tmp/kdo/templates/deployment.yaml
endif
	mkdir -p bin
	cd bin && go run .. package /tmp/kdo

kdo:: bin/kdo

VERSION_FLAGS := "-X github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo.version=${VERSION} -X github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo.kubeVersion=${KUBERNETES_VERSION}"

bin/kdo: $(GO_FILES)  go.sum go.mod Makefile
	CGO_ENABLED=0 GOARCH=amd64 GO111MODULE=on go build -ldflags ${VERSION_FLAGS} -o bin/kdo . 

docker-context/kdo:  bin/linux/kdo
	mkdir -p docker-context/
	cp bin/linux/kdo docker-context/kdo


define BOZO
bin/$(1)/kdo: $(GO_FILES)  go.sum go.mod Makefile
	mkdir -p bin/$(1)
	CGO_ENABLED=0 GOOS=$(1) GOARCH=amd64 GO111MODULE=on go build -ldflags ${VERSION_FLAGS} -o bin/$(1)/kdo .
bin/kdo-binary-$(1).tgz: bin/$(1)/kdo
	cd bin/$(1) &&  tar czf ../kdo-binary-$(1).tgz kdo
endef

$(foreach i,linux darwin windows,$(eval $(call BOZO,$(i))))

binaries: $(foreach i,linux darwin windows,bin/kdo-binary-$(i).tgz)

formula: homebrew-tap/kdo.rb

homebrew-tap/kdo.rb: bin/kdo-binary-darwin.tgz bin/kdo-binary-linux.tgz
	@mkdir -p homebrew-tap
	@sed  \
	-e "s/{{sha256-darwin}}/$$(shasum -b -a 256 bin/kdo-binary-darwin.tgz  | awk '{print $$1}')/g" \
	-e "s/{{sha256-linux}}/$$(shasum -b -a 256 bin/kdo-binary-linux.tgz  | awk '{print $$1}')/g" \
	-e "s/{{version}}/$(VERSION)/g" homebrew-formula.rb \
	> homebrew-tap/kdo.rb

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
