package v1alpha1

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type AbstractVolumeSource corev1.VolumeSource

func (in AbstractVolumeSource) getURNParts() (kind string, name string) {
	if v := in.ConfigMap; v != nil {
		return "configmap", v.Name
	} else if v := in.PersistentVolumeClaim; v != nil {
		return "persistentvolumeclaim", v.ClaimName
	} else if v := in.Secret; v != nil {
		return "secret", v.SecretName
	}
	panic(fmt.Errorf("un-suppported volume source %v", in))
}

func (in AbstractVolumeSource) GenURN(cluster, namespace string) string {
	kind, name := in.getURNParts()
	return fmt.Sprintf("urn:dataflow:volume:%s:%s.%s.%s.%s", strings.ToLower(kind), name, kind, namespace, cluster)
}
