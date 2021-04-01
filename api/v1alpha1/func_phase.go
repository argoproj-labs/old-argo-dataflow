package v1alpha1

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
