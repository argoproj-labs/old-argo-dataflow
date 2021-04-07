package v1alpha1

import corev1 "k8s.io/api/core/v1"

type StepSpec struct {
	// +kubebuilder:default=default
	Name      string     `json:"name" protobuf:"bytes,6,opt,name=name"`
	Container *Container `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Handler   *Handler   `json:"handler,omitempty" protobuf:"bytes,7,opt,name=handler"`
	Git       *Git       `json:"git,omitempty" protobuf:"bytes,12,opt,name=git"`
	Filter    Filter     `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter,casttype=Filter"`
	Map       Map        `json:"map,omitempty" protobuf:"bytes,9,opt,name=map,casttype=Map"`
	Group     *Group     `json:"group,omitempty" protobuf:"bytes,11,opt,name=group"`
	Replicas  *Replicas  `json:"replicas,omitempty" protobuf:"bytes,2,opt,name=replicas"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sources []Source `json:"sources,omitempty" protobuf:"bytes,3,rep,name=sources"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sinks []Sink `json:"sinks,omitempty" protobuf:"bytes,4,rep,name=sinks"`
	// +kubebuilder:default=OnFailure
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,5,opt,name=restartPolicy,casttype=k8s.io/api/core/v1.RestartPolicy"`
	Terminator    bool                 `json:"terminator,omitempty" protobuf:"varint,10,opt,name=terminator"` // if this step terminates, terminate all steps in the pipeline
	// +patchStrategy=merge
	// +patchMergeKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,13,rep,name=volumes"`
	// +kubebuilder:default=pipeline-runner
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,14,opt,name=serviceAccountName"`
}

func (in *StepSpec) GetReplicas() Replicas {
	if in.Replicas != nil {
		return *in.Replicas
	}
	return Replicas{Min: 1}
}

func (in *StepSpec) GetIn() *Interface {
	if in.Container != nil {
		return in.Container.In
	}
	return DefaultInterface
}

func (in *StepSpec) GetContainer(runnerImage string, policy corev1.PullPolicy, mnt corev1.VolumeMount) corev1.Container {
	return in.getType().getContainer(getContainerReq{
		runnerImage:     runnerImage,
		imagePullPolicy: policy,
		volumeMount:     mnt,
	})
}

func (in *StepSpec) getType() containerSupplier {
	if x := in.Container; x != nil {
		return x
	} else if x := in.Filter; x != "" {
		return x
	} else if x := in.Git; x != nil {
		return x
	} else if x := in.Group; x != nil {
		return x
	} else if x := in.Handler; x != nil {
		return x
	} else if x := in.Map; x != "" {
		return x
	} else {
		panic("invalid step spec")
	}
}
