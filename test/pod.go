// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podInterface = kubernetesInterface.CoreV1().Pods(namespace)

func PodRunningAndReady(p *corev1.Pod) bool {
	return p.GetDeletionTimestamp() == nil && p.Status.Phase == corev1.PodRunning && PodReady(p)
}

func PodReady(p *corev1.Pod) bool {
	for _, c := range p.Status.Conditions {
		if p.GetDeletionTimestamp() == nil && c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
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

func WaitForPod(opts ...interface{}) {
	// by default, wait for any pod to be ready
	var (
		listOptions = metav1.ListOptions{LabelSelector: KeyPipelineName}
		f           = PodRunningAndReady
		timeout     = 30 * time.Second
	)
	for _, o := range opts {
		switch v := o.(type) {
		case string:
			listOptions.LabelSelector = "" // we don't always use this for pipeline pods
			listOptions.FieldSelector = "metadata.name=" + v
		case func(*corev1.Pod) bool:
			f = v
		case time.Duration:
			timeout = v
		default:
			panic(fmt.Errorf("un-supported option type: %T", o))
		}
	}
	log.Printf("waiting for pod %q %q in %v\n", sharedutil.MustJSON(listOptions), sharedutil.GetFuncName(f), timeout.String())
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	w, err := podInterface.Watch(ctx, listOptions)
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("failed to wait for pod: %w", ctx.Err()))
		case e := <-w.ResultChan():
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
			log.Printf("pod %q is %s %q %q\n", p.Name, s.Phase, s.Message, y)
			if f(p) {
				return
			}
		}
	}
}

func DeletePod(podName string) {
	log.Printf("deleting pod %q\n", podName)
	if err := podInterface.Delete(context.Background(), podName, metav1.DeleteOptions{}); err != nil {
		panic(fmt.Errorf("failed to delete %q: %w", podName, err))
	}
}
