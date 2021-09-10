package v1alpha1

import "context"

type VolumeSink struct {
	VolumeSource AbstractVolumeSource `json:",inline" protobuf:"bytes,1,opt,name=volumeSource"`
}

func (in VolumeSink) GetURN(ctx context.Context) string {
	return in.VolumeSource.GetURN(ctx)
}
