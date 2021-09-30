// +build test

package examples

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/test"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml
//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/stan.yaml

func Test_101_hello_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/101-hello-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_101_two_node_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/101-two-node-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_102_dedupe_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/102-dedupe-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_102_filter_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/102-filter-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_102_flatten_expand_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/102-flatten-expand-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_102_map_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/102-map-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_103_autoscaling_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/103-autoscaling-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_103_scaling_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/103-scaling-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_104_golang1_16_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/104-golang1-16-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_104_java16_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/104-java16-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_104_node16_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/104-node16-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_104_python3_9_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/104-python3-9-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_106_git_go_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/106-git-go-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_106_git_nodejs_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/106-git-nodejs-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_106_git_python_generator_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/106-git-python-generator-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_106_git_python_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/106-git-python-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_107_completion_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/107-completion-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilCompleted, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_107_terminator_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/107-terminator-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilCompleted, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_108_container_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/108-container-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilCompleted, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_108_fifos_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/108-fifos-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_109_group_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/109-group-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_301_cron_log_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/301-cron-log-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

//go:generate kubectl -n argo-dataflow-system apply -f ../../examples/dataflow-103-http-main-source-default-secret.yaml

func Test_301_http_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/301-http-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_301_kafka_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/301-kafka-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_301_two_sinks_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/301-two-sinks-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func Test_301_two_sources_pipeline(t *testing.T) {
	defer Setup(t)()

	CreatePipelineFromFile("../../examples/301-two-sources-pipeline.yaml")

	WaitForPipeline()
	WaitForPipeline(UntilRunning, 90*time.Second)

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
