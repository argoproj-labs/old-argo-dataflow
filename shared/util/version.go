package util

var (
	version string
	logger  = NewLogger()
)

func init() {
	logger.Info("version", "version", version)
}

func Version() string { return version }
