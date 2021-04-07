package v1alpha1

// +kubebuilder:validation:Enum="";JSONBytesArray;JSONStringArray
type GroupFormat string

const (
	GroupFormatUnknown         GroupFormat = ""                // all messages are sent one by one - probably not what you want
	GroupFormatJSONBytesArray  GroupFormat = "JSONBytesArray"  // messages are sent as an array where each element is a base 64 encoded
	GroupFormatJSONStringArray GroupFormat = "JSONStringArray" // messages are sent as an array where each element is a string
)
