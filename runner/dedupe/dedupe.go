package dedupe

import (
	"context"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var (
	logger     = sharedutil.NewLogger()
	db         = &uniqItems{ids: map[string]*item{}}
	mu         = sync.Mutex{}
	duplicates = 0
)

func Exec(ctx context.Context, x string, maxSize resource.Quantity) error {

	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "duplicate_messages",
		Help: "Duplicates messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#duplicate_messages",
	}, func() float64 { return float64(duplicates) })

	prog, err := expr.Compile(x)
	if err != nil {
		return fmt.Errorf("failed to compile %q: %w", x, err)
	}
	int64MaxSize, ok := maxSize.AsInt64()
	if !ok {
		return fmt.Errorf("max size %v must be int64", maxSize)
	}

	go wait.JitterUntil(func() {
		size := resource.MustParse(strconv.FormatInt(int64(reflect.TypeOf(db).Size()), 10))
		logger.Info("garbage collection", "size", size, "maxSize", maxSize, "duplicates", duplicates)
		for ; int64(reflect.TypeOf(db).Size()) > int64MaxSize; {
			mu.Lock()
			db.shrink()
			mu.Unlock()
		}
	}, 10*time.Second, 1.2, true, ctx.Done())

	return util.Do(ctx, func(msg []byte) ([]byte, error) {
		r, err := expr.Run(prog, util.ExprEnv(msg))
		if err != nil {
			return nil, fmt.Errorf("failed to execute program: %w", err)
		}
		id, ok := r.(string)
		if !ok {
			return nil, fmt.Errorf("expression %q did not evaluate to string", x)
		}
		mu.Lock()
		defer mu.Unlock()
		dupe := db.update(id)
		logger.Info("exec", "message", msg, "dupe", dupe, "id", id)
		if dupe {
			duplicates++
			return nil, nil
		}
		return msg, nil
	})
}
