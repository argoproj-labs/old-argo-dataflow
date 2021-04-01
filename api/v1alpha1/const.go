package v1alpha1

const (
	// container names
	CtrInit    = "dataflow-init"
	CtrSidecar = "dataflow-sidecar"
	// env vars
	EnvFunc         = "ARGO_DATAFLOW_FUNC"
	EnvPipelineName = "ARGO_DATAFLOW_PIPELINE_NAME"
	EnvReplica      = "ARGO_DATAFLOW_REPLICA"
	// label/annotation keys
	KeyFuncName     = "dataflow.argoproj.io/func-name"
	KeyPipelineName = "dataflow.argoproj.io/pipeline-name"
	KeyReplica      = "dataflow.argoproj.io/replica"
	// paths
	PathVarRun         = "/var/run/argo-dataflow"
	PathVarRunRuntimes = "/var/run/argo-dataflow/runtimes"
)
