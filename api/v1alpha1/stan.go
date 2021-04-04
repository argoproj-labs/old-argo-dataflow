package v1alpha1

type STAN struct {
	Name      string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATSURL   string `json:"natsUrl,omitempty" protobuf:"bytes,4,opt,name=natsUrl"`
	ClusterID string `json:"clusterId,omitempty" protobuf:"bytes,5,opt,name=clusterId"`
	Subject   string `json:"subject" protobuf:"bytes,3,opt,name=subject"`
}
