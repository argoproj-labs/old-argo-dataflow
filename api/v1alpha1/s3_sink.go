package v1alpha1

type S3Sink struct {
	S3 `json:",inline" protobuf:"bytes,4,opt,name=s3"`
}
