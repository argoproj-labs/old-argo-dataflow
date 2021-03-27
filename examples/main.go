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
		const path = "examples/"
		data, err := ioutil.ReadFile(path + fn)
		if err != nil {
			panic(err)
		}
		println(fn)
		text := strings.Split(string(data), "---")[0]
		pipeline := &dfv1.Pipeline{}
		if err := yaml.Unmarshal([]byte(text), pipeline); err != nil {
			panic(err)
		}
		if data, err := yaml.Marshal(pipeline); err != nil {
			panic(err)
		} else {
			if err := ioutil.WriteFile(path+fn, data, 0600); err != nil {
				panic(err)
			}
		}
		annotations := pipeline.GetAnnotations()
		_, _ = fmt.Printf("### [%s](%s)\n\n", annotations["dataflow.argoproj.io/name"], fn)
		_, _ = fmt.Printf("%s\n\n", annotations["dataflow.argoproj.io/description"])
		_, _ = fmt.Printf("```\nkubectl apply -f https://raw.githunatsercontent.com/argoproj-labs/argo-dataflow/main/examples/%s\n```\n\n", fn)
	}
}
