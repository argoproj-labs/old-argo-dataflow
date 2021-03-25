
# Image URL to use all building/pushing image targets
TAG ?= latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests docker-build
	cd config/manager && kustomize edit set image argoproj/dataflow-controller=argoproj/dataflow-controller:$(TAG)
	kustomize build config/default | kubectl apply -f -

redeploy: deploy
	kubectl -n argo-dataflow-system rollout restart deploy/controller-manager

logs:
	stern -n argo-dataflow-system .

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	kustomize build config/default | sed 's/argoproj\//alexcollinsintuit\//' | sed 's/:latest/:$(TAG)/' > install/default.yaml

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build . --target sidecar    --tag argoproj/dataflow-sidecar:$(TAG)
	docker build . --target cat        --tag argoproj/dataflow-cat:$(TAG)
	docker build . --target controller --tag argoproj/dataflow-controller:$(TAG)

# Push the docker image
docker-push:
	docker push argoproj/dataflow-sidecar:$(TAG)
	docker push argoproj/dataflow-cat:$(TAG)
	docker push argoproj/dataflow-controller:$(TAG)

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

version:=2.3.2
name:=darwin
arch:=amd64

kubebuilder:
	# download the release
	curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(version)/kubebuilder_$(version)_$(name)_$(arch).tar.gz"

	# extract the archive
	tar -zxvf kubebuilder_$(version)_$(name)_$(arch).tar.gz
	mv kubebuilder_$(version)_$(name)_$(arch) kubebuilder && sudo mv kubebuilder /usr/local/

kafka:
	kubectl get ns kafka || kubectl create ns kafka
	kubectl apply -k github.com/Yolean/kubernetes-kafka/variants/dev-small/?ref=v6.0.3
	kubectl port-forward -n kafka svc/broker 9092:9092

example:
	kubectl -n argo-dataflow-system delete pipeline --all
	sleep 5s
	kubectl -n argo-dataflow-system apply -f example-pipeline.yaml
	kubectl get -n argo-dataflow-system pipeline -w

lint:
	golangci-lint run --fix