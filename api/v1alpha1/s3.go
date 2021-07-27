package v1alpha1

type S3 struct {
	// +kubebuilder:default=default
	Name        string          `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Bucket      string          `json:"bucket" protobuf:"bytes,2,opt,name=bucket"`
	Region      string          `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`
	Credentials *AWSCredentials `json:"credentials,omitempty" protobuf:"bytes,4,opt,name=credentials"`
	Endpoint    *AWSEndpoint    `json:"endpoint,omitempty" protobuf:"bytes,5,opt,name=endpoint"`
}
