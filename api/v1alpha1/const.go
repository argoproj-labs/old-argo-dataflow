package v1alpha1

const (
	// container names
	CtrInit    = "dataflow-init"
	CtrSidecar = "dataflow-sidecar"
	// env vars
	EnvPipelineName = "ARGO_DATAFLOW_PIPELINE_NAME"
	EnvStep         = "ARGO_DATAFLOW_STEP"
	EnvReplica      = "ARGO_DATAFLOW_REPLICA"
	// label/annotation keys
	KeyPipelineName = "dataflow.argoproj.io/pipeline-name"
	KeyStepName     = "dataflow.argoproj.io/step-name"
	KeyReplica      = "dataflow.argoproj.io/replica"
	// paths
	PathVarRun         = "/var/run/argo-dataflow"
	PathVarRunRuntimes = "/var/run/argo-dataflow/runtimes"
)
