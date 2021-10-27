package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKafkaSource_GetAutoOffsetReset(t *testing.T) {
	source := KafkaSource{}
	assert.Equal(t, "latest", source.GetAutoOffsetReset())
}

func TestKafkaSource_GetFetchMinBytes(t *testing.T) {
	v := resource.MustParse("1Ki")
	s := KafkaSource{FetchMin: &v}
	assert.Equal(t, 1024, s.GetFetchMinBytes())
}

func TestKafkaSource_GetFetchWaitMaxMs(t *testing.T) {
	s := KafkaSource{FetchWaitMax: &metav1.Duration{Duration: time.Second}}
	assert.Equal(t, 1000, s.GetFetchWaitMaxMs())
}

func TestKafkaSource_GetGroupID(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := &KafkaSource{}
		assert.Equal(t, "foo", s.GetGroupID("foo"))
	})
	t.Run("Specified", func(t *testing.T) {
		s := &KafkaSource{GroupID: "bar"}
		assert.Equal(t, "bar", s.GetGroupID("foo"))
	})
}
