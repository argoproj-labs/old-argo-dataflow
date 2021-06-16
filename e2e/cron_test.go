// +build e2e

package e2e

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/dsls/golang"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	"testing"
)

func TestCron(t *testing.T) {

	Setup(t)
	defer Teardown(t)

	pl := Pipeline("cron").
		Step(
			Cron("*/3 * * * *").
				Cat("main").
				Log(),
		).
		Run()

	assert.NotEmpty(t, pl.ResourceVersion)

	WatchPipeline(pl.Namespace, pl.Name, func(pl dfv1.Pipeline) bool {
		return meta.FindStatusCondition(pl.Status.Conditions, dfv1.ConditionSunkMessages) != nil
	})

}
