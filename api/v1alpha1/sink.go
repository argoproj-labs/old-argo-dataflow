package v1alpha1

import (
	"context"
	"fmt"
)

type Sink struct {
	// +kubebuilder:default=default
	Name   string      `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	STAN   *STAN       `json:"stan,omitempty" protobuf:"bytes,2,opt,name=stan"`
	Kafka  *KafkaSink  `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
	Log    *Log        `json:"log,omitempty" protobuf:"bytes,4,opt,name=log"`
	HTTP   *HTTPSink   `json:"http,omitempty" protobuf:"bytes,5,opt,name=http"`
	S3     *S3Sink     `json:"s3,omitempty" protobuf:"bytes,6,opt,name=s3"`
	DB     *DBSink     `json:"db,omitempty" protobuf:"bytes,7,opt,name=db"`
	Volume *VolumeSink `json:"volume,omitempty" protobuf:"bytes,8,opt,name=volume"`
}

func (s Sink) get() urner {
	if v := s.DB; v != nil {
		return v
	} else if v := s.HTTP; v != nil {
		return v
	} else if v := s.Kafka; v != nil {
		return v
	} else if v := s.Log; v != nil {
		return v
	} else if v := s.S3; v != nil {
		return v
	} else if v := s.STAN; v != nil {
		return v
	} else if v := s.Volume; v != nil {
		return v
	}
	panic(fmt.Errorf("invalid sink %q", s.Name))
}

func (s Sink) GetURN(ctx context.Context) string {
	return s.get().GetURN(ctx)
}
