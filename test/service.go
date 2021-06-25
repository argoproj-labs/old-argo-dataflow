// +build test

package test

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/url"
	"time"
)

var (
	serviceInterface = kubernetesInterface.CoreV1().Services(namespace)
)

func WaitForService() {
	WaitForPod()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	list, err := serviceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to watch services: %w", err))
	}
	for _, x := range list.Items {
		if _, ok := x.Spec.Selector[KeyPipelineName]; ok {
			log.Printf("waiting for service %q\n", x.Name)
			InvokeTestAPI("/http/wait-for?url=%s", url.QueryEscape("http://"+x.Name))
		}
	}
}

func DeleteService(serviceName string) (restoreService func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc, err := serviceInterface.Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to get service %q: %w", serviceName, err))
	}
	if err := serviceInterface.Delete(ctx, serviceName, metav1.DeleteOptions{}); err != nil {
		panic(fmt.Errorf("failed to delete service %q: %w", serviceName, err))
	}
	return func() {
		if _, err := serviceInterface.Create(ctx, svc, metav1.CreateOptions{}); err != nil {
			panic(fmt.Errorf("failed to create service %q: %w", serviceName, err))
		}
	}

}
