package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten_getContainer(t *testing.T) {
	x := Flatten{
		AbstractStep{Resources:standardResources},
	}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"flatten"}, c.Args)
	assert.Equal(t, c.Resources, standardResources)
}
