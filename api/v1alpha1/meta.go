package v1alpha1

import (
	"context"
	"fmt"
	"time"
)

type meta struct{ string }

func (m meta) String() string { return m.string }

var (
	// MetaCluster the Kubernetes cluster name.
	MetaCluster = meta{"cluster"}
	// MetaID is a unique ID for the message.
	// Required.
	// https://github.com/cloudevents/spec/blob/master/spec.md#id
	MetaID = meta{"id"}
	// MetaNamespace the Kubernetes namespace.
	MetaNamespace = meta{"namespace"}
	// MetaSequenceNumber is a unique and atomically increasing sequence number for a massage.
	// Combine this with the MetaSource to make it globally unique.
	// Optional.
	MetaSequenceNumber = meta{"sequenceNumber"}
	// MetaSource is the source of the messages as a Unique Resource Identifier (URI).
	// Required.
	// https://github.com/cloudevents/spec/blob/master/spec.md#source-1
	MetaSource = meta{"source"}
	// MetaTime is the time of the message. As meta-data, this might be different to the event-time (which might be within the message).
	// For example, it might be the last-modified time of a file, but the file itself was created at another time.
	// Optional.
	// https://github.com/cloudevents/spec/blob/master/spec.md#time
	MetaTime = meta{"time"}
)

func ContextWithCluster(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, MetaCluster, v)
}

func ContextWithNamespace(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, MetaNamespace, v)
}

func ContextWithMeta(ctx context.Context, source, id string, opts ...interface{}) context.Context {
	ctx = context.WithValue(ctx, MetaSource, source)
	ctx = context.WithValue(ctx, MetaID, id)
	for _, opt := range opts {
		switch v := opt.(type) {
		case uint64:
			ctx = context.WithValue(ctx, MetaSequenceNumber, v)
		case time.Time:
			ctx = context.WithValue(ctx, MetaTime, v)
		default:
			panic(fmt.Errorf("un-supported option type %T", opt))
		}
	}
	return ctx
}

func GetMetaCluster(ctx context.Context) string   { return ctx.Value(MetaCluster).(string) }
func GetMetaID(ctx context.Context) string        { return ctx.Value(MetaID).(string) }
func GetMetaNamespace(ctx context.Context) string { return ctx.Value(MetaNamespace).(string) }
func GetMetaSource(ctx context.Context) string    { return ctx.Value(MetaSource).(string) }
