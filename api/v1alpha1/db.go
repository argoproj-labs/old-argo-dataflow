package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type Database struct {
	// +kubebuilder:default=default
	Driver     string        `json:"driver,omitempty" protobuf:"bytes,1,opt,name=driver"`
	DataSource *DBDataSource `json:"dataSource,omitempty" protobuf:"bytes,2,opt,name=dataSource"`
}

func (in Database) GenURN(cluster, namespace string) string {
	if in.DataSource.Value != "" {
		return fmt.Sprintf("urn:dataflow:db:%s", in.DataSource.Value)
	} else {
		return fmt.Sprintf("urn:dataflow:db:%s.secret.%s.%s:%s", in.DataSource.ValueFrom.SecretKeyRef.Name, namespace, cluster, in.DataSource.ValueFrom.SecretKeyRef.Key)
	}
}

type DBDataSource struct {
	Value     string            `json:"value,omitempty" protobuf:"bytes,1,opt,name=value"`
	ValueFrom *DBDataSourceFrom `json:"valueFrom,omitempty" protobuf:"bytes,2,opt,name=valueFrom"`
}

type DBDataSourceFrom struct {
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,1,opt,name=secretKeyRef"`
}

type SQLStatement struct {
	SQL  string   `json:"sql,omitempty" protobuf:"bytes,1,opt,name=sql"`
	Args []string `json:"args,omitempty" protobuf:"bytes,2,rep,name=args"`
}

type SQLAction struct {
	SQLStatement     `json:",inline" protobuf:"bytes,1,opt,name=statement"`
	OnRecordNotFound *SQLStatement `json:"onRecordNotFound,omitempty" protobuf:"bytes,2,opt,name=onRecordNotFound"`
	OnError          *SQLStatement `json:"onError,omitempty" protobuf:"bytes,3,opt,name=onError"`
}
