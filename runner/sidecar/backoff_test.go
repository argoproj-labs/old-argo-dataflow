package sidecar

import (
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_newBackoff(t *testing.T) {
	type want struct {
		steps int
		step  time.Duration
	}
	for _, test := range []struct {
		name    string
		backoff Backoff
		want    []want
	}{
		{"Empty", Backoff{}, []want{{0, 100 * time.Millisecond}}},
		{"Default", Backoff{Steps: 2}, []want{
			{2, 100 * time.Millisecond},
			{1, 120 * time.Millisecond},
			{0, 144 * time.Millisecond},
		}},
		{"Cap", Backoff{Steps: 3, Cap: metav1.Duration{Duration: 120 * time.Millisecond}}, []want{
			{3, 100 * time.Millisecond},
			{2, 120 * time.Millisecond},
			{0, 120 * time.Millisecond}, //cap is not what I originally thought, i.e. not a total cap
		}},
	} {
		t.Run(test.name, func(t *testing.T) {

			b := newBackoff(test.backoff)
			for _, w := range test.want {
				assert.Equal(t, w.steps, b.Steps)
				assert.Equal(t, w.step, b.Step())
			}
		})
	}
}
