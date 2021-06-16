package sidecar

import (
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
}
