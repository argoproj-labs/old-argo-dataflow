package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSTAN_GenURN(t *testing.T) {
	urn := STAN{NATSURL: "my-url", Subject: "my-subject"}.GenURN(cluster, namespace)
	assert.Equal(t, "urn:dataflow:stan:my-url:my-subject", urn)
}
