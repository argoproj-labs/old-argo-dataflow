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

type ReplicaSetSpec struct {
	Replicas *int32                 `json:"replicas,omitempty"`
	Template corev1.PodTemplateSpec `json:"template"`
}

func (in *ReplicaSetSpec) GetReplicas() int {
	if in.Replicas != nil {
		return int(*in.Replicas)
	}
	return 1
}

// +kubebuilder:validation:Enum="";Pending;Running;Succeeded;Failed
type ReplicaSetPhase string

const (
	ReplicaSetUnknown   ReplicaSetPhase = ""
	ReplicaSetPending   ReplicaSetPhase = "Pending"
	ReplicaSetRunning   ReplicaSetPhase = "Running"
	ReplicaSetSucceeded ReplicaSetPhase = "Succeeded"
	ReplicaSetFailed    ReplicaSetPhase = "Failed"
)

func MinReplicaSetPhase(v ...ReplicaSetPhase) ReplicaSetPhase {
	for _, p := range []ReplicaSetPhase{ReplicaSetFailed, ReplicaSetPending, ReplicaSetRunning, ReplicaSetSucceeded} {
		for _, x := range v {
			if x == p {
				return p
			}
		}
	}
	return ReplicaSetUnknown
}

type ReplicaSetStatus struct {
	Phase ReplicaSetPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=rs
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
type ReplicaSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReplicaSetSpec    `json:"spec"`
	Status *ReplicaSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ReplicaSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicaSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReplicaSet{}, &ReplicaSetList{})
}
