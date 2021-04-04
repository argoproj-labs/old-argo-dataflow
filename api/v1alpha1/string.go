package v1alpha1

func StringOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
