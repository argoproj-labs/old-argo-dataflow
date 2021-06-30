package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	t.Run("NotRecent", func(t *testing.T) {
		x := SourceStatus{
			LastError: &Error{Time: metav1.Time{}},
			Metrics:   map[string]Metrics{"": {Errors: 1}},
		}
		assert.Equal(t, uint64(1), x.GetErrors())
		assert.False(t, x.RecentErrors())
	})
	t.Run("Recent", func(t *testing.T) {
		x := SourceStatus{
			LastError: &Error{Time: metav1.Now()},
			Metrics:   map[string]Metrics{"": {Errors: 1}},
		}
		assert.Equal(t, uint64(1), x.GetErrors())
		assert.True(t, x.RecentErrors())
	})
}

func TestSourceStatus_GetRetryCount(t *testing.T){
	t.Run("None", func(t *testing.T) {
		x := SourceStatus{}
		assert.Equal(t, uint64(0), x.GetRetryCount())
	})
	t.Run("One", func(t *testing.T) {
		x := SourceStatus{
			Metrics:   map[string]Metrics{"one": {Retries: 1}},
		}
		assert.Equal(t, uint64(1), x.GetRetryCount())
	})
	t.Run("Two", func(t *testing.T) {
		x := SourceStatus{
			Metrics:   map[string]Metrics{"one": {Retries: 1}, "two": {Retries: 1}},
		}
		assert.Equal(t, uint64(2), x.GetRetryCount())
	})
}