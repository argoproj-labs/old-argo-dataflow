package main

import (
	"github.com/argoproj-labs/argo-dataflow/sdks/golang"
)

func main() {
	golang.Start(Handler)
}
