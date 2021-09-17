package dedupe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/antonmedv/expr"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	"github.com/argoproj-labs/argo-dataflow/shared/builtin"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	logger     = sharedutil.NewLogger()
	db         = &uniqItems{ids: map[string]*item{}}
	mu         = sync.Mutex{}
)

func New(ctx context.Context, uid string, maxSize resource.Quantity) (builtin.Process, error) {
	duplicates := promauto.NewCounter(prometheus.CounterOpts{
		Name: "duplicate_messages",
		Help: "Duplicates messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#duplicate_messages",
	})

	prog, err := expr.Compile(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to compile %q: %w", uid, err)
	}

	int64MaxSize, ok := maxSize.AsInt64()
	if !ok {
		return nil, fmt.Errorf("max size %v must be int64", maxSize)
	}

	go wait.JitterUntil(func() {
		size := db.size()
		logger.Info("garbage collection", "size", resource.NewQuantity(int64(size), resource.DecimalSI), "maxSize", maxSize, "duplicates", duplicates)
		for db.size() > int(int64MaxSize) {
			func() {
				mu.Lock()
				defer mu.Unlock()
				db.shrink()
			}()
		}
	}, 15*time.Second, 1.2, true, ctx.Done())
	return func(ctx context.Context, msg []byte) ([]byte, error) {
		env, err := util.ExprEnv(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create expr env: %w", err)
		}
		r, err := expr.Run(prog, env)
		if err != nil {
			return nil, fmt.Errorf("failed to execute program: %w", err)
		}
		id, ok := r.(string)
		if !ok {
			return nil, fmt.Errorf("expression did not evaluate to string")
		}
		mu.Lock()
		defer mu.Unlock()
		dupe := db.update(id)
		if dupe {
			duplicates.Inc()
			return nil, nil
		}
		return msg, nil
	}, nil
}
