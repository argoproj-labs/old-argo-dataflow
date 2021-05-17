package v1alpha1

type Cron struct {
	Schedule string `json:"schedule" protobuf:"bytes,1,opt,name=schedule"`
	// +kubebuilder:default="2006-01-02T15:04:05Z07:00"
	Layout string `json:"layout,omitempty" protobuf:"bytes,2,opt,name=layout"`
}
