package util

var (
	message string
	logger  = NewLogger()
)

func init() {
	logger.Info("version", "message", message)
}
