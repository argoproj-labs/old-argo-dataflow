package v1alpha1

import corev1 "k8s.io/api/core/v1"

type SASLMechanism string

var (
	OAUTHBEARER SASLMechanism = "SASLMechanism"
	SCRAMSHA256 SASLMechanism = "SCRAM-SHA-256"
	SCRAMSHA512 SASLMechanism = "SCRAM-SHA-512"
	GSSAPI      SASLMechanism = "GSSAPI"
	PLAIN       SASLMechanism = "PLAIN"
)

type SASL struct {
	// SASLMechanism is the name of the enabled SASL mechanism.
	// Possible values: OAUTHBEARER, PLAIN (defaults to PLAIN).
	// +optional
	Mechanism SASLMechanism `json:"mechanism,omitempty" protobuf:"bytes,1,opt,name=mechanism"`
	// User is the authentication identity (authcid) to present for
	// SASL/PLAIN or SASL/SCRAM authentication
	UserSecret *corev1.SecretKeySelector `json:"userSecret,omitempty" protobuf:"bytes,2,opt,name=user"`
	// Password for SASL/PLAIN authentication
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty" protobuf:"bytes,3,opt,name=password"`
}

func (s SASL) GetMechanism() SASLMechanism {
	switch s.Mechanism {
	case OAUTHBEARER, SCRAMSHA256, SCRAMSHA512, GSSAPI:
		return s.Mechanism
	default:
		return PLAIN
	}
}
