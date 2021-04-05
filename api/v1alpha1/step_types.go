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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StepSpec struct {
	Name      string     `json:"name,omitempty" protobuf:"bytes,6,opt,name=name"`
	Container *Container `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Handler   *Handler   `json:"handler,omitempty" protobuf:"bytes,7,opt,name=handler"`
	Git       *Git       `json:"git,omitempty" protobuf:"bytes,12,opt,name=git"`
	Filter    Filter     `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter,casttype=Filter"`
	Map       Map        `json:"map,omitempty" protobuf:"bytes,9,opt,name=map,casttype=Map"`
	Group     *Group     `json:"group,omitempty" protobuf:"bytes,11,opt,name=group"`
	Replicas  *Replicas  `json:"replicas,omitempty" protobuf:"bytes,2,opt,name=replicas"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sources []Source `json:"sources,omitempty" protobuf:"bytes,3,rep,name=sources"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sinks         []Sink               `json:"sinks,omitempty" protobuf:"bytes,4,rep,name=sinks"`
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,5,opt,name=restartPolicy,casttype=k8s.io/api/core/v1.RestartPolicy"`
	Terminator    bool                 `json:"terminator,omitempty" protobuf:"varint,10,opt,name=terminator"` // if this step terminates, terminate all steps in the pipeline
}

func (in *StepSpec) GetReplicas() Replicas {
	if in.Replicas != nil {
		return *in.Replicas
	}
	return Replicas{Min: 1}
}

func (in *StepSpec) GetRestartPolicy() corev1.RestartPolicy {
	if in.RestartPolicy != "" {
		return in.RestartPolicy
	}
	return corev1.RestartPolicyOnFailure
}

func (in *StepSpec) GetIn() *Interface {
	if in.Container != nil {
		return in.Container.In
	}
	return DefaultInterface
}

func (in *StepSpec) GetVolumes() []corev1.Volume {
	if in != nil && in.Container != nil {
		return in.Container.Volumes
	}
	return nil
}

func (in *StepSpec) GetContainer(runnerImage string, policy corev1.PullPolicy, mnt corev1.VolumeMount) corev1.Container {
	if c := in.Container; c != nil {
		return c.GetContainer(policy, mnt)
	} else if h := in.Handler; h != nil {
		return h.GetContainer(policy, mnt)
	} else if g := in.Git; g != nil {
		return g.GetContainer(policy)
	} else if m := in.Map; m != "" {
		return m.GetContainer(runnerImage, policy)
	} else if f := in.Filter; f != "" {
		return f.GetContainer(runnerImage, policy)
	} else {
		panic("invalid step spec")
	}
}

type StepStatus struct {
	Phase         StepPhase      `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=StepPhase"`
	Message       string         `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Replicas      uint32         `json:"replicas" protobuf:"varint,5,opt,name=replicas"`
	LastScaleTime *metav1.Time   `json:"lastScaleTime,omitempty" protobuf:"bytes,6,opt,name=lastScaleTime"`
	SourceStatues SourceStatuses `json:"sourceStatuses,omitempty" protobuf:"bytes,3,rep,name=sourceStatuses"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	SinkStatues SinkStatuses `json:"sinkStatuses,omitempty" protobuf:"bytes,4,rep,name=sinkStatuses"`
}

func (m *StepStatus) GetSourceStatues() SourceStatuses {
	if m == nil {
		return nil
	}
	return m.SourceStatues
}

func (m *StepStatus) GetReplicas() int {
	if m == nil {
		return -1
	}
	return int(m.Replicas)
}

func (m *StepStatus) GetLastScaleTime() time.Time {
	if m == nil || m.LastScaleTime == nil {
		return time.Time{}
	}
	return m.LastScaleTime.Time
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

const (
	scalingQuietDuration = time.Minute // how long to wait since the last scaling event to scale
	peekQuietDuration    = time.Minute // how long to wait since the last scaling to peek
)

func (in *Step) GetTargetReplicas(pending int) int {

	targetReplicas := in.Spec.Replicas.Calculate(pending)
	lastScaleTime := in.Status.GetLastScaleTime()
	currentReplicas := in.Status.GetReplicas()

	if currentReplicas == 0 && targetReplicas == 0 && time.Since(lastScaleTime) > peekQuietDuration {
		targetReplicas = 1
	}
	if time.Since(lastScaleTime) < scalingQuietDuration {
		targetReplicas = currentReplicas
	}

	return targetReplicas
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
