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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func (in *Step) GetTargetReplicas(pending int) int {

	targetReplicas := in.Spec.GetReplicas().Calculate(pending)
	lastScaleTime := in.Status.GetLastScaleTime()
	currentReplicas := in.Status.GetReplicas() // can be -1

	if currentReplicas == 0 && targetReplicas == 0 && time.Since(lastScaleTime) > getPeekQuietDuration() {
		targetReplicas = 1
	}
	if time.Since(lastScaleTime) < getScalingQuietDuration() {
		targetReplicas = currentReplicas
	}

	return targetReplicas
}

func RequeueAfter(currentReplicas, targetReplicas int) time.Duration {
	if currentReplicas == 0 && targetReplicas == 0 {
		return getPeekQuietDuration()
	}
	return 0
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
