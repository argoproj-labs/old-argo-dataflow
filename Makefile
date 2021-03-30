
# Image URL to use all building/pushing image targets
TAG ?= latest
NS ?= $(shell kubectl config view --minify --output 'jsonpath={..namespace}')
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build: generate manifests
	go build ./...

# Run tests
test: build
	go test ./... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
start: generate install $(GOBIN)/kafka-console-consumer $(GOBIN)/kafka-console-producer
	KAFKA_PEERS=kafka-0.broker.kafka.svc.cluster.local:9092 goreman -set-ports=false -logtime=false start
logs: $(GOBIN)/stern
	stern --tail=3 .

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

undeploy: manifests
	kustomize build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	kustomize build config/default | sed 's/argoproj\//alexcollinsintuit\//' | sed 's/:latest/:$(TAG)/' > install/default.yaml

generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	go run ./examples > examples/README.md

proto: api/v1alpha1/generated.pb.go api/v1alpha1/generated.proto

api/v1alpha1/generated.%: api/v1alpha1/pipeline_types.go
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
	kubectl apply --dry-run=server -f examples

.PHONY: controller
controller:
	docker build . --target controller --tag argoproj/dataflow-controller:$(TAG)
.PHONY: runner
runner:
	docker build . --target runner --tag argoproj/dataflow-runner:$(TAG)

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

$(GOBIN)/kafka-console-consumer:
	go install github.com/Shopify/sarama/tools/kafka-console-consumer
$(GOBIN)/kafka-console-producer:
	go install github.com/Shopify/sarama/tools/kafka-console-producer

flood:
	go run ./kafka/ -topic input-topic -message flood-%d -sleep 0ms pump-topic

version:=2.3.2
name:=darwin
arch:=amd64

kubebuilder:
	# download the release
	curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(version)/kubebuilder_$(version)_$(name)_$(arch).tar.gz"

	# extract the archive
	tar -zxvf kubebuilder_$(version)_$(name)_$(arch).tar.gz
	mv kubebuilder_$(version)_$(name)_$(arch) kubebuilder && sudo mv kubebuilder /usr/local/

.PHONY: kafka
kafka:
	kubectl get ns kafka || kubectl create ns kafka
	kubectl -n kafka apply -k github.com/Yolean/kubernetes-kafka/variants/dev-small/?ref=v6.0.3
kafka-9092: kafka
	kubectl -n kafka port-forward svc/broker 9092:9092
unkafka:
	kubectl delete ns kafka

nats:
	kubectl -n $(NS) apply -f https://raw.githubusercontent.com/nats-io/k8s/master/nats-server/single-server-nats.yml
	kubectl -n $(NS) apply -f https://raw.githubusercontent.com/nats-io/k8s/master/nats-streaming-server/single-server-stan.yml
nats-4222: nats
	kubectl port-forward svc/nats 4222:4222
nats-8222: nats
	kubectl port-forward svc/nats 8222:8222
unnats:
	kubectl -n $(NS) delete -f https://raw.githubusercontent.com/nats-io/k8s/master/nats-streaming-server/single-server-stan.yml
	kubectl -n $(NS) delete -f https://raw.githubusercontent.com/nats-io/k8s/master/nats-server/single-server-nats.yml

examples/%.yaml: /dev/null
	@kubectl delete pipeline --all --cascade=foreground > /dev/null
	@echo " ▶ RUN $@"
	@kubectl apply -f $@
	@kubectl wait pipeline --all --for condition=Running
	@echo
test-examples: $(shell ls examples/*.yaml | sort)