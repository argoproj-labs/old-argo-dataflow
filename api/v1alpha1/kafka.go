package v1alpha1

import (
	"fmt"
)

type KafkaNET struct {
	TLS  *TLS  `json:"tls,omitempty" protobuf:"bytes,1,opt,name=tls"`
	SASL *SASL `json:"sasl,omitempty" protobuf:"bytes,2,opt,name=sasl"`
}

func (in *KafkaNET) GetSecurityProtocol() string {
	if in.SASL == nil && in.TLS == nil {
		return "plaintext"
	} else if in.SASL == nil && in.TLS != nil {
		return "ssl"
	} else if in.TLS == nil {
		return "sasl_plaintext"
	}
	return "sasl_ssl"
}

type KafkaConfig struct {
	Brokers         []string  `json:"brokers,omitempty" protobuf:"bytes,1,rep,name=brokers"`
	NET             *KafkaNET `json:"net,omitempty" protobuf:"bytes,3,opt,name=net"`
	MaxMessageBytes int32     `json:"maxMessageBytes,omitempty" protobuf:"varint,4,opt,name=maxMessageBytes"`
}

type Kafka struct {
	// +kubebuilder:default=default
	Name        string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	KafkaConfig `json:",inline" protobuf:"bytes,4,opt,name=kafkaConfig"`
	Topic       string `json:"topic" protobuf:"bytes,3,opt,name=topic"`
}

func (in Kafka) GenURN(cluster, namespace string) string {
	return fmt.Sprintf("urn:dataflow:kafka:%s:%s", in.Brokers[0], in.Topic)
}
