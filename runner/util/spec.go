package util

import (
	"encoding/json"
	"os"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func UnmarshallSpec() (*dfv1.StepSpec, error) {
	spec := &dfv1.StepSpec{}
	return spec, json.Unmarshal([]byte(os.Getenv(dfv1.EnvStepSpec)), spec)
}
