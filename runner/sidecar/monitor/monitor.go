package monitor

import (
	"context"
)

type Interface interface {
	// Accept determine if the message should be processed. It is not a duplicate.
	Accept(ctx context.Context, sourceName, sourceURN string, partition int32, offset int64) bool
	Close(context.Context)
}
