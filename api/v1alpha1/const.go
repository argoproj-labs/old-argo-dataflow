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
	// paths
	PathVarRun         = "/var/run/argo-dataflow"
	PathVarRunRuntimes = "/var/run/argo-dataflow/runtimes"
)
