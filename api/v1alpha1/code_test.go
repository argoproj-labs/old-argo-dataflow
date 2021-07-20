package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCode_getContainer(t *testing.T) {
	x := Code{
		Runtime: "my-runtime",
	}
	c := x.getContainer(getContainerReq{imageFormat: "fmt-%s"})
	assert.Equal(t, "fmt-dataflow-my-runtime", c.Image)
}
