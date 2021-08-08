package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type PipelineSpec struct {
	// +patchStrategy=merge
	// +patchMergeKey=name
	Steps []StepSpec `json:"steps,omitempty" protobuf:"bytes,1,rep,name=steps"`

	// ImagePullSecrets is a list of references to secrets in the same namespace to use for pulling any images
	// in pods that reference this ServiceAccount. ImagePullSecrets are distinct from Secrets because Secrets
	// can be mounted in the pod, but ImagePullSecrets are only accessed by the kubelet.
	// More info: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	// +patchStrategy=merge
	// +patchMergeKey=name
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,opt,name=imagePullSecrets"`
}

func (in *PipelineSpec) HasStep(name string) bool {
	for _, step := range in.Steps {
		if step.Name == name {
			return true
		}
	}
	return false
}
