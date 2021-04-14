package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinPipelinePhase(t *testing.T) {
	assert.Equal(t, PipelineFailed, MinPipelinePhase(PipelineFailed, PipelineRunning))
}
