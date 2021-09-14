// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var serviceInterface = kubernetesInterface.CoreV1().Services(namespace)

func WaitForService() {
	WaitForPod()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	list, err := serviceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to watch services: %w", err))
	}
	for _, x := range list.Items {
		x.Spec.ClusterIP == "None" {
			continue
		}
		if _, ok := x.Spec.Selector[KeyPipelineName]; ok {
			log.Printf("waiting for service %q\n", x.Name)
			InvokeTestAPI("/http/wait-for?url=%s", url.QueryEscape("https://"+x.Name))
		}
	}
}
