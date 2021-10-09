package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGit_getContainer(t *testing.T) {
	x := Git{
		Image:   "my-image",
		Command: []string{"my-command"},
		Env:     []corev1.EnvVar{{Name: "my-env"}},
	}
	c := x.getContainer(getContainerReq{})

	assert.Equal(t, x.Image, c.Image)
	assert.Equal(t, append([]string{"/var/run/argo-dataflow/runner", "sidecar"}, x.Command...), c.Command)
	assert.Equal(t, x.Env, c.Env)
	assert.Equal(t, PathWorkingDir, c.WorkingDir)
}
