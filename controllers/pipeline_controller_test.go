package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	Context("When TODO", func() {
		It("Should TODO", func() {
			By("By creating a new CronJob")
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
			}
			Expect(k8sClient.Create(ctx, p)).Should(Succeed())
		})
	})
})
