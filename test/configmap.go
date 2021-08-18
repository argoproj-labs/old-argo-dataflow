// +build test

package test

import (
	"context"
	"log"

	"github.com/argoproj-labs/argo-dataflow/shared/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateConfigMap(x corev1.ConfigMap) {
	log.Printf("creating config map %q\n", x.Name)
	_, err := kubernetesInterface.CoreV1().ConfigMaps(namespace).Create(context.Background(), &x, metav1.CreateOptions{})
	if util.IgnoreAlreadyExists(err) != nil {
		panic(err)
	}
}
