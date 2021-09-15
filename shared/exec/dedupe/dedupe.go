package dedupe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/antonmedv/expr/vm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/antonmedv/expr"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	logger     = sharedutil.NewLogger()
	db         = &uniqItems{ids: map[string]*item{}}
	mu         = sync.Mutex{}
	duplicates = 0
)

type Impl struct {
	uid     string
	maxSize resource.Quantity
	prog    *vm.Program
}

func New(uid string, maxSize resource.Quantity) *Impl { return &Impl{uid: uid, maxSize: maxSize} }

func (i *Impl) Init(ctx context.Context) error {
	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "duplicate_messages",
		Help: "Duplicates messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#duplicate_messages",
	}, func() float64 { return float64(duplicates) })

	prog, err := expr.Compile(i.uid)
	if err != nil {
		return fmt.Errorf("failed to compile %q: %w", i.uid, err)
	}
	i.prog = prog
	int64MaxSize, ok := i.maxSize.AsInt64()
	if !ok {
		return fmt.Errorf("max size %v must be int64", i.maxSize)
	}

	go wait.JitterUntil(func() {
		size := db.size()
		logger.Info("garbage collection", "size", resource.NewQuantity(int64(size), resource.DecimalSI), "maxSize", i.maxSize, "duplicates", duplicates)
		for db.size() > int(int64MaxSize) {
			func() {
				mu.Lock()
				defer mu.Unlock()
				db.shrink()
			}()
		}
	}, 15*time.Second, 1.2, true, ctx.Done())
	return nil
}

func (i *Impl) Exec(ctx context.Context, msg []byte) ([]byte, error) {
	env, err := util.ExprEnv(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create expr env: %w", err)
	}
	r, err := expr.Run(i.prog, env)
	if err != nil {
		return nil, fmt.Errorf("failed to execute program: %w", err)
	}
	id, ok := r.(string)
	if !ok {
		return nil, fmt.Errorf("expression %q did not evaluate to string", i.uid)
	}
	mu.Lock()
	defer mu.Unlock()
	dupe := db.update(id)
	if dupe {
		duplicates++
		return nil, nil
	}
	return msg, nil
}
