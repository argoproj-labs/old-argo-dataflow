package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {
	infos, err := ioutil.ReadDir("examples")
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Printf("### Examples\n\n")
	for _, i := range infos {
		f := i.Name()
		if !strings.HasSuffix(f, ".yaml") {
			continue
		}
		const path = "examples/"
		data, err := ioutil.ReadFile(path + f)
		if err != nil {
			panic(err)
		}
		println(f)
		var formatted []string
		for i, text := range strings.Split(string(data), "---") {
			if i == 0 {
				un := &unstructured.Unstructured{}
				if err := yaml.Unmarshal([]byte(text), un); err != nil {
					panic(err)
				}
				annotations := un.GetAnnotations()
				if name, ok := annotations["dataflow.argoproj.io/name"]; ok {
					_, _ = fmt.Printf("### [%s](%s)\n\n", name, f)
					_, _ = fmt.Printf("%s\n\n", annotations["dataflow.argoproj.io/description"])
					_, _ = fmt.Printf("```\nkubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s\n```\n\n", f)
				}
			}
			un := &unstructured.Unstructured{}
			if err := yaml.Unmarshal([]byte(text), un); err != nil {
				panic(err)
			}
			if data, err := yaml.Marshal(un); err != nil {
				panic(err)
			} else {
				formatted = append(formatted, string(data))
			}
		}
		if err := ioutil.WriteFile(path+f, []byte(strings.Join(formatted, "\n---\n")), 0600); err != nil {
			panic(err)
		}
	}
}
