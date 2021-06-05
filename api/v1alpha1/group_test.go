package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGroup_getContainer(t *testing.T) {
	x := Group{
		Key:        "my-key",
		EndOfGroup: "my-eog",
		Format:     "my-fmt",
		Storage: &Storage{
			Name:    "my-storage",
			SubPath: "my-sub-path",
		},
	}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"group", "my-key", "my-eog", "my-fmt"}, c.Args)
	assert.Contains(t, c.VolumeMounts, corev1.VolumeMount{Name: "my-storage", MountPath: "/var/run/argo-dataflow/groups", SubPath: "my-sub-path"})
}
