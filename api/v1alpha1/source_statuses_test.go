package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSourceStatuses_Set(t *testing.T) {
	ss := SourceStatuses{}

	ss.IncrTotal("bar", 1, resource.MustParse("1"), 1)

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(1), s.Metrics["1"].Total)
			assert.Equal(t, resource.MustParse("1"), s.Metrics["1"].Rate)
			assert.Equal(t, uint64(1), s.Metrics["1"].TotalBytes)
		}
	}

	ss.IncrTotal("bar", 1, resource.MustParse("1"), 1)

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(2), s.Metrics["1"].Total)
			assert.Equal(t, resource.MustParse("1"), s.Metrics["1"].Rate)
			assert.Equal(t, uint64(2), s.Metrics["1"].TotalBytes)
		}
	}

	ss.IncrTotal("bar", 0, resource.MustParse("1"), 1)

	if assert.Len(t, ss, 1) {
		s := ss["bar"]
		if assert.Len(t, s.Metrics, 2) {
			assert.Equal(t, uint64(1), s.Metrics["0"].Total)
			assert.Equal(t, uint64(2), s.Metrics["1"].Total)
			assert.Equal(t, uint64(2), s.Metrics["1"].TotalBytes)
		}
	}

	ss.IncrTotal("baz", 0, resource.MustParse("1"), 1)

	if assert.Len(t, ss, 2) {
		s := ss["baz"]
		if assert.Len(t, s.Metrics, 1) {
			assert.Equal(t, uint64(1), s.Metrics["0"].Total)
		}
	}
}

func TestSourceStatuses_IncErrors(t *testing.T) {
	ss := SourceStatuses{}
	ss.IncrErrors("foo", 0)
	assert.Equal(t, uint64(1), ss["foo"].Metrics["0"].Errors)
	ss.IncrErrors("foo", 0)
	assert.Equal(t, uint64(2), ss["foo"].Metrics["0"].Errors)
	ss.IncrErrors("bar", 0)
	assert.Equal(t, uint64(1), ss["bar"].Metrics["0"].Errors)
}

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

func TestSourceStatuses_IncrRetries(t *testing.T) {
	sources := SourceStatuses{}
	sources.IncrRetries("one", 1)
	assert.Equal(t, uint64(1), sources.Get("one").GetRetries())
	sources.IncrRetries("one", 2)
	assert.Equal(t, uint64(2), sources.Get("one").GetRetries())
	sources.IncrRetries("one", 1)
	assert.Equal(t, uint64(3), sources.Get("one").GetRetries())
	sources.IncrRetries("two", 1)
	assert.Equal(t, uint64(3), sources.Get("one").GetRetries())
	assert.Equal(t, uint64(1), sources.Get("two").GetRetries())
}
