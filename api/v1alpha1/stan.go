package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type NATSAuthStrategy string

var (
	NATSAuthNone  NATSAuthStrategy = "None"
	NATSAuthToken NATSAuthStrategy = "Token"
)

type STAN struct {
	// +kubebuilder:default=default
	Name              string        `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATSURL           string        `json:"natsUrl,omitempty" protobuf:"bytes,4,opt,name=natsUrl"`
	NATSMonitoringURL string        `json:"natsMonitoringUrl,omitempty" protobuf:"bytes,8,opt,name=natsMonitoringUrl"`
	ClusterID         string        `json:"clusterId,omitempty" protobuf:"bytes,5,opt,name=clusterId"`
	Subject           string        `json:"subject" protobuf:"bytes,3,opt,name=subject"`
	SubjectPrefix     SubjectPrefix `json:"subjectPrefix,omitempty" protobuf:"bytes,6,opt,name=subjectPrefix,casttype=SubjectPrefix"`
	Auth              *NATSAuth     `json:"auth,omitempty" protobuf:"bytes,7,opt,name=auth"`
	// Max inflight messages when subscribing to the stan server, which means how many messages
	// between commits, therefore potential duplicates during disruption
	// +kubebuilder:default=20
	MaxInflight uint32 `json:"maxInflight,omitempty" protobuf:"bytes,9,opt,name=maxInflight"`
}

func (s STAN) GenURN(cluster, namespace string) string {
	return fmt.Sprintf("urn:dataflow:stan:%s:%s", s.NATSURL, s.Subject)
}

type NATSAuth struct {
	Token *corev1.SecretKeySelector `json:"token,omitempty" protobuf:"bytes,1,opt,name=token"`
}

func (s *STAN) AuthStrategy() NATSAuthStrategy {
	if s.Auth != nil {
		if s.Auth.Token != nil {
			return NATSAuthToken
		}
	}
	return NATSAuthNone
}

func (s *STAN) GetMaxInflight() int {
	if s.MaxInflight < 1 {
		return CommitN
	}
	return int(s.MaxInflight)
}
