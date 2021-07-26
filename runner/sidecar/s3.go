package sidecar

import (
	"context"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func s3SourceFromSecret(x *dfv1.S3Source, secret *corev1.Secret) error {
	if x.Region == "" {
		x.Region = string(secret.Data["region"])
	}
	if x.Credentials == nil {
		x.Credentials = &dfv1.AWSCredentials{
			AccessKeyID: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: string(secret.Data["credentials.accessKeyId.name"]),
				},
				Key: string(secret.Data["credentials.accessKeyId.key"]),
			},
			SecretAccessKey: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: string(secret.Data["credentials.secretAccessKey.name"]),
				},
				Key: string(secret.Data["credentials.secretAccessKey.key"]),
			},
		}
	}
	if v, ok := secret.Data["endpoint.url"]; ok && x.Endpoint == nil {
		x.Endpoint = &dfv1.AWSEndpoint{URL: string(v)}
	}
	return nil
}

func enrichS3Source(ctx context.Context, secrets v1.SecretInterface, x *dfv1.S3Source) error {
	secret, err := secrets.Get(ctx, "dataflow-s3-"+x.Name, metav1.GetOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			return err
		}
	} else if err := s3SourceFromSecret(x, secret); err != nil {
		return err
	}
	return nil
}
