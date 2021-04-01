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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StepSpec struct {
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

func (in *StepSpec) GetReplicas() Replicas {
	if in.Replicas != nil {
		return *in.Replicas
	}
	return Replicas{}
}

func (in *StepSpec) GetRestartPolicy() corev1.RestartPolicy {
	if in.RestartPolicy != "" {
		return in.RestartPolicy
	}
	return corev1.RestartPolicyOnFailure
}

func (m *StepSpec) GetOut() *Interface {
	if m == nil {
		return nil
	} else if m.Container != nil {
		return m.Container.GetOut()
	} else if m.Handler != nil {
		return m.Handler.GetOut()
	}
	return nil
}

func (m *StepSpec) GetIn() *Interface {
	if m == nil {
		return nil
	} else if m.Container != nil {
		return m.Container.GetIn()
	} else if m.Handler != nil {
		return m.Handler.GetIn()
	}
	return nil
}

func (m *StepSpec) GetVolumes() []corev1.Volume {
	if m != nil && m.Container != nil {
		return m.Container.Volumes
	}
	return nil
}

func (m *StepSpec) GetContainer() corev1.Container {
	if c := m.Container; c != nil {
		return c.GetContainer()
	} else if h := m.Handler; h != nil {
		return h.GetContainer()
	} else {
		panic("invalid step spec")
	}
}

type StepStatus struct {
	Phase    StepPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=StepPhase"`
	Message  string    `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Replicas uint64    `json:"replicas,omitempty" protobuf:"varint,5,opt,name=replicas"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	SourceStatues SourceStatuses `json:"sourceStatuses,omitempty" protobuf:"bytes,3,rep,name=sourceStatuses"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	SinkStatues SinkStatuses `json:"sinkStatuses,omitempty" protobuf:"bytes,4,rep,name=sinkStatuses"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.status.replicas`
type Step struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   StepSpec    `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status *StepStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true

type StepList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Step `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&Step{}, &StepList{})
}
