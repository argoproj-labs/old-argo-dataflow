// +build test

package test

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

var (
	podInterface = kubernetesInterface.CoreV1().Pods(namespace)
)

func ToBeReady(p *corev1.Pod) bool {
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func WaitForPodsToBeDeleted() {
	log.Printf("waiting for pods to be deleted\n")

	// pods MUST exit within 30s, because 30s after SIGTERM, they'll be SIGKILLed which will result in data loss
	// so we need to be tougher on how long this is allowed to take
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("failed to wait for pods to be deleted: %w", ctx.Err()))
		default:
			list, err := podInterface.List(ctx, metav1.ListOptions{LabelSelector: KeyPipelineName})
			if err != nil {
				panic(err)
			}
			if len(list.Items) == 0 {
				return
			}
			time.Sleep(time.Second)
		}
	}
}

func WaitForPod(podName string, f func(*corev1.Pod) bool) {
	log.Printf("waiting for pod %q %q\n", podName, getFuncName(f))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	w, err := podInterface.Watch(ctx, metav1.ListOptions{FieldSelector: "metadata.name=" + podName})
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for e := range w.ResultChan() {
		p, ok := e.Object.(*corev1.Pod)
		if !ok {
			panic(errors.FromObject(e.Object))
		}
		s := p.Status
		var y []string
		for _, c := range s.Conditions {
			if c.Status == corev1.ConditionTrue {
				y = append(y, string(c.Type))
			}
		}
		log.Printf("pod %q has status %s %q conditions %q\n", p.Name, s.Phase, s.Message, y)
		if f(p) {
			return
		}
	}
}
