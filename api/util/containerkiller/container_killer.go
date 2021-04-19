package containerkiller

import (
	"fmt"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
)

var logger = zap.New()

type Interface interface {
	KillContainer(namespace, pod, container string, image string) error
}

type containerKiller struct {
	kubernetes.Interface
	*rest.Config
}

func New(k kubernetes.Interface, r *rest.Config) Interface {
	return &containerKiller{Interface: k, Config: r}
}

func (k *containerKiller) KillContainer(namespace, pod, container string, image string) error {
	// running the wrong command appears to bork the container, so we need to make a call about which command to run
	commands := []string{"/bin/sh", "-c", "kill 1"}
	if strings.Contains(image, "dataflow-runner") {
		commands = []string{"/bin/kill", "1"}
	}
	x := k.Interface.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		Param("container", container).
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false")
	for _, command := range commands {
		x = x.Param("command", command)
	}
	logger.Info("killing container", "pod", pod, "container", container, "commands", commands)
	exec, err := remotecommand.NewSPDYExecutor(k.Config, "POST", x.URL())
	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}
	if err := exec.Stream(remotecommand.StreamOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	}); dfv1.IgnoreNotFound(dfv1.IgnoreContainerNotFound(err)) != nil {
		return fmt.Errorf("failed to stream: %w", err)
	}
	return nil
}
