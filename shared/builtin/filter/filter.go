package filter

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin"
)

func New(expression string) (builtin.Process, error) {
	prog, err := expr.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to compile %q: %w", expression, err)
	}
	return func(ctx context.Context, msg []byte) ([]byte, error) {
		env, err := util.ExprEnv(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create expr env: %w", err)
		}
		res, err := expr.Run(prog, env)
		if err != nil {
			return nil, fmt.Errorf("failed to run program: %w", err)
		}
		accept, ok := res.(bool)
		if !ok {
			return nil, fmt.Errorf("must return bool")
		}
		if accept {
			return msg, nil
		} else {
			return nil, nil
		}
	}, nil
}
