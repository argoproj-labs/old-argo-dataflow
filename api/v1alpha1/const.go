package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// conditions.
	ConditionCompleted   = "Completed"   // the pipeline completed
	ConditionRunning     = "Running"     // added if any step is currently running
	ConditionTerminating = "Terminating" // added if any terminator step terminated
	// container names.
	CtrInit = "init"
	CtrMain = "main"
	// env vars.
	EnvCluster          = "ARGO_DATAFLOW_CLUSTER"
	EnvDebug            = "ARGO_DATAFLOW_DEBUG"              // enable debug flags, maybe "true" or CSV, e.g. "pprof,kafka.generic"
	EnvUnixDomainSocket = "ARGO_DATAFLOW_UNIX_DOMAIN_SOCKET" // use Unix Domain Socket, default "true"
	EnvImagePrefix      = "ARGO_DATAFLOW_IMAGE_PREFIX"       // default "quay.io/argoproj"
	EnvNamespace        = "ARGO_DATAFLOW_NAMESPACE"
	EnvPipelineName     = "ARGO_DATAFLOW_PIPELINE_NAME"
	EnvPod              = "ARGO_DATAFLOW_POD"
	EnvReplica          = "ARGO_DATAFLOW_REPLICA"
	EnvStep             = "ARGO_DATAFLOW_STEP"
	EnvPeekDelay        = "ARGO_DATAFLOW_PEEK_DELAY"         // how long between peeking (default 4m)
	EnvPullPolicy       = "ARGO_DATAFLOW_PULL_POLICY"        // default ""
	EnvScalingDelay     = "ARGO_DATAFLOW_SCALING_DELAY"      // how long to wait between any scaling events (including peeking) default "4m"
	EnvUpdateInterval   = "ARGO_DATAFLOW_UPDATE_INTERVAL"    // default "15s"
	EnvImagePullSecrets = "ARGO_DATAFLOW_IMAGE_PULL_SECRETS" // allows providing a list of imagePullSecrets as a comma delimited string (eg. "secret1,secret2")
	// label/annotation keys.
	KeyDefaultContainer = "kubectl.kubernetes.io/default-container"
	KeyDescription      = "dataflow.argoproj.io/description"
	KeyFinalizer        = "dataflow.argoproj.io/finalizer"
	KeyOwner            = "dataflow.argoproj.io/owner"
	KeyPipelineName     = "dataflow.argoproj.io/pipeline-name"
	KeyReplica          = "dataflow.argoproj.io/replica"
	KeyStepName         = "dataflow.argoproj.io/step-name" // the step name without pipeline name prefix
	KeyHash             = "dataflow.argoproj.io/hash"      // hash of the object
	// paths.
	PathAuthorization = "/var/run/argo-dataflow/authorization" // the authorization header which must be used by the main container to speak to the sidecar
	PathCheckout      = "/var/run/argo-dataflow/checkout"
	PathFIFOIn        = "/var/run/argo-dataflow/in"
	PathFIFOOut       = "/var/run/argo-dataflow/out"
	PathGroups        = "/var/run/argo-dataflow/groups"
	PathHandlerFile   = "/var/run/argo-dataflow/handler"
	PathKill          = "/var/run/argo-dataflow/kill"
	PathWorkingDir    = "/var/run/argo-dataflow/wd"
	PathVarRun        = "/var/run/argo-dataflow"
	PathRunner        = "/var/run/argo-dataflow/runner"
	// other const.
	CommitN = 20 // how many messages between commits, therefore potential duplicates during disruption
)

// the standard resources used by the `init`, `sidecar` and built-in step containers.
var standardResources = corev1.ResourceRequirements{
	Limits: corev1.ResourceList{
		"cpu":    resource.MustParse("200m"),
		"memory": resource.MustParse("256Mi"),
	},
	Requests: corev1.ResourceList{
		"cpu":    resource.MustParse("100m"),
		"memory": resource.MustParse("64Mi"),
	},
}
