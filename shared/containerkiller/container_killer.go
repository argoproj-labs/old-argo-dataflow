package containerkiller

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/podexec"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
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
	commands := []string{"/bin/sh", "-c", "kill 1"}
	if text, ok := pod.Annotations[dfv1.KeyKillCmd(container)]; ok {
		util.MustUnJSON(text, &commands)
	}
	return k.Exec(pod.Namespace, pod.Name, container, commands)
}
