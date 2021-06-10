# Disable built-in rules.
.SUFFIXES:

# Image URL to use all building/pushing image targets
TAG ?= latest
CONFIG ?= dev
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
K3D ?= $(shell [ "`command -v kubectl`" != '' ] && [ "`kubectl config current-context`" = k3d-k3s-default ] && echo true || echo false)


# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build: generate manifests
	go build ./...

# Run tests
test:
	go test -v ./... -coverprofile cover.out

pre-commit: codegen test install lint start

codegen: generate manifests proto config/ci.yaml config/default.yaml config/dev.yaml config/kafka-default.yaml config/quick-start.yaml config/stan-default.yaml examples CHANGELOG.md $(GOBIN)/pie
	go generate ./...

$(GOBIN)/pie:
	go install github.com/elliotchance/pie@v1.38.2
$(GOBIN)/goreman:
	go install github.com/mattn/goreman@v0.3.7

# Run against the configured Kubernetes cluster in ~/.kube/config
start: generate deploy $(GOBIN)/goreman
	kubectl config set-context --current --namespace=argo-dataflow-system
	goreman -set-ports=false -logtime=false start
wait:
	kubectl -n argo-dataflow-system get pod
	kubectl -n argo-dataflow-system wait pod --all --for=condition=Ready --timeout=2m
logs: $(GOBIN)/stern
	stern -n argo-dataflow-system --tail=3 .

# Install CRDs into a cluster
install:
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall:
	kustomize build config/crd | kubectl delete --ignore-not-found -f -

images: controller runner runtimes

config/ci.yaml:
config/default.yaml:
config/dev.yaml:
config/kafka-default.yaml:
config/quick-start.yaml:
config/stan-default.yaml:
config/%.yaml: /dev/null
	kustomize build --load_restrictor=none config/$* -o $@
	sed -i '' "s/:latest/:$(TAG)/" $@

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install
	grep -o 'image: .*' config/$(CONFIG).yaml | grep -v dataflow | sort -u | cut -c 8- | xargs -L 1
	kubectl apply --force -f config/$(CONFIG).yaml

undeploy:
	kubectl delete --ignore-not-found -f config/$(CONFIG).yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(GOBIN)/controller-gen $(shell find api -name '*.go' -not -name '*generated*')
	$(GOBIN)/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: $(GOBIN)/controller-gen
	$(GOBIN)/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

docs/EXAMPLES.md: $(shell find examples -name '*.yaml') examples/main.go
	go run ./examples > docs/EXAMPLES.md

.PHONY: CHANGELOG.md
CHANGELOG.md: /dev/null
	./hack/changelog.sh > CHANGELOG.md

# not dependant on api/v1alpha1/generated.proto because it often does not change when this target runs, so results in remakes when they are not needed
proto: api/v1alpha1/generated.pb.go

$(GOBIN)/go-to-protobuf:
	go install k8s.io/code-generator/cmd/go-to-protobuf@v0.19.6

api/v1alpha1/generated.pb.go:
api/v1alpha1/generated.%: $(shell find api/v1alpha1 -type f -name '*.go' -not -name '*generated*' -not -name groupversion_info.go) $(GOBIN)/go-to-protobuf
	[ ! -e api/v1alpha1/groupversion_info.go ] || mv api/v1alpha1/groupversion_info.go api/v1alpha1/groupversion_info.go.0
	go-to-protobuf \
		--go-header-file=./hack/boilerplate.go.txt \
  		--packages=github.com/argoproj-labs/argo-dataflow/api/v1alpha1 \
		--apimachinery-packages=+k8s.io/apimachinery/pkg/util/intstr,+k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime/schema,+k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1
	mv api/v1alpha1/groupversion_info.go.0 api/v1alpha1/groupversion_info.go
	go mod tidy

lint:
	go mod tidy
	golangci-lint run --fix
	kubectl apply --dry-run=client -f examples

.PHONY: controller
controller: controller-image

.PHONY: runner

runner: runner-image

.PHONY: runtimes
runtimes: go1-16 java16 python3-9

go1-16: go1-16-image
java16: java16-image
python3-9: python3-9-image

%-image:
	docker buildx build . --target $* --tag quay.io/argoproj/dataflow-$*:$(TAG) --load
ifeq ($(K3D),true)
	k3d image import quay.io/argoproj/dataflow-$*:$(TAG)
endif
scan-%:
	docker scan --severity=high quay.io/argoproj/dataflow-$*:$(TAG)

$(GOBIN)/controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1

version:=2.3.2
name:=darwin
arch:=amd64

kubebuilder:
	# download the release
	curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(version)/kubebuilder_$(version)_$(name)_$(arch).tar.gz"

	# extract the archive
	tar -zxvf kubebuilder_$(version)_$(name)_$(arch).tar.gz
	mv kubebuilder_$(version)_$(name)_$(arch) kubebuilder && sudo mv kubebuilder /usr/local/

config/kafka/dev-small.yaml:
	kustomize build -o config/kafka/dev-small.yaml github.com/Yolean/kubernetes-kafka/variants/dev-small/?ref=v6.0.3

config/nats/single-server-nats.yml:
	curl -o config/nats/single-server-nats.yml https://raw.githubusercontent.com/nats-io/k8s/v0.7.4/nats-server/single-server-nats.yml

config/stan/single-server-stan.yml:
	curl -o config/stan/single-server-stan.yml https://raw.githubusercontent.com/nats-io/k8s/v0.7.4/nats-streaming-server/single-server-stan.yml

examples: $(shell find examples -name '*-pipeline.yaml' | sort) docs/EXAMPLES.md

install-dsls: /dev/null
	pip3 install dsls/python

examples/%-pipeline.yaml: examples/%-pipeline.py dsls/python/*.py install-dsls
	cd examples && python3 $*-pipeline.py

.PHONY: test-examples
test-examples: examples
	go test -timeout 20m -tags examples -v -count 1 ./examples

argocli:
	cd ../../argoproj/argo-workflows && git checkout dev-dataflow && make ./dist/argo DEV_BRANCH=true && ./dist/argo server --secure=false --namespaced --auth-mode=server --namespace=argo-dataflow-system
ui:
	killall node || true
	cd ../../argoproj/argo-workflows && yarn --cwd ui install && yarn --cwd ui start
