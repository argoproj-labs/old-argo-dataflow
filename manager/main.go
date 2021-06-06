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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/argoproj-labs/argo-dataflow/shared/util"

	dataflowv1alpha1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/manager/controllers"
	"github.com/argoproj-labs/argo-dataflow/shared/containerkiller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = dataflowv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":9090", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(util.NewLogger())

	restConfig := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "1c03be80.my.domain",
		Namespace:          os.Getenv(dataflowv1alpha1.EnvNamespace),
	})
	if err != nil {
		panic(fmt.Errorf("unable to start manager: %w", err))
	}

	clientset := kubernetes.NewForConfigOrDie(restConfig)
	containerKiller := containerkiller.New(clientset, restConfig)
	if err = (&controllers.PipelineReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Pipeline"),
		Scheme:          mgr.GetScheme(),
		ContainerKiller: containerKiller,
	}).SetupWithManager(mgr); err != nil {
		panic(fmt.Errorf("unable to create controller manager: %w", err))
	}

	if err = (&controllers.StepReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Step"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("step-reconciler"),
		ContainerKiller: containerKiller,
	}).SetupWithManager(mgr); err != nil {
		panic(fmt.Errorf("unable to create controller manager: %w", err))
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		panic(fmt.Errorf("problem running manager: %w", err))
	}
}
