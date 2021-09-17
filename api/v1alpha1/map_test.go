package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap_getContainer(t *testing.T) {
	x := &Map{
		Expression:   "my-expr",
		AbstractStep: AbstractStep{Resources: standardResources},
	}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"map", "my-expr"}, c.Args)
	assert.Equal(t, c.Resources, standardResources)
}
