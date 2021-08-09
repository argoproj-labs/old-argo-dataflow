package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type StepSpec struct {
	// +kubebuilder:default=default
	Name string `json:"name" protobuf:"bytes,6,opt,name=name"`

	Cat       *Cat       `json:"cat,omitempty" protobuf:"bytes,15,opt,name=cat"`
	Container *Container `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Dedupe    *Dedupe    `json:"dedupe,omitempty" protobuf:"bytes,27,opt,name=dedupe"`
	Expand    *Expand    `json:"expand,omitempty" protobuf:"bytes,26,opt,name=expand"`
	Filter    Filter     `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter,casttype=Filter"`
	Flatten   *Flatten   `json:"flatten,omitempty" protobuf:"bytes,25,opt,name=flatten"`
	Git       *Git       `json:"git,omitempty" protobuf:"bytes,12,opt,name=git"`
	Group     *Group     `json:"group,omitempty" protobuf:"bytes,11,opt,name=group"`
	Code      *Code      `json:"code,omitempty" protobuf:"bytes,7,opt,name=code"`
	Map       Map        `json:"map,omitempty" protobuf:"bytes,9,opt,name=map,casttype=Map"`

	// +kubebuilder:default=1
	Replicas uint32 `json:"replicas,omitempty" protobuf:"varint,23,opt,name=replicas"`
	Scale    *Scale `json:"scale,omitempty" protobuf:"bytes,24,opt,name=scale"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sources Sources `json:"sources,omitempty" protobuf:"bytes,3,rep,name=sources"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sinks []Sink `json:"sinks,omitempty" protobuf:"bytes,4,rep,name=sinks"`
	// +kubebuilder:default=OnFailure
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,5,opt,name=restartPolicy,casttype=k8s.io/api/core/v1.RestartPolicy"`
	Terminator    bool                 `json:"terminator,omitempty" protobuf:"varint,10,opt,name=terminator"` // if this step terminates, terminate all steps in the pipeline
	// +patchStrategy=merge
	// +patchMergeKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,13,rep,name=volumes"`
	// +kubebuilder:default=pipeline
	ServiceAccountName string              `json:"serviceAccountName,omitempty" protobuf:"bytes,14,opt,name=serviceAccountName"`
	Metadata           *Metadata           `json:"metadata,omitempty" protobuf:"bytes,16,opt,name=metadata"`
	NodeSelector       map[string]string   `json:"nodeSelector,omitempty" protobuf:"bytes,17,rep,name=nodeSelector"`
	Affinity           *corev1.Affinity    `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	Tolerations        []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,19,rep,name=tolerations"`
	// +kubebuilder:default={resources: {limits: {"cpu": "500m", "memory": "256Mi"}, requests: {"cpu": "100m", "memory": "64Mi"}}}
	Sidecar Sidecar `json:"sidecar,omitempty" protobuf:"bytes,28,opt,name=sidecar"`
}

func (in StepSpec) GetIn() *Interface {
	if in.Container != nil {
		return in.Container.GetIn()
	}
	return DefaultInterface
}

func (in StepSpec) getType() containerSupplier {
	if x := in.Cat; x != nil {
		return x
	} else if x := in.Container; x != nil {
		return x
	} else if x := in.Dedupe; x != nil {
		return x
	} else if x := in.Expand; x != nil {
		return x
	} else if x := in.Filter; x != "" {
		return x
	} else if x := in.Flatten; x != nil {
		return x
	} else if x := in.Git; x != nil {
		return x
	} else if x := in.Group; x != nil {
		return x
	} else if x := in.Code; x != nil {
		return x
	} else if x := in.Map; x != "" {
		return x
	} else {
		panic("invalid step spec")
	}
}

func (in StepSpec) CalculateReplicas(pending int) int {
	if in.Scale == nil {
		return -1
	}
	return in.Scale.Calculate(pending)
}
