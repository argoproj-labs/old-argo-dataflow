package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectPrefixOr(t *testing.T) {
	assert.Equal(t, SubjectPrefixNone, SubjectPrefixOr("", SubjectPrefixNone))
}
