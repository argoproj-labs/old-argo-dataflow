// +build e2e

package e2e

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"reflect"
	"time"
)

var (
	podInterface = kubernetesInterface.CoreV1().Pods(namespace)
	toBeReady    = func(p *corev1.Pod) bool {
		for _, c := range p.Status.Conditions {
			if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	}
)

func waitForPipelinePodsToBeDeleted() {
	log.Printf("waiting for pipeline pods to be deleted\n")
	ctx := context.Background()
	for {
		if list, err := podInterface.List(ctx, metav1.ListOptions{LabelSelector: KeyPipelineName}); err != nil {
			panic(err)
		} else if len(list.Items) == 0 {
			return
		}
		time.Sleep(time.Second)
	}
}

func waitForPod(podName string, f func(*corev1.Pod) bool) {
	log.Printf("watching pod %q\n", podName)
	w, err := podInterface.Watch(context.Background(), metav1.ListOptions{FieldSelector: "metadata.name=" + podName})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for e := range w.ResultChan() {
		p, ok := e.Object.(*corev1.Pod)
		if !ok {
			panic(fmt.Errorf("expected *corev1.Pod, got %q", reflect.TypeOf(e.Object).Name()))
		}
		log.Println(fmt.Sprintf("pod %q has status %s %q", p.Name, p.Status.Phase, p.Status.Message))
		if f(p) {
			return
		}
	}
}
