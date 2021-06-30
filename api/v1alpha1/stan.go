package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type STANAuthStrategy string

var (
	STANAuthNone  STANAuthStrategy = "None"
	STANAuthToken STANAuthStrategy = "Token"
)

type STAN struct {
	// +kubebuilder:default=default
	Name              string        `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATSURL           string        `json:"natsUrl,omitempty" protobuf:"bytes,4,opt,name=natsUrl"`
	NATSMonitoringURL string        `json:"natsMonitoringUrl,omitempty" protobuf:"bytes,8,opt,name=natsMonitoringUrl"`
	ClusterID         string        `json:"clusterId,omitempty" protobuf:"bytes,5,opt,name=clusterId"`
	Subject           string        `json:"subject" protobuf:"bytes,3,opt,name=subject"`
	SubjectPrefix     SubjectPrefix `json:"subjectPrefix,omitempty" protobuf:"bytes,6,opt,name=subjectPrefix,casttype=SubjectPrefix"`
	Auth              *STANAuth     `json:"auth,omitempty" protobuf:"bytes,7,opt,name=auth"`
}

type STANAuth struct {
	Token *corev1.SecretKeySelector `json:"token,omitempty" protobuf:"bytes,1,opt,name=token"`
}

func (s *STAN) AuthStrategy() STANAuthStrategy {
	if s.Auth != nil {
		if s.Auth.Token != nil {
			return STANAuthToken
		}
	}
	return STANAuthNone
}
