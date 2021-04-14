package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var DefaultResources = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		"cpu":    resource.MustParse("50m"),
		"memory": resource.MustParse("32Mi"),
	},
	Requests: corev1.ResourceList{
		"cpu":    resource.MustParse("50m"),
		"memory": resource.MustParse("32Mi"),
	},
}
