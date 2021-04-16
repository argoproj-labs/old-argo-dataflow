package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	"sigs.k8s.io/yaml"
)

func main() {
	const dirname = "docs/examples"
	files, err := util.ReadDir(dirname)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Printf("### Examples\n\n")
	for _, info := range files {
		println(info.Name())
		formatted :=make([]string,len(info.Items))
		for i, un := range info.Items {
			if i == 0 {
				annotations := un.GetAnnotations()
				if name, ok := annotations["dataflow.argoproj.io/name"]; ok {
					title := name
					_, _ = fmt.Printf("### [%s](examples/%s)\n\n", title, info.Name())
					_, _ = fmt.Printf("%s\n\n", annotations["dataflow.argoproj.io/description"])
					_, _ = fmt.Printf("```\nkubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/%s\n```\n\n", info.Name())
				}
			}
			if data, err := yaml.Marshal(un.Object); err != nil {
				panic(err)
			} else {
				formatted[i] = string(data)
			}
		}
		if err := ioutil.WriteFile(filepath.Join(dirname, info.Name()), []byte(strings.Join(formatted, "\n---\n")), 0600); err != nil {
			panic(err)
		}
	}
}
