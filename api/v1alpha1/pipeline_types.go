/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Replicas struct {
	Value *int32 `json:"value" protobuf:"varint,1,opt,name=value"`
}

func (in Replicas) GetValue() int {
	if in.Value != nil {
		return int(*in.Value)
	}
	return 1
}

type HTTP struct {
}

type Interface struct {
	FIFO bool  `json:"fifo,omitempty" protobuf:"varint,1,opt,name=fifo"`
	HTTP *HTTP `json:"http,omitempty" protobuf:"bytes,2,opt,name=http"`
}

type Container struct {
	corev1.Container `json:",inline" protobuf:"bytes,1,opt,name=container"`
	Volumes          []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,2,rep,name=volumes"`
	In               *Interface      `json:"in,omitempty" protobuf:"bytes,3,opt,name=in"`
	Out              *Interface      `json:"out,omitempty" protobuf:"bytes,4,opt,name=out"`
}

func (in *Container) GetContainer() corev1.Container {
	return in.Container
}

func (in *Container) GetOut() *Interface {
	return in.Out
}

func (in *Container) GetIn() *Interface {
	return in.In
}

type Handler struct {
	Runtime Runtime `json:"runtime" protobuf:"bytes,4,opt,name=runtime,casttype=Runtime"`
	URL     string  `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	Code    string  `json:"code,omitempty" protobuf:"bytes,3,opt,name=code"`
}

func (in *Handler) GetContainer() corev1.Container {
	return in.Runtime.GetContainer()
}

func (in *Handler) GetOut() *Interface {
	return &Interface{HTTP: &HTTP{}}
}

func (in *Handler) GetIn() *Interface {
	return &Interface{HTTP: &HTTP{}}
}

type FuncSpec struct {
	Name      string     `json:"name,omitempty" protobuf:"bytes,6,opt,name=name"`
	Container *Container `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Handler   *Handler   `json:"handler,omitempty" protobuf:"bytes,7,opt,name=handler"`
	Replicas  *Replicas  `json:"replicas,omitempty" protobuf:"bytes,2,opt,name=replicas"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sources []Source `json:"sources,omitempty" protobuf:"bytes,3,rep,name=sources"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sinks         []Sink               `json:"sinks,omitempty" protobuf:"bytes,4,rep,name=sinks"`
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,5,opt,name=restartPolicy,casttype=k8s.io/api/core/v1.RestartPolicy"`
}

func (in *FuncSpec) GetReplicas() Replicas {
	if in.Replicas != nil {
		return *in.Replicas
	}
	return Replicas{}
}

func (in *FuncSpec) GetRestartPolicy() corev1.RestartPolicy {
	if in.RestartPolicy != "" {
		return in.RestartPolicy
	}
	return corev1.RestartPolicyOnFailure
}

func (m *FuncSpec) GetOut() *Interface {
	if m == nil {
		return nil
	} else if m.Container != nil {
		return m.Container.GetOut()
	} else if m.Handler != nil {
		return m.Handler.GetOut()
	}
	return nil
}

func (m *FuncSpec) GetIn() *Interface {
	if m == nil {
		return nil
	} else if m.Container != nil {
		return m.Container.GetIn()
	} else if m.Handler != nil {
		return m.Handler.GetIn()
	}
	return nil
}

func (m *FuncSpec) GetVolumes() []corev1.Volume {
	if m != nil && m.Container != nil {
		return m.Container.Volumes
	}
	return nil
}

func (m *FuncSpec) GetContainer() corev1.Container {
	if c := m.Container; c != nil {
		return c.GetContainer()
	} else if h := m.Handler; h != nil {
		return h.GetContainer()
	} else {
		panic("invalid func spec")
	}
}

type Kafka struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	URL   string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	Topic string `json:"topic" protobuf:"bytes,3,opt,name=topic"`
}

type NATS struct {
	Name    string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	URL     string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	Subject string `json:"subject" protobuf:"bytes,3,opt,name=subject"`
}

type Source struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATS  *NATS  `json:"nats,omitempty" protobuf:"bytes,2,opt,name=nats"`
	Kafka *Kafka `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
}

func Json(in interface{}) string {
	data, _ := json.Marshal(in)
	return string(data)
}

type Sink struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	NATS  *NATS  `json:"nats,omitempty" protobuf:"bytes,2,opt,name=nats"`
	Kafka *Kafka `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
}

type PipelineSpec struct {
	// +patchStrategy=merge
	// +patchMergeKey=name
	Funcs []FuncSpec `json:"funcs,omitempty" protobuf:"bytes,1,rep,name=funcs"`
}

// +kubebuilder:validation:Enum="";Pending;Running;Succeeded;Failed
type PipelinePhase string

const (
	PipelineUnknown   PipelinePhase = ""
	PipelinePending   PipelinePhase = "Pending"
	PipelineRunning   PipelinePhase = "Running"
	PipelineSucceeded PipelinePhase = "Succeeded"
	PipelineFailed    PipelinePhase = "Failed"
)

func MinPipelinePhase(v ...PipelinePhase) PipelinePhase {
	for _, p := range []PipelinePhase{PipelineFailed, PipelinePending, PipelineRunning, PipelineSucceeded} {
		for _, x := range v {
			if x == p {
				return p
			}
		}
	}
	return PipelineUnknown
}

type PipelineStatus struct {
	Phase      PipelinePhase      `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PipelinePhase"`
	Message    string             `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Conditions []metav1.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=pl
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   PipelineSpec    `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status *PipelineStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Pipeline `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
