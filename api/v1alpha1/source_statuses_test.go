package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceStatuses_Set(t *testing.T) {

	ss := SourceStatuses{}

	ss.Set("bar", 1, "foo")

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.NotNil(t, s.LastMessage) {
			assert.Equal(t, "foo", s.LastMessage.Data)
		}
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(1), s.Metrics["1"].Total)
		}
	}

	ss.Set("bar", 1, "bar")

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.NotNil(t, s.LastMessage) {
			assert.Equal(t, "bar", s.LastMessage.Data)
		}
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(2), s.Metrics["1"].Total)
		}
	}

	ss.Set("bar", 0, "foo")

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.NotNil(t, s.LastMessage) {
			assert.Equal(t, "foo", s.LastMessage.Data)
		}
		if assert.Len(t, s.Metrics, 2) {
			assert.Equal(t, uint64(1), s.Metrics["0"].Total)
			assert.Equal(t, uint64(2), s.Metrics["1"].Total)
		}
	}

	ss.Set("baz", 0, "foo")

	if assert.Len(t, ss, 2) {
		s := ss["baz"]
		if assert.NotNil(t, s.LastMessage) {
			assert.Equal(t, "foo", s.LastMessage.Data)
		}
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(1), s.Metrics["0"].Total)
		}
	}
}

func TestSourceStatuses_IncErrors(t *testing.T) {
	ss := SourceStatuses{}
	ss.IncErrors("foo", 0)
	assert.Equal(t, uint64(1), ss["foo"].Metrics["0"].Errors)
	ss.IncErrors("foo", 0)
	assert.Equal(t, uint64(2), ss["foo"].Metrics["0"].Errors)
	ss.IncErrors("bar", 0)
	assert.Equal(t, uint64(1), ss["bar"].Metrics["0"].Errors)
}

func TestSourceStatuses_SetPending(t *testing.T) {

	ss := SourceStatuses{}

	ss.SetPending("bar", 23)

	if assert.Len(t, ss, 1) {
		assert.Equal(t, uint64(23), ss["bar"].Pending)
	}

	ss.SetPending("bar", 34)

	if assert.Len(t, ss, 1) {
		assert.Equal(t, uint64(34), ss["bar"].Pending)
	}

	ss.SetPending("bar", 45)

	if assert.Len(t, ss, 1) {
		assert.Equal(t, uint64(45), ss["bar"].Pending)
	}
}

func TestSourceStatus_GetPending(t *testing.T) {
	assert.Equal(t, 0, SourceStatuses{}.GetPending())
	assert.Equal(t, 1, SourceStatuses{"0": {Pending: 1}}.GetPending())
}

func TestSourceStatus_AnyErrors(t *testing.T) {
	assert.False(t, SourceStatuses{}.AnyErrors())
	assert.True(t, SourceStatuses{"0": {Metrics: map[string]Metrics{"0": {Errors: 1}}}}.AnyErrors())
}
