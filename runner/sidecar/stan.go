package sidecar

import (
	"context"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func subjectiveStan(x *dfv1.STAN) {
	switch x.SubjectPrefix {
	case dfv1.SubjectPrefixNamespaceName:
		x.Subject = fmt.Sprintf("%s.%s", namespace, x.Subject)
	case dfv1.SubjectPrefixNamespacedPipelineName:
		x.Subject = fmt.Sprintf("%s.%s.%s", namespace, pipelineName, x.Subject)
	}
}

func stanFromSecret(s *dfv1.STAN, secret *corev1.Secret) {
	s.NATSURL = dfv1.StringOr(s.NATSURL, string(secret.Data["natsUrl"]))
	s.NATSMonitoringURL = dfv1.StringOr(s.NATSMonitoringURL, string(secret.Data["natsMonitoringUrl"]))
	s.ClusterID = dfv1.StringOr(s.ClusterID, string(secret.Data["clusterId"]))
	s.SubjectPrefix = dfv1.SubjectPrefixOr(s.SubjectPrefix, dfv1.SubjectPrefix(secret.Data["subjectPrefix"]))
	if _, ok := secret.Data["authToken"]; ok {
		s.Auth = &dfv1.STANAuth{
			Token: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
				Key: "authToken",
			},
		}
	}
}

func enrichSTAN(ctx context.Context, secrets v1.SecretInterface, x *dfv1.STAN) error {
	secret, err := secrets.Get(ctx, "dataflow-stan-"+x.Name, metav1.GetOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			return err
		}
	} else {
		stanFromSecret(x, secret)
	}
	subjectiveStan(x)
	return nil
}

func getSTANAuthToken(ctx context.Context, x *dfv1.STAN) (string, error) {
	if x.AuthStrategy() != dfv1.STANAuthToken {
		return "", fmt.Errorf("auth strategy is not token but %s", x.AuthStrategy())
	}
	if x.Auth == nil || x.Auth.Token == nil {
		return "", fmt.Errorf("token secret selector is nil")
	}
	secrets := kubernetesInterface.CoreV1().Secrets(namespace)
	secret, err := secrets.Get(ctx, x.Auth.Token.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if token, ok := secret.Data[x.Auth.Token.Key]; ok {
		return string(token), nil
	}
	return "", fmt.Errorf("key %s not found in secret %s", x.Auth.Token.Key, x.Auth.Token.Name)
}
