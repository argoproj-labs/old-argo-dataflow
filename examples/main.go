package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	const dirname = "examples"
	files, err := sharedutil.ReadDir(dirname)
	if err != nil {
		panic(err)
	}
	switch os.Args[1] {
	case "docs":
		_, _ = fmt.Printf("### Examples\n\n")
		for _, file := range files {
			println(file.Name())
			for i, un := range file.Items {
				if i == 0 {
					annotations := un.GetAnnotations()
					if _, ok := annotations["dataflow.argoproj.io/description"]; ok {
						_, _ = fmt.Printf("### [%s](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s)\n\n", un.GetName(), file.Name())
						_, _ = fmt.Printf("%s\n\n", annotations["dataflow.argoproj.io/description"])
						_, _ = fmt.Printf("```\nkubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s\n```\n\n", file.Name())
					}
				}
			}
		}
	case "tests":
		_, _ = fmt.Printf("// +build test\n\n")
		_, _ = fmt.Printf("package examples\n\n")
		_, _ = fmt.Printf("import (\n")
		_, _ = fmt.Printf("\t\"testing\"\n")
		_, _ = fmt.Printf("\t\"time\"\n")
		_, _ = fmt.Printf("\n")
		_, _ = fmt.Printf("\t. \"github.com/argoproj-labs/argo-dataflow/test\"\n")
		_, _ = fmt.Printf(")\n\n")
		_, _ = fmt.Printf("//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/kafka.yaml\n")
		_, _ = fmt.Printf("//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/stan.yaml\n\n")
		for _, file := range files {
			un := file.Items[0]
			if un.GetKind() == "Pipeline" && un.GetAnnotations()["dataflow.argoproj.io/test"] != "false" {
				if v := un.GetAnnotations()["dataflow.argoproj.io/needs"]; v != "" {
					_, _ = fmt.Printf("//go:generate kubectl -n argo-dataflow-system apply -f ../../examples/%s\n\n", v)
				}
				_, _ = fmt.Printf("func Test_%s(t *testing.T) {\n", strings.ReplaceAll(strings.TrimSuffix(file.Name(), ".yaml"), "-", "_"))
				_, _ = fmt.Printf("\tdefer Setup(t)()\n\n")
				_, _ = fmt.Printf("\tCreatePipelineFromFile(%q)\n\n", filepath.Join("..", "..", "examples", file.Name()))
				_, _ = fmt.Printf("\tWaitForPipeline()\n")
				_, _ = fmt.Printf("\tWaitForPipeline(Until%s, 90*time.Second)\n", getWaitFor(&un))
				_, _ = fmt.Printf("}\n\n")
			}
		}
	default:
		panic(fmt.Errorf("un-expected arguments %q", os.Args))
	}
}

func getWaitFor(un metav1.Object) string {
	if v := un.GetAnnotations()["dataflow.argoproj.io/wait-for"]; v != "" {
		return v
	}
	return ConditionSunkMessages
}
