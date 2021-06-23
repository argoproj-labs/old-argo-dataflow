// +build test

package test

import (
	"bufio"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"log"
	"regexp"
)

func ExpectLogLine(podName, containerName, pattern string) {
	log.Printf("expect pod %q container %q to log pattern %q\n", podName, containerName, pattern)
	ctx := context.Background()
	stream, err := kubernetesInterface.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: containerName}).Stream(ctx)
	if err != nil {
		panic(err)
	}
	defer func() { _ = stream.Close() }()
	for s := bufio.NewScanner(stream); s.Scan(); {
		match, err := regexp.Match(pattern, s.Bytes())
		if err != nil {
			panic(err)
		} else if match {
			return
		}
	}
	panic(fmt.Errorf("no log lines matched %q", pattern))
}

func TailLogs(podName, containerName string) {
	log.Printf("dumping logs for %q/%q\n", podName, containerName)
	ctx := context.Background()
	stream, err := kubernetesInterface.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: containerName}).Stream(ctx)
	if err != nil {
		panic(err)
	}
	defer func() { _ = stream.Close() }()
	for s := bufio.NewScanner(stream); s.Scan(); {
		log.Println(s.Text())
	}
}
