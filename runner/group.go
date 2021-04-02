package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/antonmedv/expr"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func Group(x string) error {
	prog, err := expr.Compile(x)
	if err != nil {
		return fmt.Errorf("failed to compile %q: %w", x, err)
	}
	return do(func(msg []byte) ([][]byte, error) {
		res, err := expr.Run(prog, exprEnv(msg))
		if err != nil {
			return nil, err
		}
		group, ok := res.(string)
		if !ok {
			return nil, fmt.Errorf("must return string")
		}
		return nil, ioutil.WriteFile(filepath.Join(dfv1.PathVarRun, "groups", group, string(sha256.New().Sum(msg))), msg, 0600)
	})
}
