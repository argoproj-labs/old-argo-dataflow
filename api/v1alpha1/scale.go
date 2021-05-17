package v1alpha1

type Scale struct {
	MinReplicas  int32   `json:"minReplicas,omitempty" protobuf:"varint,1,opt,name=minReplicas"`
	MaxReplicas  *uint32 `json:"maxReplicas,omitempty" protobuf:"varint,2,opt,name=maxReplicas"` // takes precedence over min
	ReplicaRatio uint32  `json:"replicaRatio,omitempty" protobuf:"varint,3,opt,name=replicaRatio"`
}

// Used to calculate the number of replicas.
// min(r.max, max(r.min, pending/ratio))
// Example:
// min=1, max=4, ratio=100
// pending=0, replicas=1
// pending=100, replicas=1
// pending=200, replicas=2
// pending=300, replicas=3
// pending=400, replicas=4
// pending=500, replicas=4
func (in Scale) Calculate(pending int) int {
	n := 0
	if in.ReplicaRatio > 0 {
		n = pending / int(in.ReplicaRatio)
	}
	if n < int(in.MinReplicas) {
		n = int(in.MinReplicas)
	}
	if in.MaxReplicas != nil && n > int(*in.MaxReplicas) {
		n = int(*in.MaxReplicas)
	}
	return n
}
