package v1alpha1

import corev1 "k8s.io/api/core/v1"

type TLS struct {
	// CACertSecret refers to the secret that contains the CA cert
	CACertSecret *corev1.SecretKeySelector `json:"caCertSecret,omitempty" protobuf:"bytes,1,opt,name=caCertSecret"`
	// CertSecret refers to the secret that contains the cert
	CertSecret *corev1.SecretKeySelector `json:"clientCertSecret,omitempty" protobuf:"bytes,2,opt,name=certSecret"`
	// KeySecret refers to the secret that contains the key
	KeySecret *corev1.SecretKeySelector `json:"clientKeySecret,omitempty" protobuf:"bytes,3,opt,name=keySecret"`
}
