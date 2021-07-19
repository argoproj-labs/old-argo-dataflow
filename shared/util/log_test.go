package util

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	NewLogger().Info("test", "a", 1, "b", "c")
}
