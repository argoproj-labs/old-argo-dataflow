package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"sigs.k8s.io/yaml"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func main() {
	infos, err := ioutil.ReadDir("examples")
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Printf("### Examples\n\n")
	for _, i := range infos {
		fn := i.Name()
		if !strings.HasSuffix(fn, ".yaml") {
			continue
		}
		data, err := ioutil.ReadFile("examples/" + fn)
		if err != nil {
			panic(err)
		}
		println(fn)
		text := strings.Split(string(data), "---")[0]
		pipeline := &dfv1.Pipeline{}
		if err := yaml.Unmarshal([]byte(text), pipeline); err != nil {
			panic(err)
		}
		annotations := pipeline.GetAnnotations()
		_, _ = fmt.Printf("### [%s](%s)\n\n", annotations["dataflow.argoproj.io/name"], fn)
		_, _ = fmt.Printf("%s\n\n", annotations["dataflow.argoproj.io/description"])
		_, _ = fmt.Printf("```\nkubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s\n```\n\n", fn)
	}
}
