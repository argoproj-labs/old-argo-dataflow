package sidecar

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_newBackoff(t *testing.T) {
	type want struct {
		steps int
		step  time.Duration
	}
	d := metav1.Duration{Duration: 100 * time.Millisecond}
	for _, test := range []struct {
		name    string
		backoff Backoff
		want    []want
	}{
		{"Empty", Backoff{Duration: d}, []want{{0, 100 * time.Millisecond}}},
		{"Default", Backoff{Duration: d, Steps: 2, FactorPercentage: 200}, []want{
			{2, 100 * time.Millisecond},
			{1, 200 * time.Millisecond},
			{0, 400 * time.Millisecond},
		}},
		{"Cap", Backoff{Duration: d, Steps: 3, FactorPercentage: 200, Cap: metav1.Duration{Duration: 220 * time.Millisecond}}, []want{
			{3, 100 * time.Millisecond},
			{2, 200 * time.Millisecond},
			{0, 220 * time.Millisecond}, // cap is not what I originally thought, i.e. not a total cap
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
