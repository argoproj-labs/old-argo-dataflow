package v1alpha1

import "fmt"

type JetStreamSource struct {
	JetStream `json:",inline" protobuf:"bytes,1,opt,name=jetstream"`
}

func (j JetStreamSource) GenURN(cluster, namespace string) string {
	return fmt.Sprintf("urn:dataflow:jetstream:%s:%s", j.NATSURL, j.Subject)
}
