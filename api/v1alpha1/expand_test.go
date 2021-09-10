package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpand_getContainer(t *testing.T) {
	x := Expand{
		AbstractStep{Resources: standardResources},
	}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"expand"}, c.Args)
	assert.Equal(t, c.Resources, standardResources)
}
