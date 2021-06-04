package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceStatus_GetTotal(t *testing.T) {
	assert.Equal(t, uint64(0), (SourceStatus{}).GetTotal())
	assert.Equal(t, uint64(1), (SourceStatus{
		Metrics: map[string]Metrics{"": {Total: 1}},
	}).GetTotal())
}

func TestSourceStatus_GetErrors(t *testing.T) {
	assert.Equal(t, uint64(0), (SourceStatus{}).GetErrors())
	assert.Equal(t, uint64(1), (SourceStatus{
		Metrics: map[string]Metrics{"": {Errors: 1}},
	}).GetErrors())
}
