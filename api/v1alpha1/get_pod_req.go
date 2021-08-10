package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type GetPodSpecReq struct {
	ClusterName    string            `protobuf:"bytes,9,opt,name=clusterName"`
	PipelineName   string            `protobuf:"bytes,1,opt,name=pipelineName"`
	Namespace      string            `protobuf:"bytes,2,opt,name=namespace"`
	Replica        int32             `protobuf:"varint,3,opt,name=replica"`
	ImageFormat    string            `protobuf:"bytes,4,opt,name=imageFormat"`
	RunnerImage    string            `protobuf:"bytes,5,opt,name=runnerImage"`
	PullPolicy     corev1.PullPolicy `protobuf:"bytes,6,opt,name=pullPolicy,casttype=k8s.io/api/core/v1.PullPolicy"`
	UpdateInterval time.Duration     `protobuf:"varint,7,opt,name=updateInterval,casttype=time.Duration"`
	StepStatus     StepStatus        `protobuf:"bytes,8,opt,name=stepStatus"`
	Sidecar        Sidecar           `protobuf:"bytes,10,opt,name=sidecar"`
}
