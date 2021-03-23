package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// +kubebuilder:scaffold:imports

	"github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var _ = Describe("Pipeline controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		Name      = "test"
		Namespace = "test"
	)

	Context("When creating pipeline", func() {
		It("Should create a new deployment", func() {
			By("By creating a new Pipeline")
			ctx := context.Background()
			p := &v1alpha1.Pipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "dataflow.argoproj.io/v1alpha1",
					Kind:       "Pipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
				Spec: v1alpha1.PipelineSpec{
					Processors: []v1alpha1.Processor{
						{Name: "main", Image: "docker/whalesay:latest"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, p)).Should(Succeed())

			Eventually(func() []v1.Deployment {
				list := &v1.DeploymentList{}
				Expect(k8sClient.List(ctx, list)).Should(Succeed())
				return list.Items
			}).Should(HaveLen(1))
		})
	})
})
