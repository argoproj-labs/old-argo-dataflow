package v1alpha1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/assert"
)

func TestCat_getContainer(t *testing.T) {
	x := Cat{}
	c := x.getContainer(getContainerReq{})
	assert.Equal(t, []string{"cat"}, c.Args)
	resource := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}

	x = Cat{AbstractStep: AbstractStep{StandardResources: resource}}
	c = x.getContainer(getContainerReq{})
	assert.Equal(t, resource.Requests.Memory(), c.Resources.Requests.Memory())
}
