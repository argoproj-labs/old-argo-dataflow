package v1alpha1

type S3Sink struct {
	S3 `json:",inline" protobuf:"bytes,4,opt,name=s3"`
	// an expression over the message, e.g. `"subpath/"+string(msg)`
	Key string `json:"key" protobuf:"bytes,3,opt,name=bucket"`
}
