package v1alpha1

const (
	// conditions
	ConditionRunning    = "Running"
	ConditionTerminated = "Terminated"
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
	KeyStepName     = "dataflow.argoproj.io/step-name"
	KeyReplica      = "dataflow.argoproj.io/replica"
	KeyRestart      = "dataflow.argoproj.io/restart"
	// paths
	PathCheckout       = "/var/run/argo-dataflow/checkout"
	PathGroups         = "/var/run/argo-dataflow/groups"
	PathFIFOIn         = "/var/run/argo-dataflow/in"
	PathFIFOOut        = "/var/run/argo-dataflow/out"
	PathVarRunRuntimes = "/var/run/argo-dataflow/runtimes"
	PathWorkingDir     = "/var/run/argo-dataflow/wd"
)
