// +build test

package test

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var secretsInterface = kubernetesInterface.CoreV1().Secrets(namespace)

func GetAuthorization() string {
	ctx := context.Background()
	pl := GetPipeline()
	for _, step := range pl.Spec.Steps {
		for _, source := range step.Sources {
			if source.HTTP != nil {
				secret, err := secretsInterface.Get(ctx, fmt.Sprintf("%s-%s", pl.Name, step.Name), metav1.GetOptions{})
				if err != nil {
					panic(err)
				}
				return string(secret.Data[fmt.Sprintf("sources.%s.http.authorization", source.Name)])
			}
		}
	}
	panic("not HTTP source")
}
