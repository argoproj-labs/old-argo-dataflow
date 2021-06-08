package v1alpha1

import (
	"fmt"
)

const (
	// conditions
	ConditionCompleted    = "Completed"    // the pipeline completed
	ConditionErrors       = "Errors"       // added if any step encounters an error
	ConditionRunning      = "Running"      // added if any step is currently running
	ConditionSunkMessages = "SunkMessages" // added if any messages have been written to a sink for any step
	ConditionTerminating  = "Terminating"  // added if any terminator step terminated
	// container names
	CtrInit    = "init"
	CtrMain    = "main"
	CtrSidecar = "sidecar"
	// env vars
	EnvImageFormat         = "ARGO_DATAFLOW_IMAGE_FORMAT" // default "quay.io/argoproj/%s:latest"
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
)

var KeyKillCmd = func(x string) string {
	return fmt.Sprintf("dataflow.argoproj.io/kill-cmd.%s", x)
}
