package v1alpha1

// +kubebuilder:validation:Enum="";None;NamespaceName;NamespacedPipelineName
type SubjectPrefix string

const (
	SubjectPrefixNone                   SubjectPrefix = "None"
	SubjectPrefixNamespaceName          SubjectPrefix = "NamespaceName"
	SubjectPrefixNamespacedPipelineName SubjectPrefix = "NamespacedPipelineName"
)

func SubjectPrefixOr(a, b SubjectPrefix) SubjectPrefix {
	if a != "" {
		return a
	}
	return b
}
