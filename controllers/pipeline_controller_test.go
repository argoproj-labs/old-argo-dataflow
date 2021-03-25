package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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
		It("Should create a new deployment", func() {
			By("By creating a new Pipeline")
			ctx := context.Background()

			Eventually(func() error {
				return k8sClient.Create(ctx, &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: Namespace},
				})
			}).Should(Succeed())

			replicas := pointer.Int32Ptr(2)
			p := &dfv1.Pipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "dataflow.argoproj.io/dfv1",
					Kind:       "Pipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
				Spec: dfv1.PipelineSpec{
					Nodes: []dfv1.Node{
						{
							Name:     "my-proc",
							Image:    "docker/whalesay:latest",
							Replicas: &dfv1.Replicas{Value: replicas},
							Sources:  []dfv1.Source{{}},
							Sinks:    []dfv1.Sink{{}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, p)).Should(Succeed())

			d := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: Namespace, Name: "my-pipeline-my-proc"}, d)
			}).
				Should(Succeed())

			Expect(d.Spec.Replicas).Should(Equal(replicas))
		})
	})
})
