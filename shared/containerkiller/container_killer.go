package containerkiller

import (
	"github.com/argoproj-labs/argo-dataflow/shared/podexec"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	KillContainer(pod corev1.Pod, container string) error
}

type containerKiller struct {
	podexec.Interface
}

func New(k kubernetes.Interface, r *rest.Config) Interface {
	return &containerKiller{podexec.New(k, r)}
}

func (k *containerKiller) KillContainer(pod corev1.Pod, container string) error {
	for _, s := range pod.Status.ContainerStatuses {
		if s.Name == container && s.State.Running == nil {
			return nil
		}
	}
	return k.Exec(pod.Namespace, pod.Name, container, []string{})
}
