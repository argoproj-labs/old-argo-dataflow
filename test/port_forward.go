// +build test

package test

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func StartPortForward(opts ...interface{}) (stopPortForward func()) {
	port := 3569
	podName := ""
	for _, opt := range opts {
		switch v := opt.(type) {
		case int:
			port = v
		case string:
			podName = v
		default:
			panic("unknown option")
		}
	}
	if podName != "" {
		WaitForPod(podName)
	} else {
		podName = WaitForPod()
	}
	log.Printf("starting port-forward to pod %q on %d\n", podName, port)
	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		panic(err)
	}
	x, err := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", restConfig.Host, namespace, podName))
	if err != nil {
		panic(err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", x)
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", port, port)}, stopChan, readyChan, os.Stdout, os.Stderr)
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
	log.Printf("started port-forward to %q on %d\n", podName, port)
	return func() {
		stopChan <- struct{}{}
		// not needed
		// forwarder.Close()
		log.Printf("stopped port-forward to %q on %d\n", podName, port)
	}
}
