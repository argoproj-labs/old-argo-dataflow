package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestContainer_getContainer(t *testing.T) {
	x := Container{
		Image:        "my-image",
		VolumeMounts: []corev1.VolumeMount{{Name: "my-vm"}},
		Command:      []string{"my-cmd"},
		Args:         []string{"my-args"},
		Env:          []corev1.EnvVar{{Name: "my-envvar"}},
		Resources: corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{
				"cpu": resource.MustParse("2"),
			},
		},
	}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, x.Image, c.Image)
	assert.Contains(t, c.VolumeMounts, c.VolumeMounts[0])
	assert.Equal(t, x.Command, c.Command)
	assert.Equal(t, x.Args, c.Args)
	assert.Equal(t, x.Env, c.Env)
	assert.Equal(t, corev1.ResourceRequirements{Requests: map[corev1.ResourceName]resource.Quantity{"cpu": resource.MustParse("2")}}, c.Resources)
}
