package v1alpha1

const (
	// conditions
	ConditionRunning     = "Running"
	ConditionTerminating = "Terminating"
	// container names
	CtrInit    = "init"
	CtrMain    = "main"
	CtrSidecar = "sidecar"
	// env vars
	EnvPipelineName = "ARGO_DATAFLOW_PIPELINE_NAME"
	EnvNamespace    = "ARGO_DATAFLOW_NAMESPACE"
	EnvStepSpec     = "ARGO_DATAFLOW_STEP_SPEC"
	EnvReplica      = "ARGO_DATAFLOW_REPLICA"
	// label/annotation keys
	KeyPipelineName = "dataflow.argoproj.io/pipeline-name"
	KeyStepName     = "dataflow.argoproj.io/step-name" // the step name without pipeline name prefix
	KeySpecHash     = "dataflow.argoproj.io/spec-hash" // hash of the spec
	KeyReplica      = "dataflow.argoproj.io/replica"
	KeyRestart      = "dataflow.argoproj.io/restart"
	// paths
	PathCheckout   = "/var/run/argo-dataflow/checkout"
	PathGroups     = "/var/run/argo-dataflow/groups"
	PathFIFOIn     = "/var/run/argo-dataflow/in"
	PathFIFOOut    = "/var/run/argo-dataflow/out"
	PathVarRun     = "/var/run/argo-dataflow"
	PathRuntimes   = "/var/run/argo-dataflow/runtimes"
	PathWorkingDir = "/var/run/argo-dataflow/wd"
)
