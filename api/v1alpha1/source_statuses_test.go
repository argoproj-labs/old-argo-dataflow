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

func TestSourceStatuses_SetPending(t *testing.T) {

	ss := SourceStatuses{}

	ss.SetPending("bar", 1, 23)

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(23), s.Metrics["0"].Pending)
		}
	}

	ss.SetPending("bar", 1, 34)

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(34), s.Metrics["0"].Pending)
		}
	}

	ss.SetPending("bar", 0, 45)

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.Len(t, s.Metrics, 2) {
			assert.Equal(t, uint64(45), s.Metrics["1"].Pending)
		}
	}
}
