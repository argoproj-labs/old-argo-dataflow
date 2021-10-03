package monitor

import (
	"context"
)

type Interface interface {
	// Accept determine if the message should be processed. It is not a duplicate.
	Accept(sourceName, sourceURN string, partition int32, offset int64) bool
	Mark(sourceURN string, partition int32, offset int64)
	AssignedPartition(ctx context.Context, sourceURN string, partition int32)
	RevokedPartition(ctx context.Context, sourceURN string, partition int32)
	Close(context.Context)
}
