package bus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_imageName(t *testing.T) {
	name, err := imageName("nats:foo")
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/nats:foo", name)
}
