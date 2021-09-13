package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCron_GenURN(t *testing.T) {
	urn := Cron{Schedule: "* * * * *"}.GenURN(ctx)
	assert.Equal(t, "urn:dataflow:cron:* * * * *", urn)
}
