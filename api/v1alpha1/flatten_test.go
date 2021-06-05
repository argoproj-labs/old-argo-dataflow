package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten_getContainer(t *testing.T) {
	x := Flatten{}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"flatten"}, c.Args)
}
