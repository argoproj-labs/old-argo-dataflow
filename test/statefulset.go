//go:build test
// +build test

package test

import (
	"context"
	"fmt"
	"log"
	"time"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func RestartStatefulSet(name string) {
	log.Printf("restarting stateful set %q\n", name)
	data := sharedutil.MustJSON(map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	})

	_, err := kubernetesInterface.AppsV1().StatefulSets(namespace).Patch(context.Background(), name, types.MergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to restart %q: %w", name, err))
	}
}
