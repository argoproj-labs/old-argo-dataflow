package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSourceUID(t *testing.T) {
	uniqueID := GetSourceUID("cluster", "default", "pipeline", "stepName", "source")
	assert.Equal(t, "dataflow-clu-def-pip-ste-sou-7c07c91b03ebf978f5dda8b77130662e016493600b8ca4e6ffe12ec5183e3d25", uniqueID)
}
