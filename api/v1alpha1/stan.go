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
	// Max inflight messages when subscribing to the stan server, which means how many messages
	// between commits, therefore potential duplicates during disruption
	// +kubebuilder:default=20
	MaxInflight uint32 `json:"maxInflight,omitempty" protobuf:"bytes,9,opt,name=maxInflight"`
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

func (s *STAN) GetMaxInflight() int {
	if s.MaxInflight < 1 {
		return CommitN
	}
	return int(s.MaxInflight)
}
