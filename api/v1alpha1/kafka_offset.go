package v1alpha1

// +kubebuilder:validation:Enum=First;Last
type KafkaOffset string

func (k KafkaOffset) GetAutoOffsetReset() string {
	switch k {
	case "First":
		return "earliest"
	default:
		return "latest"
	}
}
