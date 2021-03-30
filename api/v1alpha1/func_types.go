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

type Message struct {
	Data string      `json:"data" protobuf:"bytes,1,opt,name=data"`
	Time metav1.Time `json:"time" protobuf:"bytes,2,opt,name=time"`
}

type Metrics struct {
	Replica uint32 `json:"replica" protobuf:"varint,4,opt,name=replica"`
	Total   uint64 `json:"total,omitempty" protobuf:"varint,1,opt,name=total"`
	Pending uint64 `json:"pending,omitempty" protobuf:"varint,3,opt,name=pending"`
}

type SourceStatus struct {
	Name        string   `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	LastMessage *Message `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	// +patchStrategy=merge
	// +patchMergeKey=replica
	Metrics []Metrics `json:"metrics,omitempty" protobuf:"bytes,3,rep,name=metrics"`
}

type SinkStatus struct {
	Name        string   `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	LastMessage *Message `json:"lastMessage,omitempty" protobuf:"bytes,2,opt,name=lastMessage"`
	// +patchStrategy=merge
	// +patchMergeKey=replica
	Metrics []Metrics `json:"metrics,omitempty" protobuf:"bytes,3,rep,name=metrics"`
}

type SourceStatuses []SourceStatus

func (s *SourceStatuses) Set(name string, replica int, short string) {
	m := &Message{Data: short, Time: metav1.Now()}
	for i, x := range *s {
		if x.Name == name {
			x.LastMessage = m
			for j := len(x.Metrics); j <= replica; j++ {
				x.Metrics = append(x.Metrics, Metrics{Replica: uint32(j)})
			}
			x.Metrics[i].Total++
			(*s)[i] = x
			return
		}
	}
	*s = append(*s, SourceStatus{Name: name, LastMessage: m})
}

func (s *SourceStatuses) SetPending(name string, replica int, pending int64) {
	for i, x := range *s {
		if x.Name == name {
			for j := len(x.Metrics); j <= replica; j++ {
				x.Metrics = append(x.Metrics, Metrics{Replica: uint32(j)})
			}
			x.Metrics[i].Pending = uint64(pending)
			(*s)[i] = x
			return
		}
	}
	metrics := make([]Metrics, replica+1)
	metrics[replica].Replica = uint32(replica)
	metrics[replica].Pending = uint64(pending)
	*s = append(*s, SourceStatus{Metrics: metrics})
}

type SinkStatuses []SinkStatus

func (s *SinkStatuses) Set(name string, replica int, short string) {
	m := &Message{Data: short, Time: metav1.Now()}
	for i, x := range *s {
		if x.Name == name {
			x.LastMessage = m
			for i := len(x.Metrics); i <= replica; i++ {
				x.Metrics = append(x.Metrics, Metrics{Replica: uint32(i)})
			}
			x.Metrics[i].Total++
			(*s)[i] = x
			return
		}
	}
	*s = append(*s, SinkStatus{Name: name, LastMessage: m})
}

type FuncStatus struct {
	Phase    FuncPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=FuncPhase"`
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
// +kubebuilder:resource:shortName=fn
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.status.replicas`
type Func struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   FuncSpec    `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status *FuncStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true

type FuncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Func `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&Func{}, &FuncList{})
}
