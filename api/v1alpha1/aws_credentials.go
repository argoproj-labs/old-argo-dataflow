package v1alpha1

import corev1 "k8s.io/api/core/v1"

type AWSCredentials struct {
	AccessKeyID     corev1.SecretKeySelector `json:"accessKeyId" protobuf:"bytes,1,opt,name=accessKeyId"`
	SecretAccessKey corev1.SecretKeySelector `json:"secretAccessKey" protobuf:"bytes,2,opt,name=secretAccessKey"`
	SessionToken    corev1.SecretKeySelector `json:"sessionToken" protobuf:"bytes,3,opt,name=sessionToken"`
}
