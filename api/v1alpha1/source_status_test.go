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
	t.Run("None", func(t *testing.T) {
		x := SourceStatus{}
		assert.Equal(t, uint64(0), x.GetErrors())
	})
	t.Run("Some", func(t *testing.T) {
		x := SourceStatus{
			Metrics: map[string]Metrics{"": {Errors: 1}},
		}
		assert.Equal(t, uint64(1), x.GetErrors())
	})
}

func TestSourceStatus_GetRetries(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		x := SourceStatus{}
		assert.Equal(t, uint64(0), x.GetRetries())
	})
	t.Run("One", func(t *testing.T) {
		x := SourceStatus{
			Metrics: map[string]Metrics{"one": {Retries: 1}},
		}
		assert.Equal(t, uint64(1), x.GetRetries())
	})
	t.Run("Two", func(t *testing.T) {
		x := SourceStatus{
			Metrics: map[string]Metrics{"one": {Retries: 1}, "two": {Retries: 1}},
		}
		assert.Equal(t, uint64(2), x.GetRetries())
	})
}
