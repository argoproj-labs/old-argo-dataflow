//go:build test
// +build test

package test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/argoproj-labs/argo-dataflow/shared/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podsInterface = kubernetesInterface.CoreV1().Pods(namespace)

func ExpectLogLine(step string, opts ...interface{}) {
	var (
		ctx           = context.Background()
		timeout       = time.Minute
		containerName = "sidecar"
		matcher       func([]byte) bool
	)
	for _, opt := range opts {
		switch v := opt.(type) {
		case context.Context:
			ctx = v
		case time.Duration:
			timeout = v
		case string:
			matcher = LogMatches(v)
		case func([]byte) bool:
			matcher = v
		default:
			panic(fmt.Errorf("unknown option type %T", opt))
		}
	}
	log.Printf("expect step %q container %q to match %q\n", step, containerName, util.GetFuncName(matcher))
	labelSelector := fmt.Sprintf("dataflow.argoproj.io/step-name=%s", step)
	podList, err := podsInterface.List(ctx, metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: "status.phase=Running"})
	if err != nil {
		panic(fmt.Errorf("error getting step pods: %w", err))
	}
	if !podsLogMatches(ctx, podList, containerName, matcher, timeout) {
		panic(fmt.Errorf("no log lines matched %q", util.GetFuncName(matcher)))
	}
}

func podsLogMatches(ctx context.Context, podList *corev1.PodList, containerName string, match func([]byte) bool, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errChan := make(chan error)
	resultChan := make(chan bool)
	for _, p := range podList.Items {
		for _, s := range p.Status.ContainerStatuses {
			if s.Name == "sidecar" && s.State.Running != nil {
				go func(podName string) {
					contains, err := podLogContains(ctx, podName, containerName, match)
					if err != nil {
						errChan <- err
						return
					}
					if contains {
						resultChan <- true
					}
				}(p.Name)
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
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

func LogMatches(pattern string) func([]byte) bool {
	exp, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return func(bytes []byte) bool {
		return exp.Match(bytes)
	}
}

func podLogContains(ctx context.Context, podName, containerName string, match func([]byte) bool) (bool, error) {
	log.Printf("expect pod %q container %q log to match %q\n", podName, containerName, util.GetFuncName(match))
	stream, err := podsInterface.GetLogs(podName, &corev1.PodLogOptions{Container: containerName, Follow: true, Timestamps: true}).Stream(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = stream.Close() }()

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
			if match(data) {
				return true, nil
			}
		}
	}
}
