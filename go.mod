module github.com/argoproj-labs/argo-dataflow

go 1.15

require (
	github.com/Shopify/sarama v1.28.0
	github.com/argoproj/argo-events v1.2.3
	github.com/go-logr/logr v0.3.0
	github.com/nats-io/nats.go v1.9.1
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	go.uber.org/zap v1.15.0
	k8s.io/api v0.19.6
	k8s.io/apimachinery v0.19.6
	k8s.io/client-go v0.19.6
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/controller-runtime v0.7.0
)
