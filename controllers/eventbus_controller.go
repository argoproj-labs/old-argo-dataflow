/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-events/common"
	"github.com/argoproj/argo-events/controllers/eventbus/installer"
	ev1 "github.com/argoproj/argo-events/pkg/apis/eventbus/v1alpha1"
	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

// EventBusReconciler reconciles a EventBus object
type EventBusReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=eventbus,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataflow.argoproj.io,resources=eventbus/status,verbs=get;update;patch

func (r *EventBusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("eventbus", req.NamespacedName)

	var bus v1alpha1.EventBus
	if err := r.Client.Get(ctx, req.NamespacedName, &bus); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if _, err := installer.NewNATSInstaller(r.Client, &ev1.EventBus{
		ObjectMeta: bus.ObjectMeta,
		Spec: ev1.EventBusSpec{
			NATS: &ev1.NATSBus{
				Native: &ev1.NativeStrategy{
					Replicas: 3,
					Auth:     &ev1.AuthStrategyNone,
				},
			},
		},
	}, "nats-streaming:0.17.0", "synadia/prometheus-nats-exporter:0.6.2", map[string]string{
		"controller":          "pipeline-controller",
		"eventbus-name":       bus.Name,
		common.LabelOwnerName: bus.Name,
	}, zap.L().Sugar()).Install(); IgnoreAlreadyExists(err) != nil {
		return ctrl.Result{}, fmt.Errorf("failed to install NATS: %w", err)
	}

	log.Info("installed NATS (if not already installed)")

	return ctrl.Result{}, nil
}

func (r *EventBusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.EventBus{}).
		Complete(r)
}
