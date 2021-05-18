package v1alpha1

func trunc(msg string) string {
	if len(msg) > 32 {
		return msg[0:15] + "..." + msg[len(msg)-14:]
	}
	return msg
}
