package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestKafkaSink_GetBatchSize(t *testing.T) {
	v := resource.MustParse("1Ki")
	s := KafkaSink{BatchSize: &v}
	assert.Equal(t, 1024, s.GetBatchSize())
}

func TestKafkaSink_GetLingerMs(t *testing.T) {
	t.Run("Sync", func(t *testing.T) {
		s := KafkaSink{}
		assert.Equal(t, 0, s.GetLingerMs())
	})
	t.Run("ASync", func(t *testing.T) {
		s := KafkaSink{Async: true}
		assert.Equal(t, 5, s.GetLingerMs())
	})
	t.Run("Specified", func(t *testing.T) {
		s := KafkaSink{Linger: &metav1.Duration{Duration: time.Second}}
		assert.Equal(t, 1000, s.GetLingerMs())
	})
}

func TestKafkaSink_GetAcks(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		v := intstr.FromString("all")
		s := KafkaSink{Acks: &v}
		assert.Equal(t, "all", s.GetAcks())
	})
	t.Run("1", func(t *testing.T) {
		v := intstr.FromInt(1)
		s := KafkaSink{Acks: &v}
		assert.Equal(t, 1, s.GetAcks())
	})
}

func TestKafkaSink_GetMessageMaxBytes(t *testing.T) {
	s := KafkaSink{
		Kafka: Kafka{
			KafkaConfig: KafkaConfig{
				MaxMessageBytes: 1,
			},
		},
	}
	assert.Equal(t, 1, s.GetMessageMaxBytes())
}
