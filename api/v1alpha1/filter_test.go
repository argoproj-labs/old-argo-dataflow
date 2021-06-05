package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_getContainer(t *testing.T) {
	const x Filter = "my-filter"
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"filter", "my-filter"}, c.Args)
}
