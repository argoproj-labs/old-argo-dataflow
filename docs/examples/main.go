package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {
	const path = "docs/examples"
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Printf("### Examples\n\n")
	sort.Slice(infos, func(i, j int) bool {
		return strings.Compare(infos[i].Name(), infos[j].Name()) < 0
	})
	for _, i := range infos {
		f := i.Name()
		if !strings.HasSuffix(f, ".yaml") {
			continue
		}
		file := filepath.Join(path, f)
		data, err := ioutil.ReadFile(file)
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
					title := name
					_, _ = fmt.Printf("### [%s](examples/%s)\n\n", title, f)
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
		if err := ioutil.WriteFile(file, []byte(strings.Join(formatted, "\n---\n")), 0600); err != nil {
			panic(err)
		}
	}
}
