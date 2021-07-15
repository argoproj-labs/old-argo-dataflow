package podexec

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Interface interface {
	Exec(namespace, podName, container string, commands []string) error
}

type podExec struct {
	kubernetes.Interface
	*rest.Config
}

func New(k kubernetes.Interface, r *rest.Config) Interface {
	return &podExec{Interface: k, Config: r}
}

func (k *podExec) Exec(namespace, pod, container string, commands []string) error {
	// running the wrong command appears to bork the container, so we need to make a call about which command to run
	x := k.Interface.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		Param("container", container).
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "true")
	for _, command := range commands {
		x = x.Param("command", command)
	}
	exec, err := remotecommand.NewSPDYExecutor(k.Config, "POST", x.URL())
	if err != nil {
		return fmt.Errorf("failed to exec %q: %w", commands, err)
	}
	if err := exec.Stream(remotecommand.StreamOptions{Stdout: os.Stdout, Stderr: os.Stderr, Tty: true}); err != nil {
		return fmt.Errorf("failed to stream %q: %w", commands, err)
	}
	return nil
}
