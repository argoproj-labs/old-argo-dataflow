package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCat_getContainer(t *testing.T) {
	x := Cat{}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"cat"}, c.Args)
}
