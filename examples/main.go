package main

import (
	"fmt"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
)

func main() {
	const dirname = "examples"
	files, err := util.ReadDir(dirname)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Printf("### Examples\n\n")
	for _, info := range files {
		println(info.Name())
		for i, un := range info.Items {
			if i == 0 {
				annotations := un.GetAnnotations()
				if _, ok := annotations["dataflow.argoproj.io/description"]; ok {
					_, _ = fmt.Printf("### [%s](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s)\n\n", un.GetName(), info.Name())
					_, _ = fmt.Printf("%s\n\n", annotations["dataflow.argoproj.io/description"])
					_, _ = fmt.Printf("```\nkubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s\n```\n\n", info.Name())
				}
			}
		}
	}
}
