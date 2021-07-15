package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// conditions
	ConditionCompleted    = "Completed"    // the pipeline completed
	ConditionRecentErrors = "RecentErrors" // added if any step encountered an error recently
	ConditionRunning      = "Running"      // added if any step is currently running
	ConditionSunkMessages = "SunkMessages" // added if any messages have been written to a sink for any step
	ConditionTerminating  = "Terminating"  // added if any terminator step terminated
	// container names
	CtrInit    = "init"
	CtrMain    = "main"
	CtrSidecar = "sidecar"
	// env vars
	EnvImagePrefix         = "ARGO_DATAFLOW_IMAGE_PREFIX"   // default "quay.io/argoproj"
	EnvDeletionDelay       = "ARGO_DATAFLOW_DELETION_DELAY" // default "720h" ~= "30d"
	EnvNamespace           = "ARGO_DATAFLOW_NAMESPACE"
	EnvPipelineName        = "ARGO_DATAFLOW_PIPELINE_NAME"
	EnvReplica             = "ARGO_DATAFLOW_REPLICA"
	EnvStep                = "ARGO_DATAFLOW_STEP"
	EnvPeekDelay           = "ARGO_DATAFLOW_PEEK_DELAY"      // how long between peeking (default 4m)
	EnvPullPolicy          = "ARGO_DATAFLOW_PULL_POLICY"     // default ""
	EnvScalingDelay        = "ARGO_DATAFLOW_SCALING_DELAY"   // how long to wait between any scaling events (including peeking) default "4m"
	EnvUpdateInterval      = "ARGO_DATAFLOW_UPDATE_INTERVAL" // default "1m"
	EnvDataflowBearerToken = "ARGO_DATAFLOW_BEARER_TOKEN"
	// label/annotation keys
	KeyDefaultContainer = "kubectl.kubernetes.io/default-container"
	KeyDescription      = "dataflow.argoproj.io/description"
	KeyFinalizer        = "dataflow.argoproj.io/finalizer"
	KeyOwner            = "dataflow.argoproj.io/owner"
	KeyPipelineName     = "dataflow.argoproj.io/pipeline-name"
	KeyReplica          = "dataflow.argoproj.io/replica"
	KeyStepName         = "dataflow.argoproj.io/step-name" // the step name without pipeline name prefix
	KeyHash             = "dataflow.argoproj.io/hash"      // hash of the object
	// paths
	PathCheckout    = "/var/run/argo-dataflow/checkout"
	PathFIFOIn      = "/var/run/argo-dataflow/in"
	PathFIFOOut     = "/var/run/argo-dataflow/out"
	PathGroups      = "/var/run/argo-dataflow/groups"
	PathHandlerFile = "/var/run/argo-dataflow/handler"
	PathKill        = "/var/run/argo-dataflow/kill"
	PathPreStop     = "/var/run/argo-dataflow/prestop"
	PathWorkingDir  = "/var/run/argo-dataflow/wd"
	PathVarRun      = "/var/run/argo-dataflow"
	// other const
	CommitN = 20 // how many messages between commits, therefore potential duplicates during disruption
)

var KeyKillCmd = func(x string) string {
	return fmt.Sprintf("dataflow.argoproj.io/kill-cmd.%s", x)
}

// the standard resources used by the `init`, `sidecar` and built-in step containers
var standardResources = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		"cpu":    resource.MustParse("500m"),
		"memory": resource.MustParse("256Mi"),
	},
	Requests: corev1.ResourceList{
		"cpu":    resource.MustParse("100m"),
		"memory": resource.MustParse("64Mi"),
	},
}
