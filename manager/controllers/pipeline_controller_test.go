package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// +kubebuilder:scaffold:imports
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var _ = Describe("Pipeline controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		Name      = "my-pipeline"
		Namespace = "my-ns"
	)

	Context("When creating pipeline", func() {
		It("Should create a new step", func() {
			By("By creating a new Pipeline")
			ctx := context.Background()

			Eventually(func() error {
				return k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: Namespace},
				})
			}).Should(Succeed())

			p := &dfv1.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
				Spec: dfv1.PipelineSpec{
					Steps: []dfv1.StepSpec{
						{
							Name: "my-step",
							Container: &dfv1.Container{
								Image: "docker/whalesay:latest",
							},
							Replicas: &dfv1.Replicas{Min: 2},
							Sources:  []dfv1.Source{{}},
							Sinks:    []dfv1.Sink{{}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, p)).Should(Succeed())

			step := &dfv1.Step{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: "my-pipeline-my-step"}, step)
			}).
				Should(Succeed())
		})
	})
})
