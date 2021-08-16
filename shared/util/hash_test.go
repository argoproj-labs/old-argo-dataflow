package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustHash(t *testing.T) {
	assert.Equal(t, "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae", MustHash([]byte("foo")))
}

func TestGetUniquePipelineID(t *testing.T) {
	uniqueID := GetUniquePipelineID("cluster", "default", "pipeline", "stepName", "source")
	assert.Equal(t, "dataflow-clu-def-pip-ste-sou-7c07c91b03ebf978f5dda8b77130662e016493600b8ca4e6ffe12ec5183e3d25", uniqueID)
}
