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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Replicas struct {
	Value *int32 `json:"value"`
}

type HTTP struct {
}

type Interface struct {
	HTTP HTTP `json:"http"`
}

type Node struct {
	Name     string    `json:"name"`
	Image    string    `json:"image"`
	Replicas *Replicas `json:"replicas,omitempty"`
	From     Interface `json:"from"`
	To       Interface `json:"to"`
	Source   Source    `json:"source"`
	Sink     Sink      `json:"sink"`
}

func (in *Node) GetReplicas() Replicas {
	if in.Replicas != nil {
		return *in.Replicas
	}
	return Replicas{}
}

type Kafka struct {
	URL   string `json:"url"`
	Topic string `json:"topic"`
}

type Bus struct {
	Subject string `json:"subject"`
}

type Source struct {
	Bus   *Bus   `json:"bus,omitempty"`
	Kafka *Kafka `json:"kafka,omitempty"`
}

func NewSource(x string) Source {
	v := Source{}
	_ = json.Unmarshal([]byte(x), &v)
	return v
}

func (in *Source) Json() string { return Json(in) }

func Json(in interface{}) string {
	data, _ := json.Marshal(in)
	return string(data)
}

func NewSink(x string) Sink {
	v := Sink{}
	_ = json.Unmarshal([]byte(x), &v)
	return v
}

type Sink struct {
	Bus   *Bus   `json:"bus,omitempty"`
	Kafka *Kafka `json:"kafka,omitempty"`
}

func (in *Sink) Json() string { return Json(in) }

type PipelineSpec struct {
	// +patchStrategy=merge
	// +patchMergeKey=name
	Nodes []Node `json:"nodes,omitempty"`
}

// +kubebuilder:validation:Enum=Pending;Running;Error
type Phase string

const (
	PipelinePending Phase = "Pending"
	PipelineRunning Phase = "Running"
	PipelineError   Phase = "Error"
)

type PipelineStatus struct {
	Phase   Phase  `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=pl
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
