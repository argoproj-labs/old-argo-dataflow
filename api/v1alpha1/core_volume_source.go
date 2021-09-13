package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type AbstractVolumeSource corev1.VolumeSource

func (in AbstractVolumeSource) getURNParts() (kind string, name string) {
	if v := in.ConfigMap; v != nil {
		return "ConfigMap", v.Name
	} else if v := in.PersistentVolumeClaim; v != nil {
		return "PersistentVolumeClaim", v.ClaimName
	} else if v := in.Secret; v != nil {
		return "Secret", v.SecretName
	}
	panic(fmt.Errorf("un-suppported volume source %v", in))
}

func (in AbstractVolumeSource) GenURN(ctx context.Context) string {
	kind, name := in.getURNParts()
	return fmt.Sprintf("urn:dataflow:volume:%s:%s", strings.ToLower(kind), dnsName(ctx, kind, name))
}
