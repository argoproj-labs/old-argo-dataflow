package v1alpha1

func trunc(msg string) string {
	return truncN(msg, 64)
}

func truncN(msg string, n int) string {
	x := n / 2
	if len(msg) > n {
		return msg[0:x-1] + "..." + msg[len(msg)-x+2:]
	}
	return msg
}
