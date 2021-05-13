package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_imageName(t *testing.T) {
	name, err := imageName("nats:foo")
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/nats:foo", name)
}
