//go:build test
// +build test

package test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerName string

var podsInterface = kubernetesInterface.CoreV1().Pods(namespace)

func ExpectLogLine(step, pattern string, opts ...interface{}) {
	var (
		ctx           = context.Background()
		timeout       = time.Minute
		containerName = "sidecar"
	)
	for _, opt := range opts {
		switch v := opt.(type) {
		case context.Context:
			ctx = v
		case time.Duration:
			timeout = v
		case ContainerName:
			containerName = string(v)
		default:
			panic(fmt.Errorf("unknown option time %T", opt))
		}
	}
	log.Printf("expect step %q container %q to log pattern %q\n", step, containerName, pattern)
	labelSelector := fmt.Sprintf("dataflow.argoproj.io/step-name=%s", step)
	podList, err := podsInterface.List(ctx, metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: "status.phase=Running"})
	if err != nil {
		panic(fmt.Errorf("error getting step pods: %w", err))
	}
	if !podsLogContains(ctx, podList, containerName, pattern, timeout) {
		panic(fmt.Errorf("no log lines matched %q", pattern))
	}
}

func podsLogContains(ctx context.Context, podList *corev1.PodList, containerName, pattern string, timeout time.Duration) bool {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errChan := make(chan error)
	resultChan := make(chan bool)
	for _, p := range podList.Items {
		go func(podName string) {
			log.Printf("watching pod: %s\n", podName)
			contains, err := podLogContains(cctx, podName, containerName, pattern)
			if err != nil {
				errChan <- err
				return
			}
			if contains {
				resultChan <- true
			}
		}(p.Name)
	}

	for {
		select {
		case <-cctx.Done():
			return false
		case result := <-resultChan:
			if result {
				return true
			}
		case err := <-errChan:
			fmt.Printf("error: %v", err)
		}
	}
}

func podLogContains(ctx context.Context, podName, containerName, pattern string) (bool, error) {
	log.Printf("expect pod %q container %q to log pattern %q\n", podName, containerName, pattern)
	stream, err := podsInterface.GetLogs(podName, &corev1.PodLogOptions{Container: containerName, Follow: true}).Stream(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = stream.Close() }()

	exp, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	s := bufio.NewScanner(stream)
	for {
		select {
		case <-ctx.Done():
			return false, nil
		default:
			if !s.Scan() {
				return false, s.Err()
			}
			data := s.Bytes()
			if exp.Match(data) {
				return true, nil
			}
		}
	}
}

func TailLogs() {
	ctx := context.Background()
	list, err := podsInterface.List(ctx, metav1.ListOptions{LabelSelector: KeyPipelineName})
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}
	for _, item := range list.Items {
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()
			tailLogs(podName, "init")
			tailLogs(podName, "sidecar")
			tailLogs(podName, "main")
		}(item.Name)
	}
	wg.Wait()
}

func tailLogs(podName, containerName string) {
	ctx := context.Background()
	log.Printf("tailing logs for %q/%q\n", podName, containerName)
	stream, err := podsInterface.GetLogs(podName, &corev1.PodLogOptions{Container: containerName}).Stream(ctx)
	if err != nil {
		panic(err)
	}
	defer func() { _ = stream.Close() }()
	for s := bufio.NewScanner(stream); s.Scan(); {
		log.Println(s.Text())
	}
}
