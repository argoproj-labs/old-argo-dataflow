package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type GetPodSpecReq struct {
	Cluster          string                        `protobuf:"bytes,1,opt,name=cluster"`
	Debug            bool                          `protobuf:"varint,2,opt,name=debug"`
	PipelineName     string                        `protobuf:"bytes,3,opt,name=pipelineName"`
	Replica          int32                         `protobuf:"varint,4,opt,name=replica"`
	ImageFormat      string                        `protobuf:"bytes,5,opt,name=imageFormat"`
	RunnerImage      string                        `protobuf:"bytes,6,opt,name=runnerImage"`
	PullPolicy       corev1.PullPolicy             `protobuf:"bytes,7,opt,name=pullPolicy,casttype=k8s.io/api/core/v1.PullPolicy"`
	UpdateInterval   time.Duration                 `protobuf:"varint,8,opt,name=updateInterval,casttype=time.Duration"`
	StepStatus       StepStatus                    `protobuf:"bytes,9,opt,name=stepStatus"`
	Sidecar          Sidecar                       `protobuf:"bytes,10,opt,name=sidecar"`
	ImagePullSecrets []corev1.LocalObjectReference `protobuf:"bytes,11,rep,name=imagePullSecrets"`
	Hostname         string                        `protobuf:"bytes,12,opt,name=hostname"`
	Subdomain        string                        `protobuf:"bytes,13,opt,name=subdomain"`
}
