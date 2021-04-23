package containerkiller

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var logger = zap.New()

type Interface interface {
	KillContainer(pod corev1.Pod, container string) error
}

type containerKiller struct {
	kubernetes.Interface
	*rest.Config
}

func New(k kubernetes.Interface, r *rest.Config) Interface {
	return &containerKiller{Interface: k, Config: r}
}

func (k *containerKiller) killContainer(namespace, pod, container string, commands []string) error {
	// running the wrong command appears to bork the container, so we need to make a call about which command to run
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

func (k *containerKiller) KillContainer(pod corev1.Pod, container string) error {
	for _, s := range pod.Status.ContainerStatuses {
		if s.Name == container && s.State.Running == nil {
			return nil
		}
	}
	commands := []string{"/bin/sh", "-c", "kill 1"}
	if text, ok := pod.Annotations[dfv1.KeyKillCmd(container)]; ok {
		util.MustUnJSON(text, &commands)
	}
	return k.killContainer(pod.Namespace, pod.Name, container, commands)
}
