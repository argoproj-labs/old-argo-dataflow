package v1alpha1

func StringOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func StringsOr(a, b []string) []string {
	if len(a) > 0 {
		return a
	}
	return b
}
