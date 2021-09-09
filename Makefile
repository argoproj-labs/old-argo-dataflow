# Disable built-in rules.
.SUFFIXES:

# Image URL to use all building/pushing image targets
TAG ?= latest
VERSION ?= v0.0.0-latest-0
CONFIG ?= dev
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
K3D ?= $(shell [ "`command -v kubectl`" != '' ] && [ "`kubectl config current-context`" = k3d-k3s-default ] && echo true || echo false)
UI ?= false

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build:
	go build ./...

# Run tests
.PHONY: test
test:
ifeq ($(CI),true)
	go test -v ./... -coverprofile cover.out
else
	go test -v ./...
endif

test-fmea: test-http-fmea test-kafka-fmea test-stan-fmea
test-stress: test-http-stress test-kafka-stress test-stan-stress

test-stress-2-replicas:
	env REPLICAS=2 $(MAKE) test-stress

test-stress-large-messages:
	env MESSAGE_SIZE=1000000 $(MAKE) test-stress

test-db-e2e:
test-e2e:
test-examples:
test-http-fmea:
test-http-stress:
test-hpa:
	kubectl -n kube-system apply -k config/apps/metrics-server
	kubectl -n kube-system wait deploy/metrics-server --for=condition=available
	kubectl -n argo-dataflow-system delete hpa --all
	kubectl -n argo-dataflow-system delete pipeline --all
	kubectl -n argo-dataflow-system apply -f examples/101-hello-pipeline.yaml
	kubectl -n argo-dataflow-system wait pipeline/101-hello --for=condition=running
	if [ `kubectl -n argo-dataflow-system get step 101-hello-main -o=jsonpath='{.status.replicas}'` != 1 ]; then exit 1; fi
	kubectl -n argo-dataflow-system autoscale step 101-hello-main --min 2 --max 2
	sleep 20s
	if [ `kubectl -n argo-dataflow-system get step 101-hello-main -o=jsonpath='{.status.replicas}'` != 2 ]; then exit 1; fi
test-kafka-e2e:
test-kafka-fmea:
test-kafka-stress:
test-s3-e2e:
test-stan-e2e:
test-stan-fmea:
test-stan-stress:
test-%:
	go generate $(shell find ./test/$* -name '*.go')
	kubectl -n argo-dataflow-system wait pod -l statefulset.kubernetes.io/pod-name --for condition=ready --timeout=2m
	go test -count 1 -v --tags test ./test/$*

pprof:
	go tool pprof -web http://127.0.0.1:3569/debug/pprof/allocs
	go tool pprof -web http://127.0.0.1:3569/debug/pprof/heap
	go tool pprof -web http://127.0.0.1:3569/debug/pprof/profile?seconds=10
	curl -s http://127.0.0.1:3569/debug/pprof/trace\?seconds\=10 | go tool trace /dev/stdin

pre-commit: codegen proto test install runner testapi lint start

codegen: generate manifests examples tests
	go generate ./...
	cd runtimes/golang1-16 && go get -u github.com/argoproj-labs/argo-dataflow && go mod tidy
	cd examples/git && go get -u github.com/argoproj-labs/argo-dataflow && go mod tidy

$(GOBIN)/goreman:
	go install github.com/mattn/goreman@v0.3.7

# Run against the configured Kubernetes cluster in ~/.kube/config
start: deploy build runner $(GOBIN)/goreman wait
	kubectl config set-context --current --namespace=argo-dataflow-system
	env UI=$(UI) goreman -set-ports=false -logtime=false start
wait:
	kubectl -n argo-dataflow-system get pod
	kubectl -n argo-dataflow-system wait deploy --all --for=condition=available --timeout=2m
	# kubectl wait does not work for statesfulsets, as statefulsets do not have conditions
	kubectl -n argo-dataflow-system wait pod -l statefulset.kubernetes.io/pod-name --for condition=ready
logs: $(GOBIN)/stern
	stern -n argo-dataflow-system --tail=3 -l dataflow.argoproj.io/step-name .

# Install CRDs into a cluster
install:
	kubectl kustomize config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall:
	kubectl kustomize config/crd | kubectl delete --ignore-not-found -f -

images: controller runner testapi runtimes

