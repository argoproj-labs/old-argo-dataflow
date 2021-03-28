package v1alpha1

const (
	EnvFunc         = "FUNC"
	EnvPipelineName = "PIPELINE_NAME"
	KeyFuncName     = "dataflow.argoproj.io/func-name"
	KeyPipelineName = "dataflow.argoproj.io/pipeline-name"
	KeyReplica      = "dataflow.argoproj.io/replica"
	CtrInit         = "dataflow-init"
	CtrSidecar      = "dataflow-sidecar"
)
