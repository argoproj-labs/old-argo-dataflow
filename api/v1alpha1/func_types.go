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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum="";Pending;Running;Succeeded;Failed
type FuncPhase string

func (p FuncPhase) Completed() bool {
	return p == FuncSucceeded || p == FuncFailed
}

const (
	FuncUnknown   FuncPhase = ""
	FuncPending   FuncPhase = "Pending"
	FuncRunning   FuncPhase = "Running"
	FuncSucceeded FuncPhase = "Succeeded"
	FuncFailed    FuncPhase = "Failed"
)

func MinFuncPhase(v ...FuncPhase) FuncPhase {
	for _, p := range []FuncPhase{FuncFailed, FuncPending, FuncRunning, FuncSucceeded} {
		for _, x := range v {
			if x == p {
				return p
			}
		}
	}
	return FuncUnknown
}

type FuncStatus struct {
	Phase FuncPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=fn
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
type Func struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FuncSpec    `json:"spec"`
	Status *FuncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type FuncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Func `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Func{}, &FuncList{})
}
