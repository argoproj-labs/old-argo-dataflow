package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKafkaOffset_GetAutoOffsetReset(t *testing.T) {
	t.Run("First", func(t *testing.T) {
		assert.Equal(t, "earliest", KafkaOffset("First").GetAutoOffsetReset())
	})
	t.Run("Last", func(t *testing.T) {
		assert.Equal(t, "latest", KafkaOffset("Last").GetAutoOffsetReset())
	})
}