config/%.yaml: config/$*
	kubectl kustomize --load-restrictor=LoadRestrictionsNone config/$* -o $@
	sed "s/:latest/:$(TAG)/" $@ > tmp
	mv tmp $@

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install
	kubectl apply --force -f config/$(CONFIG).yaml

undeploy:
	kubectl delete --ignore-not-found -f config/$(CONFIG).yaml

crds: $(GOBIN)/controller-gen $(shell find api -name '*.go' -not -name '*generated*')
	$(GOBIN)/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate manifests e.g. CRD, RBAC etc.
manifests: crds $(shell find config config/apps -maxdepth 1 -name '*.yaml')

generate: $(GOBIN)/controller-gen
	$(GOBIN)/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

docs/EXAMPLES.md: examples/main.go
	go run ./examples docs | grep -v 'time=' > docs/EXAMPLES.md

test/examples/examples_test.go: examples/main.go
	go run ./examples tests | grep -v 'time=' > test/examples/examples_test.go
	gofmt -w test/examples/examples_test.go

.PHONY: CHANGELOG.md
CHANGELOG.md: /dev/null
	./hack/changelog.sh > CHANGELOG.md

# not dependant on api/v1alpha1/generated.proto because it often does not change when this target runs, so results in remakes when they are not needed
proto: api/v1alpha1/generated.pb.go

$(GOBIN)/go-to-protobuf:
	go install k8s.io/code-generator/cmd/go-to-protobuf@v0.20.4

api/v1alpha1/generated.pb.go:
api/v1alpha1/generated.%: $(shell find api/v1alpha1 -type f -name '*.go' -not -name '*generated*' -not -name groupversion_info.go) $(GOBIN)/go-to-protobuf
	[ ! -e api/v1alpha1/groupversion_info.go ] || mv api/v1alpha1/groupversion_info.go api/v1alpha1/groupversion_info.go.0
	go-to-protobuf \
		--go-header-file=./hack/boilerplate.go.txt \
  		--packages=github.com/argoproj-labs/argo-dataflow/api/v1alpha1 \
		--apimachinery-packages=+k8s.io/apimachinery/pkg/util/intstr,+k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime/schema,+k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1
	mv api/v1alpha1/groupversion_info.go.0 api/v1alpha1/groupversion_info.go
	go mod tidy

$(GOPATH)/bin/gofumpt:
	go install mvdan.cc/gofumpt@v0.1.1

lint: $(GOPATH)/bin/gofumpt
	go mod tidy
	# run gofumpt outside of golangci-lint because it seems to be unreliable inside it
	gofumpt -l -w .
	golangci-lint run --fix
	kubectl apply --dry-run=client -f examples

.PHONY: controller
controller: controller-image

.PHONY: runner
runner: runner-image

.PHONY: testapi
testapi: testapi-image

.PHONY: runtimes
runtimes: golang1-16 java16 python3-9 node16

golang1-16: golang1-16-image
java16: java16-image
python3-9: python3-9-image
node16: node16-image

%-image:
	docker buildx build . --target $* --tag sarabala1979/dataflow-$*:$(TAG) --load --build-arg VERSION="$(VERSION)"
ifeq ($(K3D),true)
	k3d image import quay.io/argoprojlabs/dataflow-$*:$(TAG)
endif

scan: scan-controller scan-runner scan-testapi scan-golang1-16 scan-java16 scan-python3-9
	snyk test --severity-threshold=high

scan-%:
	docker scan --severity=high quay.io/argoprojlabs/dataflow-$*:$(TAG)

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

.PHONY: examples
examples: $(shell find examples -name '*-pipeline.yaml' | sort) docs/EXAMPLES.md

.PHONY: tests
tests: test/examples/examples_test.go

.PHONY: install-dsls
install-dsls:
	pip3 install --use-feature=in-tree-build dsls/python

examples/%-pipeline.yaml: examples/%-pipeline.py dsls/python/*.py install-dsls
	cd examples && python3 $*-pipeline.py

argocli:
	cd ../../argoproj/argo-workflows && make ./dist/argo DEV_BRANCH=true && ./dist/argo server --secure=false --namespaced --auth-mode=server --namespace=argo-dataflow-system
ui:
	killall node || true
	cd ../../argoproj/argo-workflows && yarn --cwd ui install && yarn --cwd ui start
