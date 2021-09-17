package sink

import "context"

type Interface interface {
	Sink(ctx context.Context, msg []byte) error
}
