package v1alpha1

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type AbstractVolumeSource corev1.VolumeSource

func (in AbstractVolumeSource) getURIParts(ctx context.Context) (_type, kind, name string) {
	if v := in.ConfigMap; v != nil {
		return "configmaps", "ConfigMap", v.Name
	} else if v := in.EmptyDir; v != nil {
		return "emptydir", "Pod", GetMetaPod(ctx)
	} else if v := in.PersistentVolumeClaim; v != nil {
		return "persistentvolumeclaim:%s", "PersistentVolumeClaim", v.ClaimName
	} else if v := in.Secret; v != nil {
		return "secret", "Secret", v.SecretName
	}
	panic(fmt.Errorf("un-suppported volume source %v", in))
}

func (in AbstractVolumeSource) GetURN(ctx context.Context) string {
	_type, kind, name := in.getURIParts(ctx)
	return fmt.Sprintf("urn:dataflow:volume:%s:%s", _type, dnsName(ctx, kind, name))
}
