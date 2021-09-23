package monitor

import (
	"context"
)

type Interface interface {
	Accept(ctx context.Context, sourceName, sourceURN string, partition int32, offset int64) (bool, error)
}
