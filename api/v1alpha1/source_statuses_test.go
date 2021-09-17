package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceStatuses_SetPending(t *testing.T) {
	ss := SourceStatuses{}

	ss.SetPending("bar", 23)

	if assert.Len(t, ss, 1) {
		assert.Equal(t, uint64(23), *ss["bar"].Pending)
	}

	ss.SetPending("bar", 34)

	if assert.Len(t, ss, 1) {
		assert.Equal(t, uint64(23), *ss["bar"].LastPending)
		assert.Equal(t, uint64(34), *ss["bar"].Pending)
	}

	ss.SetPending("bar", 45)

	if assert.Len(t, ss, 1) {
		assert.Equal(t, uint64(34), *ss["bar"].LastPending)
		assert.Equal(t, uint64(45), *ss["bar"].Pending)
	}
}

func TestSourceStatus_GetPending(t *testing.T) {
	assert.Equal(t, uint64(0), SourceStatuses{}.GetPending())
	v := uint64(1)
	assert.Equal(t, uint64(1), SourceStatuses{"0": {Pending: &v}}.GetPending())
}

func TestSourceStatus_GetLastPending(t *testing.T) {
	assert.Equal(t, uint64(0), SourceStatuses{}.GetLastPending())
	v := uint64(1)
	assert.Equal(t, uint64(1), SourceStatuses{"0": {LastPending: &v}}.GetLastPending())
}
