package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_containerBuilder(t *testing.T) {
	c := containerBuilder{}.
		init(getContainerReq{}).
		build()
	assert.Equal(t, "main", c.Name)
}
