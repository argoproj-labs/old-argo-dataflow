package v1alpha1

type Cron struct {
	Schedule string `json:"schedule" protobuf:"bytes,1,opt,name=schedule"`
}
