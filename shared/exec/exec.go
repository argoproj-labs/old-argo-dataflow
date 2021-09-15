package exec

import (
	"context"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/exec/cat"
	"github.com/argoproj-labs/argo-dataflow/shared/exec/dedupe"
	"github.com/argoproj-labs/argo-dataflow/shared/exec/expand"
	"github.com/argoproj-labs/argo-dataflow/shared/exec/filter"
	"github.com/argoproj-labs/argo-dataflow/shared/exec/flatten"
	"github.com/argoproj-labs/argo-dataflow/shared/exec/group"
	_map "github.com/argoproj-labs/argo-dataflow/shared/exec/map"
)

type Processor interface {
	Init(ctx context.Context) error
	Exec(ctx context.Context, msg []byte) ([]byte, error)
}

type Interface interface {
	Get(v interface{}) Processor
	Has(v interface{}) bool
}
type DB struct{}

var Instance Interface = DB{}

func (d DB) Get(i interface{}) Processor {
	switch v := i.(type) {
	case dfv1.Cat:
		return cat.New()
	case dfv1.Dedupe:
		return dedupe.New(v.UID, v.MaxSize)
	case dfv1.Expand:
		return expand.New()
	case dfv1.Filter:
		return filter.New(v.Expression)
	case dfv1.Flatten:
		return flatten.New()
	case dfv1.Group:
		return group.New(v.Key, v.EndOfGroup, v.Format)
	case dfv1.Map:
		return _map.New(v.Expression)
	}
	return nil
}

func (d DB) Has(v interface{}) bool {
	return d.Get(v) != nil
}
