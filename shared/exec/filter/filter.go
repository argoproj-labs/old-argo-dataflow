package filter

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr/vm"

	"github.com/antonmedv/expr"

	"github.com/argoproj-labs/argo-dataflow/runner/util"
)

type Impl struct {
	expression string
	prog       *vm.Program
}

func New(expression string) *Impl { return &Impl{expression: expression} }

func (i *Impl) Init(ctx context.Context) error {
	prog, err := expr.Compile(i.expression)
	if err != nil {
		return fmt.Errorf("failed to compile %q: %w", i.expression, err)
	}
	i.prog = prog
	return nil
}

func (i *Impl) Exec(ctx context.Context, msg []byte) ([]byte, error) {
	env, err := util.ExprEnv(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create expr env: %w", err)
	}
	res, err := expr.Run(i.prog, env)
	if err != nil {
		return nil, fmt.Errorf("failed to run program %x: %w", i.expression, err)
	}
	accept, ok := res.(bool)
	if !ok {
		return nil, fmt.Errorf("%q must return bool", i.expression)
	}
	if accept {
		return msg, nil
	} else {
		return nil, nil
	}
}
