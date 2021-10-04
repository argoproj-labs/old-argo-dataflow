package v1alpha1

import "fmt"

type JetStreamSource struct {
	JetStream         `json:",inline" protobuf:"bytes,1,opt,name=jetstream"`
	NATSMonitoringURL string `json:"natsMonitoringUrl,omitempty" protobuf:"bytes,2,opt,name=natsMonitoringUrl"`
	DurableName       string `json:"durableName,omitempty" protobuf:"bytes,3,opt,name=durableName"`
	// Max inflight messages when subscribing to the stan server, which means how many messages
	// between commits, therefore potential duplicates during disruption
	// +kubebuilder:default=20
	MaxInflight uint32 `json:"maxInflight,omitempty" protobuf:"bytes,4,opt,name=maxInflight"`
}

func (j JetStreamSource) GenURN(cluster, namespace string) string {
	return fmt.Sprintf("urn:dataflow:jetstream:%s:%s", j.NATSURL, j.Subject)
}

func (j *JetStreamSource) GetMaxInflight() int {
	if j.MaxInflight < 1 {
		return CommitN
	}
	return int(j.MaxInflight)
}
