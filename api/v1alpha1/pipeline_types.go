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

type HTTP struct {
}

type Interface struct {
	FIFO bool  `json:"fifo,omitempty" protobuf:"varint,1,opt,name=fifo"`
	HTTP *HTTP `json:"http,omitempty" protobuf:"bytes,2,opt,name=http"`
}

type Node struct {
	corev1.Container `json:",inline" protobuf:"bytes,1,opt,name=container"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Volumes  []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,2,rep,name=volumes"`
	Replicas *Replicas       `json:"replicas,omitempty" protobuf:"bytes,3,opt,name=replicas"`
	In       *Interface      `json:"in,omitempty" protobuf:"bytes,4,opt,name=in"`
	Out      *Interface      `json:"out,omitempty" protobuf:"bytes,5,opt,name=out"`
	Sources  []Source        `json:"sources,omitempty" protobuf:"bytes,6,rep,name=sources"`
	Sinks    []Sink          `json:"sinks,omitempty" protobuf:"bytes,7,rep,name=sinks"`
}

func (in *Node) GetReplicas() Replicas {
	if in.Replicas != nil {
		return *in.Replicas
	}
	return Replicas{}
}

type Kafka struct {
	URL   string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Topic string `json:"topic" protobuf:"bytes,2,opt,name=topic"`
}

type Bus struct {
	Subject string `json:"subject" protobuf:"bytes,1,opt,name=subject"`
}

type Source struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Bus   *Bus   `json:"bus,omitempty" protobuf:"bytes,2,opt,name=bus"`
	Kafka *Kafka `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
}

func Json(in interface{}) string {
	data, _ := json.Marshal(in)
	return string(data)
}

type Sink struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Bus   *Bus   `json:"bus,omitempty" protobuf:"bytes,2,opt,name=bus"`
	Kafka *Kafka `json:"kafka,omitempty" protobuf:"bytes,3,opt,name=kafka"`
}

type PipelineSpec struct {
	// +patchStrategy=merge
	// +patchMergeKey=name
	Nodes []Node `json:"nodes,omitempty" protobuf:"bytes,1,rep,name=nodes"`
}

// +kubebuilder:validation:Enum=Pending;Running;Error
type PipelinePhase string

const (
	PipelineUnknown PipelinePhase = ""
	PipelinePending PipelinePhase = "Pending"
	PipelineRunning PipelinePhase = "Running"
	PipelineError   PipelinePhase = "Error"
)

func MinPhase(v ...PipelinePhase) PipelinePhase {
	for _, p := range []PipelinePhase{PipelineError, PipelinePending, PipelineRunning} {
		for _, x := range v {
			if x == p {
				return p
			}
		}
	}
	return PipelineUnknown
}

// +kubebuilder:validation:Enum=Pending;Running;Error
type NodePhase string

const (
	NodePending NodePhase = "Pending"
	NodeRunning NodePhase = "Running"
)

type NodeStatus struct {
	Name    string    `json:"name" protobuf:"bytes,1,opt,name=name"`
	Phase   NodePhase `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase,casttype=NodePhase"`
	Message string    `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
}

type PipelineStatus struct {
	Phase        PipelinePhase      `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PipelinePhase"`
	Message      string             `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Conditions   []metav1.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	NodeStatuses []NodeStatus       `json:"nodeStatuses,omitempty" protobuf:"bytes,4,rep,name=nodeStatuses"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=pl
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PipelinePhase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   PipelineSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status PipelineStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
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
