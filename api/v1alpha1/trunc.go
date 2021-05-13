package v1alpha1

func trunc(msg string) string {
	if len(msg) > 32 {
		return msg[0:32]
	}
	return msg
}
