package v1alpha1

type STAN struct {
	// +kubebuilder:default=default
	Name          string        `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATSURL       string        `json:"natsUrl,omitempty" protobuf:"bytes,4,opt,name=natsUrl"`
	ClusterID     string        `json:"clusterId,omitempty" protobuf:"bytes,5,opt,name=clusterId"`
	Subject       string        `json:"subject" protobuf:"bytes,3,opt,name=subject"`
	SubjectPrefix SubjectPrefix `json:"subjectPrefix,omitempty" protobuf:"bytes,6,opt,name=subjectPrefix,casttype=SubjectPrefix"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Parallel uint32 `json:"parallel,omitempty" protobuf:"varint,7,opt,name=parallel"`
}
