// +build e2e

package e2e

import (
	"fmt"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"log"
	"net/http"
	"net/url"
	"os"
)

func portForward(podName string) func() {
	log.Println(fmt.Sprintf("starting port-forward to pod %q", podName))
	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		panic(err)
	}
	// https://kubernetes.docker.internal:6443/api/v1/namespaces/argo-dataflow-system/pods/http-main-0/portforward
	x, err := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", restConfig.Host, namespace, podName))
	if err != nil {
		panic(err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", x)
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer, []string{"3569:3569"}, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		panic(err)
	}
	go func() {
		defer runtimeutil.HandleCrash()
		if err := forwarder.ForwardPorts(); err != nil {
			panic(err)
		}
	}()
	<-readyChan
	log.Println(fmt.Sprintf("started port-forward to %q", podName))
	return func() {
		forwarder.Close()
		log.Println(fmt.Sprintf("stopped port-forward to %q", podName))
	}
}
