package v1alpha1

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
type Replicas struct {
	Min   int32  `json:"min" protobuf:"varint,1,opt,name=min"`           // this is both the min, and the initial value
	Max   *int32 `json:"max,omitempty" protobuf:"varint,2,opt,name=max"` // takes precedence over min
	Ratio uint32 `json:"ratio,omitempty" protobuf:"bytes,3,opt,name=ratio"`
}

func (in Replicas) Calculate(pending int) int {
	n := 0
	if in.Ratio > 0 {
		n = pending / int(in.Ratio)
	}
	if int(in.Min) > n {
		n = int(in.Min)
	}
	if in.Max != nil && n > int(*in.Max) {
		n = int(*in.Max)
	}
	return n
}
