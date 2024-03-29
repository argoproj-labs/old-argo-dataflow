package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_getContainer(t *testing.T) {
	x := &Filter{
		Expression: "my-filter",
		AbstractStep: AbstractStep{
			Resources: standardResources,
		},
	}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"filter", "my-filter"}, c.Args)
	assert.Equal(t, c.Resources, standardResources)
}
